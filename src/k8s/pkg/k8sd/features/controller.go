package features

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"helm.sh/helm/v3/pkg/chart/loader"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Feature interface {
	IsEnabled(config types.ClusterConfig) bool
	GetReleaseName() string
	GetNamespace() string
	GetValues(config types.ClusterConfig) map[string]any
	GetInstallableChart() helm.InstallableChart
}

type Controller struct {
	Feature
	snap snap.Snap
}

func (r *Controller) Reconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	helm := r.snap.HelmClient()

	f := r.GetInstallableChart()

	chart, err := loader.Load(filepath.Join("basedir", f.ManifestPath))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to load manifest for: %w", err)
	}

	release, err := helm.Get(ctx, r.GetReleaseName(), r.GetNamespace())
	if err != nil {
		if helm.IsNotFound(err) && r.IsEnabled(config) {
			// Feature is not installed, but it is enabled in the config.
			// Install the feature with the provided values.
			values := r.GetValues(config)
			_, err = helm.Install(ctx, chart, values)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}

	if !r.IsEnabled(config) {
		// Feature is installed, but it is disabled in the config.
		// Uninstall the feature.
		_, err = helm.Uninstall(ctx, chart)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to uninstall %s: %w", f.Name, err)
		}
		return ctrl.Result{}, nil
	}

	// Feature is installed and enabled in the config.
	// Check if the release is up to date with the desired values.
	values := r.GetValues(config)

	shouldUpgrade, err := release.ShouldUpgrade(f, values)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if %s should be upgraded: %w", f.Name, err)
	}

	if shouldUpgrade {
		// Upgrade the feature with the provided values.
		_, err = helm.Upgrade(ctx, chart, values)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to upgrade %s: %w", f.Name, err)
		}
	}

	return ctrl.Result{}, nil
}
