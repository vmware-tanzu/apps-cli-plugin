# Working with Workloads

This document describes how to create a workload from example source code with Tanzu Application Platform.

## <a id="get-started"></a> Get started with an example workload

### <a id="workload-git"></a> Create a workload from GitHub repository

Tanzu Application Platform supports creating a workload from an existing git repository by setting the flags `--git-repo`, `--git-branch`, `--git-tag` and `--git-commit`, this will allow the supply chain to get the source from the given repository to deploy the application.

To create a named workload and specify a git source code location, run:

 ```bash
tanzu apps workload create pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web  
```

Respond `Y` to prompts to complete process.

Where:

- `pet-clinic` is the name of the workload.
- `--git-repo` is the location of the code to build the workload from.
- `--git-branch` (optional) specifies which branch in the repository to pull the code from.
- `--type` is used to distinguish the workload type.

NOTE: The above command will create the workload in the default namespace of your cluster, it is also possible to specify a namespace to use when creating a workload with the `--namespace` flag. For example: `tanzu apps workload create pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --namespace NAME`

You can find the options available for specifying the workload in the command reference for [`workload create`](command-reference/tanzu_apps_workload_create.md) and find some other usage examples in [workload apply details](./commands-details/workload_create_update_apply.md).

### <a id="workload-local-source"></a> Create a workload from local source code

Tanzu Application Platform supports creating a workload from an existing local project by setting the flags `--local-path` and `--source-image`, then this will allow the supply chain to generate an image (carvel-imgpkg) and push it to the given registry to be used in the workload.

- To create a named workload and specify where the local source code is, run:

    ```bash
    tanzu apps workload create pet-clinic --local-path /path/to/my/project --source-image springio/petclinic
    ```

    Respond `Y` to the prompt about publishing local source code if the image needs to be updated.

    Where:

    - `pet-clinic` is the name of the workload.
    - `--local-path` points to the directory where the source code is located.
    - `--source-image` is the registry path where the local source code will be uploaded as an image. It can be set by an [environment variable](#env-vars)

    **Exclude Files**
    When working with local source code, you can exclude files from the source code to be uploaded within the image by creating a file `.tanzuignore` at the root of the source code. You can find the options available to specify the workload in the command reference for [`workload create`](command-reference/tanzu_apps_workload_create.md), or you can run `tanzu apps workload create --help`.
    
    The file should contain a list of filepaths to exclude from the image including the file itself and the folders should not end with the system path separator (`/` or `\`).

    If the file contains files/folders that are not in the source code, they will be ignored.

    If a line in the file starts with a `#` character, the line will be ignored.

    **Example**

    ```
    # This is a comment
    this/is/a/folder/to/exclude

    this-is-a-file.ext
    ```

### <a id="workload-image"></a> Create workload from an existing image

Tanzu Application Platform supports creating a workload from an existing image by setting the flag `--image`. This will allow the supply chain to get the given image from the registry to deploy the application.

An example on how to create a workload from image is as follows:

```console
tanzu apps workload create petclinic-image --image springcommunity/spring-framework-petclinic
```

Respond `Y` to prompts to complete process.

 Where:

- `petclinic-image` is the name of the workload.
- `--image` is an existing image, pulled from a registry, that contains the source that the workload is going to use to create the application.

### <a id="workload-maven"></a> Create a workload from Maven repository artifact

**Note:** This is currently supported only by [Tanzu Application Platform Enterprise](https://docs.vmware.com/en/VMware-Tanzu-Application-Platform/index.html).

Tanzu Application Platform supports creating a workload from a Maven repository artifact (source controller) by setting some specific properties as yaml parameter in the workload when using the supply chain.

The maven repository url is being set when the supply chain is created.

- Param name: maven
- Param value:
    - YAML:
    ```yaml
    artifactId: ...
    type: ... # default jar if not provided
    version: ...
    groupId: ...

    ``` 
    - JSON: 
    ```json
    {
        "artifactId": ...,
        "type": ..., // default jar if not provided
        "version": ...,
        "groupId": ...
    }
    ```

For example, to create a workload from a maven artifact, something like this could be done:

```bash
# YAML
tanzu apps workload create petclinic-image --param-yaml maven=$"artifactId:hello-world\ntype: jar\nversion: 0.0.1\ngroupId: carto.run"

# JSON
tanzu apps workload create petclinic-image --param-yaml maven="{"artifactId":"hello-world", "type": "jar", "version": "0.0.1", "groupId": "carto.run"}"
```

### <a id="env-vars"></a> Create and Apply environment variables

Developers will provide the same flags/values repeatedly when iterating on their application code.
Typing or *copy*/*pasting* the flag values for every workload `create`/`apply` adds friction to the developer workflow.

For this reason the apps plugin support the use some environment variables to set those values for the following flags:

- `--type`: `TANZU_APPS_TYPE`
- `--registry-ca-cert`: `TANZU_APPS_REGISTRY_CA_CERT`
- `--registry-password`: `TANZU_APPS_REGISTRY_PASSWORD`
- `--registry-username`: `TANZU_APPS_REGISTRY_USERNAME`
- `--registry-token`: `TANZU_APPS_REGISTRY_TOKEN`
- `--source-image`: `TANZU_APPS_SOURCE_IMAGE`

**Note:** Be aware that when set a supported environment value, each apps plugin command will set the flag with the value on the environment variable value

## <a id='service-binding'></a> Bind a Service to a Workload

Multiple services can be configured for each workload. The cluster supply chain is in charge of provisioning those services.

1. A way to bind a database service to a workload is:

    ```sh
    tanzu apps workload update pet-clinic --service-ref "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-db"
    ```

    Where:

    + `pet-clinic` is the name of the workload to be updated
    + `--service-ref` is the reference to the service using the format {service-ref-name}={apiVersion}:{kind}:{service-binding-name}. 

**Note:**: Check [Tanzu Application Platform documentation](https://docs-staging.vmware.com/en/draft/VMware-Tanzu-Application-Platform/1.2/tap/GUID-getting-started-consume-services.html#bind-an-application-workload-to-the-service-instance-6) to get a more detailed explanation on how to bind services to a workload.

## <a id='next-steps'></a> Next Steps

You can check workload details and status, add environment variables, export definitions, and use flags with these [commands](command-reference.md), for example:

1. To check workload status and details, use `workload get` command and to get workload logs, use `workload tail` command. For more info about these, refer to [debug workload section](debug-workload.md).

2. Add some environment variables

    ```bash
    tanzu apps workload update pet-clinic --env foo=bar
    ```

3. Export the workload definition (to check into git, or migrate to another environment)

    ```bash
    tanzu apps workload get pet-clinic --export
    ```

4. Check out the flags available for workload commands. A more detailed explanation of each flag and examples of their usage can be found in [command details folder](./commands-details/).

    ```bash
    tanzu apps workload -h
    tanzu apps workload get -h
    tanzu apps workload apply -h
    tanzu apps workload list -h
    ```
