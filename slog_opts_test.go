package dlog

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

type testWriter struct {
	levelLabel string
}

func (w *testWriter) Write(bs []byte) (n int, err error) {
	m := map[string]any{}
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return
	}
	w.levelLabel = m["level"].(string)
	return
}

func TestLevelReplacer(t *testing.T) {
	data := map[slog.Level]string{
		slog.LevelDebug: "DEBUG",
		slog.LevelInfo:  "INFO",
		slog.LevelWarn:  "WARN",
		slog.LevelError: "ERROR",
	}
	for k, v := range extraLevelLabels {
		sv, exist := data[k]
		if exist {
			t.Error("extra level", k, "equal to slog level", sv)
			t.FailNow()
		}
		data[k] = v
	}
	writer := new(testWriter)
	opts := &slog.HandlerOptions{
		Level:       LevelDebug,
		ReplaceAttr: LevelReplacer,
	}
	logger := slog.New(slog.NewJSONHandler(writer, opts))
	for level, label := range data {
		logger.Log(context.Background(), level, "test")
		if writer.levelLabel != label {
			t.Error("level", level, "label cannot be replaced by", label, "writerLabel", writer.levelLabel)
			t.FailNow()
		}
	}
}
