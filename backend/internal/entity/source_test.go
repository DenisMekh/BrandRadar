package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceIsActive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status SourceStatus
		want   bool
	}{
		{
			name:   "active",
			status: SourceStatusActive,
			want:   true,
		},
		{
			name:   "inactive",
			status: SourceStatusInactive,
			want:   false,
		},
		{
			name:   "unknown",
			status: SourceStatus("other"),
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := &Source{Status: tc.status}
			if got := s.IsActive(); got != tc.want {
				t.Fatalf("unexpected active state: got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestValidateSourceURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sourceType string
		rawURL     string
		wantErr    bool
	}{
		// telegram valid
		{"telegram valid handle", "telegram", "durov", false},
		{"telegram valid long handle", "telegram", "channel_name_123", false},
		// telegram invalid
		{"telegram empty url", "telegram", "", true},
		{"telegram no at sign", "telegram", "@durov", true},
		{"telegram starts with digit", "telegram", "1channel", false},
		{"telegram special chars", "telegram", "chan-nel", true},
		// web valid
		{"web valid https", "web", "https://example.com", false},
		{"web valid http", "web", "http://example.com", false},
		{"web valid with path", "web", "https://example.com/path?q=1", false},
		// web invalid
		{"web empty url", "web", "", true},
		{"web no scheme", "web", "example.com", true},
		{"web ftp scheme", "web", "ftp://example.com", true},
		{"web no host", "web", "https://", true},
		// rss valid
		{"rss valid https", "rss", "https://feed.example.com/rss", false},
		{"rss valid http", "rss", "http://feed.example.com/rss", false},
		// rss invalid
		{"rss empty url", "rss", "", true},
		{"rss no scheme", "rss", "feed.example.com/rss", true},
		// unknown type
		{"unknown type", "unknown", "https://example.com", true},
		{"empty type", "", "https://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateSourceURL(tt.sourceType, tt.rawURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSource_TelegramHandle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"with @", "@durov", "durov"},
		{"without @", "durov", "durov"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Source{URL: tt.url}
			assert.Equal(t, tt.want, s.TelegramHandle())
		})
	}
}
