package kubernetes

import (
	"encoding/json"
	"fmt"
	"github.com/sportshead/powergrid/coordinator/discord"
	powergridv10 "github.com/sportshead/powergrid/coordinator/pkg/apis/powergrid.sportshead.dev/v10"
	informers "github.com/sportshead/powergrid/coordinator/pkg/generated/informers/externalversions"
	"github.com/sportshead/powergrid/coordinator/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"log/slog"
	"os"
	"time"
)

const ByName = "DiscordCommandNameIndexer"

var commandInformer cache.SharedIndexInformer

type commandObject struct {
	Name string `json:"name"`
}

func updateCommands() {
	list := commandInformer.GetStore().List()

	discord.UpdateCommands(list)
}

func loadCommands() {
	factory := informers.NewSharedInformerFactoryWithOptions(powergridClient, 10*time.Minute, informers.WithNamespace(namespace))
	commandInformer = factory.Powergrid().V10().Commands().Informer()
	err := commandInformer.AddIndexers(map[string]cache.IndexFunc{
		ByName: func(obj interface{}) ([]string, error) {
			index := make([]string, 1)
			command := obj.(*powergridv10.Command)
			cmd := &commandObject{}
			err := json.Unmarshal(command.Spec.Command.Raw, cmd)
			if err != nil {
				slog.Error("failed to parse command object", utils.Tag("k8s_command_parse_failed"), utils.Error(err), slog.String("object", utils.TryMarshal(command)))
				return nil, err
			}
			index[0] = cmd.Name

			slog.Info("indexing command", utils.Tag("k8s_index_command"), slog.String("name", command.Name), slog.String("command", cmd.Name))
			return index, nil
		},
	})
	if err != nil {
		slog.Error("failed to add indexer", utils.Tag("k8s_indexer_failed"), utils.Error(err))
		os.Exit(1)
	}

	stopCh := make(chan struct{})
	factory.Start(stopCh)            // start goroutines
	factory.WaitForCacheSync(stopCh) // wait for init

	wait.Until(updateCommands, 1*time.Minute, wait.NeverStop)
}

func GetCommand(name string) (*powergridv10.Command, error) {
	commands, err := commandInformer.GetIndexer().ByIndex(ByName, name)
	if err != nil {
		return nil, err
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("no command matches the name %s", name)
	}
	if len(commands) > 1 {
		return nil, fmt.Errorf("%d commands match the name %s", len(commands), name)
	}

	return commands[0].(*powergridv10.Command), nil
}
