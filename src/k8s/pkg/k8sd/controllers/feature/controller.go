package feature

import (
	"context"
	"fmt"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	configv1alpha1 "github.com/canonical/k8s/pkg/k8sd/crds/config"
	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var customEventChan = make(chan event.TypedGenericEvent[*configv1alpha1.ClusterConfig])

type Controller struct {
	s                state.State
	snap             snap.Snap
	logger           logr.Logger
	client           client.Client
	getClusterConfig func(context.Context) (types.ClusterConfig, error)

	featureReconciler *features.FeatureReconciler
}

func NewController(
	s state.State,
	snap snap.Snap,
	logger logr.Logger,
	client client.Client,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
) *Controller {
	return &Controller{
		s:                 s,
		snap:              snap,
		logger:            logger,
		client:            client,
		getClusterConfig:  getClusterConfig,
		featureReconciler: features.NewFeatureReconciler(s, snap),
	}
}

// TODO(berkayoz): Periodic trigger to not miss events, similar to what controller-runtime does.
func (r *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("feature-controller").
		WatchesRawSource(source.Channel(customEventChan, &handler.TypedEnqueueRequestForObject[*configv1alpha1.ClusterConfig]{})).
		Complete(r)
}

// Reconcile implements the reconcile.TypedReconciler interface.
func (r *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger.Info("Reconciling ClusterConfig", "request", req)

	config, err := r.getClusterConfig(ctx)
	if err != nil {
		r.logger.Error(err, "Failed to get cluster config")
		return ctrl.Result{}, err
	}

	blocked, err := r.isBlocked(ctx, config)
	if err != nil {
		r.logger.Error(err, "Failed to check if feature controller is blocked")
		return ctrl.Result{}, err
	}
	if blocked {
		r.logger.Info("Feature controller is blocked by an in-progress upgrade, skipping reconciliation")
		// TODO(berkayoz): Consider using requeue after?
		return ctrl.Result{}, nil
	}

	return r.featureReconciler.Reconcile(ctx, config)
}

// isBlocked checks if the feature controller is blocked by an in-progress upgrade.
// If an upgrade is in progress, the feature controller will not apply any configuration changes.
func (r *Controller) isBlocked(ctx context.Context, clusterConfig types.ClusterConfig) (bool, error) {
	log := log.FromContext(ctx)

	// Skip feature reconciliation while an upgrade is in progress to avoid conflicting cluster
	// configuration changes.
	if _, ok := clusterConfig.Annotations.Get(apiv1_annotations.AnnotationDisableSeparateFeatureUpgrades); !ok {
		k8sClient, err := r.snap.KubernetesClient("")
		if err != nil {
			return false, fmt.Errorf("failed to get Kubernetes client: %w", err)
		}

		upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to check for in-progress upgrade: %w", err)
		}

		if upgrade == nil {
			return false, nil
		}

		if upgrade.Status.Phase == upgradesv1alpha.UpgradePhaseFeatureUpgrade {
			log.Info("Upgrade in progress - but in feature upgrade phase - applying configuration", "upgrade", upgrade.Name, "phase", upgrade.Status.Phase)
			return false, nil
		}

		log.Info("Upgrade in progress - feature controller blocked", "upgrade", upgrade.Name, "phase", upgrade.Status.Phase)
		return true, nil
	}

	return false, nil
}

func SendClusterConfigEvent() {
	// This function can be used to send events to the customEventChan.
	// For example, you can create a new ClusterConfig event and send it.
	clusterConfig := &configv1alpha1.ClusterConfig{
		// Populate the ClusterConfig fields as needed.
	}
	event := event.TypedGenericEvent[*configv1alpha1.ClusterConfig]{
		Object: clusterConfig,
	}
	customEventChan <- event
	// Ensure to handle the channel properly in your application to avoid blocking.
}
