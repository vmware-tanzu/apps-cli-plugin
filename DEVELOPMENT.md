# Development

This doc explains how to set up a development environment so you can get started
[contributing](CONTRIBUTING.md) to `tanzu apps plugin`. Also
take a look at:

- [Prerequisites](#Prerequisites)
- [Building](#building)
- [How to add and run tests](#testing)
- [Iterating](#iterating)
- [Uninstall](#uninstalling)

## Prerequisites

Follow the instructions below to set up your development environment. Once you
meet these requirements, you can make changes and
[deploy your own version of apps plugin](#starting-apps-plugin)!

Before submitting a PR, see also [CONTRIBUTING.md](./CONTRIBUTING.md).

### Install requirements

You must install these tools:

- [`go`](https://golang.org/doc/install): for compiling the plugin as well as other dependencies - 1.20+
- [`tanzu CLI`](https://github.com/vmware-tanzu/tanzu-framework/blob/main/docs/cli/getting-started.md#install-the-latest-release-of-tanzu-cli): for adding the plugin
- [`git`](https://help.github.com/articles/set-up-git/): for source control

### Setup your development environment
To check out this repository:

1. Create your own
   [fork of this repo](https://docs.github.com/en/get-started/quickstart/fork-a-repo)
1. Clone it on your machine:

```shell
git clone git@github.com:${YOUR_GITHUB_USERNAME}/apps-cli-plugin.git
cd apps-cli-plugin
git remote add upstream https://github.com/vmware-tanzu/apps-cli-plugin.git
git remote set-url --push upstream no_push
```

_Adding the `upstream` remote sets you up nicely for regularly
[syncing your fork](https://help.github.com/articles/syncing-a-fork/)._

Once you reach this point you are ready to do a full build and install as
described below.

## Building
Once you've [setup your development environment](#prerequisites), let's build
`apps plugin`. [go mod](https://github.com/golang/go/wiki/Modules#quick-start) is used and
required for dependencies.

**Install builder plugin**

Install the tanzu builder and test plugins. Set variable `OS` to `linux`, `darwin` or `windows` depending on your OS:

```sh
TANZU_VERSION=$(cat TANZU_VERSION)
TANZU_HOME=$HOME/tanzu
curl -Lo admin-plugins.tar.gz https://github.com/vmware-tanzu/tanzu-framework/releases/download/${TANZU_VERSION}/tanzu-framework-plugins-admin-${OS}-amd64.tar.gz
tar -xzf admin-plugins.tar.gz -C ${TANZU_HOME}
tanzu plugin install builder --local ${TANZU_HOME}/admin-plugins
tanzu plugin install test --local ${TANZU_HOME}/admin-plugins
```

**Building Apps CLI Plug-in**

To build and install the apps plugin, run: (repeat this step any time you pull new source code to get the latest)

```sh
make install
```
Note : For machines with arm64 processor run `make install GOHOSTARCH=amd64`

Verify installed plugins

```
tanzu plugin list
```

*Publishing Apps CLI Plug-in artifact to any registry*

Since Tanzu CLI supports installing plugins from an OCI repository, the generated artifact for Apps CLI Plugin release can be uploaded to be publicly accessible in any repository like, for example, GitHub Container Registry.

To publish the artifact, there are some steps that need to be done before running `make publish-oci`.

1. Login to your registry.

2. Export discovery and distribution environment variables to be used in the `make` command.
  ```sh
  export DISCOVERY_REPO=<your-registry>/USERNAME/apps-cli-plugin/discovery
  export DISTRIBUTION_REPO=<your-registry>/USERNAME/apps-cli-plugin/distribution
  ```

3. Clear the `artifacts` directory and then run `make build` to fill it with the corresponding binaries.
  If this folder has more than one Apps Plug-in version, all of them will be uploaded in the next step.

4. Run `make publish-oci` command to push to the registry.
  ```sh
  make publish-oci
  ```
  The output of this last step is similar to the following:
  ```bash
  2023-03-09T11:59:51-05:00 [ℹ]  Publishing plugin: apps to <your-registry>/USERNAME/apps-cli-plugin/distribution/apps-darwin-amd64:v0.11.1-dev-4ea1dafb
  2023-03-09T12:00:22-05:00 [✔]  Successfully published plugin: apps
  2023-03-09T12:00:22-05:00 [ℹ]  Publishing discovery image to: <your-registry>/USERNAME/apps-cli-plugin/discovery
  2023-03-09T12:00:29-05:00 [✔]  Successfully published CLIPlugin resources to discovery image: <your-registry>/USERNAME/apps-cli-plugin/discovery
  ```

## Testing
### Unit testing

All unit tests can be run on any machine with go 1.17+ installed.

```sh
make test
```

### Acceptance testing
In order to use the CLI, the runtime dependencies need to be installed on the target Kubernetes cluster.

```sh
BUNDLE=${REGISTRY_URL/REPO:tag}
make integration-test
```

Note: `BUNDLE` env var is used by e2e tests to push source code bundle. This image is not deleted after the test.

### Add test

Any contribtuions for bug fix or feature requests will require unit and integration tests as part of the PR.

### Cluster requirement
- Create a Kubernetes cluster
- Deploy [Cartographer](https://github.com/vmware-tanzu/cartographer#installation)

## Iterating
As you make changes to the code-base, there are several special cases to be aware
of:

- **If you change/add Cartographer APIs**, then you must run
  `make prepare` to update generated code. You might also require creating fakes via [dies](https://pkg.go.dev/dies.dev/diegen) for easier testing.

- **If you change/add dependencies** (including adding an external dependency),
  then you must run `make prepare` command. In some cases, if newer dependencies are required, you need to run "go get" manually.

- **If you change/add any help text for flags on any command**, then you must run `make docs` command. This will generate command documentation. CI will fail if this content is not kept up to date.

- **If you change/add code completion**, then follow the steps to ease testing of this feature. Add the prefix `__complete` to your command e.g.

  ```sh
  tanzu __complete apps workload tail spring-petclinic --component=
  ```
  
  This will output the completion suggestions and the ShellCompDirective
  Add logs using `cobra.CompDebugln(msg string, printToStdErr bool)` and set `BASH_COMP_DEBUG_FILE` env variable to a local file path to see logs without adding new entries to the suggestions. Make sure to remove any logs before sending a PR

Once the codegen, dependency is correct, reinstalling the
plugin is simply:

```shell
make install
```

Or you can [uninstall](./DEVELOPMENT.md#Uninstalling) and
[completely redeploy `apps plugin`](./DEVELOPMENT.md#starting-apps-plugin).

## Uninstalling
You can delete apps plugin with:

```sh
tanzu plugin delete apps
```
