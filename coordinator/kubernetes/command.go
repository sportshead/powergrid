package kubernetes

import (
	"context"
	"encoding/json"
	powergridv10 "github.com/sportshead/powergrid/coordinator/pkg/apis/powergrid.sportshead.dev/v10"
	"github.com/sportshead/powergrid/coordinator/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"os"
	"time"
)

var CommandMap = make(map[string]powergridv10.Command)

type commandObject struct {
	Name string `json:"name"`
}

func loadCommands() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	commands, err := powergridClient.PowergridV10().Commands(namespace).List(ctx, metav1.ListOptions{})
	cancel()
	if err != nil {
		slog.Error("failed to get commands", utils.Tag("k8s_command_list_failed"), utils.Error(err))
		os.Exit(1)
	}

	for _, command := range commands.Items {
		cmd := &commandObject{}
		err = json.Unmarshal(command.Spec.Command.Raw, cmd)
		if err != nil {
			slog.Error("failed to parse command object", utils.Tag("k8s_command_parse_failed"), utils.Error(err), slog.String("object", utils.TryMarshal(command)))
			continue
		}

		CommandMap[cmd.Name] = command

		slog.Info("got command", utils.Tag("k8s_command_loaded"), slog.String("name", command.GetName()), slog.String("command", cmd.Name), slog.String("object", utils.TryMarshal(command)))
	}
	slog.Debug("got all commands", slog.String("commandMap", utils.TryMarshal(CommandMap)))
}
