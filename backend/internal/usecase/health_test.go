package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestHealthUseCase_Check_AllHealthy(t *testing.T) {
	setupTestLogger()

	// Arrange
	checker1 := &mockrepo.MockHealthChecker{}
	checker2 := &mockrepo.MockHealthChecker{}

	checker1.On("Name").Return("postgres").Maybe()
	checker1.On("Ping", mock.Anything).Return(nil).Once()
	checker2.On("Name").Return("redis").Maybe()
	checker2.On("Ping", mock.Anything).Return(nil).Once()

	uc := NewHealthUseCase([]HealthChecker{checker1, checker2}, "test")

	// Act
	status := uc.Check(testCtx())

	// Assert
	assert.Equal(t, "ok", status.Status)
	assert.Equal(t, "ok", status.Dependencies["postgres"])
	assert.Equal(t, "ok", status.Dependencies["redis"])
	checker1.AssertExpectations(t)
	checker2.AssertExpectations(t)
}

func TestHealthUseCase_Check_PostgresDown(t *testing.T) {
	setupTestLogger()

	// Arrange
	postgres := &mockrepo.MockHealthChecker{}
	redis := &mockrepo.MockHealthChecker{}
	postgres.On("Name").Return("postgres").Maybe()
	postgres.On("Ping", mock.Anything).Return(errors.New("down")).Once()
	redis.On("Name").Return("redis").Maybe()
	redis.On("Ping", mock.Anything).Return(nil).Once()

	uc := NewHealthUseCase([]HealthChecker{postgres, redis}, "test")

	// Act
	status := uc.Check(testCtx())

	// Assert
	assert.Equal(t, "degraded", status.Status)
	assert.Equal(t, "fail", status.Dependencies["postgres"])
	assert.Equal(t, "ok", status.Dependencies["redis"])
	postgres.AssertExpectations(t)
	redis.AssertExpectations(t)
}

func TestHealthUseCase_Check_RedisDown(t *testing.T) {
	setupTestLogger()

	// Arrange
	postgres := &mockrepo.MockHealthChecker{}
	redis := &mockrepo.MockHealthChecker{}
	postgres.On("Name").Return("postgres").Maybe()
	postgres.On("Ping", mock.Anything).Return(nil).Once()
	redis.On("Name").Return("redis").Maybe()
	redis.On("Ping", mock.Anything).Return(errors.New("down")).Once()

	uc := NewHealthUseCase([]HealthChecker{postgres, redis}, "test")

	// Act
	status := uc.Check(testCtx())

	// Assert
	assert.Equal(t, "degraded", status.Status)
	assert.Equal(t, "ok", status.Dependencies["postgres"])
	assert.Equal(t, "fail", status.Dependencies["redis"])
	postgres.AssertExpectations(t)
	redis.AssertExpectations(t)
}

func TestHealthUseCase_Check_AllDown(t *testing.T) {
	setupTestLogger()

	// Arrange
	checker1 := &mockrepo.MockHealthChecker{}
	checker2 := &mockrepo.MockHealthChecker{}
	checker1.On("Name").Return("postgres").Maybe()
	checker1.On("Ping", mock.Anything).Return(errors.New("down")).Once()
	checker2.On("Name").Return("redis").Maybe()
	checker2.On("Ping", mock.Anything).Return(errors.New("down")).Once()

	uc := NewHealthUseCase([]HealthChecker{checker1, checker2}, "test")

	// Act
	status := uc.Check(testCtx())

	// Assert
	assert.Equal(t, "degraded", status.Status)
	assert.Equal(t, "fail", status.Dependencies["postgres"])
	assert.Equal(t, "fail", status.Dependencies["redis"])
	checker1.AssertExpectations(t)
	checker2.AssertExpectations(t)
}
