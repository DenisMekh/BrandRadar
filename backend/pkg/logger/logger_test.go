package logger

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInit_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		Init()
	})
}

func TestInit_SetsDebugLevel(t *testing.T) {
	Init()
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
}

func TestInit_SetsCustomFormatter(t *testing.T) {
	Init()

	formatter := logrus.StandardLogger().Formatter
	_, ok := formatter.(*Formatter)
	assert.True(t, ok, "formatter should be *logger.Formatter")
}

func TestFormatter_Format_AllLevels(t *testing.T) {
	f := &Formatter{}

	levels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			entry := &logrus.Entry{
				Level:   level,
				Message: "test message",
				Logger:  logrus.StandardLogger(),
			}

			bs, err := f.Format(entry)
			assert.NoError(t, err)
			assert.Contains(t, string(bs), "test message")
			assert.Contains(t, string(bs), "["+strings.ToUpper(level.String())+"]")
		})
	}
}

func TestGetColor_ReturnsNonEmpty(t *testing.T) {
	levels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
		logrus.TraceLevel,
	}

	for _, level := range levels {
		color := getColor(level)
		assert.NotEmpty(t, color, "color for %s should not be empty", level)
	}
}
