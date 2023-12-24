package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
)

var config *rest.Config

// TODO: typed client
var client *dynamic.DynamicClient
var namespace = v1.NamespaceDefault

var discordCommandResource = schema.GroupVersionResource{Group: "powergrid.sportshead.dev", Version: "v10", Resource: "commands"}
var CommandMap = make(map[string]*unstructured.Unstructured)

func initKubernetes() {
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		slog.Error("failed to get in cluster config", slogTag("k8s_config_create_failed"), slogError(err))
		os.Exit(1)
	}
	client, err = dynamic.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create dynamic client", slogTag("k8s_client_create_failed"), slogError(err))
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

func loadCommands() {
	commands, err := client.Resource(discordCommandResource).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		slog.Error("failed to get commands", slogTag("k8s_command_list_failed"), slogError(err))
		os.Exit(1)
	}

	for _, command := range commands.Items {
		commandName := command.Object["spec"].(map[string]interface{})["command"].(map[string]interface{})["name"].(string)

		CommandMap[commandName] = &command

		slog.Info("got command", slogTag("k8s_command_loaded"), slog.String("name", command.GetName()), slog.String("command", commandName))
	}
}

var serviceResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

func getServiceAddr(ctx context.Context, serviceName string) string {
	service, err := client.Resource(serviceResource).Namespace(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		slog.Error("failed to get service", slogTag("k8s_service_get_failed"), slogError(err))
		return ""
	}
	spec, ok := service.Object["spec"].(map[string]interface{})
	if !ok {
		slog.Error("failed to cast spec", slogTag("k8s_service_cast_failed"), slog.String("key", "spec"), slog.String("object", tryMarshal(service.Object)))
		return ""
	}
	ip, ok := spec["clusterIP"].(string)
	if !ok {
		slog.Error("failed to cast clusterIP", slogTag("k8s_service_cast_failed"), slog.String("key", "clusterIP"), slog.String("object", tryMarshal(service.Object)))
		return ""
	}
	if ip == "" || ip == "None" {
		slog.Error("service has no clusterIP", slogTag("k8s_service_missing_cluster_ip"), slog.String("name", serviceName), slog.String("object", tryMarshal(service.Object)))
		return ""
	}

	ports, ok := spec["ports"].([]interface{})
	if !ok {
		slog.Error("failed to cast ports", slogTag("k8s_service_cast_failed"), slog.String("key", "ports"), slog.String("object", tryMarshal(service.Object)))
		return ""
	}
	port := ""
	for _, _p := range ports {
		p := _p.(map[string]interface{})
		if port == "" || p["name"] == "http" {
			port = strconv.Itoa(int(p["port"].(int64)))
			break
		}
	}
	if port == "" {
		slog.Error("service has no http port", slogTag("k8s_service_missing_http_port"), slog.String("name", serviceName), slog.String("object", tryMarshal(service.Object)))
		return ""
	}

	return net.JoinHostPort(ip, port)
}

func tryMarshal(obj map[string]interface{}) string {
	if obj == nil {
		return "<nil>"
	}

	//bytes, err := json.Marshal(obj)
	//if err == nil {
	//	return string(bytes)
	//}

	return fmt.Sprintf("%#v", obj)
}
