# <a id='how-to-guides'>How-to-guides

<!-- Intro to this point -->

<!-- ## <a id='service-account-usage'> --service-account usage -->

## <a id='custom-registry'> Custom registry credentials

The Apps CLI plug-in allows users to push images to their private registry by setting some flags when creating workloads from local source code.

A user can either trust a custom certificate on their system or pass the path to the certificate via flags.

To pass the certificate via flags the user may need to specify:

- `--registry-ca-cert`, which refers to the path of the self-signed certificate needed for the custom/private registry. This is also populated with a default value through the environment variable `TANZU_APPS_REGISTRY_CA_CERT`.
- `--registry-password` which is used when the registry requires credentials to push. The value of this flag can also be specified through `TANZU_APPS_REGISTRY_PASSWORD`.
- `--registry-username` usually used with `--registry-password` to set the registry credentials. It can also be provided as the environment variable `TANZU_APPS_REGISTRY_USERNAME`. 
- `--registry-token` which is set when the registry authentication is done via token. The value of this flag can also be taken from `TANZU_APPS_REGISTRY_TOKEN` environment variable.

For example:

```bash
tanzu apps workload apply my-workload --local-path path/to/my/repo -s registry.url.nip.io/my-package/my-image --type web --registry-ca-cert path/to/cacert/mycert.nip.io.crt --registry-username my-username --registry-password my-password
❓ Publish source in "path/to/my/repo" to "registry.url.nip.io/my-package/my-image"? It may be visible to others who can pull images from that repository [Yn]: y
Publishing source in "path/to/my/repo" to "registry.url.nip.io/my-package/my-image"...
37.53 kB / 37.53 kB [-----------------------------------------------------------------------------------] 100.00% 57.67 kB p/s
📥 Published source

🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    image: registry.url.nip.io/my-package/my-image:latest@sha256:caeb7e3a0e3ae0659f74d01095b6fdfe0d3c4a12856a15ac67ad6cd3b9e43648
❓ Do you want to create this workload? [yN]:
```

Also, the same command could be run as:

```bash
export TANZU_APPS_REGISTRY_CA_CERT=path/to/cacert/mycert.nip.io.crt
export TANZU_APPS_REGISTRY_PASSWORD=my-username
export TANZU_APPS_REGISTRY_USERNAME=my-password

tanzu apps workload apply my-workload --local-path path/to/my/repo -s registry.url.nip.io/my-package/my-image
```

Using environment variables provides the added convenience of not having to enter these flag values repeatedly in the event that multiple workloads must be created, with references to the same registry, during a given terminal session.
<!-- ## <a id='control-auto-scale'> Using label annotation to control auto-scale -->

<!-- ## <a id='limit-request'>limit/request-cpu and limit/request-memory -->

<!-- ## <a id='tail-usage'> --tail usage -->

## <a id='live-updated-debug'> --live-update and --debug

`--live-update` deploys the workload with configuration which enables local source code changes to be reflected on the running workload within seconds after the source code changes have been saved. This can be particularly valuable when iterating on features which require the workload to be deployed and running to validate.

Live update is ideally situated for executing from within one of our supported IDE extensions, but it can also be utilized independently as shown in the following Spring Boot application example:

**Prerequisites:** [Tilt](https://docs.tilt.dev/install.html) must be installed on the client

1. Clone the repository by running:

   ```console
   git clone https://github.com/vmware-tanzu/application-accelerator-samples
   ```

1. Change into the `tanzu-java-web-app` directory.
1. In `Tiltfile`, first change the `SOURCE_IMAGE` variable to use your registry and project.
1. At the very end of the file add:

   ```bash
   allow_k8s_contexts('your-cluster-name')
   ```

1. Inside the directory, run:

   ```bash
   tanzu apps workload apply tanzu-java-web-app --live-update --local-path . -s
   gcr.io/my-project/tanzu-java-web-app-live-update -y
   ```

   Expected output:

   ```bash
   The files and directories listed in the .tanzuignore file are being excluded from the uploaded source code.
   Publishing source in "." to "gcr.io/my-project/tanzu-java-web-app-live-update"...
   📥 Published source
   
   🔎 Create workload:
       1 + |---
       2 + |apiVersion: carto.run/v1alpha1
       3 + |kind: Workload
       4 + |metadata:
       5 + |  name: tanzu-java-web-app
       6 + |  namespace: default
       7 + |spec:
       8 + |  params:
       9 + |  - name: live-update
      10 + |    value: "true"
      11 + |  source:
      12 + |    image: gcr.io/my-project/tanzu-java-web-app-live-update:latest@sha256:3c9fd738492a23ac532a709301fcf0c9aa2a8761b2b9347bdbab52ce9404264b
   👍 Created workload "tanzu-java-web-app"

   To see logs:   "tanzu apps workload tail tanzu-java-web-app --timestamp --since 1h"
   To get status: "tanzu apps workload get tanzu-java-web-app"

   ```

1. Run Tilt to deploy the workload.

    ```bash
    tilt up

    Tilt started on http://localhost:10350/
    v0.23.6, built 2022-01-14

    (space) to open the browser
    (s) to stream logs (--stream=true)
    (t) to open legacy terminal mode (--legacy=true)
    (ctrl-c) to exit
    Tilt started on http://localhost:10350/
    v0.23.6, built 2022-01-14

    Initial Build • (Tiltfile)
    Loading Tiltfile at: /path/to/repo/tanzu-java-web-app/Tiltfile
    Successfully loaded Tiltfile (1.500809ms)
    tanzu-java-w… │
    tanzu-java-w… │ Initial Build • tanzu-java-web-app
    tanzu-java-w… │ WARNING: Live Update failed with unexpected error:
    tanzu-java-w… │   Cannot extract live updates on this build graph structure
    tanzu-java-w… │ Falling back to a full image build + deploy
    tanzu-java-w… │ STEP 1/1 — Deploying
    tanzu-java-w… │      Objects applied to cluster:
    tanzu-java-w… │        → tanzu-java-web-app:workload
    tanzu-java-w… │
    tanzu-java-w… │      Step 1 - 8.87s (Deploying)
    tanzu-java-w… │      DONE IN: 8.87s
    tanzu-java-w… │
    tanzu-java-w… │
    tanzu-java-w… │ Tracking new pod rollout (tanzu-java-web-app-build-1-build-pod):
    tanzu-java-w… │      ┊ Scheduled       - (…) Pending
    tanzu-java-w… │      ┊ Initialized     - (…) Pending
    tanzu-java-w… │      ┊ Ready           - (…) Pending
    ...
    ```

<!-- add info regarding debug flag -->

## <a id='export-usage'> --export usage

When using the `--export` flag, which can be specified as the shorthand `-e`, user can retrieve the workload definition with all the extraneous, cluster-specific, properties/values removed (e.g. the status and metadata fields like `creationTimestamp`) so that the workload definition could be saved and applied to a different environment without having to make significant edits.

This means that the workload definition includes only the fields that were specified by the developer that created it (`--export` preserves the essence of the developer's intent for portability).

For example, if user creates a workload with:

```bash
tanzu apps workload apply rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" -t web
```

When querying the workload with `--export`, the default export format is yaml as follows:

```bash
# with yaml format
    tanzu apps workload get rmq-sample-app --export
    ---
    apiVersion: carto.run/v1alpha1
    kind: Workload
    metadata:
    labels:
        apps.tanzu.vmware.com/workload-type: web
    name: rmq-sample-app
    namespace: default
    spec:
    serviceClaims:
    - name: rmq
        ref:
        apiVersion: rabbitmq.com/v1beta1
        kind: RabbitmqCluster
        name: example-rabbitmq-cluster-1
    source:
        git:
        ref:
            branch: main
        url: https://github.com/jhvhs/rabbitmq-sample

# with json format
    tanzu apps workload get rmq-sample-app --export --output json
    {
        "apiVersion": "carto.run/v1alpha1",
        "kind": "Workload",
        "metadata": {
            "labels": {
                "apps.tanzu.vmware.com/workload-type": "web"
            },
            "name": "rmq-sample-app",
            "namespace": "default"
        },
        "spec": {
            "serviceClaims": [
                {
                    "name": "rmq",
                    "ref": {
                        "apiVersion": "rabbitmq.com/v1beta1",
                        "kind": "RabbitmqCluster",
                        "name": "example-rabbitmq-cluster-1"
                    }
                }
            ],
            "source": {
                "git": {
                    "ref": {
                        "branch": "main"
                    },
                    "url": "https://github.com/jhvhs/rabbitmq-sample"
                }
            }
        }
    }
```

If it's desired to retrieve the workload including all the cluster-specifics (with its status and all its fields) the `--output` flag should be provided.

As shown before, this flag can also be used alongside `--export` to set the export format (as `json`, `yaml` or `yml`).

```bash
# with json format
tanzu apps workload get rmq-sample-app --output json # can also be used as tanzu apps workload get rmq-sample-app -ojson
    {
        "kind": "Workload",
        "apiVersion": "carto.run/v1alpha1",
        "metadata": {
            "name": "rmq-sample-app",
            "namespace": "default",
            "uid": "3619ff6d-9e73-473a-9112-891a6d8aee9e",
            "resourceVersion": "11657434",
            "generation": 2,
            "creationTimestamp": "2022-11-28T05:10:32Z",
            "labels": {
                "apps.tanzu.vmware.com/workload-type": "web"
            },
            "managedFields": [
                {
                    "manager": "v0.10.0+dev-002cc44e",
                    "operation": "Update",
                    "apiVersion": "carto.run/v1alpha1",
                    "time": "2022-11-28T05:10:32Z",
                    "fieldsType": "FieldsV1",
                    "fieldsV1": {
                        "f:metadata": {
                            "f:labels": {
                                ".": {},
                                "f:apps.tanzu.vmware.com/workload-type": {}
                            }
                        },
                        ...
                    }
                },
                ...
            ]
        },
        ...
            "status": {
            "observedGeneration": 2,
            "conditions": [
                {
                    "type": "SupplyChainReady",
                    "status": "True",
                    "lastTransitionTime": "2022-11-28T05:10:32Z",
                    "reason": "Ready",
                    "message": ""
                },
                {
                    "type": "ResourcesSubmitted",
                    "status": "True",
                    "lastTransitionTime": "2022-11-28T05:13:33Z",
                    "reason": "ResourceSubmissionComplete",
                    "message": ""
                },
                ...
            ],
            "supplyChainRef": {
                "kind": "ClusterSupplyChain",
                "name": "source-to-url"
            },
            "resources": [
                {
                    "name": "source-provider",
                    "stampedRef": {
                        "kind": "GitRepository",
                        "namespace": "default",
                        "name": "rmq-sample-app",
                        "apiVersion": "source.toolkit.fluxcd.io/v1beta1",
                        "resource": "gitrepositories.source.toolkit.fluxcd.io"
                    },
                    "templateRef": {
                        "kind": "ClusterSourceTemplate",
                        "name": "source-template",
                        "apiVersion": "carto.run/v1alpha1"
                    },
                ...
                }
            ...
            ]
            ...
        }
        ...
    }

## with yaml format
tanzu apps workload get rmq-sample-app --output yaml # can also be used as tanzu apps workload get rmq-sample-app -oyaml
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "2022-11-28T05:10:32Z"
  generation: 2
  labels:
    apps.tanzu.vmware.com/workload-type: web
  managedFields:
  - apiVersion: carto.run/v1alpha1
      ...
    manager: v0.10.0+dev-002cc44e
    operation: Update
    time: "2022-11-28T05:10:32Z"
  - apiVersion: carto.run/v1alpha1
    fieldsType: FieldsV1
    ...
    manager: cartographer
    operation: Update
    subresource: status
    time: "2022-11-28T05:10:36Z"
  name: rmq-sample-app
  namespace: default
  resourceVersion: "11657434"
  uid: 3619ff6d-9e73-473a-9112-891a6d8aee9e
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
      url: https://github.com/jhvhs/rabbitmq-sample
status:
  conditions:
  - lastTransitionTime: "2022-11-28T05:10:32Z"
    message: ""
    reason: Ready
    status: "True"
    type: SupplyChainReady
  ...
  observedGeneration: 2
  resources:
  ...
    name: source-provider
    outputs:
    - digest: sha256:97b2cb779b4ea31339595cd204a3fec0053805eeacbbd6d6dd23af7d3000a6ae
      lastTransitionTime: "2022-11-28T05:16:01Z"
      name: url
      preview: |
        http://fluxcd-source-controller.flux-system.svc.cluster.local./gitrepository/default/rmq-sample-app/73c6311eefbf724fee9ad6f4524fa24ec842ff34.tar.gz
    - digest: sha256:e7884b071fe1bbb2551d42a171043d061a7591e744705572136e689c2a154b7a
      lastTransitionTime: "2022-11-28T05:16:01Z"
      name: revision
      preview: |
        HEAD/73c6311eefbf724fee9ad6f4524fa24ec842ff34
    stampedRef:
      apiVersion: source.toolkit.fluxcd.io/v1beta1
      kind: GitRepository
      name: rmq-sample-app
      namespace: default
      resource: gitrepositories.source.toolkit.fluxcd.io
    templateRef:
      apiVersion: carto.run/v1alpha1
      kind: ClusterSourceTemplate
      name: source-template
  - conditions:
    - lastTransitionTime: "2022-11-28T05:13:25Z"
      message: ""
      reason: ResourceSubmissionComplete
      status: "True"
      type: ResourceSubmitted
    ...
    inputs:
    - name: source-provider
```

## <a id='subpath-usage'> --sub-path

This flag is provided to support use cases where more than one application is included in a single project or repository.

- Using `--sub-path` when creating a workload from a Git repository

    ```bash
    tanzu apps workload apply subpathtester --git-repo https://github.com/path-to-repo/my-repo --git-branch main --type web --sub-path my-subpath

    🔎 Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: subpathtester
        8 + |  namespace: default
        9 + |spec:
       10 + |  source:
       11 + |    git:
       12 + |      ref:
       13 + |        branch: main
       14 + |      url: https://github.com/path-to-repo/my-repo
       15 + |    subPath: my-subpath
    ❓ Do you want to create this workload? [yN]:
    ```

- Using `--sub-path` when creating a workload from local source code
  - In the directory of the project you want to create the workload from

      ```bash
      tanzu apps workload apply my-workload --local-path . -s gcr.io/my-registry/my-workload-image --sub-path subpath_folder
      ❓ Publish source in "." to "gcr.io/my-registry/my-workload-image"? It may be visible to others who can pull images from that repository Yes
      Publishing source in "." to "gcr.io/my-registry/my-workload-image"...
      📥 Published source
      
      🔎 Create workload:
            1 + |---
            2 + |apiVersion: carto.run/v1alpha1
            3 + |kind: Workload
            4 + |metadata:
            5 + |  name: myworkload
            6 + |  namespace: default
            7 + |spec:
            8 + |  source:
            9 + |    image: gcr.io/my-registry/my-workload-image:latest@sha256:f28c5fedd0e902800e6df9605ce5e20a8e835df9e87b1a0aa256666ea179fc3f
           10 + |    subPath: subpath_folder
      ❓ Do you want to create this workload? [yN]:

      ```

**Note:** in cases where a workload must be created from local source code it's recommended to just set the `--local-path` value to point directly to the directory containing the code to be deployed rather than using `--sub-path` to reduce the total amount of code that must be uploaded to create the workload.

## <a id='tanzuignore-file-usage'> .tanzuignore file

As more systems around us adopt the "as code" approaches, application developers will increasingly have files in their projects that have nothing to do with actually running code (those files don't end up in the running container).

When creating a workload from local source code, these unused files can be added to the `.tanzuignore` file so there won't be an unnecessary consumption of resources when uploading the source.

Additionally, and perhaps more importantly, when iterating on code with `--live-update` enabled, changes which are triggered automatically and/or manually to certain directories/files specified in `.tanzuignore`, will not trigger the automatic re-deployment of the source code (making the iteration loop tighter than it would be if those directories/files were not specified in the `.tanzuignore` file).

Lastly, it's recommended that the `.tanzuignore` file include a reference to itself given it provides no value when deployed.

Folders/directories are supported (these must not end with the system separator, e.g. `/` or `\`).

Individual files can be listed.

And comments (which start with `#`) can be included.

If the `.tanzuignore` file contains files or directories that are not found in the source code, they will be ignored.

**Example of a .tanzuignore file**
```bash
    .tanzuignore # must contain itself in order to be ignored
    # This is a comment
    this/is/a/folder/to/exclude

    this-is-a-file.ext
```

<!-- ## <a id='maximize-efficiency'> Leveraging ENVs to maximize efficiency

## <a id='param-yaml-usage'> --param-yaml

## <a id='app-usage'> --app

## <a id='type-usage'> --type

## <a id='labels-anns-env'> --labels, --annotations, --env, --build-env -->

## <a id='dry-run-usage'> --dry-run

The main goal of `--dry-run` flag is to prepare all the steps to submit a workload to the cluster and stop before sending it, showing
an output of the final structure of the workload.

For example, when applying a workload from Git source:

```bash
tanzu apps workload apply rmq-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" -t web --dry-run
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-app
  namespace: default
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
      url: https://github.com/jhvhs/rabbitmq-sample
status:
  supplyChainRef: {}
```

This will allow the user to check how a workload *would be* created/updated in the cluster based on the current specifications passed in via `--file workload.yaml` and/or command flags.

If there would be an error when trying to apply the workload, this would be shown with the `--dry-run` flag.

```bash
tanzu apps workload create rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" -t web --dry-run
Error: workload "default/rmq-sample-app" already exists
```

## <a id='update-strategy-usage'> --update-strategy

The `--update-strategy` flag accepts two values (`merge` (default) and `replace`).

This flag can be used to control whether configuration properties/values passed in via `--file workload.yaml` for an existing workload will `merge` with, or completely `replace` (or overwrite), existing on-cluster properties/values set for a workload.

**Note:** All `tanzu apps workload apply` commands will employ the `merge` update strategy if not specified otherwise by explicitly using the flag and setting its value to `replace`.

With the default `merge`:

If the `--file workload.yaml` deletes an existing on-cluster property/value, that property will not be removed from the on-cluster definition.
If the `--file workload.yaml` includes a new property/value - it will be added to the on-cluster workload properties/values.
If the `--file workload.yaml` updates an existing value for a property, that property's value on-cluster will be updated.

With `replace`:

The on-cluster workload will be updated to exactly what has been specified in the `--file workload.yaml` definition.

The intent of the current default merge strategy is to prevent unintentional deletions of critical properties from existing workloads.

However, it has been found that this default strategy is counter-intuitive for users and `merge` has been deprecated as of TAP v1.4.0. The default update strategy will be switched to `replace` in 2024.

Examples of the outcomes of both `merge` and `replace` update strategies are provided below:

- ```bash
  # Export workload if there is no previous yaml definition
  tanzu apps workload get spring-petclinic --export > spring-petclinic.yaml

  # modify the workload definition
  vi rmq-sample-app.yaml
  ---
  apiVersion: carto.run/v1alpha1
  kind: Workload
  metadata:
    name: spring-petclinic
    labels:
      app.kubernetes.io/part-of: spring-petclinic
      apps.tanzu.vmware.com/workload-type: web
  spec:
    resources:
      requests:
        memory: 1Gi
      limits:           # delete this line
        memory: 1Gi     # delete this line
        cpu: 500m       # delete this line
    source:
      git:
        url: https://github.com/sample-accelerators/spring-petclinic
        ref:
          tag: tap-1.1
  ```

After saving the file, to check how both of the update strategy options behave, run:

```bash
tanzu apps workload apply -f ./spring-petclinic.yaml --update-strategy merge # if flag is not specified, merge is taken as default
```

Executing the command above will produce the following output:

```bash
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

Workload is unchanged, skipping update
```

On the other hand, use `replace` as follows:

```bash
tanzu apps workload apply -f ./spring-petclinic.yaml --update-strategy replace
```

Executing the `replace` command above will produce the following output:

```bash
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
 11, 11   |  resources:
 12     - |    limits:
 13     - |      cpu: 500m
 14     - |      memory: 1Gi
 15, 12   |    requests:
 16, 13   |      memory: 1Gi
 17, 14   |  source:
 18, 15   |    git:
...
❓ Really update the workload "spring-petclinic"? [yN]:
```

Note that with this last, the lines that were deleted in the yaml file are going to be deleted as well in the workload running in the cluster.

**Note**: It's important to highlight that the only fields that will remain exactly as they were created are the system populated metadata fields (`resourceVersion`, `uuid`, `generation`, `creationTimestamp`, `deletionTimestamp`).

## <a id='apply-output-usage'> Output workload after create/apply

`tanzu apps workload create/apply` commands can be used with `--output` flag which prints the workload once the process of
its creation or update happens.

Since the usage of this flag is mainly for scripting processes, all the prompts can be skipped with the usage of `--yes` flag as follows:

```bash
tanzu apps workload apply rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" --type web --output yaml --yes
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "2023-04-04T16:15:41Z"
  generation: 1
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
  resourceVersion: "184277312"
  uid: faf6e581-a2a3-47ab-b2d3-4160513c72df
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
      url: https://github.com/jhvhs/rabbitmq-sample
status:
  supplyChainRef: {}
```

If it is not used with `--yes` flag, all the prompts will be printed and the workload definition will be shown at the end.

```bash
tanzu apps workload apply rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" --type web --output yaml
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: rmq-sample-app
      8 + |  namespace: default
      9 + |spec:
     10 + |  serviceClaims:
     11 + |  - name: rmq
     12 + |    ref:
     13 + |      apiVersion: rabbitmq.com/v1beta1
     14 + |      kind: RabbitmqCluster
     15 + |      name: example-rabbitmq-cluster-1
     16 + |  source:
     17 + |    git:
     18 + |      ref:
     19 + |        branch: main
     20 + |      url: https://github.com/jhvhs/rabbitmq-sample
❓ Do you want to create this workload? [yN]: y
👍 Created workload "rmq-sample-app"

To see logs:   "tanzu apps workload tail rmq-sample-app --timestamp --since 1h"
To get status: "tanzu apps workload get rmq-sample-app"

---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "2023-04-04T15:18:13Z"
  generation: 1
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
  resourceVersion: "184169566"
  uid: 6588d398-b803-47e3-b31a-23d9a1a633a9
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
      url: https://github.com/jhvhs/rabbitmq-sample
status:
  supplyChainRef: {}
```

This flag can also be used with `--wait` or `--tail` if the intention is to retrieve the workload with everything and its status.
The behavior is the same regarding the prompts: if `--yes` flag is not used, then the workload definition and the surveys are displayed and it remains waiting until the workload is in status `ready`. Otherwise, it does not show anything until the workload is ready in the cluster.

It should be made clear that, if `--tail` is used, its logs are not going to be suppressed despite the usage of `--yes` flag.

```bash
tanzu apps workload apply rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" --type web --output yaml --yes --wait
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "2023-04-04T16:22:29Z"
  generation: 1
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
  resourceVersion: "184296857"
  uid: 7fa58a71-0b41-4975-b816-781b87d02cde
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
      url: https://github.com/jhvhs/rabbitmq-sample
status:
  conditions:
  ...
  - lastTransitionTime: "2023-04-04T16:26:03Z"
    message: ""
    reason: Ready
    status: "True"
    type: Ready
  observedGeneration: 1
  resources:
  - conditions:
    - lastTransitionTime: "2023-04-04T16:22:36Z"
      message: ""
      reason: ResourceSubmissionComplete
      status: "True"
      type: ResourceSubmitted
      ...
    name: source-provider
    outputs:
    - digest: sha256:1982253401b1be2236786e3da433216d36d289d0b158fbc9ca6477ac94879e60
      lastTransitionTime: "2023-04-04T16:22:36Z"
      name: url
      preview: |
        http://fluxcd-source-controller.flux-system.svc.cluster.local./gitrepository/default/rmq-sample-app/103fde37882b5510e9b3974e5fe209161b54f675.tar.gz
    ...
    stampedRef:
      apiVersion: source.toolkit.fluxcd.io/v1beta1
      kind: GitRepository
      name: rmq-sample-app
      namespace: default
      resource: gitrepositories.source.toolkit.fluxcd.io
    templateRef:
      apiVersion: carto.run/v1alpha1
      kind: ClusterSourceTemplate
      name: source-template
  - conditions:
    ...
    - lastTransitionTime: "2023-04-04T16:25:45Z"
      message: ""
      reason: Ready
      status: "True"
      type: Ready
    inputs:
    - name: source-provider
    ...
  - conditions:
    ...
    - lastTransitionTime: "2023-04-04T16:25:52Z"
      message: ""
      reason: Ready
      status: "True"
      type: Ready
    inputs:
    - name: image-provider
    name: config-provider
    outputs:
    - digest: sha256:0549f3f3fe5ef817af62ae6357465e6df1a6c901e5a7abc17468ee3f3e16c1a1
      lastTransitionTime: "2023-04-04T16:25:52Z"
      name: config
      preview: |-
        metadata:
            annotations:
                boot.spring.io/version: 2.4.9
                conventions.carto.run/applied-conventions: |-
                    spring-boot-convention/auto-configure-actuators-check
                    ...
                developer.conventions/target-containers: workload
                services.conventions.carto.run/rabbitmq: amqp-client/5.10.0
            labels:
                app.kubernetes.io/component: run
                apps.tanzu.vmware.com/auto-configure-actuators: "false"
                apps.tanzu.vmware.com/workload-type: web
                car
    ...
  supplyChainRef:
    kind: ClusterSupplyChain
    name: source-to-url
```

And with `tail`:

```bash
tanzu apps workload apply rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1" --type web --output yaml --yes --tail
+ rmq-sample-app-build-1-build-pod › prepare
rmq-sample-app-build-1-build-pod[prepare] Build reason(s): CONFIG
rmq-sample-app-build-1-build-pod[prepare] CONFIG:
rmq-sample-app-build-1-build-pod[prepare] 	+ env:
rmq-sample-app-build-1-build-pod[prepare] 	+ - name: BP_OCI_SOURCE
rmq-sample-app-build-1-build-pod[prepare] 	+   value: main/103fde37882b5510e9b3974e5fe209161b54f675
rmq-sample-app-build-1-build-pod[prepare] 	resources: {}
rmq-sample-app-build-1-build-pod[prepare] 	- source: {}
rmq-sample-app-build-1-build-pod[prepare] 	+ source:
rmq-sample-app-build-1-build-pod[prepare] 	+   blob:
rmq-sample-app-build-1-build-pod[prepare] 	+     url: http://fluxcd-source-controller.flux-system.svc.cluster.local./gitrepository/default/rmq-sample-app/103fde37882b5510e9b3974e5fe209161b54f675.tar.gz
rmq-sample-app-build-1-build-pod[prepare] Loading secret for "gcr.io" from secret "registry-credentials" at location "/var/build-secrets/registry-credentials"
rmq-sample-app-build-1-build-pod[prepare] Loading secret for "registry.tanzu.vmware.com" from secret "registry-credentials" at location "/var/build-secrets/registry-credentials"
rmq-sample-app-build-1-build-pod[prepare] Loading cluster credential helpers
rmq-sample-app-build-1-build-pod[prepare] Downloading fluxcd-source-controller.flux-system.svc.cluster.local./gitrepository/default/rmq-sample-app/103fde37882b5510e9b3974e5fe209161b54f675.tar.gz...
...
...
...
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "2023-04-04T16:22:29Z"
  generation: 1
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
  resourceVersion: "184296857"
  uid: 7fa58a71-0b41-4975-b816-781b87d02cde
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
      url: https://github.com/jhvhs/rabbitmq-sample
status:
  conditions:
  ...
  - lastTransitionTime: "2023-04-04T16:26:03Z"
    message: ""
    reason: Ready
    status: "True"
    type: Ready
  observedGeneration: 1
  resources:
  - conditions:
    - lastTransitionTime: "2023-04-04T16:22:36Z"
      message: ""
      reason: ResourceSubmissionComplete
      status: "True"
      type: ResourceSubmitted
      ...
    name: source-provider
    outputs:
    - digest: sha256:1982253401b1be2236786e3da433216d36d289d0b158fbc9ca6477ac94879e60
      lastTransitionTime: "2023-04-04T16:22:36Z"
      name: url
      preview: |
        http://fluxcd-source-controller.flux-system.svc.cluster.local./gitrepository/default/rmq-sample-app/103fde37882b5510e9b3974e5fe209161b54f675.tar.gz
    ...
    stampedRef:
      apiVersion: source.toolkit.fluxcd.io/v1beta1
      kind: GitRepository
      name: rmq-sample-app
      namespace: default
      resource: gitrepositories.source.toolkit.fluxcd.io
    templateRef:
      apiVersion: carto.run/v1alpha1
      kind: ClusterSourceTemplate
      name: source-template
  - conditions:
    ...
    - lastTransitionTime: "2023-04-04T16:25:45Z"
      message: ""
      reason: Ready
      status: "True"
      type: Ready
    inputs:
    - name: source-provider
    ...
  - conditions:
    ...
    - lastTransitionTime: "2023-04-04T16:25:52Z"
      message: ""
      reason: Ready
      status: "True"
      type: Ready
    inputs:
    - name: image-provider
    name: config-provider
    outputs:
    - digest: sha256:0549f3f3fe5ef817af62ae6357465e6df1a6c901e5a7abc17468ee3f3e16c1a1
      lastTransitionTime: "2023-04-04T16:25:52Z"
      name: config
      preview: |-
        metadata:
            annotations:
                boot.spring.io/version: 2.4.9
                conventions.carto.run/applied-conventions: |-
                    spring-boot-convention/auto-configure-actuators-check
                    ...
                developer.conventions/target-containers: workload
                services.conventions.carto.run/rabbitmq: amqp-client/5.10.0
            labels:
                app.kubernetes.io/component: run
                apps.tanzu.vmware.com/auto-configure-actuators: "false"
                apps.tanzu.vmware.com/workload-type: web
                car
    ...
  supplyChainRef:
    kind: ClusterSupplyChain
    name: source-to-url
```

## <a id='unsetting-git-fields'> Un-setting Git fields

There are various ways to update a workload. It can be by changing its fields through flags or create a `yaml` file with the changes and run `tanzu apps workload apply` command with the `--update-strategy` set as `replace` (check [`--update-strategy`](./how-to-guides.md#update-strategy) for a better usage explanation).

However, for fields deletion, there is an easier way supported for the `--git-*` flags in which, through setting them as empty string in the command, the `workload.spec.source.git` fields get removed.

For example, if there is a workload that specifies `--git-tag`, `--git-commit` and `--git-branch`, to remove any of these the only thing that needs to be done is use empty string right after setting them.

```bash
## Existing workload definition
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
        commit: dec60a68190a4a8ebd3644806962002983ded69e
        tag: v0.1.0
      url: https://github.com/jhvhs/rabbitmq-sample

## Update the workload to remove one of its git fields

tanzu apps workload apply rmq-sample-app --git-tag ""
🔎 Update workload:
...
 17, 17   |    git:
 18, 18   |      ref:
 19, 19   |        branch: main
 20, 20   |        commit: dec60a68190a4a8ebd3644806962002983ded69e
 21     - |        tag: v0.1.0
 22, 21   |      url: https://github.com/jhvhs/rabbitmq-sample
❓ Really update the workload "rmq-sample-app"? [yN]: y
👍 Updated workload "rmq-sample-app"

To see logs:   "tanzu apps workload tail rmq-sample-app --timestamp --since 1h"
To get status: "tanzu apps workload get rmq-sample-app"

## Export the workload to see that `spec.source.git.ref.tag` is not part of the definition

tanzu apps workload get rmq-sample-app --export

---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
  source:
    git:
      ref:
        branch: main
        commit: dec60a68190a4a8ebd3644806962002983ded69e
      url: https://github.com/jhvhs/rabbitmq-sample
```

**NOTE**: If `--git-repo` is set to empty, then the whole git section is going to be removed from the workload definition.

```bash
tanzu apps workload apply rmq-sample-app --git-repo ""
🔎 Update workload:
...
 12, 12   |    ref:
 13, 13   |      apiVersion: rabbitmq.com/v1beta1
 14, 14   |      kind: RabbitmqCluster
 15, 15   |      name: example-rabbitmq-cluster-1
 16     - |  source:
 17     - |    git:
 18     - |      ref:
 19     - |        branch: main
 20     - |        commit: dec60a68190a4a8ebd3644806962002983ded69e
 21     - |      url: https://github.com/jhvhs/rabbitmq-sample
❗ NOTICE: no source code or image has been specified for this workload.
❓ Really update the workload "rmq-sample-app"? [yN]: y
👍 Updated workload "rmq-sample-app"

To see logs:   "tanzu apps workload tail rmq-sample-app --timestamp --since 1h"
To get status: "tanzu apps workload get rmq-sample-app"

## Export the workload and check that the git source section does not exist

tanzu apps workload get rmq-sample-app --export
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: rmq-sample-app
  namespace: default
spec:
  serviceClaims:
  - name: rmq
    ref:
      apiVersion: rabbitmq.com/v1beta1
      kind: RabbitmqCluster
      name: example-rabbitmq-cluster-1
```

## <a id='no-color-usage'> Remove color from output

Most of Tanzu Apps Plug-in commands have emojis and colored output with the intention to be more user-friendly.

However, sometimes color, emojis and other characters are not needed (e.g. scripting) or even well interpreted in certain terminals and the best way to suppress them is using the `--no-color` flag. <!-- TODO: give a better explanation -->

So, for example, in a workload that is created through local path, which usually shows emojis and a progress bar, these special characters would be avoid by using `--no-color`.

```bash
tanzu apps workload apply my-workload --local-path path/to/my/source -s my-registry.ext/my-project/my-workload --type web --no-color
The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
Publishing source in "path/to/my/source" to "my-registry.ext/my-project/my-workload"...
Published source

Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    image: my-registry.ext/my-project/my-workload:latest@sha256:724bcd14c3a84fc7a918cd8ee7a6a987de1699617a17c5af166e8c689a2becf7
? Do you want to create this workload? [yN]:
```

To avoid having color in general for Tanzu Apps Plug-in, the best way is setting the `NO_COLOR` environment variable. This will suppress color, emojis and progress bar for all the Plug-in related commands.

```bash
export NO_COLOR=true
```

<!-- ## <a id='wait-usage'> --wait

## <a id='wait-timeout-usage'> --wait-timeout 

## <a id='airgapped-env'> Apps plug-in usage in an airgapped environment -->

**NOTE: Additional how-to's will be added soon**
