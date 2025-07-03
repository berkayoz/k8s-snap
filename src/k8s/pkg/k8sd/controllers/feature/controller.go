package feature

import (
	"context"

	configv1alpha1 "github.com/canonical/k8s/pkg/k8sd/crds/config"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var customEventChan = make(chan event.TypedGenericEvent[*configv1alpha1.ClusterConfig])

type Controller struct {
	logger           logr.Logger
	client           client.Client
	getClusterConfig func(context.Context) (types.ClusterConfig, error)
}

func NewController(
	logger logr.Logger,
	client client.Client,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
) *Controller {
	return &Controller{
		logger:           logger,
		client:           client,
		getClusterConfig: getClusterConfig,
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
	// TODO: Add your reconciliation logic here.
	return ctrl.Result{}, nil
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
