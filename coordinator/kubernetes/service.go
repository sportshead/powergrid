package kubernetes

import (
	"context"
	"github.com/sportshead/powergrid/coordinator/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"net"
	"strconv"
)

func GetServiceAddr(ctx context.Context, serviceName string) string {
	service, err := kubernetesClient.CoreV1().Services(namespace).Get(ctx, serviceName, v1.GetOptions{})
	if err != nil {
		slog.Error("failed to get service", utils.Tag("k8s_service_get_failed"), utils.Error(err))
		return ""
	}

	ip := service.Spec.ClusterIP
	if ip == "" || ip == "None" {
		slog.Error("service has no clusterIP", utils.Tag("k8s_service_missing_cluster_ip"), slog.String("name", serviceName), slog.String("object", utils.TryMarshal(service)))
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
		slog.Error("service has no http port", utils.Tag("k8s_service_missing_http_port"), slog.String("name", serviceName), slog.String("object", utils.TryMarshal(service)))
		return ""
	}

	return net.JoinHostPort(ip, port)
}
