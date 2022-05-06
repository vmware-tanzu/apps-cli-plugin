- [Local builds](#local-builds)
- [Command documentation](#command-documentation)
- [Unit testing](#unit-testing)
- [Acceptance testing](#acceptance-testing)
  - [Installing Cartographer](#installing-cartographer)
- [Debug Code Completion](#debug-code-completion)

---

## Local builds

Building the CLI plugin locally requires both go 1.18 and the [Tanzu CLI](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli#installation) with the builder plugin installed.

Install the Tanzu builder and test plugins:

```sh
tanzu init
```

```sh
tanzu plugin repo add -b tanzu-cli-admin-plugins -n admin -p artifacts-admin
```

```sh
tanzu plugin install builder
```

```sh
tanzu plugin install test
```

To build and install the apps plugin, run: (repeat this step any time you pull new source code to get the latest)

```sh
make install
```

## Command documentation

Command argument and flag documentation is available in the [docs](./docs/tanzu_apps.md) directory. CI will fail if this content is not kept up to date.

To regenerate the docs, run:

```sh
make docs
```

## Unit testing

All unit tests can be run on any machine with go 1.16+ installed.

```sh
make test
```

## Acceptance testing

In order to use the CLI, the runtime dependencies need to be installed into the target cluster.

Dependencies:
- Cartographer

### Installing Cartographer

See Cartographer install documentation [here](https://github.com/vmware-tanzu/cartographer#installation) 

## Debug Code Completion

To enable code completion follow the instructions in `tanzu help completion`.

To ease testing add the prefix `__complete` to your command e.g.
```sh
tanzu __complete apps workload tail spring-petclinic --component=
```
This will output the completion suggestions and the ShellCompDirective
Add logs using `cobra.CompDebugln(msg string, printToStdErr bool)` and set `BASH_COMP_DEBUG_FILE` env variable to a local file path to see logs without adding new entries to the suggestions
Make sure to remove any logs before sending a PR