package bootstrap

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	frameworkplugin "github.com/leeforge/framework/plugin"
	frameworkruntime "github.com/leeforge/framework/runtime"
	"github.com/leeforge/plugins/tenant"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	core "github.com/leeforge/core"
)

type noopDomainWriter struct{}

func (noopDomainWriter) ResolveDomain(context.Context, string, string) (*core.ResolvedDomain, error) {
	return nil, nil
}
func (noopDomainWriter) ResolveDomainByID(context.Context, uuid.UUID) (*core.ResolvedDomain, error) {
	return nil, nil
}
func (noopDomainWriter) CheckMembership(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}
func (noopDomainWriter) GetUserDefaultDomain(context.Context, uuid.UUID) (*core.ResolvedDomain, error) {
	return nil, nil
}
func (noopDomainWriter) GetDomainString(string, string) string { return "" }
func (noopDomainWriter) ListUserDomains(context.Context, uuid.UUID) ([]*core.UserDomainInfo, error) {
	return nil, nil
}
func (noopDomainWriter) EnsureDomain(context.Context, string, string, string) (*core.ResolvedDomain, error) {
	return nil, nil
}
func (noopDomainWriter) AddMembership(context.Context, uuid.UUID, uuid.UUID, string, bool) error {
	return nil
}
func (noopDomainWriter) RemoveMembership(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestRegisterExamplePlugins_RegistersTenantFactoryBeforeBootstrap(t *testing.T) {
	rt := frameworkruntime.NewRuntime(frameworkruntime.Config{
		Router: chi.NewRouter(),
		Logger: zap.NewNop(),
	})
	services := rt.Services()
	require.NotNil(t, services)
	require.NoError(t, services.Register("domain.service", core.DomainWriter(noopDomainWriter{})))

	require.NoError(t, registerExamplePlugins(rt, services, zap.NewNop()))
	require.True(t, services.Has(tenant.ServiceKeyTenantFactory))

	require.NoError(t, rt.Bootstrap(context.Background()))
	state, ok := rt.GetPluginState("tenant")
	require.True(t, ok)
	require.Equal(t, frameworkplugin.StateEnabled, state)
}
