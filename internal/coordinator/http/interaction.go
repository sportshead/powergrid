package http

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sportshead/powergrid/internal/coordinator/discord"
	"github.com/sportshead/powergrid/pkg/utils"
	"github.com/sportshead/powergrid/pkg/version"
	"io"
	"log/slog"
	"net/http"
)

const (
	// InteractionResponsePongJSON is the JSON representation of a discordgo.InteractionResponsePong
	InteractionResponsePongJSON = `{"type":1}`
	// InteractionResponseDeferredChannelMessageWithSourceJSON is the JSON representation of a discordgo.InteractionResponseDeferredChannelMessageWithSource
	InteractionResponseDeferredChannelMessageWithSourceJSON = `{"type":5}`
)

func forwardInteraction(log *slog.Logger, w http.ResponseWriter, req *http.Request, shouldDefer bool, interaction *discordgo.Interaction) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("failed to forward request", utils.Tag("failed_forward_request"), utils.Error(err), slog.String("interaction", utils.TryMarshal(interaction)))
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
			slog.String("interaction", utils.TryMarshal(interaction)),
			slog.String("response", fmt.Sprint(io.ReadAll(req.Body))),
		)
		if !shouldDefer {
			writeMessage(w, fmt.Sprintf(UpstreamErrorMessage, res.StatusCode, res.Status))
		} else {
			_, err = discord.Session.FollowupMessageCreate(interaction, false, &discordgo.WebhookParams{
				Content: fmt.Sprintf(UpstreamErrorMessage, res.StatusCode, res.Status),
			})
			if err != nil {
				log.Error("failed to send followup message", utils.Tag("failed_send_followup"), utils.Error(err))
			}
		}
		return
	}

	if !shouldDefer {
		contentType := res.Header.Get("Content-Type")
		if contentType == "" {
			contentType = utils.MimeTypeJSON
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		if version.Debug {
			r := io.TeeReader(res.Body, w)
			resp, _ := io.ReadAll(r)
			log.Debug("response", slog.String("ctype", contentType), slog.String("response", string(resp)))
		} else {
			_, err = io.Copy(w, res.Body)
		}

		if err != nil {
			log.Error("failed to copy response body", utils.Tag("failed_write_body"), utils.Error(err), slog.String("interaction", utils.TryMarshal(interaction)))
			return
		}
	}

	log.Info("handled interaction", utils.Tag("interaction_handled"))
	return
}
