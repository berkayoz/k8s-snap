package cilium

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type GatewayAPI struct {
	*component.BaseComponent
}

func NewGatewayAPI(s state.State, snap snap.Snap) *GatewayAPI {
	return &GatewayAPI{
		BaseComponent: component.NewBaseComponent(s, snap, "ck-gateway", "kube-system", chartGateway),
	}
}

func (g *GatewayAPI) IsEnabled(config types.ClusterConfig) bool {
	return *config.Gateway.Enabled
}

func (g *GatewayAPI) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	return nil, nil
}

func (g *GatewayAPI) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayAPI) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayAPI) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayAPI) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayAPI) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayAPI) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayAPI) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
