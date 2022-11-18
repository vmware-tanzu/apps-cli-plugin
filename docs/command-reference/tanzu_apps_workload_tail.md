## tanzu apps workload tail

Watch workload related logs

### Synopsis

Stream logs for a workload until canceled. To cancel, press Ctl-c in
the shell or stop the process. As new workload pods are started, the logs
are displayed. To show historical logs use --since.

```
tanzu apps workload tail <name> [flags]
```

### Examples

```
tanzu apps workload tail my-workload
tanzu apps workload tail my-workload --since 1h
```

### Options

```
      --component name   workload component name (e.g. build)
  -h, --help             help for tail
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
      --since duration   time duration to start reading logs from (default 1m0s)
  -t, --timestamp        print timestamp for each log line
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

