package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMention_Fields(t *testing.T) {
	m := &Mention{
		ID:        uuid.New(),
		Sentiment: SENTIMENT_POSITIVE,
	}
	assert.Equal(t, SENTIMENT_POSITIVE, m.Sentiment)
}
