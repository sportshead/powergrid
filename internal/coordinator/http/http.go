package http

import (
	"bytes"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/sportshead/powergrid/internal/coordinator/env"
	"github.com/sportshead/powergrid/internal/coordinator/kubernetes"
	powergridv10 "github.com/sportshead/powergrid/pkg/apis/powergrid.sportshead.dev/v10"
	"github.com/sportshead/powergrid/pkg/utils"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func HandleHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			slog.Error("http handler panicked", utils.Tag("http_panic"), slog.Any("error", err))
		}
	}()

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
		if utils.WriteJSONString(w, InteractionResponsePongJSON) {
			log.Info("responding to ping", utils.Tag("pong"), slog.String("ip", utils.GetIP(r)))
		}

	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
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

		req := makeRequest(r, addr, body)

		var shouldDefer bool
		shouldDefer = cmd.Spec.ShouldSendDeferred && interaction.Type != discordgo.InteractionApplicationCommandAutocomplete
		log = log.With(slog.Bool("deferred", shouldDefer))
		if shouldDefer {
			utils.WriteJSONString(w, InteractionResponseDeferredChannelMessageWithSourceJSON)
			go forwardInteraction(log, w, req, shouldDefer, interaction)
			return
		}
		forwardInteraction(log, w, req, shouldDefer, interaction)

	case discordgo.InteractionMessageComponent:
		data := interaction.Data.(discordgo.MessageComponentInteractionData)
		log = log.With(
			slog.String("component", data.CustomID),
			slog.String("user", interaction.Member.User.ID),
			slog.String("guild", interaction.GuildID),
			slog.String("channel", interaction.ChannelID),
		)
		log.Info("message component interaction received", utils.Tag("component_received"))

		handleMessageOrModal(log, w, r, body, interaction, data.CustomID)

	case discordgo.InteractionModalSubmit:
		data := interaction.Data.(discordgo.ModalSubmitInteractionData)
		log = log.With(
			slog.String("component", data.CustomID),
			slog.String("user", interaction.Member.User.ID),
			slog.String("guild", interaction.GuildID),
			slog.String("channel", interaction.ChannelID),
		)
		log.Info("modal submit interaction received", utils.Tag("modal_received"))

		handleMessageOrModal(log, w, r, body, interaction, data.CustomID)
	}
}

func handleMessageOrModal(log *slog.Logger, w http.ResponseWriter, r *http.Request, body []byte, interaction *discordgo.Interaction, id string) {
	service := strings.Split(id, "/")[0]
	log = log.With(slog.String("service", service))

	addr := kubernetes.GetServiceAddr(log, service)
	if addr == "" {
		log.Error("failed to get service address", utils.Tag("failed_get_service_address"))

		writeMessage(w, MissingServiceMessage)
		return
	}

	log = log.With(slog.String("addr", addr))

	req := makeRequest(r, addr, body)
	forwardInteraction(log, w, req, false, interaction)
}

func makeRequest(r *http.Request, addr string, body []byte) *http.Request {
	req := r.Clone(context.Background())
	req.URL.Scheme = "http"
	req.URL.Host = addr
	req.URL.Path = "/"
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.RequestURI = ""
	return req
}
