package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	core "github.com/leeforge/core"
	corecore "github.com/leeforge/core/core"
	postmodule "leeforge-example-service/modules/post"
)

type App struct {
	echo   *echo.Echo
	logger *zap.Logger
	rt     core.Runtime
}

func NewApp() (*App, error) {
	return newApp(zap.NewNop(), core.RuntimeOptions{
		ConfigPath:           resolveConfigPath(),
		PluginRegistrar:      registerExamplePlugins,
		ExtraModuleFactories: []corecore.ModuleFactory{postmodule.NewPostModule},
	})
}

func NewAppForTest() (*App, error) {
	return newApp(zap.NewNop(), core.RuntimeOptions{
		ConfigPath:       resolveConfigPath(),
		ResourceProvider: noopResourceProvider{},
		SkipPlugins:      true,
		SkipMigrate:      true,
	})
}

func (a *App) Echo() *echo.Echo {
	return a.echo
}

func newApp(logger *zap.Logger, opts core.RuntimeOptions) (*App, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	rt, err := core.BuildRuntime(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return &App{
		echo:   buildEcho(rt),
		logger: logger,
		rt:     rt,
	}, nil
}

func buildEcho(rt core.Runtime) *echo.Echo {
	e := echo.New()
	e.Any("/*", echo.WrapHandler(rt.Handler()))
	return e
}

func resolveConfigPath() string {
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		return env
	}
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../configs"))
}

type noopResourceProvider struct{}

func (noopResourceProvider) Build(context.Context, core.ResourceInput) (*core.RuntimeResources, error) {
	return &core.RuntimeResources{}, nil
}
