package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestEventUseCase_Create_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockEventRepository{}
	uc := NewEventUseCase(repo)

	event := &entity.Event{
		Type:    entity.EventMentionCreated,
		Payload: []byte(`{"id":"1"}`),
	}

	repo.On("Create", mock.Anything, event).Return(nil).Once()

	// Act
	err := uc.Create(testCtx(), event)

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestEventUseCase_List_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockEventRepository{}
	uc := NewEventUseCase(repo)
	items := []entity.Event{{Type: entity.EventMentionCreated}}
	repo.On("List", mock.Anything, (*string)(nil), 50, 0).Return(items, int64(1), nil).Once()

	// Act
	gotItems, total, err := uc.List(testCtx(), "", 50, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, gotItems, 1)
	assert.Equal(t, 1, total)
	repo.AssertExpectations(t)
}

func TestEventUseCase_List_FilterByType(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockEventRepository{}
	uc := NewEventUseCase(repo)
	eventType := string(entity.EventAlertFired)
	repo.On("List", mock.Anything, &eventType, 10, 0).Return([]entity.Event{}, int64(0), nil).Once()

	// Act
	items, total, err := uc.List(testCtx(), eventType, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
	repo.AssertExpectations(t)
}

func TestEventUseCase_List_Empty(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockEventRepository{}
	uc := NewEventUseCase(repo)
	repo.On("List", mock.Anything, (*string)(nil), 50, 0).Return([]entity.Event{}, int64(0), nil).Once()

	// Act
	items, total, err := uc.List(testCtx(), "", 0, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, items, 0)
	assert.Equal(t, 0, total)
	repo.AssertExpectations(t)
}
