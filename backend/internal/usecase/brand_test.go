package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestBrandUseCase_Create_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)

	brand := &entity.Brand{
		Name:     " Brand A ",
		Keywords: []string{"brand", "news"},
	}

	repo.On("Create", mock.Anything, brand).Return(nil).Once()
	repo.On("Count", mock.Anything).Return(int64(1), nil).Once()

	// Act
	err := uc.Create(testCtx(), brand)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, brand.ID)
	assert.Equal(t, "Brand A", brand.Name)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Create_EmptyName(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)

	brand := &entity.Brand{
		Name:     "  ",
		Keywords: []string{"brand"},
	}

	// Act
	err := uc.Create(testCtx(), brand)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Create_NoKeywords(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)

	brand := &entity.Brand{
		Name:     "Бренд B",
		Keywords: []string{},
	}

	// Act
	err := uc.Create(testCtx(), brand)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Create_Duplicate(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)

	brand := &entity.Brand{
		Name:     "Brand C",
		Keywords: []string{"brand"},
	}
	repo.On("Create", mock.Anything, brand).Return(entity.ErrDuplicate).Once()

	// Act
	err := uc.Create(testCtx(), brand)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrDuplicate)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_GetByID_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	id := uuid.New()
	expected := &entity.Brand{ID: id, Name: "Brand"}
	repo.On("GetByID", mock.Anything, id).Return(expected, nil).Once()

	// Act
	got, err := uc.GetByID(testCtx(), id)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_GetByID_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return((*entity.Brand)(nil), entity.ErrNotFound).Once()

	// Act
	got, err := uc.GetByID(testCtx(), id)

	// Assert
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_List_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	items := []entity.Brand{{ID: uuid.New(), Name: "A"}, {ID: uuid.New(), Name: "B"}}
	repo.On("List", mock.Anything, 20, 0).Return(items, int64(2), nil).Once()

	// Act
	gotItems, total, err := uc.List(testCtx(), 20, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, gotItems, 2)
	assert.Equal(t, 2, total)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_List_EmptyResult(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	repo.On("List", mock.Anything, 20, 0).Return([]entity.Brand{}, int64(0), nil).Once()

	// Act
	gotItems, total, err := uc.List(testCtx(), 20, 0)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, gotItems)
	assert.Equal(t, 0, total)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_List_Pagination(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	repo.On("List", mock.Anything, 2, 0).Return([]entity.Brand{}, int64(0), nil).Once()

	// Act
	_, _, err := uc.List(testCtx(), 2, 0)

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Update_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	brand := &entity.Brand{
		ID:       uuid.New(),
		Name:     " Brand ",
		Keywords: []string{"k1"},
	}
	repo.On("Update", mock.Anything, brand).Return(nil).Once()

	// Act
	err := uc.Update(testCtx(), brand)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Brand", brand.Name)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Update_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	brand := &entity.Brand{
		ID:       uuid.New(),
		Name:     "Brand",
		Keywords: []string{"k1"},
	}
	repo.On("Update", mock.Anything, brand).Return(entity.ErrNotFound).Once()

	// Act
	err := uc.Update(testCtx(), brand)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Delete_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil).Once()
	repo.On("Count", mock.Anything).Return(int64(0), nil).Once()

	// Act
	err := uc.Delete(testCtx(), id)

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Delete_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(entity.ErrNotFound).Once()

	// Act
	err := uc.Delete(testCtx(), id)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_List_DefaultLimit(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	repo.On("List", mock.Anything, 20, 0).Return([]entity.Brand{}, int64(0), nil).Once()

	_, _, err := uc.List(testCtx(), 0, 0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_List_MaxLimit(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	repo.On("List", mock.Anything, 100, 0).Return([]entity.Brand{}, int64(0), nil).Once()

	_, _, err := uc.List(testCtx(), 500, 0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_List_NegativeOffset(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	repo.On("List", mock.Anything, 20, 0).Return([]entity.Brand{}, int64(0), nil).Once()

	_, _, err := uc.List(testCtx(), 0, -5)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBrandUseCase_Update_EmptyName(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	brand := &entity.Brand{ID: uuid.New(), Name: "  ", Keywords: []string{"k"}}

	err := uc.Update(testCtx(), brand)

	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestBrandUseCase_Update_NoKeywords(t *testing.T) {
	setupTestLogger()

	repo := &mockrepo.MockBrandRepository{}
	uc := NewBrandUseCase(repo)
	brand := &entity.Brand{ID: uuid.New(), Name: "Brand", Keywords: []string{}}

	err := uc.Update(testCtx(), brand)

	assert.ErrorIs(t, err, entity.ErrValidation)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
