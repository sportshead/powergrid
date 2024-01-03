package main

import (
	"context"
	"encoding/json"
	powergridv10 "github.com/sportshead/powergrid/coordinator/pkg/apis/powergrid.sportshead.dev/v10"
	clientset "github.com/sportshead/powergrid/coordinator/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var config *rest.Config

var powergridClient *clientset.Clientset
var kubernetesClient *kubernetes.Clientset
var namespace = corev1.NamespaceDefault

var CommandMap = make(map[string]powergridv10.Command)

func initKubernetes() {
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		slog.Error("failed to get in cluster config", slogTag("k8s_config_create_failed"), slogError(err))
		os.Exit(1)
	}
	powergridClient, err = clientset.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create k8s client", slogTag("k8s_client_create_failed"), slogError(err), slog.String("client", "powergrid"))
		os.Exit(1)
	}
	kubernetesClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create k8s client", slogTag("k8s_client_create_failed"), slogError(err), slog.String("client", "kubernetes"))
		os.Exit(1)
	}

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		slog.Warn("failed to read namespace file", slogTag("k8s_namespace_read_failed"), slogError(err))
	} else {
		namespace = strings.TrimSpace(string(data))
	}

	slog.Info("initiated kubernetes client",
		slogTag("k8s_client_created"),
		slog.String("namespace", namespace),
		slog.String("host", config.Host))

	go loadCommands()
}

type commandObject struct {
	Name string `json:"name"`
}

func loadCommands() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	commands, err := powergridClient.PowergridV10().Commands(namespace).List(ctx, metav1.ListOptions{})
	cancel()
	if err != nil {
		slog.Error("failed to get commands", slogTag("k8s_command_list_failed"), slogError(err))
		os.Exit(1)
	}

	for _, command := range commands.Items {
		cmd := &commandObject{}
		err = json.Unmarshal(command.Spec.Command.Raw, cmd)
		if err != nil {
			slog.Error("failed to parse command object", slogTag("k8s_command_parse_failed"), slogError(err), slog.String("object", tryMarshal(command)))
			continue
		}

		CommandMap[cmd.Name] = command

		slog.Info("got command", slogTag("k8s_command_loaded"), slog.String("name", command.GetName()), slog.String("command", cmd.Name), slog.String("object", tryMarshal(command)))
	}
	slog.Debug("got all commands", slog.String("commandMap", tryMarshal(CommandMap)))
}

func getServiceAddr(ctx context.Context, serviceName string) string {
	service, err := kubernetesClient.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		slog.Error("failed to get service", slogTag("k8s_service_get_failed"), slogError(err))
		return ""
	}

	ip := service.Spec.ClusterIP
	if ip == "" || ip == "None" {
		slog.Error("service has no clusterIP", slogTag("k8s_service_missing_cluster_ip"), slog.String("name", serviceName), slog.String("object", tryMarshal(service)))
		return ""
	}

	ports := service.Spec.Ports
	port := ""
	for _, p := range ports {
		if port == "" || p.Name == "http" {
			port = strconv.Itoa(int(p.Port))
			break
		}
	}
	if port == "" {
		slog.Error("service has no http port", slogTag("k8s_service_missing_http_port"), slog.String("name", serviceName), slog.String("object", tryMarshal(service)))
		return ""
	}

	return net.JoinHostPort(ip, port)
}
