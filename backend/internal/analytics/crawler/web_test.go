package crawler

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ref      string
		base     string
		expected string
	}{
		{
			name:     "empty ref",
			ref:      "",
			base:     "https://example.com/news/",
			expected: "",
		},
		{
			name:     "absolute url",
			ref:      "https://other.example/post",
			base:     "https://example.com/news/",
			expected: "https://other.example/post",
		},
		{
			name:     "relative path",
			ref:      "/post/1",
			base:     "https://example.com/news/",
			expected: "https://example.com/post/1",
		},
		{
			name:     "invalid base",
			ref:      "post/1",
			base:     "://bad base",
			expected: "post/1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, resolveURL(tc.ref, tc.base))
		})
	}
}

func TestWebCrawlerParse(t *testing.T) {
	t.Parallel()

	crawler := &WebCrawler{PostsLimit: 2}
	html := `
<html><body>
  <article>
    <h2><a href="/post/1">  News 1  </a></h2>
    <div class="meta">Опубликовано: 2026-01-01</div>
    <p> First text </p>
  </article>
  <article>
    <h2><a href="/post/2">News 2</a></h2>
    <div class="meta">2026-01-02</div>
    <p>Second text</p>
  </article>
  <article>
    <h2><a href="/post/3">News 3</a></h2>
    <p>Third text</p>
  </article>
</body></html>`

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(html)),
	}

	items, err := crawler.parse(resp, "https://example.com")
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "https://example.com/post/1", items[0].Link)
}

func TestWebCrawlerParse_BadStatus(t *testing.T) {
	t.Parallel()

	crawler := &WebCrawler{PostsLimit: 1}
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("bad")),
	}

	_, err := crawler.parse(resp, "https://example.com")
	require.Error(t, err)
}
