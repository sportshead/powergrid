package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"log/slog"
	"net/http"
	"os"
)

// DISCORD_PUBLIC_KEY
var DiscordPublicKey ed25519.PublicKey

// DISCORD_APPLICATION_ID
var DiscordApplicationId string

// DISCORD_BOT_TOKEN
var DiscordBotToken string

// DISCORD_OAUTH_SECRET
var DiscordOauthSecret string

// replaced at build time
var BuildCommitHash = "dev"
var serverHeader = "coordinator/" + BuildCommitHash

const JSONMimeType = "application/json"

func main() {
	if BuildCommitHash == "dev" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}

	discordPublicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	if discordPublicKey == "" {
		slog.Error("missing env variable", slogTag("invalid_env"), slog.String("key", "DISCORD_PUBLIC_KEY"))
		os.Exit(1)
	}
	var err error
	DiscordPublicKey, err = hex.DecodeString(discordPublicKey)
	if err != nil {
		slog.Error("failed to parse hex", slogTag("invalid_env"), slogError(err), slog.String("key", "DISCORD_PUBLIC_KEY"))
		os.Exit(1)
	}

	DiscordApplicationId = os.Getenv("DISCORD_APPLICATION_ID")
	if DiscordApplicationId == "" {
		slog.Error("missing env variable", slogTag("invalid_env"), slog.String("key", "DISCORD_APPLICATION_ID"))
		os.Exit(1)
	}

	DiscordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	if DiscordBotToken == "" {
		slog.Error("missing env variable", slogTag("invalid_env"), slog.String("key", "DISCORD_BOT_TOKEN"))
		os.Exit(1)
	}

	DiscordOauthSecret = os.Getenv("DISCORD_OAUTH_SECRET")
	if DiscordOauthSecret == "" {
		slog.Error("missing env variable", slogTag("invalid_env"), slog.String("key", "DISCORD_OAUTH_SECRET"))
		os.Exit(1)
	}

	initKubernetes()

	server := http.NewServeMux()
	server.HandleFunc("/", handleHTTP)
	server.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\nrunning " + BuildCommitHash))
	})

	slog.Info("public http server listening", slogTag("http_listen"), slog.String("addr", "0.0.0.0:8000"))
	err = http.ListenAndServe("0.0.0.0:8000", server)
	slog.Error("public http server died", slogTag("http_died"), slogError(err))
	os.Exit(1)
}

const (
	MissingHandlerMessage = "**Error**: Unknown command"
	MissingServiceMessage = "**Error**: Failed to get service address"
	ForwardFailedMessage  = "**Error**: Failed to forward request"
	UpstreamErrorMessage  = "**Error**: Upstream returned error `%d %s`"
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
			slog.Error("http handler panicked", slogTag("http_panic"), slog.Any("error", err))
		}
	}()
	w.Header().Set("Server", serverHeader)
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !discordgo.VerifyInteraction(r, DiscordPublicKey) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", slogTag("failed_read_body"), slogError(err))
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	var interaction = &discordgo.Interaction{}
	err = interaction.UnmarshalJSON(body)
	if err != nil {
		slog.Error("failed to unmarshal json", slogTag("failed_unmarshal_json"), slogError(err), slog.String("body", string(body)))
		http.Error(w, "failed to unmarshal json", http.StatusInternalServerError)
		return
	}

	switch interaction.Type {
	case discordgo.InteractionPing:
		if writeJSONString(w, InteractionResponsePongJSON) {
			slog.Info("responding to ping", slogTag("pong"), slog.String("ip", getIP(r)))
		}

	case discordgo.InteractionApplicationCommand:
		data := interaction.Data.(discordgo.ApplicationCommandInteractionData)
		slog.Info("application command executed", slogTag("command_executed"), slog.String("command", data.Name), slog.String("user", interaction.Member.User.ID))
		if cmd, ok := CommandMap[data.Name]; ok {
			addr := getServiceAddr(r.Context(), cmd.Object["spec"].(map[string]interface{})["serviceName"].(string))
			if addr == "" {
				slog.Error("failed to get service address", slogTag("failed_get_service_address"), slog.String("command", data.Name), slog.String("user", interaction.Member.User.ID), slog.String("body", string(body)))

				writeMessage(w, MissingServiceMessage)
				return
			}

			req := r.Clone(context.Background())
			req.URL.Scheme = "http"
			req.URL.Host = addr
			req.URL.Path = "/"
			req.Body = io.NopCloser(bytes.NewReader(body))
			req.RequestURI = ""

			var shouldDefer bool
			shouldDefer, ok = cmd.Object["spec"].(map[string]interface{})["shouldSendDeferred"].(bool)
			if !ok {
				shouldDefer = false
			}
			if shouldDefer {
				writeJSONString(w, InteractionResponseDeferredChannelMessageWithSourceJSON)
				go handleApplicationCommand(w, req, shouldDefer, data, interaction, addr, body)
				return
			}
			handleApplicationCommand(w, req, shouldDefer, data, interaction, addr, body)
		} else {
			slog.Error("missing handler for command", slogTag("unknown_command"), slog.String("command", data.Name), slog.String("user", interaction.Member.User.ID), slog.String("body", string(body)))
			writeMessage(w, MissingHandlerMessage)
			return
		}
	}
}

func handleApplicationCommand(w http.ResponseWriter, req *http.Request, shouldDefer bool, data discordgo.ApplicationCommandInteractionData, interaction *discordgo.Interaction, addr string, body []byte) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to forward request", slogTag("failed_forward_request"), slogError(err), slog.String("body", string(body)))
		if !shouldDefer {
			writeMessage(w, ForwardFailedMessage)
		}
		return
	}
	if res.StatusCode != http.StatusOK {
		slog.Error("upstream returned error",
			slogTag("upstream_error"),
			slog.Int("status", res.StatusCode),
			slog.String("status_text", res.Status),
			slog.String("command", data.Name),
			slog.String("user", interaction.Member.User.ID),
			slog.String("addr", addr),
			slog.String("body", string(body)),
			slog.String("response", fmt.Sprint(io.ReadAll(req.Body))))
		if !shouldDefer {
			writeMessage(w, fmt.Sprintf(UpstreamErrorMessage, res.StatusCode, res.Status))
		}
		return
	}

	if !shouldDefer {
		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, res.Body)

		if err != nil {
			slog.Error("failed to copy response body", slogTag("failed_write_body"), slogError(err), slog.String("body", string(body)))
			return
		}
	}

	slog.Info("handled command", slogTag("command_executed"), slog.String("command", data.Name), slog.String("user", interaction.Member.User.ID), slog.String("addr", addr), slog.Int("status", res.StatusCode), slog.Bool("deferred", shouldDefer))
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
		slog.Error("failed to write JSON", slogTag("failed_write_json"), slogError(err), slog.String("data", data))
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
		slog.Error("failed to write message", slogTag("failed_write_message"), slogError(err), slog.String("message", message))
		return
	}
}
