package usecase

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestMentionUseCase_List_CacheHit(t *testing.T) {
	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	cache := &mockrepo.MockMentionCache{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, cache)

	filter := entity.MentionFilter{Limit: 10, Offset: 0}
	cachedItems := []entity.Mention{{ID: uuid.New()}}
	cache.On("GetList", mock.Anything, filter).Return(cachedItems, int64(1), nil).Once()

	items, total, err := uc.List(testCtx(), filter)

	assert.NoError(t, err)
	assert.Equal(t, cachedItems, items)
	assert.Equal(t, 1, total)
	mentionRepo.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
	cache.AssertExpectations(t)
}

func TestMentionUseCase_List_CacheMiss(t *testing.T) {
	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	cache := &mockrepo.MockMentionCache{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, cache)

	filter := entity.MentionFilter{Limit: 10, Offset: 5}
	repoItems := []entity.Mention{{ID: uuid.New()}}
	cache.On("GetList", mock.Anything, filter).Return(nil, int64(0), nil).Once()
	mentionRepo.On("List", mock.Anything, filter).Return(repoItems, int64(1), nil).Once()
	cache.On("SetList", mock.Anything, filter, repoItems, int64(1)).Return(nil).Once()

	items, total, err := uc.List(testCtx(), filter)

	assert.NoError(t, err)
	assert.Equal(t, repoItems, items)
	assert.Equal(t, 1, total)
	mentionRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestMentionUseCase_List_CacheErrorFallback(t *testing.T) {
	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	cache := &mockrepo.MockMentionCache{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, cache)

	filter := entity.MentionFilter{Limit: 20, Offset: 0}
	repoItems := []entity.Mention{{ID: uuid.New()}}
	cache.On("GetList", mock.Anything, filter).Return(nil, int64(0), errors.New("redis down")).Once()
	mentionRepo.On("List", mock.Anything, filter).Return(repoItems, int64(1), nil).Once()
	cache.On("SetList", mock.Anything, filter, repoItems, int64(1)).Return(nil).Once()

	items, total, err := uc.List(testCtx(), filter)

	assert.NoError(t, err)
	assert.Equal(t, repoItems, items)
	assert.Equal(t, 1, total)
	mentionRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestMentionUseCase_List_NilCache(t *testing.T) {
	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, nil)

	filter := entity.MentionFilter{Limit: 0, Offset: -10}
	expectedFilter := entity.MentionFilter{Limit: 20, Offset: 0}
	repoItems := []entity.Mention{{ID: uuid.New()}}
	mentionRepo.On("List", mock.Anything, expectedFilter).Return(repoItems, int64(1), nil).Once()

	items, total, err := uc.List(testCtx(), filter)

	assert.NoError(t, err)
	assert.Equal(t, repoItems, items)
	assert.Equal(t, 1, total)
	mentionRepo.AssertExpectations(t)
}

func TestMentionUseCase_Create_InvalidatesCache(t *testing.T) {
	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	cache := &mockrepo.MockMentionCache{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, cache)

	brandID := uuid.New()
	mention := &entity.Mention{
		BrandID:   brandID,
		SourceID:  uuid.New(),
		Text:      "text",
		Sentiment: entity.SENTIMENT_NEGATIVE,
	}

	mentionRepo.On("Create", mock.Anything, mention).Return(nil).Once()
	eventRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *entity.Event) bool {
		return e.Type == entity.EventMentionCreated
	})).Return(nil).Once()
	cache.On("InvalidateByBrand", mock.Anything, brandID).Return(nil).Once()

	result, err := uc.Create(testCtx(), mention)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	cache.AssertExpectations(t)
}
