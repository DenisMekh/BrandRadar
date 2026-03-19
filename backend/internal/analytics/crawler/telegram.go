package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/config"
)

type Crawler interface {
	Crawl(ctx context.Context) []CrawlResult
}

type telegramSource struct {
	ID     *uuid.UUID
	Handle string
}

// getChannels читает telegram-каналы из БД (активные источники).
// Фолбэк на конфиг если sourceReader не задан.
func (c *TelegramCrawler) getChannels(ctx context.Context) []telegramSource {
	if c.sourceReader != nil {
		sources, err := c.sourceReader.ListActive(ctx)
		if err != nil {
			logrus.Warnf("TelegramCrawler: failed to read sources from DB: %v", err)
		} else {
			var entries []telegramSource
			for i := range sources {
				if sources[i].Type != "telegram" {
					continue
				}
				handle := sources[i].TelegramHandle()
				if handle != "" {
					id := sources[i].ID
					entries = append(entries, telegramSource{ID: &id, Handle: handle})
				}
			}
			if len(entries) > 0 {
				return entries
			}
		}
	}

	// Фолбэк на конфиг (source_id будет nil)
	cfg := config.Get()
	channels := cfg.Crawler.TelegramChannels
	if len(channels) == 0 {
		channels = []string{"brand_radar_case"}
	}
	entries := make([]telegramSource, len(channels))
	for i, ch := range channels {
		entries[i] = telegramSource{Handle: ch}
	}
	return entries
}

const (
	DefaultTimeout    = 30
	DefaultPostsLimit = 10
)

type TelegramCrawler struct {
	Client       *http.Client
	Deduplicator Deduplicator
	PostsLimit   int
	sourceReader SourceReader
}

func NewTelegramCrawler(deduplicator Deduplicator, sourceReader SourceReader) *TelegramCrawler {
	cfg := config.Get()
	postsLimit := cfg.Crawler.PostsLimit
	if postsLimit <= 0 {
		postsLimit = DefaultPostsLimit
	}
	timeout := cfg.Crawler.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return &TelegramCrawler{
		Client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		Deduplicator: deduplicator,
		PostsLimit:   postsLimit,
		sourceReader: sourceReader,
	}
}

func (c *TelegramCrawler) fetch(ctx context.Context, channelURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, channelURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

func (c *TelegramCrawler) parse(resp *http.Response) ([]Item, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			println("error closing response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var items []Item

	doc.Find(".tgme_widget_message_wrap").Each(func(i int, s *goquery.Selection) {
		if len(items) >= c.PostsLimit {
			return
		}

		msgWidget := s.Find(".tgme_widget_message")
		if msgWidget.Length() == 0 {
			return
		}

		item := Item{}

		if dataPost, exists := msgWidget.Attr("data-post"); exists {
			item.Link = "https://t.me/" + dataPost
		}

		textEl := msgWidget.Find(".tgme_widget_message_text")
		if textEl.Length() > 0 {
			item.Text = strings.TrimSpace(textEl.Text())
		}

		timeEl := msgWidget.Find(".tgme_widget_message_date time")
		if timeEl.Length() > 0 {
			if datetime, exists := timeEl.Attr("datetime"); exists {
				item.Datetime, err = time.Parse(time.RFC3339Nano, datetime)
				if err != nil {
					return
				}
			}
		}

		item.Uuid = uuid.New()

		if item.Text != "" || item.Link != "" {
			items = append(items, item)
		}
	})

	return items, nil
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func (c *TelegramCrawler) Crawl(ctx context.Context) []CrawlResult {
	sources := c.getChannels(ctx)
	if len(sources) == 0 {
		return nil
	}

	var (
		wg      sync.WaitGroup
		results = make([]CrawlResult, 0, len(sources))
		mu      sync.Mutex
	)

	for _, src := range sources {
		wg.Add(1)
		go func(s telegramSource) {
			defer wg.Done()

			channelURL := "https://t.me/s/" + s.Handle

			resp, err := c.fetch(ctx, channelURL)
			if err != nil {
				mu.Lock()
				results = append(results, CrawlResult{Source: s.Handle, Error: err})
				mu.Unlock()
				return
			}

			defer func() {
				err = resp.Body.Close()
				if err != nil {
					println("error closing response body")
				}
			}()

			items, err := c.parse(resp)
			if err != nil {
				mu.Lock()
				results = append(results, CrawlResult{Source: s.Handle, Error: err})
				mu.Unlock()
				return
			}

			for i := range items {
				items[i].SourceID = s.ID
			}

			if c.Deduplicator != nil {
				items, err = c.Deduplicator.FilterDuplicates(ctx, items)
				if err != nil {
					mu.Lock()
					results = append(results, CrawlResult{Source: s.Handle, Error: err})
					mu.Unlock()
					return
				}
			}

			mu.Lock()
			results = append(results, CrawlResult{Source: s.Handle, Items: items})
			mu.Unlock()
		}(src)
	}

	wg.Wait()
	return results
}
