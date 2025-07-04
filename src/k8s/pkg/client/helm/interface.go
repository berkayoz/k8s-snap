package helm

import (
	"context"
	"fmt"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
)

// Client handles the lifecycle of charts (manifests + config) on the cluster.
type Client interface {
	// Apply ensures the state of a InstallableChart on the cluster.
	// When state is StatePresent, Apply will install or upgrade the chart using the specified values as configuration. Apply returns true if the chart was not installed, or any values were changed.
	// When state is StateUpgradeOnly, Apply will upgrade the chart using the specified values as configuration. Apply returns true if the chart was not installed, or any values were changed. An error is returned if the chart is not already installed.
	// When state is StateDeleted, Apply will ensure that the chart is removed. If the chart is not installed, this is a no-op. Apply returns true if the chart was previously installed.
	// Apply returns an error in case of failure.
	Apply(ctx context.Context, f InstallableChart, desired State, values map[string]any) (bool, error)

	// Get retrieves the latest Helm release with the specified name and namespace.
	Get(ctx context.Context, release string, namespace string) (*Release, error)

	// Install installs a new release with the given chart and values.
	Install(ctx context.Context, chart *chart.Chart, values map[string]any) (*Release, error)

	// Upgrade upgrades an existing release with the given chart and values.
	Upgrade(ctx context.Context, chart *chart.Chart, values map[string]any) (*Release, error)

	// Uninstall removes a release from the cluster.
	Uninstall(ctx context.Context, chart *chart.Chart) (*release.UninstallReleaseResponse, error)

	// IsNotFound checks if the error is a not found error.
	IsNotFound(err error) bool
}

type Release struct {
	*release.Release
}

func (r *Release) ShouldUpgrade(f InstallableChart, values map[string]any) (bool, error) {
	// If the release is nil, we cannot upgrade.
	if r.Release == nil {
		return false, fmt.Errorf("cannot upgrade release %s: release is nil", f.Name)
	}
	chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
	if err != nil {
		return false, fmt.Errorf("failed to load manifest for: %w", err)
	}
	// NOTE(Angelos): oldConfig and values are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
	// NOTE(Hue) (KU-3592): We are ignoring the values that are overwritten by the user.
	// The user can change some values in the chart, but we will revert them back upon an upgrade.
	// NOTE(Hue): We clone the values map to avoid modifying the original user provided values.
	clonedValues, err := cloneMap(values)
	if err != nil {
		return false, fmt.Errorf("failed to clone values %s: %w", err)
	}
	mergedValues := chartutil.CoalesceTables(clonedValues, r.Config)
	sameValues := jsonEqual(r.Config, mergedValues)
	// NOTE(Hue): For the charts that we manage (e.g. ck-loadbalancer), we need to make
	// sure we bump the version manually. Otherwise, they'll not be applied unless
	// we're lucky and providing different extra values.
	sameVersions := r.Chart.Metadata.Version == chart.Metadata.Version

	// Check if the values have changed.
	return !sameValues || !sameVersions, nil
}
