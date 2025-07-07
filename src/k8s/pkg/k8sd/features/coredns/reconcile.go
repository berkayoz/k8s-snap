package coredns

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CoreDNS struct {
	*component.BaseComponent
}

func NewCoreDNS(s state.State, snap snap.Snap) *CoreDNS {
	return &CoreDNS{
		BaseComponent: component.NewBaseComponent(s, snap, "ck-dns", "kube-system", Chart),
	}
}

func (c *CoreDNS) IsEnabled(config types.ClusterConfig) bool {
	return *config.DNS.Enabled
}

func (c *CoreDNS) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	values := map[string]any{
		"image": map[string]any{
			"repository": imageRepo,
			"tag":        ImageTag,
		},
		"service": map[string]any{
			"name":      "coredns",
			"clusterIP": config.Kubelet.GetClusterDNS(),
		},
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "coredns",
		},
		"deployment": map[string]any{
			"name": "coredns",
		},
		"servers": []map[string]any{
			{
				"zones": []map[string]any{
					{"zone": "."},
				},
				"port": 53,
				"plugins": []map[string]any{
					{"name": "errors"},
					{"name": "health", "configBlock": "lameduck 5s"},
					{"name": "ready"},
					{
						"name":        "kubernetes",
						"parameters":  fmt.Sprintf("%s in-addr.arpa ip6.arpa", config.Kubelet.GetClusterDomain()),
						"configBlock": "pods insecure\nfallthrough in-addr.arpa ip6.arpa\nttl 30",
					},
					{"name": "prometheus", "parameters": "0.0.0.0:9153"},
					{"name": "forward", "parameters": fmt.Sprintf(". %s", strings.Join(config.DNS.GetUpstreamNameservers(), " "))},
					{"name": "cache", "parameters": "30"},
					{"name": "loop"},
					{"name": "reload"},
					{"name": "loadbalance"},
				},
			},
		},
		// TODO(berkayoz): Adjust the rock to support a stricter security context
		// Below is the workaround to revert https://github.com/coredns/helm/pull/184/
		"securityContext": map[string]any{
			"allowPrivilegeEscalation": true,
			"readOnlyRootFilesystem":   false,
			"capabilities": map[string]any{
				"drop": []string{},
			},
		},
	}

	return values, nil
}

func (c *CoreDNS) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (c *CoreDNS) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (c *CoreDNS) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	snap := c.Snap()
	s := c.State()

	client, err := snap.KubernetesClient("")
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to retrieve the coredns service: %w", err)
	}

	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
			Kubelet: types.Kubelet{ClusterDNS: utils.Pointer(dnsIP)},
		}); err != nil {
			return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
	}

	// DNS IP has changed, notify node config controller
	// a.NotifyUpdateNodeConfigController()

	return ctrl.Result{}, nil
}

func (c *CoreDNS) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (c *CoreDNS) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (c *CoreDNS) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (c *CoreDNS) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
