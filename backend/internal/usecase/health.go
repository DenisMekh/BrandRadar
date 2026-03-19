package usecase

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthStatus — агрегированный результат проверки здоровья системы.
type HealthStatus struct {
	Status       string            `json:"status"`
	Dependencies map[string]string `json:"dependencies"`
	Version      string            `json:"version"`
	Uptime       int64             `json:"uptime_seconds"`
}

// HealthUseCase — бизнес-логика проверки доступности зависимостей.
type HealthUseCase struct {
	checkers  []HealthChecker
	startTime time.Time
	version   string
}

// NewHealthUseCase — конструктор usecase health-check.
func NewHealthUseCase(
	checkers []HealthChecker,
	version string,
) *HealthUseCase {
	return &HealthUseCase{
		checkers:  checkers,
		startTime: time.Now(),
		version:   version,
	}
}

// Check выполняет проверку всех внешних зависимостей.
func (uc *HealthUseCase) Check(ctx context.Context) HealthStatus {
	deps := make(map[string]string)
	overall := "ok"

	for _, checker := range uc.checkers {
		if err := checker.Ping(ctx); err != nil {
			deps[checker.Name()] = "fail"
			overall = "degraded"
			logrus.Warnf("health check failed: dependency=%s, error=%v", checker.Name(), err)
		} else {
			deps[checker.Name()] = "ok"
		}
	}

	return HealthStatus{
		Status:       overall,
		Dependencies: deps,
		Version:      uc.version,
		Uptime:       int64(time.Since(uc.startTime).Seconds()),
	}
}
