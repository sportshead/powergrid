package kubernetes

import (
	"context"
	"github.com/sportshead/powergrid/coordinator/env"
	"github.com/sportshead/powergrid/coordinator/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"log/slog"
	"os"
	"time"
)

func startLeader() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanupGroup.Add(1)

	go func() {
		<-stop
		cancel()
		cleanupGroup.Done()
	}()

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      env.DeploymentName,
			Namespace: namespace,
		},
		Client: kubernetesClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: env.Hostname,
		},
	}

	slog.Info("joining leader election", utils.Tag("lead_join"), slog.String("id", env.Hostname))
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   30 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				slog.Info("started leading", utils.Tag("lead_start"), slog.String("id", env.Hostname))
				wait.NonSlidingUntilWithContext(ctx, updateCommands, time.Minute)
			},
			OnStoppedLeading: func() {
				slog.Error("stopped leading", utils.Tag("lead_lost"), slog.String("id", env.Hostname))
				os.Exit(0)
			},
			OnNewLeader: func(leader string) {
				if leader == env.Hostname {
					// we're leading
					return
				}
				slog.Info("observed new leader", utils.Tag("lead_changed"), slog.String("leader", leader))
			},
		},
	})
}
