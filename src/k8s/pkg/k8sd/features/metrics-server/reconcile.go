package metrics_server

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type MetricsServer struct {
	*features.BaseComponent
}

func NewMetricsServer(s state.State, snap snap.Snap) *MetricsServer {
	return &MetricsServer{
		BaseComponent: features.NewBaseComponent(s, snap, "metrics-server", "kube-system", chart),
	}
}

func (m *MetricsServer) IsEnabled(config types.ClusterConfig) bool {
	return *config.MetricsServer.Enabled
}

func (m *MetricsServer) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	mcnf := internalConfig(config.Annotations)

	values := map[string]any{
		"image": map[string]any{
			"repository": mcnf.imageRepo,
			"tag":        mcnf.imageTag,
		},
		"securityContext": map[string]any{
			// ROCKs with Pebble as the entrypoint do not work with this option.
			"readOnlyRootFilesystem": false,
		},
	}

	return values, nil
}

func (m *MetricsServer) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (m *MetricsServer) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (m *MetricsServer) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (m *MetricsServer) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (m *MetricsServer) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (m *MetricsServer) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (m *MetricsServer) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
