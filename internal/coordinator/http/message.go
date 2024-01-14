package http

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/sportshead/powergrid/pkg/utils"
	"log/slog"
	"net/http"
)

const (
	MissingHandlerMessage = "**Error**: Unknown command"
	MissingServiceMessage = "**Error**: Failed to get service address"
	ForwardFailedMessage  = "**Error**: Failed to forward request"
	UpstreamErrorMessage  = "**Error**: Upstream server returned error `%d`: `%s`"
)

func writeMessage(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", utils.MimeTypeJSON)
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
