# Apps Plugin for the Tanzu CLI

![CI](https://github.com/vmware-tanzu/apps-cli-plugin/workflows/CI/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/vmware-tanzu/apps-cli-plugin/branch/main/graph/badge.svg?token=LYP76S1UI4)](https://codecov.io/gh/vmware-tanzu/apps-cli-plugin)

## Overview

This Tanzu CLI plugin provides the ability to create, view, update, and delete application workloads on any Kubernetes cluster that has the Tanzu Application Platform components installed.

### <a id='About'></a>About Workloads

A Workload enables developers to choose application specifications such as the location of their repository, environment variables, and service claims.

Tanzu Application Platform can support range of possible workloads including, a serverless process that spins up on demand, a constellation of microservices that work together as a logical application, or a small hello-world test app.

## Documentation

The [documentation](docs) provides a usage guide, information about working with workloads, and the full command reference.

## Getting Started

### Prerequisites

The [Tanzu CLI](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli#installation) is required to use the `Apps` CLI plugin.

### From a pre-built distribution

Download the `tanzu-apps-plugin.tar.gz` from the most recent release listed on the [Apps Plugin for the Tanzu CLI releases](https://github.com/vmware-tanzu/apps-cli-plugin/releases) page. If you're looking for the latest dev build, you may be able to find the artifact from a [recent CI build](https://github.com/vmware-tanzu/apps-cli-plugin/actions/workflows/ci.yaml?query=branch%3Amain+event%3Apush).

Extract the archive to a local directory:

```sh
tar -zxvf tanzu-apps-plugin.tar.gz
```

Install the apps plugin:

```sh
APPS_VERSION=v0.0.0-dev
tanzu plugin install apps --local ./artifacts --version $APPS_VERSION
```

### Build from source

See the [development guide](./DEVELOPMENT.md#local-builds).

## Contributing

Thanks for taking the time to join our community and start contributing! We welcome pull requests. Feel free to dig through the [issues](https://github.com/vmware-tanzu/apps-cli-plugin/issues) and jump in.


## <a id='feedback'></a>Feedback

For questions or feeadback you can reach us at tanzu-application-platform-beta@vmware.com
