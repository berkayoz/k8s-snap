package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	"github.com/canonical/k8s/pkg/k8sd/features/metallb"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type FeatureReconciler struct {
	s    state.State
	snap snap.Snap

	componentReconciler *component.ComponentReconciler

	// The reconciler will handle multiple features, each represented by a component.
	cilium        *cilium.Cilium
	metallb       *metallb.MetalLB
	lbCRs         *metallb.LBCRs
	localpv       *localpv.LocalPV
	coredns       *coredns.CoreDNS
	gatewayapi    *cilium.GatewayAPI
	gatewayClass  *cilium.GatewayClass
	metricsServer *metrics_server.MetricsServer
}

func NewFeatureReconciler(s state.State, snap snap.Snap) *FeatureReconciler {
	return &FeatureReconciler{
		s:                   s,
		snap:                snap,
		componentReconciler: component.NewComponentReconciler(snap),
		cilium:              cilium.NewCilium(s, snap),
		metallb:             metallb.NewMetalLB(s, snap),
		lbCRs:               metallb.NewLBCRs(s, snap),
		localpv:             localpv.NewLocalPV(s, snap),
		coredns:             coredns.NewCoreDNS(s, snap),
		gatewayapi:          cilium.NewGatewayAPI(s, snap),
		gatewayClass:        cilium.NewGatewayClass(s, snap),
		metricsServer:       metrics_server.NewMetricsServer(s, snap),
	}
}

func (fr *FeatureReconciler) Reconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	result, err := fr.componentReconciler.Reconcile(ctx, fr.gatewayapi, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile GatewayAPI: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.gatewayClass, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile GatewayClass: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.cilium, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile Cilium: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.coredns, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile CoreDNS: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.localpv, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile LocalPV: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.metallb, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile MetalLB: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.lbCRs, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile MetalLB CRs: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	result, err = fr.componentReconciler.Reconcile(ctx, fr.metricsServer, config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile Metrics Server: %w", err)
	}
	if !result.IsZero() {
		return result, nil
	}

	return ctrl.Result{}, nil
}

// StatusChecks implements the Canonical Kubernetes built-in feature status checks.
var StatusChecks StatusInterface = &statusChecks{
	checkNetwork: cilium.CheckNetwork,
	checkDNS:     coredns.CheckDNS,
}

var Cleanup CleanupInterface = &cleanup{
	cleanupNetwork: cilium.CleanupNetwork,
}
