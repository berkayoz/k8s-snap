package calico

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Calico struct {
	*component.BaseComponent
}

func NewCalico(s state.State, snap snap.Snap) *Calico {
	return &Calico{
		BaseComponent: component.NewBaseComponent(s, snap, "ck-network", "kube-system", ChartCalico),
	}
}

func (c *Calico) IsEnabled(config types.ClusterConfig) bool {
	return *config.Network.Enabled
}

func (c *Calico) Values(config types.ClusterConfig) (map[string]any, error) {
	cnf, err := internalConfig(config.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotations: %w", err)
	}

	podIpPools := []map[string]any{}
	ipv4PodCIDR, ipv6PodCIDR, err := utils.SplitCIDRStrings(config.Network.GetPodCIDR())
	if err != nil {
		return nil, fmt.Errorf("invalid pod cidr: %w", err)

	}
	if ipv4PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name":          "ipv4-ippool",
			"cidr":          ipv4PodCIDR,
			"encapsulation": cnf.encapsulationV4,
		})
	}
	if ipv6PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name":          "ipv6-ippool",
			"cidr":          ipv6PodCIDR,
			"encapsulation": cnf.encapsulationV6,
		})
	}

	serviceCIDRs := []string{}
	ipv4ServiceCIDR, ipv6ServiceCIDR, err := utils.SplitCIDRStrings(config.Network.GetServiceCIDR())
	if err != nil {
		return nil, fmt.Errorf("invalid service cidr: %w", err)
	}
	if ipv4ServiceCIDR != "" {
		serviceCIDRs = append(serviceCIDRs, ipv4ServiceCIDR)
	}
	if ipv6ServiceCIDR != "" {
		serviceCIDRs = append(serviceCIDRs, ipv6ServiceCIDR)
	}

	calicoNetworkValues := map[string]any{
		"ipPools": podIpPools,
	}

	if cnf.autodetectionV4 != nil {
		calicoNetworkValues["nodeAddressAutodetectionV4"] = cnf.autodetectionV4
	}

	if cnf.autodetectionV6 != nil {
		calicoNetworkValues["nodeAddressAutodetectionV6"] = cnf.autodetectionV6
	}

	values := map[string]any{
		"tigeraOperator": map[string]any{
			"registry": imageRepo,
			"image":    tigeraOperatorImage,
			"version":  tigeraOperatorVersion,
		},
		"calicoctl": map[string]any{
			"image": calicoCtlImage,
			"tag":   calicoCtlTag,
		},
		"installation": map[string]any{
			"calicoNetwork": calicoNetworkValues,
			"registry":      imageRepo,
		},
		"apiServer": map[string]any{
			"enabled": cnf.apiServerEnabled,
		},
		"serviceCIDRs": serviceCIDRs,
	}

	return values, nil
}

func (c *Calico) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (c *Calico) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (c *Calico) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (c *Calico) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (c *Calico) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (c *Calico) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (c *Calico) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
