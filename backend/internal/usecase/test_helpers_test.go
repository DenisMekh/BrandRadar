package usecase

import (
	"context"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

func testCtx() context.Context {
	return context.Background()
}

var testLoggerOnce sync.Once

func setupTestLogger() {
	testLoggerOnce.Do(func() {
		logger := logrus.New()
		logger.SetLevel(logrus.PanicLevel)
		logger.SetOutput(io.Discard)

		logrus.SetLevel(logger.GetLevel())
		logrus.SetOutput(io.Discard)
	})
}
