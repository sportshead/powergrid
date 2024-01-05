package discord

import (
	"context"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/sportshead/powergrid/coordinator/env"
	powergridv10 "github.com/sportshead/powergrid/coordinator/pkg/apis/powergrid.sportshead.dev/v10"
	"github.com/sportshead/powergrid/coordinator/utils"
	"log/slog"
	"reflect"
	"slices"
)

type commandObject struct {
	Name string `json:"name"`
}

// https://stackoverflow.com/a/37335777
func removeFromList(s []interface{}, i int) []interface{} {
	s[len(s)-1], s[i] = 0, s[len(s)-1]
	return s[:len(s)-1]
}

func UpdateCommands(ctx context.Context, list []interface{}) {
	commands, err := session.ApplicationCommands(env.DiscordApplicationID, env.DiscordGuildID, discordgo.WithContext(ctx))
	if err != nil {
		slog.Error("failed to get commands", utils.Tag("discord_commands_failed"), utils.Error(err))
		return
	}

	for _, oldCommand := range commands {
		log := slog.With(slog.String("command", oldCommand.Name), slog.String("id", oldCommand.ID), slog.String("version", oldCommand.Version))
		i := slices.IndexFunc(list, func(i interface{}) bool {
			powergridCommand := i.(*powergridv10.Command)
			cmd := &commandObject{}
			err = json.Unmarshal(powergridCommand.Spec.Command.Raw, cmd)
			if err != nil {
				log.Error("failed to parse command object", utils.Tag("k8s_command_parse_failed"), utils.Error(err), slog.String("object", utils.TryMarshal(powergridCommand)))
				return false
			}
			return oldCommand.Name == cmd.Name
		})

		if i == -1 {
			err = session.ApplicationCommandDelete(oldCommand.ApplicationID, oldCommand.GuildID, oldCommand.ID, discordgo.WithContext(ctx))
			if err != nil {
				log.Error("failed to delete command", utils.Tag("discord_command_delete_failed"), utils.Error(err))
				continue
			}
			log.Info("deleted command", utils.Tag("discord_command_delete"))
			continue
		}

		powergridCommand := list[i].(*powergridv10.Command)
		newCommand := &discordgo.ApplicationCommand{
			ID:            oldCommand.ID,
			ApplicationID: oldCommand.ApplicationID,
			GuildID:       oldCommand.GuildID,
			Version:       oldCommand.Version,
		}
		err = json.Unmarshal(powergridCommand.Spec.Command.Raw, newCommand)
		if err != nil {
			log.Error("failed to parse command object", utils.Tag("k8s_command_parse_failed"), utils.Error(err), slog.String("object", utils.TryMarshal(powergridCommand)))
			continue
		}

		// set defaults
		if newCommand.Type == 0 {
			newCommand.Type = discordgo.ChatApplicationCommand
		}
		if newCommand.NSFW == nil {
			newCommand.NSFW = utils.Ptr(false)
		}

		if oldCommand.Type != newCommand.Type || !reflect.DeepEqual(oldCommand, newCommand) {
			var edited *discordgo.ApplicationCommand
			edited, err = session.ApplicationCommandEdit(oldCommand.ApplicationID, oldCommand.GuildID, oldCommand.ID, newCommand, discordgo.WithContext(ctx))
			if err != nil {
				log.Error("failed to edit command", utils.Tag("discord_command_edit_failed"), utils.Error(err))
				continue
			}
			log.Info("updated command", utils.Tag("discord_command_updated"), slog.String("new_version", edited.Version))
		} else {
			log.Debug("command unchanged", utils.Tag("discord_command_unchanged"))
		}
		list = removeFromList(list, i)
	}

	for _, i := range list {
		powergridCommand := i.(*powergridv10.Command)
		log := slog.With(slog.String("name", powergridCommand.Name))
		newCommand := &discordgo.ApplicationCommand{}
		err = json.Unmarshal(powergridCommand.Spec.Command.Raw, newCommand)
		if err != nil {
			log.Error("failed to parse command object", utils.Tag("k8s_command_parse_failed"), utils.Error(err), slog.String("object", utils.TryMarshal(powergridCommand)))
			continue
		}
		var created *discordgo.ApplicationCommand
		created, err = session.ApplicationCommandCreate(env.DiscordApplicationID, env.DiscordGuildID, newCommand, discordgo.WithContext(ctx))
		if err != nil {
			log.Error("failed to create command", utils.Tag("discord_command_create_failed"), utils.Error(err))
			continue
		}
		log.Info("created command", utils.Tag("discord_command_created"), slog.String("command", created.Name), slog.String("id", created.ID))
	}
}
