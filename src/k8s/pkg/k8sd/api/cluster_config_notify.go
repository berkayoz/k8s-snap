package api

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

type ClusterConfigNotifyResponse struct{}

func (e *Endpoints) postClusterConfigNotify(s state.State, r *http.Request) response.Response {
	if err := e.provider.OnClusterConfigChanged(r.Context(), s); err != nil {
		return response.InternalError(fmt.Errorf("failed to handle cluster config change: %w", err))
	}
	return response.SyncResponse(true, &ClusterConfigNotifyResponse{})
}
