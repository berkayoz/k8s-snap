package features

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Component represents a feature component with lifecycle hooks.
type Component interface {
	IsEnabled(config types.ClusterConfig) bool
	ReleaseName() string
	Namespace() string
	Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error)
	InstallableChart() helm.InstallableChart

	LifecycleHooks
}

// LifecycleHooks defines lifecycle hook methods for a component.
type LifecycleHooks interface {
	// PreReconcile is called before the reconciliation process.
	// It can be used to check for certain conditions before any changes are applied.
	PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)

	// PreInstall is called before the installation process.
	PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)

	// PostInstall is called after the installation process.
	// It can be used to perform any necessary actions after the component is installed.
	PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)

	// PreUninstall is called before the uninstallation process.
	PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)

	// PostUninstall is called after the uninstallation process.
	// It can be used to clean up resources or perform any necessary actions after the component is
	PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)

	// PreUpgrade is called before the upgrade process.
	PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)

	// PostUpgrade is called after the upgrade process.
	// It can be used to perform any necessary actions after the component is upgraded.
	PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error)
}

// BaseComponent provides a base implementation for Component.
type BaseComponent struct {
	releaseName      string
	namespace        string
	installableChart helm.InstallableChart

	snap  snap.Snap
	state state.State
}

// NewBaseComponent returns a new BaseComponent.
func NewBaseComponent(s state.State, snap snap.Snap, releaseName, namespace string, installableChart helm.InstallableChart) *BaseComponent {
	return &BaseComponent{
		releaseName:      releaseName,
		namespace:        namespace,
		installableChart: installableChart,
		snap:             snap,
		state:            s,
	}
}

func (c *BaseComponent) State() state.State {
	return c.state
}

func (c *BaseComponent) Snap() snap.Snap {
	return c.snap
}

// ReleaseName returns the release name.
func (c *BaseComponent) ReleaseName() string {
	return c.releaseName
}

// Namespace returns the namespace.
func (c *BaseComponent) Namespace() string {
	return c.namespace
}

// InstallableChart returns the installable chart.
func (c *BaseComponent) InstallableChart() helm.InstallableChart {
	return c.installableChart
}
