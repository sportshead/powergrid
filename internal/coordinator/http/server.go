package http

import (
	"errors"
	"github.com/sportshead/powergrid/pkg/utils"
	"github.com/sportshead/powergrid/pkg/version"
	"log/slog"
	"net/http"
	"sync"
)

func Init(stop chan struct{}, cleanupGroup *sync.WaitGroup) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", HandleHTTP)
	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", utils.MimeTypeText)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\nrunning " + version.String))
	})

	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: version.Middleware("coordinator", serveMux),
	}

	cleanupGroup.Add(1)
	go func() {
		defer cleanupGroup.Done()
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server died", utils.Tag("http_died"), utils.Error(err))
			close(stop)
		}
	}()

	slog.Info("http server listening", utils.Tag("http_listen"), slog.String("addr", server.Addr))
}
