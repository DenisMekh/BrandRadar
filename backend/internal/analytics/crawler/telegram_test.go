package crawler

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "abc",
			maxLen:   5,
			expected: "abc",
		},
		{
			name:     "long string",
			input:    "abcdefghijklmnopqrstuvwxyz",
			maxLen:   5,
			expected: "abcde...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, truncate(tc.input, tc.maxLen))
		})
	}
}

func TestTelegramCrawlerParse(t *testing.T) {
	t.Parallel()

	crawler := &TelegramCrawler{PostsLimit: 2}
	html := `
<html><body>
  <div class="tgme_widget_message_wrap">
    <div class="tgme_widget_message" data-post="brand/1">
      <div class="tgme_widget_message_text">First message</div>
      <div class="tgme_widget_message_date"><time datetime="2026-01-01T00:00:00Z"></time></div>
    </div>
  </div>
  <div class="tgme_widget_message_wrap">
    <div class="tgme_widget_message" data-post="brand/2">
      <div class="tgme_widget_message_text">Second message</div>
      <div class="tgme_widget_message_date"><time datetime="2026-01-02T00:00:00Z"></time></div>
    </div>
  </div>
  <div class="tgme_widget_message_wrap">
    <div class="tgme_widget_message" data-post="brand/3">
      <div class="tgme_widget_message_text">Third message</div>
      <div class="tgme_widget_message_date"><time datetime="2026-01-03T00:00:00Z"></time></div>
    </div>
  </div>
</body></html>`

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(html)),
	}

	items, err := crawler.parse(resp)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "https://t.me/brand/1", items[0].Link)
}

func TestTelegramCrawlerParse_BadStatus(t *testing.T) {
	t.Parallel()

	crawler := &TelegramCrawler{PostsLimit: 1}
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("bad")),
	}

	_, err := crawler.parse(resp)
	require.Error(t, err)
}
