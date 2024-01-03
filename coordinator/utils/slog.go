package utils

import (
	"log/slog"
)

func Tag(tag string) slog.Attr {
	return slog.String("tag", tag)
}

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}
