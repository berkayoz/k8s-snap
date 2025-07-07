package localpv

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type LocalPV struct {
	*component.BaseComponent
}

func NewLocalPV(s state.State, snap snap.Snap) *LocalPV {
	return &LocalPV{
		BaseComponent: component.NewBaseComponent(s, snap, "ck-storage", "kube-system", Chart),
	}
}

func (l *LocalPV) IsEnabled(config types.ClusterConfig) bool {
	return *config.LocalStorage.Enabled
}

func (l *LocalPV) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	values := map[string]any{
		"storageClass": map[string]any{
			"enabled":       true,
			"isDefault":     config.LocalStorage.GetDefault(),
			"reclaimPolicy": config.LocalStorage.GetReclaimPolicy(),
		},
		"serviceMonitor": map[string]any{
			"enabled": false,
		},
		"controller": map[string]any{
			"csiDriverArgs": []string{"--args", "rawfile", "csi-driver", "--disable-metrics"},
			"image": map[string]any{
				"repository": imageRepo,
				"tag":        ImageTag,
			},
		},
		"node": map[string]any{
			"image": map[string]any{
				"repository": imageRepo,
				"tag":        ImageTag,
			},
			"storage": map[string]any{
				"path": config.LocalStorage.GetLocalPath(),
			},
		},
		"images": map[string]any{
			"csiNodeDriverRegistrar": csiNodeDriverImage,
			"csiProvisioner":         csiProvisionerImage,
			"csiResizer":             csiResizerImage,
			"csiSnapshotter":         csiSnapshotterImage,
		},
	}

	return values, nil
}

func (l *LocalPV) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (l *LocalPV) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (l *LocalPV) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (l *LocalPV) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (l *LocalPV) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (l *LocalPV) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (l *LocalPV) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
