# Working with Workloads

## <a id='Creating'></a> Creating a Workload 

This document describes how to create a workload from example source code with the Tanzu Application Platform.
### Getting Started with an Example Workload

Start by naming the workload and specifying a source code location for the workload to be created from.

1. Run:

    ```sh
    tanzu apps workload create pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web  
    ```

    Respond `Y` to prompts to complete process

    Where:

     + `pet-clinic` is the name that will be given to the workload
     + `--git-repo` is the location of the code to build the workload from
     + `--git-branch` (optional) specifies which branch in the repo to pull the code from
     + `--type` is used to distinguish the workload type

    NOTE: The above command will create the workload in the default namespace of your cluster, it is also possible to specify a namespace to use when creating a workload with the `--namespace` flag. For example: `tanzu apps workload create pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --namespace NAME`  

    The options available for specifying the workload are available in the command reference for [`workload create`](command-reference/tanzu_apps_workload_create.md) or by running `tanzu apps workload create --help`


### <a id='workload-tail'></a> Check Build Logs

Once the workload has been successfully created, you can tail the workload to view the build and runtime logs:

1. Check logs by running:

    ```sh
    tanzu apps workload tail pet-clinic --since 10m --timestamp
    ```

    Where:

     + `pet-clinic` is the name you gave the workload
     + `--since` (optional) is how long ago to start streaming logs from (default 1s)
     + `--timestamp` (optional) prints the timestamp with each log line

### <a id='workload-get'></a> Get the Workload Status and Details

Once the workload build process has completed, a knative service will be created to run the workload.
You can view workload details at anytime in the process but some details such as the workload URL will only be available once the workload is running:

1. Check the workload details by running:

    ```sh
    tanzu apps workload get pet-clinic
    ```

    Where:

     + `pet-clinic` is the name of the workload you would like details from.

2. See the running workload
When the workload has been created, `tanzu apps workload get` will include the URL for the running workload.
Depending on the terminal you use you may be able to `ctrl`+click on the URL to view. If that doesn't work you can copy-paste the URL into your web browser to see the workload.

### <a id='workload-local-source'></a> Create a Workload from Local Source

Additionally, it is possible to create a workload using code from a local folder.

1. Inside the folder that contains the source code, execute the following:

    ```sh
    tanzu apps workload create pet-clinic --local-path . --source-image springio/petclinic
    ```

    Respond `Y` to the prompt about publishing local source if the image needs to be updated.

    Where:

    + `pet-clinic` is the name that will be given to the workload
    + `--local-path` is pointing to the folder where the source code is located
    + `--source-image` is the registry path for the local source code

## <a id='service-binding'></a> Bind a Service to a Workload

Multiple services can be configured for each workload. The cluster supply chain is in charge of provisioning those services.

1. A way to bind a database service to a workload is:

    ```sh
    tanzu apps workload update pet-clinic --service-ref "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-db"
    ```

    Where:

    + `pet-clinic` is the name of the workload to be updated
    + `--service-ref` is the reference to the service using the format {name}={apiVersion}:{kind}:{name}. For more details, refer to [update command](command-reference/tanzu_apps_workload_update.md#update-options)

## <a id='next-steps'></a> Next Steps

You can add environment variables, export definitions, and use flags with these [commands](command-reference.md), for example:

1. Add some environment variables

    ```sh
    tanzu apps workload update pet-clinic --env foo=bar
    ```

2. Export the workload definition (to check into git, or migrate to another environment)

    ```sh
    tanzu apps workload get pet-clinic --export
    ```

3. Check out the flags available for the workload commands

    ```sh
    tanzu apps workload -h
    tanzu apps workload get -h
    tanzu apps workload create -h
    ```
