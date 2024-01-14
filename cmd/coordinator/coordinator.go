package main

import (
	"github.com/sportshead/powergrid/internal/coordinator/discord"
	"github.com/sportshead/powergrid/internal/coordinator/http"
	"github.com/sportshead/powergrid/internal/coordinator/kubernetes"
	"github.com/sportshead/powergrid/pkg/utils"
	"github.com/sportshead/powergrid/pkg/version"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var stop = make(chan struct{})
var cleanupGroup = &sync.WaitGroup{}

func main() {
	slog.Info("starting coordinator", utils.Tag("start"), slog.String("version", version.String))

	discord.Init()
	kubernetes.Init(stop, cleanupGroup)

	http.Init(stop, cleanupGroup)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<-ch
	slog.Info("gracefully stopping coordinator", utils.Tag("stopping"))
	close(stop)

	if !utils.WaitTimeout(cleanupGroup, 10*time.Second) {
		slog.Error("cleanup timed out", utils.Tag("cleanup_timeout"))
		os.Exit(1)
	} else {
		slog.Info("stopped coordinator", utils.Tag("stopped"))
		os.Exit(0)
	}
}
