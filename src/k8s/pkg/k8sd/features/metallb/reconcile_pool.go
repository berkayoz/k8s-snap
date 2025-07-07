package metallb

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type LBCRs struct {
	*component.BaseComponent
}

func NewLBCRs(s state.State, snap snap.Snap) *LBCRs {
	return &LBCRs{
		BaseComponent: component.NewBaseComponent(s, snap, "metallb-loadbalancer", "metallb-system", ChartMetalLBLoadBalancer),
	}
}

func (l *LBCRs) IsEnabled(config types.ClusterConfig) bool {
	return *config.LoadBalancer.Enabled
}

func (l *LBCRs) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	cidrs := []map[string]any{}
	for _, cidr := range config.LoadBalancer.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range config.LoadBalancer.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	values := map[string]any{
		"driver": "metallb",
		"l2": map[string]any{
			"enabled":    config.LoadBalancer.GetL2Mode(),
			"interfaces": config.LoadBalancer.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": cidrs,
		},
		"bgp": map[string]any{
			"enabled":  config.LoadBalancer.GetBGPMode(),
			"localASN": config.LoadBalancer.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": config.LoadBalancer.GetBGPPeerAddress(),
					"peerASN":     config.LoadBalancer.GetBGPPeerASN(),
					"peerPort":    config.LoadBalancer.GetBGPPeerPort(),
				},
			},
		},
	}

	return values, nil
}

func (l *LBCRs) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (l *LBCRs) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	snap := l.Snap()

	// TODO(berkayoz): This should not wait for anything but only check if the CRDs are present.
	if err := waitForRequiredLoadBalancerCRDs(ctx, snap, config.LoadBalancer.GetBGPMode()); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to wait for required MetalLB CRDs: %w", err)
	}

	return ctrl.Result{}, nil
}

func (l *LBCRs) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (l *LBCRs) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (l *LBCRs) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (l *LBCRs) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (l *LBCRs) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
