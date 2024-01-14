package utils

import (
	"encoding/json"
	"fmt"
	"github.com/sportshead/powergrid/pkg/version"
	"log/slog"
	"os"
)

func init() {
	if version.Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
}

func Tag(tag string) slog.Attr {
	return slog.String("tag", tag)
}

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

func TryMarshal(obj any) string {
	if obj == nil {
		return "<nil>"
	}

	bytes, err := json.Marshal(obj)
	if err == nil {
		return string(bytes)
	}

	return fmt.Sprintf("%#v", obj)
}
