package dlog

import "log/slog"

const (
	LevelDebug               = slog.LevelDebug
	LevelInfo                = slog.LevelInfo
	LevelWarn                = slog.LevelWarn
	LevelError               = slog.LevelError
	LevelFatal    slog.Level = 12
	LevelDisaster slog.Level = 16
)

var extraLevelLabels = map[slog.Level]string{
	LevelFatal:    "FATAL",
	LevelDisaster: "DISASTER",
}

func LevelLabel(a slog.Attr) (label string, ok bool) {
	if a.Key != slog.LevelKey {
		return "", false
	}
	level, ok := a.Value.Any().(slog.Level)
	if !ok {
		return "", false
	}
	levelLabel, exist := extraLevelLabels[level]
	if !exist {
		return level.String(), true
	}
	return levelLabel, true
}
