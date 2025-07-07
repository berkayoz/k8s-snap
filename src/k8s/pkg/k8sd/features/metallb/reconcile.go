package metallb

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type MetalLB struct {
	*component.BaseComponent
}

func NewMetalLB(s state.State, snap snap.Snap) *MetalLB {
	return &MetalLB{
		BaseComponent: component.NewBaseComponent(s, snap, "metallb", "metallb-system", ChartMetalLB),
	}
}

func (m *MetalLB) IsEnabled(config types.ClusterConfig) bool {
	return *config.LoadBalancer.Enabled
}

func (m *MetalLB) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	values := map[string]any{
		"controller": map[string]any{
			"image": map[string]any{
				"repository": controllerImageRepo,
				"tag":        ControllerImageTag,
			},
			"command": "/controller",
		},
		"speaker": map[string]any{
			"image": map[string]any{
				"repository": speakerImageRepo,
				"tag":        speakerImageTag,
			},
			"command": "/speaker",
			// TODO(neoaggelos): make frr enable/disable configurable through an annotation
			// We keep it disabled by default
			"frr": map[string]any{
				"enabled": false,
				"image": map[string]any{
					"repository": frrImageRepo,
					"tag":        frrImageTag,
				},
			},
		},
	}

	return values, nil
}

func (m *MetalLB) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (m *MetalLB) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (m *MetalLB) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (m *MetalLB) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (m *MetalLB) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (m *MetalLB) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (m *MetalLB) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
