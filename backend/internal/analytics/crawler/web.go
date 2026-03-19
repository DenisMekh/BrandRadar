package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/config"
)

const (
	DefaultWebTimeout    = 30
	DefaultWebPostsLimit = 1
)

type WebCrawler struct {
	Client       *http.Client
	Deduplicator Deduplicator
	URLs         []string
	PostsLimit   int
	sourceReader SourceReader
}

func NewCrawler(deduplicator Deduplicator, sourceReader SourceReader, postsLimit int) *WebCrawler {
	cfg := config.Get()
	timeout := cfg.Crawler.Timeout
	if timeout <= 0 {
		timeout = DefaultWebTimeout
	}
	limit := postsLimit
	if limit <= 0 {
		limit = DefaultWebPostsLimit
	}

	return &WebCrawler{
		Client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		Deduplicator: deduplicator,
		PostsLimit:   limit,
		sourceReader: sourceReader,
	}
}

type webSource struct {
	ID  *uuid.UUID
	URL string
}

// getURLs читает web/rss URL из БД (активные источники).
// Фолбэк на конфиг если sourceReader не задан.
func (c *WebCrawler) getURLs(ctx context.Context) []webSource {
	sources, err := c.sourceReader.ListActive(ctx)
	if err != nil {
		logrus.Warnf("WebCrawler: failed to read sources from DB: %v", err)
	} else {
		var entries []webSource
		for i := range sources {
			if sources[i].Type == "web" || sources[i].Type == "rss" {
				if sources[i].URL != "" {
					id := sources[i].ID
					entries = append(entries, webSource{ID: &id, URL: sources[i].URL})
				}
			}
		}
		if len(entries) > 0 {
			return entries
		}
	}
	return make([]webSource, 0)
}

func (c *WebCrawler) fetch(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BrandRadar WebCrawler/1.0)")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

func (c *WebCrawler) parse(resp *http.Response, baseURL string) ([]Item, error) {
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

	doc.Find("article").Each(func(i int, s *goquery.Selection) {
		if len(items) >= c.PostsLimit {
			return
		}

		item := Item{}

		titleEl := s.Find("h2 a")
		if titleEl.Length() > 0 {
			if href, exists := titleEl.Attr("href"); exists {
				item.Link = resolveURL(href, baseURL)
			}
		}

		metaEl := s.Find(".meta")
		if metaEl.Length() > 0 {
			metaText := strings.TrimSpace(metaEl.Text())
			if strings.HasPrefix(metaText, "Опубликовано:") {
				metaText = strings.TrimSpace(strings.TrimPrefix(metaText, "Опубликовано:"))
			}

			for _, layout := range []string{time.RFC3339Nano, time.RFC3339, time.DateOnly} {
				if t, parseErr := time.Parse(layout, metaText); parseErr == nil {
					item.Datetime = t
					break
				}
			}
		}

		pEl := s.Find("p")
		if pEl.Length() > 0 {
			item.Text = strings.TrimSpace(pEl.Text())
		}

		item.Uuid = uuid.New()
		if item.Text != "" || item.Link != "" {
			items = append(items, item)
		}
	})

	return items, nil
}

func resolveURL(ref, base string) string {
	if ref == "" {
		return ""
	}

	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}

	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	resolved := baseURL.ResolveReference(refURL)
	return resolved.String()
}

func (c *WebCrawler) Crawl(ctx context.Context) []CrawlResult {
	sources := c.getURLs(ctx)

	var (
		wg      sync.WaitGroup
		results = make([]CrawlResult, 0, len(sources))
		mu      sync.Mutex
	)

	for _, src := range sources {
		wg.Add(1)
		go func(s webSource) {
			defer wg.Done()

			resp, err := c.fetch(ctx, s.URL)
			if err != nil {
				mu.Lock()
				results = append(results, CrawlResult{Source: s.URL, Error: err})
				mu.Unlock()
				return
			}

			defer func() {
				err = resp.Body.Close()
				if err != nil {
					println("error closing response body")
				}
			}()

			items, err := c.parse(resp, s.URL)
			if err != nil {
				mu.Lock()
				results = append(results, CrawlResult{Source: s.URL, Error: err})
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
					results = append(results, CrawlResult{Source: s.URL, Error: err})
					mu.Unlock()
					return
				}
			}

			mu.Lock()
			results = append(results, CrawlResult{Source: s.URL, Items: items})
			mu.Unlock()
		}(src)
	}

	wg.Wait()
	return results
}
