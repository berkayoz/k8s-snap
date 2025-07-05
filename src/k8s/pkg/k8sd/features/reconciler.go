// Package features provides interfaces and base implementations for feature components.
package features

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"helm.sh/helm/v3/pkg/chart/loader"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ComponentReconciler provides a base implementation for a reconciler.
type ComponentReconciler struct {
	Snap snap.Snap
}

// Reconcile reconciles a component's release state.
func (r *ComponentReconciler) Reconcile(ctx context.Context, component Component, config types.ClusterConfig) (ctrl.Result, error) {
	helm := r.Snap.HelmClient()

	f := component.InstallableChart()

	chart, err := loader.Load(filepath.Join("basedir", f.ManifestPath))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to load manifest for: %w", err)
	}

	result, err := component.PreReconcile(ctx, config)
	if err != nil {
		return result, fmt.Errorf("pre-reconcile failed for %s: %w", f.Name, err)
	}
	if !result.IsZero() {
		return result, nil
	}

	release, err := helm.Get(ctx, component.ReleaseName(), component.Namespace())
	if err != nil {
		if helm.IsNotFound(err) && component.IsEnabled(config) {
			// Feature is not installed, but it is enabled in the config.
			// Install the feature with the provided values.

			result, err = component.PreInstall(ctx, config)
			if err != nil {
				return result, fmt.Errorf("pre-install failed for %s: %w", f.Name, err)
			}
			if !result.IsZero() {
				return result, nil
			}

			values, err := component.Values(ctx, config)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to get values for %s: %w", f.Name, err)
			}

			_, err = helm.Install(ctx, chart, values)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to install %s: %w", f.Name, err)
			}

			result, err = component.PostInstall(ctx, config)
			if err != nil {
				return result, fmt.Errorf("post-install failed for %s: %w", f.Name, err)
			}
			if !result.IsZero() {
				return result, nil
			}
		}
		return ctrl.Result{}, err
	}

	// TODO(berkayoz): Pending / broken releases might be handled here.

	if !component.IsEnabled(config) {
		// Feature is installed, but it is disabled in the config.
		// Uninstall the feature.
		result, err = component.PreUninstall(ctx, config)
		if err != nil {
			return result, fmt.Errorf("pre-uninstall failed for %s: %w", f.Name, err)
		}
		if !result.IsZero() {
			return result, nil
		}

		_, err = helm.Uninstall(ctx, chart)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to uninstall %s: %w", f.Name, err)
		}

		result, err = component.PostUninstall(ctx, config)
		if err != nil {
			return result, fmt.Errorf("post-uninstall failed for %s: %w", f.Name, err)
		}
		if !result.IsZero() {
			return result, nil
		}

		return ctrl.Result{}, nil
	}

	// Feature is installed and enabled in the config.
	// Check if the release is up to date with the desired values.
	values, err := component.Values(ctx, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get values for %s: %w", f.Name, err)
	}

	shouldUpgrade, err := release.ShouldUpgrade(chart, values)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if %s should be upgraded: %w", f.Name, err)
	}

	if shouldUpgrade {
		result, err = component.PreUpgrade(ctx, config)
		if err != nil {
			return result, fmt.Errorf("pre-upgrade failed for %s: %w", f.Name, err)
		}
		if !result.IsZero() {
			return result, nil
		}

		// Upgrade the feature with the provided values.
		_, err = helm.Upgrade(ctx, chart, values)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to upgrade %s: %w", f.Name, err)
		}

		result, err = component.PostUpgrade(ctx, config)
		if err != nil {
			return result, fmt.Errorf("post-upgrade failed for %s: %w", f.Name, err)
		}
		if !result.IsZero() {
			return result, nil
		}
	}

	return ctrl.Result{}, nil
}
