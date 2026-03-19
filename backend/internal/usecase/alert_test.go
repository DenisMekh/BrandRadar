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

func TestAlertUseCase_CreateConfig_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	uc := NewAlertUseCase(configRepo, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)

	cfg := &entity.AlertConfig{
		BrandID:         uuid.New(),
		WindowMinutes:   15,
		CooldownMinutes: 30,
	}

	configRepo.On("Create", mock.Anything, cfg).Return(nil).Once()

	// Act
	err := uc.CreateConfig(testCtx(), cfg)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, cfg.ID)
	configRepo.AssertExpectations(t)
}

func TestAlertUseCase_CreateConfig_Validation(t *testing.T) {
	setupTestLogger()

	// Arrange — nil BrandID triggers validation error
	uc := NewAlertUseCase(&mockrepo.MockAlertConfigRepository{}, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)

	// Act
	err := uc.CreateConfig(testCtx(), &entity.AlertConfig{
		BrandID:         uuid.Nil,
		WindowMinutes:   15,
		CooldownMinutes: 30,
	})

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrValidation)
}

func TestAlertUseCase_CreateConfig_Duplicate(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	uc := NewAlertUseCase(configRepo, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)
	cfg := &entity.AlertConfig{
		BrandID:         uuid.New(),
		WindowMinutes:   15,
		CooldownMinutes: 30,
	}
	configRepo.On("Create", mock.Anything, cfg).Return(entity.ErrDuplicate).Once()

	// Act
	err := uc.CreateConfig(testCtx(), cfg)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrDuplicate)
	configRepo.AssertExpectations(t)
}

func TestAlertUseCase_UpdateConfig_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	uc := NewAlertUseCase(configRepo, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)
	cfg := &entity.AlertConfig{
		ID:              uuid.New(),
		BrandID:         uuid.New(),
		WindowMinutes:   15,
		CooldownMinutes: 20,
	}
	configRepo.On("Update", mock.Anything, cfg).Return(nil).Once()

	// Act
	err := uc.UpdateConfig(testCtx(), cfg)

	// Assert
	assert.NoError(t, err)
	configRepo.AssertExpectations(t)
}

func TestAlertUseCase_UpdateConfig_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	uc := NewAlertUseCase(configRepo, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)
	cfg := &entity.AlertConfig{
		ID:              uuid.New(),
		BrandID:         uuid.New(),
		WindowMinutes:   15,
		CooldownMinutes: 20,
	}
	configRepo.On("Update", mock.Anything, cfg).Return(entity.ErrNotFound).Once()

	// Act
	err := uc.UpdateConfig(testCtx(), cfg)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	configRepo.AssertExpectations(t)
}

func TestAlertUseCase_GetConfigByBrand_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	uc := NewAlertUseCase(configRepo, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)
	brandID := uuid.New()
	cfg := &entity.AlertConfig{ID: uuid.New(), BrandID: brandID}
	configRepo.On("GetByBrandID", mock.Anything, brandID).Return(cfg, nil).Once()

	// Act
	got, err := uc.GetConfigByBrand(testCtx(), brandID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, cfg, got)
	configRepo.AssertExpectations(t)
}

func TestAlertUseCase_GetConfigByBrand_NotFound(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	uc := NewAlertUseCase(configRepo, &mockrepo.MockAlertRepository{}, &mockrepo.MockMentionRepository{}, &mockrepo.MockBrandRepository{}, &mockrepo.MockCooldownCache{}, &mockrepo.MockEventRepository{}, time.Minute, nil)
	brandID := uuid.New()

	configRepo.On("GetByBrandID", mock.Anything, brandID).Return((*entity.AlertConfig)(nil), entity.ErrNotFound).Once()

	// Act
	got, err := uc.GetConfigByBrand(testCtx(), brandID)

	// Assert
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
	configRepo.AssertExpectations(t)
}

func TestAlertUseCase_CheckAndFire_NoConfig(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	alertRepo := &mockrepo.MockAlertRepository{}
	mentionRepo := &mockrepo.MockMentionRepository{}
	brandRepo := &mockrepo.MockBrandRepository{}
	cooldown := &mockrepo.MockCooldownCache{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewAlertUseCase(configRepo, alertRepo, mentionRepo, brandRepo, cooldown, eventRepo, time.Minute, nil)
	brandID := uuid.New()
	configRepo.On("GetByBrandID", mock.Anything, brandID).Return((*entity.AlertConfig)(nil), entity.ErrNotFound).Once()

	// Act
	err := uc.CheckAndFire(testCtx(), brandID)

	// Assert
	assert.NoError(t, err)
	configRepo.AssertExpectations(t)
	alertRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAlertUseCase_CheckAndFire_WithAlert(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	alertRepo := &mockrepo.MockAlertRepository{}
	mentionRepo := &mockrepo.MockMentionRepository{}
	brandRepo := &mockrepo.MockBrandRepository{}
	cooldown := &mockrepo.MockCooldownCache{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewAlertUseCase(configRepo, alertRepo, mentionRepo, brandRepo, cooldown, eventRepo, time.Minute, nil)
	brandID := uuid.New()
	cfg := &entity.AlertConfig{
		ID:              uuid.New(),
		BrandID:         brandID,
		WindowMinutes:   15,
		CooldownMinutes: 30,
		Enabled:         true,
	}
	configRepo.On("GetByBrandID", mock.Anything, brandID).Return(cfg, nil).Once()
	mentionRepo.On("CountByBrandSince", mock.Anything, brandID, mock.Anything, "").Return(int64(15), nil).Once()
	cooldown.On("TryLock", mock.Anything, cfg.ID, cfg.CooldownMinutes).Return(true, nil).Once()
	alertRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Alert")).Return(nil).Once()
	eventRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *entity.Event) bool {
		return e.Type == entity.EventAlertFired
	})).Return(nil).Once()
	brandRepo.On("GetByID", mock.Anything, brandID).Return(&entity.Brand{Name: "TestBrand"}, nil).Once()

	// Act
	err := uc.CheckAndFire(testCtx(), brandID)

	// Assert
	assert.NoError(t, err)
	configRepo.AssertExpectations(t)
	mentionRepo.AssertExpectations(t)
	cooldown.AssertExpectations(t)
	alertRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestAlertUseCase_CheckAndFire_Cooldown(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	alertRepo := &mockrepo.MockAlertRepository{}
	mentionRepo := &mockrepo.MockMentionRepository{}
	brandRepo := &mockrepo.MockBrandRepository{}
	cooldown := &mockrepo.MockCooldownCache{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewAlertUseCase(configRepo, alertRepo, mentionRepo, brandRepo, cooldown, eventRepo, time.Minute, nil)

	brandID := uuid.New()
	cfg := &entity.AlertConfig{
		ID:              uuid.New(),
		BrandID:         brandID,
		WindowMinutes:   15,
		CooldownMinutes: 30,
		Enabled:         true,
	}

	configRepo.On("GetByBrandID", mock.Anything, brandID).Return(cfg, nil).Once()
	mentionRepo.On("CountByBrandSince", mock.Anything, brandID, mock.Anything, "").Return(int64(15), nil).Once()
	cooldown.On("TryLock", mock.Anything, cfg.ID, cfg.CooldownMinutes).Return(false, nil).Once()

	// Act
	err := uc.CheckAndFire(testCtx(), brandID)

	// Assert
	assert.NoError(t, err)
	alertRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	configRepo.AssertExpectations(t)
	mentionRepo.AssertExpectations(t)
	cooldown.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestAlertUseCase_ListAlerts_Success(t *testing.T) {
	setupTestLogger()

	// Arrange
	configRepo := &mockrepo.MockAlertConfigRepository{}
	alertRepo := &mockrepo.MockAlertRepository{}
	mentionRepo := &mockrepo.MockMentionRepository{}
	cooldown := &mockrepo.MockCooldownCache{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := NewAlertUseCase(configRepo, alertRepo, mentionRepo, &mockrepo.MockBrandRepository{}, cooldown, eventRepo, time.Minute, nil)
	brandID := uuid.New()
	alerts := []entity.Alert{{ID: uuid.New(), BrandID: brandID}}
	alertRepo.On("ListByBrandID", mock.Anything, brandID, 10, 0).Return(alerts, int64(1), nil).Once()

	// Act
	items, total, err := uc.ListAlerts(testCtx(), brandID, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, total)
	alertRepo.AssertExpectations(t)
	configRepo.AssertExpectations(t)
	mentionRepo.AssertExpectations(t)
	cooldown.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}
