// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package engine

import (
	"context"
	"github.com/google/wire"
	"github.com/windmilleng/tilt/internal/analytics"
	"github.com/windmilleng/tilt/internal/build"
	"github.com/windmilleng/tilt/internal/docker"
	"github.com/windmilleng/tilt/internal/dockercompose"
	"github.com/windmilleng/tilt/internal/dockerfile"
	"github.com/windmilleng/tilt/internal/k8s"
	"github.com/windmilleng/tilt/internal/logger"
	"github.com/windmilleng/tilt/internal/minikube"
	"github.com/windmilleng/tilt/internal/synclet"
	"github.com/windmilleng/wmclient/pkg/dirs"
)

// Injectors from wire.go:

func provideBuildAndDeployer(ctx context.Context, docker2 docker.Client, kClient k8s.Client, dir *dirs.WindmillDir, env k8s.Env, updateMode UpdateModeFlag, sCli synclet.SyncletClient, dcc dockercompose.DockerComposeClient, clock build.Clock, kp KINDPusher, analytics2 *analytics.TiltAnalytics) (BuildAndDeployer, error) {
	syncletManager := NewSyncletManagerForTests(kClient, sCli)
	runtime := k8s.ProvideContainerRuntime(ctx, kClient)
	engineUpdateMode, err := ProvideUpdateMode(updateMode, env, runtime)
	if err != nil {
		return nil, err
	}
	syncletBuildAndDeployer := NewSyncletBuildAndDeployer(syncletManager, kClient, engineUpdateMode)
	containerUpdater := build.NewContainerUpdater(docker2)
	localContainerBuildAndDeployer := NewLocalContainerBuildAndDeployer(containerUpdater, analytics2, env)
	labels := _wireLabelsValue
	dockerImageBuilder := build.NewDockerImageBuilder(docker2, labels)
	imageBuilder := build.DefaultImageBuilder(dockerImageBuilder)
	cacheBuilder := build.NewCacheBuilder(docker2)
	client := minikube.ProvideMinikubeClient()
	dockerEnv, err := docker.ProvideEnv(ctx, env, runtime, client)
	if err != nil {
		return nil, err
	}
	execCustomBuilder := build.NewExecCustomBuilder(docker2, dockerEnv, clock)
	imageBuildAndDeployer := NewImageBuildAndDeployer(imageBuilder, cacheBuilder, execCustomBuilder, kClient, env, analytics2, engineUpdateMode, clock, runtime, kp)
	engineImageAndCacheBuilder := NewImageAndCacheBuilder(imageBuilder, cacheBuilder, execCustomBuilder, engineUpdateMode)
	dockerComposeBuildAndDeployer := NewDockerComposeBuildAndDeployer(dcc, docker2, engineImageAndCacheBuilder, clock)
	buildOrder := DefaultBuildOrder(syncletBuildAndDeployer, localContainerBuildAndDeployer, imageBuildAndDeployer, dockerComposeBuildAndDeployer, env, engineUpdateMode, runtime)
	compositeBuildAndDeployer := NewCompositeBuildAndDeployer(buildOrder)
	return compositeBuildAndDeployer, nil
}

var (
	_wireLabelsValue = dockerfile.Labels{}
)

func provideImageBuildAndDeployer(ctx context.Context, docker2 docker.Client, kClient k8s.Client, env k8s.Env, dir *dirs.WindmillDir, clock build.Clock, kp KINDPusher, analytics2 *analytics.TiltAnalytics) (*ImageBuildAndDeployer, error) {
	labels := _wireLabelsValue
	dockerImageBuilder := build.NewDockerImageBuilder(docker2, labels)
	imageBuilder := build.DefaultImageBuilder(dockerImageBuilder)
	cacheBuilder := build.NewCacheBuilder(docker2)
	runtime := k8s.ProvideContainerRuntime(ctx, kClient)
	client := minikube.ProvideMinikubeClient()
	dockerEnv, err := docker.ProvideEnv(ctx, env, runtime, client)
	if err != nil {
		return nil, err
	}
	execCustomBuilder := build.NewExecCustomBuilder(docker2, dockerEnv, clock)
	updateModeFlag := _wireUpdateModeFlagValue
	updateMode, err := ProvideUpdateMode(updateModeFlag, env, runtime)
	if err != nil {
		return nil, err
	}
	imageBuildAndDeployer := NewImageBuildAndDeployer(imageBuilder, cacheBuilder, execCustomBuilder, kClient, env, analytics2, updateMode, clock, runtime, kp)
	return imageBuildAndDeployer, nil
}

var (
	_wireUpdateModeFlagValue = UpdateModeFlag(UpdateModeAuto)
)

func provideDockerComposeBuildAndDeployer(ctx context.Context, dcCli dockercompose.DockerComposeClient, dCli docker.Client, dir *dirs.WindmillDir) (*DockerComposeBuildAndDeployer, error) {
	labels := _wireLabelsValue
	dockerImageBuilder := build.NewDockerImageBuilder(dCli, labels)
	imageBuilder := build.DefaultImageBuilder(dockerImageBuilder)
	cacheBuilder := build.NewCacheBuilder(dCli)
	env := _wireEnvValue
	portForwarder := k8s.ProvidePortForwarder()
	clientConfig := k8s.ProvideClientConfig()
	namespace := k8s.ProvideConfigNamespace(clientConfig)
	config, err := k8s.ProvideKubeConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	kubeContext, err := k8s.ProvideKubeContext(config)
	if err != nil {
		return nil, err
	}
	int2 := provideKubectlLogLevelInfo()
	kubectlRunner := k8s.ProvideKubectlRunner(kubeContext, int2)
	client := k8s.ProvideK8sClient(ctx, env, portForwarder, namespace, kubectlRunner, clientConfig)
	runtime := k8s.ProvideContainerRuntime(ctx, client)
	minikubeClient := minikube.ProvideMinikubeClient()
	dockerEnv, err := docker.ProvideEnv(ctx, env, runtime, minikubeClient)
	if err != nil {
		return nil, err
	}
	clock := build.ProvideClock()
	execCustomBuilder := build.NewExecCustomBuilder(dCli, dockerEnv, clock)
	updateModeFlag := _wireEngineUpdateModeFlagValue
	updateMode, err := ProvideUpdateMode(updateModeFlag, env, runtime)
	if err != nil {
		return nil, err
	}
	engineImageAndCacheBuilder := NewImageAndCacheBuilder(imageBuilder, cacheBuilder, execCustomBuilder, updateMode)
	dockerComposeBuildAndDeployer := NewDockerComposeBuildAndDeployer(dcCli, dCli, engineImageAndCacheBuilder, clock)
	return dockerComposeBuildAndDeployer, nil
}

var (
	_wireEnvValue                  = k8s.Env(k8s.EnvNone)
	_wireEngineUpdateModeFlagValue = UpdateModeFlag(UpdateModeAuto)
)

// wire.go:

var DeployerBaseWireSet = wire.NewSet(wire.Value(dockerfile.Labels{}), wire.Value(UpperReducer), minikube.ProvideMinikubeClient, docker.ProvideEnv, build.DefaultImageBuilder, build.NewCacheBuilder, build.NewDockerImageBuilder, build.NewExecCustomBuilder, wire.Bind(new(build.CustomBuilder), new(build.ExecCustomBuilder)), NewImageBuildAndDeployer, build.NewContainerUpdater, NewSyncletBuildAndDeployer,
	NewLocalContainerBuildAndDeployer,
	NewDockerComposeBuildAndDeployer,
	NewImageAndCacheBuilder,
	DefaultBuildOrder, wire.Bind(new(BuildAndDeployer), new(CompositeBuildAndDeployer)), NewCompositeBuildAndDeployer,
	ProvideUpdateMode,
)

var DeployerWireSetTest = wire.NewSet(
	DeployerBaseWireSet,
	NewSyncletManagerForTests,
)

var DeployerWireSet = wire.NewSet(
	DeployerBaseWireSet,
	NewSyncletManager,
)

func provideKubectlLogLevelInfo() k8s.KubectlLogLevel {
	return k8s.KubectlLogLevel(logger.InfoLvl)
}
