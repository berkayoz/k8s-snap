package cilium

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type GatewayClass struct {
	*component.BaseComponent
}

func NewGatewayClass(s state.State, snap snap.Snap) *GatewayClass {
	return &GatewayClass{
		BaseComponent: component.NewBaseComponent(s, snap, "ck-gateway-class", "kube-system", chartGatewayClass),
	}
}

func (g *GatewayClass) IsEnabled(config types.ClusterConfig) bool {
	return *config.Gateway.Enabled
}

func (g *GatewayClass) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	return nil, nil
}

func (g *GatewayClass) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayClass) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayClass) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayClass) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayClass) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayClass) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (g *GatewayClass) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
