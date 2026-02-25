package bootstrap

import (
	"fmt"

	coreent "github.com/leeforge/core/server/ent"
	frameplugin "github.com/leeforge/framework/plugin"
	frameworkruntime "github.com/leeforge/framework/runtime"
	ou "github.com/leeforge/plugins/ou"
	oufactory "github.com/leeforge/plugins/ou/factory"
	tenant "github.com/leeforge/plugins/tenant"
	tenantfactory "github.com/leeforge/plugins/tenant/factory"
	"go.uber.org/zap"
)

func registerExamplePlugins(
	rt *frameworkruntime.Runtime,
	services *frameplugin.ServiceRegistry,
	logger *zap.Logger,
) error {
	if rt == nil {
		return fmt.Errorf("plugin runtime is nil")
	}
	if services == nil {
		return fmt.Errorf("plugin service registry is nil")
	}

	// Resolve the ent client registered by core bootstrapPlugins.
	client, _ := frameplugin.Resolve[*coreent.Client](services, "core.ent.client")

	if !services.Has(tenant.ServiceKeyTenantFactory) {
		if err := services.Register(tenant.ServiceKeyTenantFactory, tenantfactory.NewEntFactory(client)); err != nil {
			return fmt.Errorf("register tenant service factory: %w", err)
		}
	}

	if err := rt.Register(&tenant.TenantPlugin{}); err != nil {
		return fmt.Errorf("register plugin tenant: %w", err)
	}

	// --- OU Plugin ---
	if !services.Has(ou.ServiceKeyOUFactory) {
		if err := services.Register(ou.ServiceKeyOUFactory, oufactory.NewEntFactory(client)); err != nil {
			return fmt.Errorf("register ou service factory: %w", err)
		}
	}

	if err := rt.Register(&ou.OUPlugin{}); err != nil {
		return fmt.Errorf("register plugin ou: %w", err)
	}

	if logger != nil {
		logger.Info("plugins registered via examples registrar",
			zap.Strings("plugins", []string{"tenant", "ou"}),
		)
	}

	return nil
}
