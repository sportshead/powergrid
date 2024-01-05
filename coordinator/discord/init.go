package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sportshead/powergrid/coordinator/env"
	"github.com/sportshead/powergrid/coordinator/utils"
	"log/slog"
	"os"
)

// session is a discordgo session for use with the Discord REST API.
// !! DO NOT CALL .Open() !!
// Gateway connection is handled in powergrid/gateway, not in powergrid/coordinator
var session *discordgo.Session

func Init() {
	var err error

	token := env.DiscordBotToken
	if token[:4] != "Bot " {
		token = "Bot " + token
	}

	session, err = discordgo.New(token)
	if err != nil {
		slog.Error("failed to create discord session", utils.Tag("discord_session_failed"), utils.Error(err))
		os.Exit(1)
	}
}
