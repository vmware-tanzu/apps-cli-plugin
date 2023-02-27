# Apps Plugin for the Tanzu CLI

[![CI](https://github.com/vmware-tanzu/apps-cli-plugin/actions/workflows/ci.yaml/badge.svg)](https://github.com/vmware-tanzu/apps-cli-plugin/actions/workflows/ci.yaml)
[![GoDoc](https://godoc.org/github.com/vmware-tanzu/apps-cli-plugin?status.svg)](https://godoc.org/github.com/vmware-tanzu/apps-cli-plugin)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmware-tanzu/apps-cli-plugin)](https://goreportcard.com/report/github.com/vmware-tanzu/apps-cli-plugin)
[![codecov](https://codecov.io/gh/vmware-tanzu/apps-cli-plugin/branch/main/graph/badge.svg)](https://codecov.io/gh/vmware-tanzu/apps-cli-plugin)


Apps plugin for [Tanzu CLI](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli#installation) provides the ability to create, view, update, and delete application workloads on any Kubernetes cluster that has [Cartographer](https://cartographer.sh/) installed. It also provides commands to list and view ClusterSupplychain resources.

### <a id='About'></a>About Workloads and ClusterSupplychain

With a [ClusterSupplyChain](https://cartographer.sh/docs/v0.4.0/reference/ClusterSupplyChain), app operators describe which “shape of applications” they deal with (via spec.selector), and what series of resources are responsible for creating an artifact that delivers it (via spec.resources).

[Workload](https://cartographer.sh/docs/v0.4.0/reference/workload/) allows the developer to pass information about the app to be delivered through the supply chain. 

For further information about Cartographer resources, read the [official docs](https://github.com/vmware-tanzu/cartographer)

## Getting Started
### Prerequisites
[Tanzu CLI](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli#installation) is required to use the `Apps` CLI plugin.

### From a pre-built distribution
Download the `tanzu-apps-plugin-<OS>-amd64-${VERSION}.tar.gz` from the most recent release listed on the [Apps Plugin for the Tanzu CLI releases](https://github.com/vmware-tanzu/apps-cli-plugin/releases) page.

#### macOS 
Download binary executable(`tanzu-apps-plugin-darwin-amd64-${VERSION}.tar.gz`) for CLI apps plugin. Following are the instructions for installing plugin version v0.7.0.

```
VERSION=v0.7.0
tar -xvf tanzu-apps-plugin-darwin-amd64-${VERSION}.tar.gz
tanzu plugin install apps --local ./tanzu-apps-plugin-darwin-amd64-${VERSION} --version ${VERSION}
```

#### Linux
Download binary executable(`tanzu-apps-plugin-linux-amd64-${VERSION}.tar.gz`) for CLI apps plugin. Following are the instructions for installing plugin version v0.7.0.

```
VERSION=v0.7.0
tar -xvf tanzu-apps-plugin-linux-amd64-${VERSION}.tar.gz
tanzu plugin install apps --local ./tanzu-apps-plugin-linux-amd64-${VERSION} --version ${VERSION}
```

#### Windows
Download binary executable(`tanzu-apps-plugin-windows-amd64-${VERSION}.tar.gz`) for CLI apps plugin. Unzip the file tanzu-apps-plugin-windows-amd64-${VERSION}.tar.gz. Following are the instructions for installing plugin version v0.7.0 

```
tanzu plugin install apps --local . --version v0.7.0
```

**NOTE**: --local should point to the directory which has discovery and distribution folder in it after unzipping.

### Build from CI artifact

If you're looking for the latest dev build, you may be able to find the artifact from a [recent CI build](https://github.com/vmware-tanzu/apps-cli-plugin/actions/workflows/ci.yaml?query=branch%3Amain+event%3Apush). Follow Github documentation to locate and download the artifact for [certain build](https://docs.github.com/en/actions/managing-workflow-runs/downloading-workflow-artifacts).

### Build from source

See the [development guide](./DEVELOPMENT.md) for instructions to build from source.

## Documentation

Detailed documentation for commands in Apps Plugin for the Tanzu CLI can be found in the [docs](./docs/README.md) folder of this repository. Documentation provides usage guide, information about working with workloads, and the full command reference.

## Community and Support

Join us!

If you have questions or want to get the latest project news, you can connect with us in the following ways:
- Checkout our github issue/ PR section

## Contributing

Pull Requests and feedback on issues are very welcome! See the [issue tracker](https://github.com/vmware-tanzu/apps-cli-plugin/issues) if you're unsure where to start, especially the Good first issue label, and also feel free to reach out to discuss.

If you are ready to jump in and test, add code, or help with documentation, please follow the instructions on our [Contribution Guidelines](./CONTRIBUTING.md) to get started and - at all times- follow our [Code of Conduct](./CODE_OF_CONDUCT.md)

## License

Apache 2.0. Refer to LICENSE for details.