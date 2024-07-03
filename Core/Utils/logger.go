package Utils

import (
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
)

func LogError(message string, key string, value string) {
	log.ErrorLevelStyle = lipgloss.NewStyle().SetString("ERROR").Foreground(lipgloss.Color("203"))
	Logger.Error(message, key, value)
}

func LogSuccess(message string, key string, value string) {
	log.InfoLevelStyle = lipgloss.NewStyle().SetString("SUCCESS").Foreground(lipgloss.Color("#6c02ff"))
	Logger.Info(message, key, value)
}

func LogInfo(message string, key string, value string) {
	log.WarnLevelStyle = lipgloss.NewStyle().SetString("INFO").Foreground(lipgloss.Color("#51d1f6"))
	Logger.Warn(message, key, value)
}

func LogPanic(message string, key string, value string) {
	Logger.Fatal(message, key, value)
}
