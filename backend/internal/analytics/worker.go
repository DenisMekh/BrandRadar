package analytics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"prod-pobeda-2026/internal/analytics/crawler"
	"prod-pobeda-2026/internal/client"
	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
)

// each N minutes
// crawl
// store in db crawl result
// get all brands + results
// items * (text, brand title,keywords)
// send 1 by 1 into sentiment_ml
// prepare batch
// store in db sentiment_ml result, itemId fkey, brandId fkey

type Worker struct {
	webCrawler        *crawler.WebCrawler
	telegramCrawler   *crawler.TelegramCrawler
	brandUseCase      *usecase.BrandUseCase
	sentimentMLClient client.SentimentMLClient
	analyticsRepo     usecase.AnalyticsRepository
	interval          time.Duration
	triggerCh         chan struct{}
	running           atomic.Bool
}

func NewWorker(webCrawler *crawler.WebCrawler, telegramCrawler *crawler.TelegramCrawler, brandUseCase *usecase.BrandUseCase, sentimentMLClient client.SentimentMLClient, analyticsRepo usecase.AnalyticsRepository, interval time.Duration) *Worker {
	return &Worker{
		webCrawler:        webCrawler,
		telegramCrawler:   telegramCrawler,
		brandUseCase:      brandUseCase,
		sentimentMLClient: sentimentMLClient,
		analyticsRepo:     analyticsRepo,
		interval:          interval,
		triggerCh:         make(chan struct{}, 1),
	}
}

// Trigger запускает цикл обработки немедленно, если он ещё не запущен.
func (w *Worker) Trigger() {
	select {
	case w.triggerCh <- struct{}{}:
	default: // уже есть сигнал в буфере — не дублируем
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tryProcess(ctx)
		case <-w.triggerCh:
			w.tryProcess(ctx)
		}
	}
}

func (w *Worker) tryProcess(ctx context.Context) {
	if !w.running.CompareAndSwap(false, true) {
		logrus.Info("worker already running, skipping cycle")
		return
	}
	defer w.running.Store(false)

	logger.Info("starting new processing cycle")
	if err := w.process(ctx); err != nil {
		logrus.Error(err)
	}
}

func (w *Worker) process(ctx context.Context) error {
	if err := w.sentimentMLClient.HealthCheck(); err != nil {
		logrus.WithError(err).Warn("ML service unavailable, skipping analytics tick")
		return err
	}

	var (
		mu     sync.Mutex
		allRes []crawler.CrawlResult
	)

	g, gCtx := errgroup.WithContext(ctx)
	for _, c := range []crawler.Crawler{w.webCrawler, w.telegramCrawler} {
		g.Go(func() error {
			res := c.Crawl(gCtx)
			mu.Lock()
			allRes = append(allRes, res...)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	brands, _, err := w.brandUseCase.List(ctx, 100, 0)
	if err != nil {
		return fmt.Errorf("list brands: %w", err)
	}
	logrus.Infof("found %d brands, %d crawl results", len(brands), len(allRes))
	return w.processResults(ctx, allRes, brands)
}

type processedPair struct {
	item   crawler.Item
	brand  entity.Brand
	output entity.SentimentMLOutput
}

func (w *Worker) processResults(
	ctx context.Context,
	results []crawler.CrawlResult,
	brands []entity.Brand,
) error {
	g, _ := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 10)

	var (
		mu       sync.Mutex
		allPairs []processedPair
	)

	for _, res := range results {
		if res.Error != nil {
			logrus.Warn("error in item: ", res.Error)
			continue
		}
		for _, item := range res.Items {
			for _, brand := range brands {
				g.Go(func() error {
					sem <- struct{}{}
					defer func() { <-sem }()

					pair := w.processOnePair(item, brand)
					if pair != nil {
						mu.Lock()
						allPairs = append(allPairs, *pair)
						mu.Unlock()
					}
					return nil
				})
			}
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	items, sentimentResults := w.prepareForBatchInsert(allPairs)
	logrus.Infof("preparing to batch insert %d items, %d sentiment results with sentiment results", len(items), len(sentimentResults))
	if err := w.analyticsRepo.BatchInsertItems(ctx, items, sentimentResults); err != nil {
		return fmt.Errorf("batch insert: %w", err)
	}

	logrus.Infof("batch inserted %d items with sentiment results", len(allPairs))
	return nil
}

func (w *Worker) processOnePair(
	item crawler.Item,
	brand entity.Brand,
) *processedPair {
	relevance, err := w.sentimentMLClient.Relevance(item.Text, brand.Name, brand.Keywords)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"item":  item.Text,
			"brand": brand.ID,
		}).Warn("sentiment_ml relevance failed: ", err)
		return nil
	}
	if !relevance.IsRelevant {
		// Если ML вернул neutral — проверяем risk words бренда
		if strings.EqualFold(relevance.Label, "neutral") && containsRiskWord(item.Text, brand.RiskWords) {
			logrus.Debugf("neutral relevance overridden by risk word match for brand %s", brand.Name)
		} else {
			return nil
		}
	}

	output, err := w.sentimentMLClient.Sentiment(item.Text, brand.Name)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"item":  item.Text,
			"brand": brand.ID,
		}).Warn("sentiment_ml sentiment failed: ", err)
		return nil
	}

	return &processedPair{
		item:   item,
		brand:  brand,
		output: output,
	}
}

func (w *Worker) prepareForBatchInsert(allPairs []processedPair) ([]entity.CrawlerItem, []entity.SentimentMLResult) {
	seenItems := make(map[uuid.UUID]struct{})
	items := make([]entity.CrawlerItem, 0)
	sentimentResults := make([]entity.SentimentMLResult, 0, len(allPairs))

	for _, pair := range allPairs {
		if _, seen := seenItems[pair.item.Uuid]; !seen {
			seenItems[pair.item.Uuid] = struct{}{}
			items = append(items, entity.CrawlerItem{
				ID:          pair.item.Uuid,
				Text:        pair.item.Text,
				Link:        pair.item.Link,
				PublishedAt: pair.item.Datetime,
				SourceID:    pair.item.SourceID,
			})
		}

		sentimentResults = append(sentimentResults, entity.SentimentMLResult{
			ItemID:     pair.item.Uuid,
			BrandID:    pair.brand.ID,
			Sentiment:  string(pair.output.Sentiment),
			Confidence: pair.output.Confidence,
		})
	}

	return items, sentimentResults
}

// containsRiskWord проверяет, содержит ли текст хотя бы одно risk-слово бренда (регистронезависимо).
func containsRiskWord(text string, riskWords []string) bool {
	lower := strings.ToLower(text)
	for _, word := range riskWords {
		if word != "" && strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}
