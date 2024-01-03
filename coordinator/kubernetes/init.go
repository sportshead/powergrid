package kubernetes

import (
	clientset "github.com/sportshead/powergrid/coordinator/pkg/generated/clientset/versioned"
	"github.com/sportshead/powergrid/coordinator/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log/slog"
	"os"
	"strings"
)

var config *rest.Config

var powergridClient *clientset.Clientset
var kubernetesClient *kubernetes.Clientset
var namespace = corev1.NamespaceDefault

func Init() {
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		slog.Error("failed to get in cluster config", utils.Tag("k8s_config_create_failed"), utils.Error(err))
		os.Exit(1)
	}
	powergridClient, err = clientset.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create k8s client", utils.Tag("k8s_client_create_failed"), utils.Error(err), slog.String("client", "powergrid"))
		os.Exit(1)
	}
	kubernetesClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create k8s client", utils.Tag("k8s_client_create_failed"), utils.Error(err), slog.String("client", "kubernetes"))
		os.Exit(1)
	}

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		slog.Warn("failed to read namespace file", utils.Tag("k8s_namespace_read_failed"), utils.Error(err))
	} else {
		namespace = strings.TrimSpace(string(data))
	}

	slog.Info("initiated kubernetes client",
		utils.Tag("k8s_client_created"),
		slog.String("namespace", namespace),
		slog.String("host", config.Host))

	go loadCommands()
}
