package logger

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type Formatter struct{}

func getColor(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "\033[37m" // Белый
	case logrus.InfoLevel:
		return "\033[37m" // Белый
	case logrus.WarnLevel:
		return "\033[33m" // Желтый
	case logrus.ErrorLevel:
		return "\033[31m" // Красный
	case logrus.FatalLevel:
		return "\033[35m" // Фиолетовый
	case logrus.PanicLevel:
		return "\033[35m" // Фиолетовый
	default:
		return "\033[0m" // Без цвета
	}
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	color := getColor(entry.Level)
	reset := "\033[0m"

	level := "[" + strings.ToUpper(entry.Level.String()) + "]"
	time := "[" + entry.Time.Format("2006-01-02 15:04:05") + "]"

	return []byte(color + level + " " + time + " " + entry.Message + reset + "\n"), nil
}

func Init() {
	logrus.SetFormatter(&Formatter{})
	logrus.SetLevel(logrus.DebugLevel)
}
