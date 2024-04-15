package dlog

import (
	"log/slog"
	"testing"
)

func TestLevel(t *testing.T) {
	data := map[slog.Level]slog.Level{
		LevelDebug:    slog.LevelDebug,
		LevelInfo:     slog.LevelInfo,
		LevelWarn:     slog.LevelWarn,
		LevelError:    slog.LevelError,
		LevelFatal:    slog.LevelError + 4,
		LevelDisaster: slog.LevelError + 8,
	}
	for logLevel, slogLevel := range data {
		if logLevel != slogLevel {
			t.Error(logLevel, "!=", slogLevel)
			t.FailNow()
		}
	}
}

func TestLevelLabel(t *testing.T) {
	data := map[slog.Level]string{
		slog.LevelDebug: "DEBUG",
		slog.LevelInfo:  "INFO",
		slog.LevelWarn:  "WARN",
		slog.LevelError: "ERROR",
		LevelFatal:      "FATAL",
		LevelDisaster:   "DISASTER",
	}
	for level, label := range data {
		a := slog.Any("level", level)
		levelLabel, ok := LevelLabel(a)
		if !ok || label != levelLabel {
			t.Error(label, "!=", levelLabel)
			t.FailNow()
		}
	}
}
