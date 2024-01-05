package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sportshead/powergrid/coordinator/discord"
	"github.com/sportshead/powergrid/coordinator/env"
	"github.com/sportshead/powergrid/coordinator/kubernetes"
	powergridv10 "github.com/sportshead/powergrid/coordinator/pkg/apis/powergrid.sportshead.dev/v10"
	"github.com/sportshead/powergrid/coordinator/utils"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// BuildCommitHash is the latest git commit, and is replaced at build time by CI.
var BuildCommitHash = "dev"
var serverHeader = "coordinator/" + BuildCommitHash

const JSONMimeType = "application/json"

var stop = make(chan struct{})
var cleanupGroup = &sync.WaitGroup{}

func init() {
	if BuildCommitHash == "dev" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
}

func main() {
	slog.Info("starting coordinator", utils.Tag("start"), slog.String("hash", BuildCommitHash))

	discord.Init()
	kubernetes.Init(stop, cleanupGroup)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", handleHTTP)
	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\nrunning " + BuildCommitHash))
	})

	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: serveMux,
	}

	cleanupGroup.Add(1)
	go func() {
		defer cleanupGroup.Done()
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server died", utils.Tag("http_died"), utils.Error(err))
			os.Exit(1)
		}
	}()
	slog.Info("http server listening", utils.Tag("http_listen"), slog.String("addr", server.Addr))

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<-ch
	slog.Info("gracefully stopping coordinator", utils.Tag("stopping"))
	close(stop)

	if !waitTimeout(cleanupGroup, 10*time.Second) {
		slog.Error("cleanup timed out", utils.Tag("cleanup_timeout"))
		os.Exit(1)
	} else {
		slog.Info("stopped coordinator", utils.Tag("stopped"))
	}
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
// https://stackoverflow.com/a/32843750
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

const (
	MissingHandlerMessage = "**Error**: Unknown command"
	MissingServiceMessage = "**Error**: Failed to get service address"
	ForwardFailedMessage  = "**Error**: Failed to forward request"
	UpstreamErrorMessage  = "**Error**: Upstream server returned error `%d`: `%s`"
)

const (
	// InteractionResponsePongJSON is the JSON representation of a discordgo.InteractionResponsePong
	InteractionResponsePongJSON = `{"type":1}`
	// InteractionResponseDeferredChannelMessageWithSourceJSON is the JSON representation of a discordgo.InteractionResponseDeferredChannelMessageWithSource
	InteractionResponseDeferredChannelMessageWithSourceJSON = `{"type":5}`
)

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			slog.Error("http handler panicked", utils.Tag("http_panic"), slog.Any("error", err))
		}
	}()
	w.Header().Set("Server", serverHeader)
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !discordgo.VerifyInteraction(r, env.DiscordPublicKey) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", utils.Tag("failed_read_body"), utils.Error(err))
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	var interaction = &discordgo.Interaction{}
	err = interaction.UnmarshalJSON(body)
	if err != nil {
		slog.Error("failed to unmarshal json", utils.Tag("failed_unmarshal_json"), utils.Error(err), slog.String("body", string(body)))
		http.Error(w, "failed to unmarshal json", http.StatusInternalServerError)
		return
	}
	log := slog.With(slog.String("id", interaction.ID))

	switch interaction.Type {
	case discordgo.InteractionPing:
		if writeJSONString(w, InteractionResponsePongJSON) {
			log.Info("responding to ping", utils.Tag("pong"), slog.String("ip", getIP(r)))
		}

	case discordgo.InteractionApplicationCommand:
		data := interaction.Data.(discordgo.ApplicationCommandInteractionData)
		log = log.With(
			slog.String("command", data.Name),
			slog.String("user", interaction.Member.User.ID),
			slog.String("guild", interaction.GuildID),
			slog.String("channel", interaction.ChannelID),
		)
		log.Info("application command interaction received", utils.Tag("command_received"))
		var cmd *powergridv10.Command
		cmd, err = kubernetes.GetCommand(data.Name)
		if err != nil {
			log.Error("failed to get handler for command", utils.Tag("unknown_command"), utils.Error(err), slog.String("body", string(body)))
			writeMessage(w, MissingHandlerMessage)
			return
		}

		addr := kubernetes.GetServiceAddr(log, cmd.Spec.ServiceName)
		if addr == "" {
			log.Error("failed to get service address", utils.Tag("failed_get_service_address"))

			writeMessage(w, MissingServiceMessage)
			return
		}

		log = log.With(slog.String("addr", addr))

		req := r.Clone(context.Background())
		req.URL.Scheme = "http"
		req.URL.Host = addr
		req.URL.Path = "/"
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.RequestURI = ""

		var shouldDefer bool
		shouldDefer = cmd.Spec.ShouldSendDeferred
		if shouldDefer {
			writeJSONString(w, InteractionResponseDeferredChannelMessageWithSourceJSON)
			go forwardApplicationCommand(log, w, req, shouldDefer, addr, body)
			return
		}
		forwardApplicationCommand(log, w, req, shouldDefer, addr, body)
	}
}

func forwardApplicationCommand(log *slog.Logger, w http.ResponseWriter, req *http.Request, shouldDefer bool, addr string, body []byte) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("failed to forward request", utils.Tag("failed_forward_request"), utils.Error(err), slog.String("body", string(body)))
		if !shouldDefer {
			writeMessage(w, ForwardFailedMessage)
		}
		return
	}
	log = log.With(
		slog.Int("status", res.StatusCode),
		slog.String("status_text", res.Status),
	)
	if res.StatusCode != http.StatusOK {
		log.Error("upstream returned error",
			utils.Tag("upstream_error"),
			slog.String("body", string(body)),
			slog.String("response", fmt.Sprint(io.ReadAll(req.Body))),
		)
		if !shouldDefer {
			writeMessage(w, fmt.Sprintf(UpstreamErrorMessage, res.StatusCode, res.Status))
		} else {
			// TODO: update deferred with error message
		}
		return
	}

	if !shouldDefer {
		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, res.Body)

		if err != nil {
			log.Error("failed to copy response body", utils.Tag("failed_write_body"), utils.Error(err), slog.String("body", string(body)))
			return
		}
	}

	log.Info("handled command", utils.Tag("command_executed"), slog.Bool("deferred", shouldDefer))
	return
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("CF-Connecting-IP")
	}
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

func writeJSONString(w http.ResponseWriter, data string) bool {
	w.Header().Set("Content-Type", JSONMimeType)
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte(data))
	if err != nil {
		slog.Error("failed to write JSON", utils.Tag("failed_write_json"), utils.Error(err), slog.String("data", data))
		return false
	}
	return true
}

func writeMessage(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", JSONMimeType)
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)

	res := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		},
	}

	err := encoder.Encode(res)
	if err != nil {
		slog.Error("failed to write message", utils.Tag("failed_write_message"), utils.Error(err), slog.String("message", message))
		return
	}
}
