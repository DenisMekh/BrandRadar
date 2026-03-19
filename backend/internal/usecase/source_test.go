package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestSourceUseCase_Create_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)

	source := &entity.Source{
		Type: "telegram",
		Name: "Канал",
		URL:  "durov",
	}

	repo.On("Create", mock.Anything, source).Return(nil).Once()
	repo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"telegram": 1}, nil).Once()

	// Act
	err := uc.Create(testCtx(), source)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, source.ID)
	assert.Equal(t, entity.SourceStatusActive, source.Status)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Create_InvalidType(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	source := &entity.Source{
		Type: "invalid-type",
		Name: "Bad source",
		URL:  "whatever",
	}

	// Act — fails at URL validation (unknown source type)
	err := uc.Create(testCtx(), source)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Create_InvalidTelegramHandle(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	source := &entity.Source{
		Type: "telegram",
		Name: "Bad channel",
		URL:  "not-a-handle",
	}

	err := uc.Create(testCtx(), source)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
}

func TestSourceUseCase_Create_InvalidWebURL(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	source := &entity.Source{
		Type: "web",
		Name: "Bad site",
		URL:  "not-a-url",
	}

	err := uc.Create(testCtx(), source)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
}

func TestSourceUseCase_GetByID_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	id := uuid.New()
	expected := &entity.Source{ID: id, Name: "source"}
	repo.On("GetByID", mock.Anything, id).Return(expected, nil).Once()

	// Act
	got, err := uc.GetByID(testCtx(), id)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_GetByID_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return((*entity.Source)(nil), entity.ErrNotFound).Once()

	// Act
	got, err := uc.GetByID(testCtx(), id)

	// Assert
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_List_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	items := []entity.Source{{ID: uuid.New()}, {ID: uuid.New()}}
	repo.On("List", mock.Anything, 20, 0).Return(items, int64(2), nil).Once()

	// Act
	gotItems, total, err := uc.List(testCtx(), 20, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, gotItems, 2)
	assert.Equal(t, 2, total)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_List_Empty(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	repo.On("List", mock.Anything, 20, 0).Return([]entity.Source{}, int64(0), nil).Once()

	// Act
	gotItems, total, err := uc.List(testCtx(), 20, 0)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, gotItems)
	assert.Equal(t, 0, total)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Toggle_ActiveToInactive(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)

	id := uuid.New()
	source := &entity.Source{ID: id, Status: entity.SourceStatusActive}

	repo.On("GetByID", mock.Anything, id).Return(source, nil).Once()
	repo.On("Update", mock.Anything, source).Return(nil).Once()
	repo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"telegram": 0}, nil).Once()
	eventRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *entity.Event) bool {
		return e.Type == entity.EventSourceToggled
	})).Return(nil).Once()

	// Act
	res, err := uc.Toggle(testCtx(), id)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, entity.SourceStatusInactive, res.Status)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Toggle_InactiveToActive(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)

	id := uuid.New()
	source := &entity.Source{ID: id, Status: entity.SourceStatusInactive}

	repo.On("GetByID", mock.Anything, id).Return(source, nil).Once()
	repo.On("Update", mock.Anything, source).Return(nil).Once()
	repo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"telegram": 1}, nil).Once()
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Event")).Return(nil).Once()

	// Act
	res, err := uc.Toggle(testCtx(), id)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, entity.SourceStatusActive, res.Status)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Toggle_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return((*entity.Source)(nil), entity.ErrNotFound).Once()

	// Act
	res, err := uc.Toggle(testCtx(), id)

	// Assert
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Delete_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil).Once()
	repo.On("CountActiveByType", mock.Anything).Return(map[string]int64{}, nil).Once()

	// Act
	err := uc.Delete(testCtx(), id)

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Delete_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(entity.ErrNotFound).Once()

	// Act
	err := uc.Delete(testCtx(), id)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	repo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceUseCase_Create_EmptyName(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	source := &entity.Source{Type: "web", Name: "  ", URL: "https://example.com"}

	err := uc.Create(testCtx(), source)

	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSourceUseCase_Create_EmptyType(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	source := &entity.Source{Type: "", Name: "Site", URL: "https://example.com"}

	err := uc.Create(testCtx(), source)

	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSourceUseCase_List_DefaultLimit(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	repo.On("List", mock.Anything, 20, 0).Return([]entity.Source{}, int64(0), nil).Once()

	_, _, err := uc.List(testCtx(), 0, 0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSourceUseCase_List_MaxLimit(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	repo.On("List", mock.Anything, 100, 0).Return([]entity.Source{}, int64(0), nil).Once()

	_, _, err := uc.List(testCtx(), 500, 0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSourceUseCase_List_NegativeOffset(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewSourceUseCase(repo, eventRepo)
	repo.On("List", mock.Anything, 20, 0).Return([]entity.Source{}, int64(0), nil).Once()

	_, _, err := uc.List(testCtx(), 0, -10)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
