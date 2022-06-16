# Usage

## <a id='changing-clusters'></a> Changing Clusters

The Apps CLI plugin uses the default context that is set in the kubeconfig file to connect to the cluster. To switch clusters use kubectl to set the [default context](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/).

## <a id='yaml-files'></a>Working with YAML Files

In many cases the lifecycle of workloads can be managed through CLI commands and their flags alone but there might be cases where it is desired to manage a workload using a `yaml` file and the Apps plugin supports this use case.

The plugin has been designed to manage one workload at a time. As such when a workload is being managed via a `yaml` file, that file must contain a single workload definition. In addition plugin commands support only one file per command.

For example, a valid file would be like this:

```yaml
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: spring-petclinic
  labels:
    app.kubernetes.io/part-of: spring-petclinic
    apps.tanzu.vmware.com/workload-type: java
spec:
  source:
    git:
      url: https://github.com/sample-accelerators/spring-petclinic
      ref:
        tag: tap-1.1
```

Then, to create a workload from this file, run:

```console
tanzu apps workload create -f my-workload-file.yaml
```

Another way to create a workload from `yaml` is passing the definition through `stdin`. For example, run:

```console
tanzu apps workload create -f - --yes
```

The console will remain waiting for some input, and the content with a valid `yaml` definition for a workload can be either written or pasted, then press `ctrl`+D three times to start workload creation. This can also be done with `workload update` and `workload apply` commands.

**Note**: to pass workload through `stdin`, `--yes` flag is needed. If not used, command will fail.

## <a id='autocompletion'></a> Autocompletion

To enable command autocompletion, the Tanzu CLI offers the `tanzu completion` command.

Add the following command to the shell config file according to the current setup:

### Bash

```bash
tanzu completion bash >  $HOME/.tanzu/completion.bash.inc
```

### Zsh

```sh
echo "autoload -U compinit; compinit" >> ~/.zshrc
tanzu completion zsh > "${fpath[1]}/_tanzu"
```
