package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentimentCounts_Total(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		counts   SentimentCounts
		expected int64
	}{
		{"all zeros", SentimentCounts{}, 0},
		{"positive only", SentimentCounts{Positive: 10}, 10},
		{"mixed", SentimentCounts{Positive: 5, Negative: 3, Neutral: 7}, 15},
		{"large numbers", SentimentCounts{Positive: 1000, Negative: 2000, Neutral: 3000}, 6000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.counts.Total())
		})
	}
}
