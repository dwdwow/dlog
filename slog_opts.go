package log

import "log/slog"

func LevelReplacer(groups []string, a slog.Attr) slog.Attr {
	label, ok := LevelLabel(a)
	if !ok {
		return a
	}
	a.Value = slog.StringValue(label)
	return a
}
