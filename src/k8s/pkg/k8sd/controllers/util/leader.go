package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	coordinationv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/types"
)

func IsLeader(ctx context.Context, s state.State, snap snap.Snap) (bool, error) {
	k8sclient, err := snap.KubernetesClient("")
	if err != nil {
		return false, fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	var lease coordinationv1.Lease
	if err := k8sclient.Get(ctx, types.NamespacedName{
		Name:      "oy6981cu.controller-coordinator",
		Namespace: "kube-system",
	}, &lease); err != nil {
		return false, fmt.Errorf("failed to retrieve lease: %w", err)
	}

	identity := strings.Split(*lease.Spec.HolderIdentity, "_")

	if identity[0] != s.Name() {
		return false, nil
	}
	return true, nil
}
