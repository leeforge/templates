package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	core "github.com/leeforge/core"
	"github.com/leeforge/core/host"
	postmodule "leeforge-example-service/modules/post"
)

type App struct {
	engine *gin.Engine
	logger *zap.Logger
	rt     core.Runtime
}

func NewApp() (*App, error) {
	return newApp(zap.NewNop(), core.RuntimeOptions{
		ConfigPath:      resolveConfigPath(),
		PluginRegistrar: registerExamplePlugins,
		Modules: []host.ModuleBootstrapper{
			postmodule.NewPostModuleBootstrapper,
		},
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

func (a *App) Engine() *gin.Engine {
	return a.engine
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
		engine: buildGinEngine(rt),
		logger: logger,
		rt:     rt,
	}, nil
}

func buildGinEngine(rt core.Runtime) *gin.Engine {
	engine := gin.Default()
	engine.Any("/*any", gin.WrapH(rt.Handler()))
	return engine
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
