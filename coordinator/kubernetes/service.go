package kubernetes

import (
	"github.com/sportshead/powergrid/coordinator/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log/slog"
	"net"
	"strconv"
	"time"
)

var serviceInformer cache.SharedIndexInformer

func loadServices() {
	factory := informers.NewSharedInformerFactoryWithOptions(kubernetesClient, 10*time.Minute, informers.WithNamespace(namespace))
	serviceInformer = factory.Core().V1().Services().Informer()

	stopCh := make(chan struct{})
	factory.Start(stopCh)            // start goroutines
	factory.WaitForCacheSync(stopCh) // wait for init
}

func GetServiceAddr(log *slog.Logger, serviceName string) string {
	log = log.With(slog.String("name", serviceName))
	svc, exists, err := serviceInformer.GetIndexer().GetByKey(namespace + "/" + serviceName)
	if err != nil {
		log.Error("failed to get service", utils.Tag("k8s_service_get_failed"), utils.Error(err))
		return ""
	}
	if !exists {
		log.Error("service does not exist", utils.Tag("k8s_service_missing"))
		return ""
	}

	log = log.With(slog.String("object", utils.TryMarshal(svc)))
	service, ok := svc.(*corev1.Service)
	if !ok {
		log.Error("failed to cast service", utils.Tag("k8s_service_cast_failed"))
		return ""
	}

	ip := service.Spec.ClusterIP
	if ip == "" || ip == "None" {
		log.Error("service has no clusterIP", utils.Tag("k8s_service_missing_cluster_ip"))
		return ""
	}

	ports := service.Spec.Ports
	port := ""
	// get http port, otherwise first port
	for _, p := range ports {
		if port == "" || p.Name == "http" {
			port = strconv.Itoa(int(p.Port))
			break
		}
	}
	if port == "" {
		log.Error("service has no http port", utils.Tag("k8s_service_missing_http_port"))
		return ""
	}

	return net.JoinHostPort(ip, port)
}
