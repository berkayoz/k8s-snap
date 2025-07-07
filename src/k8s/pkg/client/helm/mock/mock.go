package mock

import (
	"context"
	"errors"

	"github.com/canonical/k8s/pkg/client/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type MockApplyArguments struct {
	Context context.Context
	Chart   helm.InstallableChart
	State   helm.State
	Values  map[string]any
}

type MockGetArguments struct {
	Context     context.Context
	ReleaseName string
	Namespace   string
}

type MockInstallArguments struct {
	Context     context.Context
	ReleaseName string
	Namespace   string
	Chart       *chart.Chart
	Values      map[string]any
}

type MockUpgradeArguments struct {
	Context     context.Context
	ReleaseName string
	Namespace   string
	Chart       *chart.Chart
	Values      map[string]any
}

type MockUninstallArguments struct {
	Context     context.Context
	ReleaseName string
	Namespace   string
}

// Mock is a mock implementation of helm.Client.
type Mock struct {
	ApplyCalledWith []MockApplyArguments
	ApplyChanged    bool
	ApplyErr        error

	GetCalledWith []MockGetArguments
	GetRelease    *helm.Release
	GetErr        error

	InstallCalledWith []MockInstallArguments
	InstallRelease    *helm.Release
	InstallErr        error

	UpgradeCalledWith []MockUpgradeArguments
	UpgradeRelease    *helm.Release
	UpgradeErr        error

	UninstallCalledWith []MockUninstallArguments
	UninstallRelease    *helm.Release
	UninstallErr        error
}

// Apply implements helm.Client.
func (m *Mock) Apply(ctx context.Context, c helm.InstallableChart, desired helm.State, values map[string]any) (bool, error) {
	m.ApplyCalledWith = append(m.ApplyCalledWith, MockApplyArguments{Context: ctx, Chart: c, State: desired, Values: values})
	return m.ApplyChanged, m.ApplyErr
}

func (m *Mock) Get(ctx context.Context, releaseName string, namespace string) (*helm.Release, error) {
	m.GetCalledWith = append(m.GetCalledWith, MockGetArguments{Context: ctx, ReleaseName: releaseName, Namespace: namespace})
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.GetRelease, nil
}

func (m *Mock) Install(ctx context.Context, releaseName string, namespace string, chart *chart.Chart, values map[string]any) (*helm.Release, error) {
	m.InstallCalledWith = append(m.InstallCalledWith, MockInstallArguments{Context: ctx, ReleaseName: releaseName, Namespace: namespace, Chart: chart, Values: values})
	if m.InstallErr != nil {
		return nil, m.InstallErr
	}
	return m.InstallRelease, nil
}

func (m *Mock) Upgrade(ctx context.Context, releaseName string, namespace string, chart *chart.Chart, values map[string]any) (*helm.Release, error) {
	m.UpgradeCalledWith = append(m.UpgradeCalledWith, MockUpgradeArguments{Context: ctx, ReleaseName: releaseName, Namespace: namespace, Chart: chart, Values: values})
	if m.UpgradeErr != nil {
		return nil, m.UpgradeErr
	}
	return m.UpgradeRelease, nil
}

func (m *Mock) Uninstall(ctx context.Context, releaseName string, namespace string) (*helm.Release, error) {
	m.UninstallCalledWith = append(m.UninstallCalledWith, MockUninstallArguments{Context: ctx, ReleaseName: releaseName, Namespace: namespace})
	if m.UninstallErr != nil {
		return nil, m.UninstallErr
	}
	return m.UninstallRelease, nil
}

func (m *Mock) IsNotFound(err error) bool {
	// Check if the error is a driver.ErrReleaseNotFound or a not found error.
	if errors.Is(err, driver.ErrReleaseNotFound) {
		return true
	}

	return false
}

var _ helm.Client = &Mock{}
