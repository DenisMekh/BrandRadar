package usecase

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestMentionUseCase_Create_Success(t *testing.T) {
	setupTestLogger()

	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	configRepo := &mockrepo.MockAlertConfigRepository{}
	alertRepo := &mockrepo.MockAlertRepository{}
	cooldown := &mockrepo.MockCooldownCache{}

	brandID := uuid.New()
	sourceID := uuid.New()

	alertUC := NewAlertUseCase(configRepo, alertRepo, mentionRepo, &mockrepo.MockBrandRepository{}, cooldown, eventRepo, time.Minute, nil)
	uc := NewMentionUseCase(mentionRepo, eventRepo, alertUC, nil)

	mention := &entity.Mention{
		BrandID:   brandID,
		SourceID:  sourceID,
		Text:      "текст",
		Sentiment: entity.SENTIMENT_POSITIVE,
	}

	mentionRepo.On("Create", mock.Anything, mention).Return(nil).Once()
	eventRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *entity.Event) bool {
		return e.Type == entity.EventMentionCreated
	})).Return(nil).Once()
	configRepo.On("GetByBrandID", mock.Anything, brandID).Return((*entity.AlertConfig)(nil), entity.ErrNotFound).Once()

	result, err := uc.Create(testCtx(), mention)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID)
	mentionRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
	configRepo.AssertExpectations(t)
}

func TestMentionUseCase_Create_EmptyText(t *testing.T) {
	setupTestLogger()

	mentionRepo := &mockrepo.MockMentionRepository{}
	uc := NewMentionUseCase(mentionRepo, &mockrepo.MockEventRepository{}, nil, nil)

	mention := &entity.Mention{
		BrandID:  uuid.New(),
		SourceID: uuid.New(),
		Text:     "",
	}

	_, err := uc.Create(testCtx(), mention)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
	mentionRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mentionRepo.AssertExpectations(t)
}

func TestMentionUseCase_GetByID_Success(t *testing.T) {
	setupTestLogger()

	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, nil)
	id := uuid.New()
	expected := &entity.Mention{ID: id, Text: "mention"}
	mentionRepo.On("GetByID", mock.Anything, id).Return(expected, nil).Once()

	got, err := uc.GetByID(testCtx(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	mentionRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestMentionUseCase_GetByID_NotFound(t *testing.T) {
	setupTestLogger()

	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, nil)
	id := uuid.New()
	mentionRepo.On("GetByID", mock.Anything, id).Return((*entity.Mention)(nil), entity.ErrNotFound).Once()

	got, err := uc.GetByID(testCtx(), id)

	assert.Nil(t, got)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	mentionRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestMentionUseCase_List_Success(t *testing.T) {
	setupTestLogger()

	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, nil)
	filter := entity.MentionFilter{Limit: 20, Offset: 0}
	items := []entity.Mention{{ID: uuid.New()}, {ID: uuid.New()}}
	mentionRepo.On("List", mock.Anything, filter).Return(items, int64(2), nil).Once()

	gotItems, total, err := uc.List(testCtx(), filter)

	assert.NoError(t, err)
	assert.Len(t, gotItems, 2)
	assert.Equal(t, 2, total)
	mentionRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestMentionUseCase_List_WithFilters(t *testing.T) {
	setupTestLogger()

	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewMentionUseCase(mentionRepo, eventRepo, nil, nil)
	filter := entity.MentionFilter{
		BrandID:   uuid.New(),
		Sentiment: "negative",
		Limit:     20,
		Offset:    0,
	}
	mentionRepo.On("List", mock.Anything, filter).Return([]entity.Mention{}, int64(0), nil).Once()

	items, total, err := uc.List(testCtx(), filter)

	assert.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
	mentionRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}
