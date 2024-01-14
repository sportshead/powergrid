package env

import (
	"crypto/ed25519"
	"encoding/hex"
	"github.com/sportshead/powergrid/pkg/utils"
	"log/slog"
	"os"
)

// DISCORD_PUBLIC_KEY
var DiscordPublicKey ed25519.PublicKey

// DISCORD_APPLICATION_ID
var DiscordApplicationID string

// DISCORD_BOT_TOKEN
var DiscordBotToken string

// DISCORD_OAUTH_SECRET
var DiscordOAuthSecret string

// DISCORD_GUID_ID
var DiscordGuildID string

// DeploymentName is the name of the current deployment, used as the name of the leader election lease.
// Passed in as the DEPLOYMENT_NAME env var.
var DeploymentName string

// Hostname is the name of the current pod, used for identification in leader election.
// Passed in as the HOSTNAME env var.
var Hostname string

func init() {
	discordPublicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	if discordPublicKey == "" {
		slog.Error("missing env variable", utils.Tag("invalid_env"), slog.String("key", "DISCORD_PUBLIC_KEY"))
		os.Exit(1)
	}
	var err error
	DiscordPublicKey, err = hex.DecodeString(discordPublicKey)
	if err != nil {
		slog.Error("failed to parse hex", utils.Tag("invalid_env"), utils.Error(err), slog.String("key", "DISCORD_PUBLIC_KEY"), slog.String("hex", discordPublicKey))
		os.Exit(1)
	}
	if len(DiscordPublicKey) != ed25519.PublicKeySize {
		slog.Error("invalid public key length", utils.Tag("invalid_env"), slog.String("key", "DISCORD_PUBLIC_KEY"), slog.String("hex", discordPublicKey), slog.Int("len", len(DiscordPublicKey)))
	}

	DiscordApplicationID = os.Getenv("DISCORD_APPLICATION_ID")
	if DiscordApplicationID == "" {
		slog.Error("missing env variable", utils.Tag("invalid_env"), slog.String("key", "DISCORD_APPLICATION_ID"))
		os.Exit(1)
	}

	DiscordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	if DiscordBotToken == "" {
		slog.Error("missing env variable", utils.Tag("invalid_env"), slog.String("key", "DISCORD_BOT_TOKEN"))
		os.Exit(1)
	}

	DiscordOAuthSecret = os.Getenv("DISCORD_OAUTH_SECRET")
	if DiscordOAuthSecret == "" {
		slog.Error("missing env variable", utils.Tag("invalid_env"), slog.String("key", "DISCORD_OAUTH_SECRET"))
		os.Exit(1)
	}

	// optional
	DiscordGuildID = os.Getenv("DISCORD_GUILD_ID")

	DeploymentName = os.Getenv("DEPLOYMENT_NAME")
	if DeploymentName == "" {
		slog.Error("missing env variable", utils.Tag("invalid_env"), slog.String("key", "DEPLOYMENT_NAME"))
		os.Exit(1)
	}

	Hostname = os.Getenv("HOSTNAME")
	if Hostname == "" {
		slog.Error("missing env variable", utils.Tag("invalid_env"), slog.String("key", "HOSTNAME"))
		os.Exit(1)
	}
}
