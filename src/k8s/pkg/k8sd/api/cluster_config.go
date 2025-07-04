package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	microclusterapi "github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/v2/client"
	microclusterresponse "github.com/canonical/microcluster/v2/rest/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) putClusterConfig(s state.State, r *http.Request) response.Response {
	var req apiv1.SetClusterConfigRequest

	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	requestedConfig, err := types.ClusterConfigFromUserFacing(req.Config)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid configuration: %w", err))
	}
	if requestedConfig.Datastore, err = types.DatastoreConfigFromUserFacing(req.Datastore); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse datastore config: %w", err))
	}

	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, requestedConfig); err != nil {
			return fmt.Errorf("failed to update cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
	}

	e.provider.NotifyUpdateNodeConfigController()
	e.provider.NotifyFeatureController(
		!requestedConfig.Network.Empty(),
		!requestedConfig.Gateway.Empty(),
		!requestedConfig.Ingress.Empty(),
		!requestedConfig.LoadBalancer.Empty(),
		!requestedConfig.LocalStorage.Empty(),
		!requestedConfig.MetricsServer.Empty(),
		!requestedConfig.DNS.Empty() || !requestedConfig.Kubelet.Empty(),
	)

	// TODO(berkayoz): Maybe this should run in a goroutine?
	// Maybe https://github.com/canonical/dqlite/issues/326?
	if err := e.provider.OnClusterConfigChanged(r.Context(), s); err != nil {
		return response.InternalError(fmt.Errorf("failed to handle cluster config change: %w", err))
	}

	// TODO(berkayoz): Errors below should be ignored?
	if err := clusterConfigNotify(r.Context(), s); err != nil {
		return response.InternalError(fmt.Errorf("failed to notify cluster config change: %w", err))
	}

	return response.SyncResponse(true, &apiv1.SetClusterConfigResponse{})
}

func (e *Endpoints) getClusterConfig(s state.State, r *http.Request) response.Response {
	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}

	return response.SyncResponse(true, &apiv1.GetClusterConfigResponse{
		Config:      config.ToUserFacing(),
		Datastore:   config.Datastore.ToUserFacing(),
		PodCIDR:     config.Network.PodCIDR,
		ServiceCIDR: config.Network.ServiceCIDR,
	})
}

func clusterConfigNotify(ctx context.Context, s state.State) error {
	// Performs a request against all members except self
	cluster, err := s.Cluster(true)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	// Error is ignored here, as this is a best-effort request to notify other members
	cluster.Query(ctx, true, func(ctx context.Context, client *client.Client) error {
		resp, err := client.QueryRaw(ctx, "POST", apiv1.K8sdAPIVersion, microclusterapi.NewURL().Path(strings.Split("k8sd/cluster/config/notify", "/")...), nil)
		if err != nil {
			return fmt.Errorf("failed to request cluster config notify: %w", err)
		}
		defer resp.Body.Close()

		if _, err := microclusterresponse.ParseResponse(resp); err != nil {
			return fmt.Errorf("failed to handle cluster config notify: %w", err)
		}
		return nil
	})

	return nil
}
