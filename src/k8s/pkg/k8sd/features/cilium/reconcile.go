package cilium

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/features/component"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Cilium struct {
	*component.BaseComponent
}

func NewCilium(s state.State, snap snap.Snap) *Cilium {
	// TODO(berkayoz): Add rollout pods config values to the Cilium values.
	return &Cilium{
		BaseComponent: component.NewBaseComponent(s, snap, "ck-network", "kube-system", ChartCilium),
	}
}

func (c *Cilium) IsEnabled(config types.ClusterConfig) bool {
	return *config.Network.Enabled
}

func (c *Cilium) Values(ctx context.Context, config types.ClusterConfig) (map[string]any, error) {
	s := c.State()
	snap := c.Snap()

	cilConf, err := internalConfig(config.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotations: %w", err)
	}

	mcl, err := s.Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := mcl.GetClusterMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	// TODO(berkayoz): Still not sure if this makes sense to do here,
	// as this will only check for a single node's localhost address.
	localhostAddress, err := utils.DetermineLocalhostAddress(clusterMembers)
	if err != nil {
		return nil, fmt.Errorf("failed to determine localhost address: %w", err)
	}

	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return nil, fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
	}

	defaultCidr, err := utils.FindCIDRForIP(nodeIP)
	if err != nil {
		return nil, fmt.Errorf("failed to find cidr of default interface: %w", err)
	}

	ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(config.Network.GetPodCIDR())
	if err != nil {
		return nil, fmt.Errorf("invalid kube-proxy --cluster-cidr value: %w", err)
	}

	ciliumNodePortValues := map[string]any{
		"enabled": true,
		// kube-proxy also binds to the same port for health checks so we need to disable it
		"enableHealthCheck": false,
	}

	if cilConf.directRoutingDevice != "" {
		ciliumNodePortValues["directRoutingDevice"] = cilConf.directRoutingDevice
	}

	bpfValues := map[string]any{}
	if cilConf.vlanBPFBypass != nil {
		bpfValues["vlanBypass"] = cilConf.vlanBPFBypass
	}

	values := map[string]any{
		"bpf": bpfValues,
		"image": map[string]any{
			"repository": ciliumAgentImageRepo,
			"tag":        CiliumAgentImageTag,
			"useDigest":  false,
		},
		"socketLB": map[string]any{
			"enabled": true,
		},
		"cni": map[string]any{
			"confPath":     "/etc/cni/net.d",
			"binPath":      "/opt/cni/bin",
			"exclusive":    cilConf.cniExclusive,
			"chainingMode": "portmap",
		},
		"sctp": map[string]any{
			"enabled": cilConf.sctpEnabled,
		},
		"operator": map[string]any{
			"replicas": 1,
			"image": map[string]any{
				"repository": ciliumOperatorImageRepo,
				"tag":        ciliumOperatorImageTag,
				"useDigest":  false,
			},
		},
		"ipv4": map[string]any{
			"enabled": ipv4CIDR != "",
		},
		"ipv6": map[string]any{
			"enabled": ipv6CIDR != "",
		},
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": ipv4CIDR,
				"clusterPoolIPv6PodCIDRList": ipv6CIDR,
			},
		},
		"envoy": map[string]any{
			"enabled": false, // 1.16+ installs envoy as a standalone daemonset by default if not explicitly disabled
		},
		// https://docs.cilium.io/en/v1.15/network/kubernetes/kubeproxy-free/#kube-proxy-hybrid-modes
		"nodePort":                 ciliumNodePortValues,
		"disableEnvoyVersionCheck": true,
		// socketLB requires an endpoint to the apiserver that's not managed by the kube-proxy
		// so we point to the localhost:secureport to talk to either the kube-apiserver or the kube-apiserver-proxy
		"k8sServiceHost": strings.Trim(localhostAddress, "[]"), // Cilium already adds the brackets for ipv6 addresses, so we need to remove them
		"k8sServicePort": config.APIServer.GetSecurePort(),
		// This flag enables the runtime device detection which is set to true by default in Cilium 1.16+
		"enableRuntimeDeviceDetection": true,
		"sessionAffinity":              true,
		"loadBalancer": map[string]any{
			"protocolDifferentiation": map[string]any{
				"enabled": true,
			},
		},
		"tunnelPort": cilConf.tunnelPort,
	}

	// If we are deploying with IPv6 only, we need to set the routing mode to native
	if ipv4CIDR == "" && ipv6CIDR != "" {
		values["routingMode"] = "native"
		values["ipv6NativeRoutingCIDR"] = defaultCidr
		values["autoDirectNodeRoutes"] = true
	}

	if cilConf.devices != "" {
		values["devices"] = cilConf.devices
	}

	if snap.Strict() {
		values["bpf"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			// TODO(berkayoz): We should expect a default mount point for the bpf filesystem
			// instead of trying to determine it from a single node. Update docs when strict mode is enabled.
			// "root": bpfMnt,
		}
		values["cgroup"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			// "hostRoot": cgrMnt,
		}
	}

	// Ingress configuration
	if config.Ingress.GetEnabled() {
		values["ingressController"] = map[string]any{
			"enabled":                true,
			"loadBalancerMode":       "shared",
			"defaultSecretNamespace": "kube-system",
			"defaultSecretName":      config.Ingress.GetDefaultTLSSecret(),
			"enableProxyProtocol":    config.Ingress.GetEnableProxyProtocol(),
		}
	} else {
		values["ingressController"] = map[string]any{
			"enabled":                false,
			"loadBalancerMode":       "",
			"defaultSecretNamespace": "",
			"defaultSecretName":      "",
			"enableProxyProtocol":    false,
		}
	}

	// Gateway API configuration
	if config.Gateway.GetEnabled() {
		values["gatewayAPI"] = map[string]any{
			"enabled": true,
			"gatewayClass": map[string]any{
				// This needs to be string, not bool, as the helm chart uses a string
				// Due to the values of 'auto', 'true' and 'false'
				"create": "false",
			},
		}
	} else {
		values["gatewayAPI"] = map[string]any{
			"enabled": false,
		}
	}

	return values, nil
}

func (c *Cilium) PreReconcile(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-reconcile actions needed
	return ctrl.Result{}, nil
}

func (c *Cilium) PreInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-install actions needed
	return ctrl.Result{}, nil
}

func (c *Cilium) PostInstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-install actions needed
	return ctrl.Result{}, nil
}

func (c *Cilium) PreUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-uninstall actions needed
	return ctrl.Result{}, nil
}

func (c *Cilium) PostUninstall(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-uninstall actions needed
	return ctrl.Result{}, nil
}

func (c *Cilium) PreUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No pre-upgrade actions needed
	return ctrl.Result{}, nil
}

func (c *Cilium) PostUpgrade(ctx context.Context, config types.ClusterConfig) (ctrl.Result, error) {
	// No post-upgrade actions needed
	return ctrl.Result{}, nil
}
