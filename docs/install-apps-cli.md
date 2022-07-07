# Install Apps CLI plug-in

- [Install Apps CLI plug-in](#install-apps-cli-plug-in)
  - [Prerequisites](#prerequisites)
  - [Install From Release](#install-from-release)
  - [Install From Source](#install-from-source)
  - [Install From Tanzu Community Edition](#install-from-tanzu-community-edition)
  - [Install From Tanzu Network](#install-from-tanzu-network)

This document describes how to install the Apps CLI plug-in.


## Prerequisites

Before you install the Apps CLI plug-in:

- Follow the instructions to [Install or update the Tanzu CLI and plug-ins](../../install-tanzu-cli.md#cli-and-plugin).


## Install From Release

The latest release can be found in the [repository release page](https://github.com/vmware-tanzu/apps-cli-plugin/releases/). Each of these releases has the *Assets* section where the packages for each *system-architecture* are placed.

To install the Apps CLI plug-in:

Download binary executable `tanzu-apps-plugin-{OS_ARCH}-{version}.tar.gz`
Run the following commands(for example for macOS and plugin version v0.7.0)

```bash
tar -xvf tanzu-apps-plugin-darwin-amd64-v0.7.0.tar.gz
tanzu plugin install apps --local ./tanzu-apps-plugin-darwin-amd64-v0.7.0 --version v0.7.0
```

## Install From Source

The source code provides a way to install the Apps CLI plug-in by using a *Makefile* with target `install`, find more information in the [DEVELOPMENT](../DEVELOPMENT.md#local-builds) guide

## Install From Tanzu Community Edition

The tanzu community edition distributes the Apps CLI plug-in as a package build in the tanzu CLI, please refer to the installation instructions for the [Tanzu CLI](https://tanzucommunityedition.io/docs/edge/cli-installation/)

## Install From Tanzu Network

Find the complete guide at [Tanzu Documentation](https://docs.vmware.com/en/VMware-Tanzu-Application-Platform/1.1/tap/GUID-install-tanzu-cli.html#install-or-update-the-tanzu-cli-and-plugins-4)

To install the Apps CLI plug-in:

1. From the `$HOME/tanzu` directory, run:

    ```console
    tanzu plugin install --local ./cli apps
    ```

    This will use the downloaded tarball used when installing tanzu cli as the apps plugin is distributed within.

2. To verify that the CLI is installed correctly, run:

    ```console
    tanzu apps version
    ```

    A version should be displayed in the output.

    If the following error is displayed during installation:

    ```console
    Error: could not find plug-in "apps" in any known repositories

    ✖  could not find plug-in "apps" in any known repositories
    ```

    Verify that there is an `apps` entry in the `cli/manifest.yaml` file. It should look like this:

    ```yaml
    plugins:
    ...
        - name: apps
        description: Applications on Kubernetes
        versions: []
    ```
