package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	ctrl "sigs.k8s.io/controller-runtime"
)

const namespace = "kube-system"

type CiliumController struct {
	snap snap.Snap
}

func (r *CiliumController) Reconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	helm := r.snap.HelmClient()

	network := config.Network

	release, err := helm.Get(ctx, "cilium", namespace)
	if err != nil {
		if helm.IsNotFound(err) && *network.Enabled {
			// Cilium is not installed, but it is enabled in the config.
			// Install Cilium with the provided values.
			values := r.valuesForCilium(config)
			_, err = helm.Install(ctx, ChartCilium, values)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to install Cilium: %w", err)
			}

			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !*network.Enabled {
		// Cilium is installed, but it is disabled in the config.
		// Uninstall Cilium.
		_, err = helm.Uninstall(ctx, ChartCilium)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to uninstall Cilium: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Cilium is installed and enabled in the config.
	// Check if the release is up to date with the desired values.

	values := r.valuesForCilium(config)
	shouldUpgrade, err := release.ShouldUpgrade(ChartCilium, values)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if Cilium should be upgraded: %w", err)
	}

	if shouldUpgrade {
		// Upgrade Cilium with the provided values.
		_, err = helm.Upgrade(ctx, ChartCilium, values)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to upgrade Cilium: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *CiliumController) valuesForCilium(config types.ClusterConfig) map[string]any {
	values := map[string]any{}

	return values
}
