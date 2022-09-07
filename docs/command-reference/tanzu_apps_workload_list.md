## tanzu apps workload list

Table listing of workloads

### Synopsis

List workloads in a namespace or across all namespaces.

```
tanzu apps workload list [flags]
```

### Examples

```
tanzu apps workload list
tanzu apps workload list --all-namespaces
```

### Options

```
  -A, --all-namespaces   use all kubernetes namespaces
      --app name         application name the workload is a part of
  -h, --help             help for list
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
  -o, --output string    output the Workloads formatted. Supported formats: "json", "yaml", "yml"
```

### Options inherited from parent commands

```
      --context name      name of the kubeconfig context to use (default is current-context defined by kubeconfig)
      --kubeconfig file   kubeconfig file (default is $HOME/.kube/config)
      --no-color          deactivate color, bold, and emoji output
  -v, --verbose int32     number for the log level verbosity (default 1)
```

### SEE ALSO

* [tanzu apps workload](tanzu_apps_workload.md)	 - Workload lifecycle management

