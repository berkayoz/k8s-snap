package app

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/controllers/feature"
	controllerutil "github.com/canonical/k8s/pkg/k8sd/controllers/util"
	"github.com/canonical/microcluster/v2/state"
)

func (a *App) OnClusterConfigChanged(ctx context.Context, s state.State) error {
	snap := a.Snap()

	isLeader, err := controllerutil.IsLeader(ctx, s, snap)
	if err != nil {
		return fmt.Errorf("failed to check if node is leader: %w", err)
	}

	if isLeader {
		feature.SendClusterConfigEvent()
	}

	return nil
}
