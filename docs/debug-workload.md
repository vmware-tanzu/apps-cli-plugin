# Debugging workloads

## <a id="check-build-logs"></a> Check build logs

Once the workload is created, you can tail the workload to view the build and runtime logs. For more info about `tail` command, refer to [workload tail](command-reference/tanzu-apps-workload-tail.md) in command reference section and to check in more detail flag usage (with examples), go to [workload tail examples](./commands-details/workload_tail.md).

- Check logs by running:

    ```bash
    tanzu apps workload tail pet-clinic --since 10m --timestamp
    ```
    
    For Example

    ```bash
    tanzu apps workload tail pet-clinic

    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:52.684  INFO 1 --- [           main] org.apache.catalina.core.StandardEngine  : Starting Servlet engine: [Apache Tomcat/9.0.63]
    + pet-clinic-build-3-build-pod › export
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:52.699  INFO 1 --- [           main] o.a.c.c.C.[Tomcat-1].[localhost].[/]     : Initializing Spring embedded WebApplicationContext
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:52.699  INFO 1 --- [           main] w.s.c.ServletWebServerApplicationContext : Root WebApplicationContext: initialization completed in 131 ms
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:52.755  INFO 1 --- [           main] o.s.b.a.e.web.EndpointLinksResolver      : Exposing 13 endpoint(s) beneath base path '/actuator'
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:53.059  INFO 1 --- [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat started on port(s): 8081 (http) with context path ''
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:53.074  INFO 1 --- [           main] o.s.s.petclinic.PetClinicApplication     : Started PetClinicApplication in 8.373 seconds (JVM running for 8.993)
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:53.229  INFO 1 --- [nio-8081-exec-1] o.a.c.c.C.[Tomcat-1].[localhost].[/]     : Initializing Spring DispatcherServlet 'dispatcherServlet'
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:53.229  INFO 1 --- [nio-8081-exec-1] o.s.web.servlet.DispatcherServlet        : Initializing Servlet 'dispatcherServlet'
    pet-clinic-00004-deployment-6445565f7b-ts8l5[workload] 2022-06-14 16:28:53.231  INFO 1 --- [nio-8081-exec-1] o.s.web.servlet.DispatcherServlet        : Completed initialization in 2 ms
    ```

    Where:

    - `pet-clinic` is the name you gave the workload.
    - `--since` (optional) the amount of time to go back to begin streaming logs. The default is 1 second.
    - `--timestamp` (optional) prints the timestamp with each log line.

## <a id="workload-status"></a> Get the workload status and details

When the apps plugin is used to create or update a workload, it submits the changes to the platform and the CLI command is completed successfully. This does not necessarily mean that the change has been realized on the platform. The time it takes for the change to be executed on the backend will depend on the nature of the change requested.

After the workload build process is complete, create a Knative service to run the workload.

1. To check the workload details, run:
    ```bash
   tanzu apps workload get pet-clinic
   ```
    For example:

    ```bash
   tanzu apps workload get pet-clinic

    ---
    # pet-clinic: Ready
    ---
    Source
    type:   git
    url:    https://github.com/sample-accelerators/spring-petclinic
    tag:    tap-1.2

    Supply Chain
    name:          source-to-url
    last update:   10d
    ready:         True

    RESOURCE          READY   TIME
    source-provider
    deliverable
    image-builder
    config-provider
    app-config
    config-writer

    Issues
    No issues reported.

    Pods
    NAME                                           STATUS      RESTARTS   AGE
    pet-clinic-00001-deployment-6445565f7b-ts8l5   Running     0          102s
    pet-clinic-build-1-build-pod                   Succeeded   0          102s
    pet-clinic-config-writer-8c9zv-pod             Succeeded   0          2m7s
    Knative Services
    NAME         READY   URL
    pet-clinic   Ready   http://pet-clinic.default.apps.34.133.168.14.nip.io

    To see logs: "tanzu apps workload tail pet-clinic"

    ```

    Where:

    - `pet-clinic` is the name of the workload you want details about.

Refer to a more detailed info about the command and its flags usage in [workload get examples](./commands-details/workload_get.md)

2. You can now see the running workload. When the workload is created, `tanzu apps workload get` includes the URL for the running workload. Some terminals allow you to `ctrl`+click the URL to view it. You can also copy and paste the URL into your web browser to see the workload.
