# tanzu apps workload apply

tanzu apps workload apply is a command used to create/update workloads that will be deployed in a cluster through a supply chain.

## Default view

In the output of workload apply command, the specification for the workload will be shown as if they were in a yaml file.
For example:

- ```bash
    tanzu apps workload apply pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web

    Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  source:
       11 + |    git:
       12 + |      ref:
       13 + |        tag: tap-1.1
       14 + |      url: https://github.com/sample-accelerators/spring-petclinic

    ? Do you want to create this workload? Yes
    Created workload "pet-clinic"

    To see logs:   "tanzu apps workload tail pet-clinic"
    To get status: "tanzu apps workload get pet-clinic"

    ```

In the first part, the workload definition is displayed. After that, there is a survey asking user if the workload should be created or updated and finally, if workload is actually to be created or updated, a couple of hints to commands that can be used to do a follow up. Each flag used in this example will be explained in detail in the following section.

  <details><summary>Workload Apply flags</summary>

  - *annotation*(`--annotation`): set the annotations to be applied to the workload, to specify more than one annotation set the flag multiple times, this annotations will be passed as parameters to be processed in the supply chain.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --annotation tag=tap-1.1 --annotation name="Spring pet clinic"
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  params:
     11 + |  - name: annotations
     12 + |    value:
     13 + |      name: Spring pet clinic
     14 + |      tag: tap-1.1
     15 + |  source:
     16 + |    git:
     17 + |      ref:
     18 + |        tag: tap-1.1
     19 + |      url: https://github.com/sample-accelerators/spring-petclinic
  ```

  To delete an annotation, use `-` after its name.
  ```bash
  tanzu apps workload apply spring-pet-clinic --annotation tag-
  Update workload:
  ...
  10, 10   |  params:
  11, 11   |  - name: annotations
  12, 12   |    value:
  13, 13   |      name: Spring pet clinic
  14     - |      tag: tap-1.1
  15, 14   |  source:
  16, 15   |    git:
  17, 16   |      ref:
  18, 17   |        tag: tap-1.1
  ...

  ? Really update the workload "spring-pet-clinic"? (y/N)
  ```

  - *app name*(`--app`): the app of which the workload is part of. This will be part of the workload metadata section.
  ```bash
  tanzu apps workload apply pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --app spring-petclinic

  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
      8 + |  name: pet-clinic
      9 + |  namespace: default
     10 + |spec:
     11 + |  source:
     12 + |    git:
     13 + |      ref:
     14 + |        tag: tap-1.1
     15 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? Yes
  Created workload "pet-clinic"

  To see logs:   "tanzu apps workload tail pet-clinic"
  To get status: "tanzu apps workload get pet-clinic"

  ```

  - *build environment variables*(`--build-env`): sets environment variables to be used in the **build** phase by the build resources in the supply chain where some *build* specific behavior can be set or changed
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --build-env JAVA_VERSION=1.8
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  build:
       11 + |    env:
       12 + |    - name: JAVA_VERSION
       13 + |      value: "1.8"
       14 + |  source:
       15 + |    git:
       16 + |      ref:
       17 + |        tag: tap-1.1
       18 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload?
  ```

  To delete a build environment variable, use `-` after its name.
  ```bash
  tanzu apps workload apply spring-pet-clinic --build-env JAVA_VERSION-
  Update workload:
  ...
    6,  6   |    apps.tanzu.vmware.com/workload-type: web
    7,  7   |  name: spring-pet-clinic
    8,  8   |  namespace: default
    9,  9   |spec:
   10     - |  build:
   11     - |    env:
   12     - |    - name: JAVA_VERSION
   13     - |      value: "1.8"
   14, 10   |  source:
   15, 11   |    git:
   16, 12   |      ref:
   17, 13   |        tag: tap-1.1
  ...

  ? Really update the workload "spring-pet-clinic"? (y/N)
  ```

  - *debug*(`--debug`): <!--TODO Definition and example-->

  - *dry run*(`--dry-run`): prepares all the steps to submit the workload to the cluster but stops just before sending it, showing as output how the final structure of the workload would be.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --build-env JAVA_VERSION=1.8 --param-yaml server=$'port: 8080\nmanagement-port: 8181' --dry-run
  ---
  apiVersion: carto.run/v1alpha1
  kind: Workload
  metadata:
    creationTimestamp: null
    labels:
      apps.tanzu.vmware.com/workload-type: web
    name: spring-pet-clinic
    namespace: default
  spec:
    build:
      env:
      - name: JAVA_VERSION
        value: "1.8"
    params:
    - name: server
      value:
        management-port: 8181
        port: 8080
    source:
      git:
        ref:
          tag: tap-1.1
        url: https://github.com/sample-accelerators/spring-petclinic
  status:
    supplyChainRef: {}
  ```

  - *environment variables*(`--env`): set the environment variables to the workload so the supply chain resources can used it to properly deploy the workload application
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --env NAME="Spring Pet Clinic"
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  env:
       11 + |  - name: NAME
       12 + |    value: Spring Pet Clinic
       13 + |  source:
       14 + |    git:
       15 + |      ref:
       16 + |        tag: tap-1.1
       17 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload?
  ```

  To unset an environment variable, use `-` after its name.
  ```bash
  tanzu apps workload apply spring-pet-clinic --env NAME-
  Update workload:
  ...
    6,  6   |    apps.tanzu.vmware.com/workload-type: web
    7,  7   |  name: spring-pet-clinic
    8,  8   |  namespace: default
    9,  9   |spec:
   10     - |  env:
   11     - |  - name: NAME
   12     - |    value: Spring Pet Clinic
   13, 10   |  source:
   14, 11   |    git:
   15, 12   |      ref:
   16, 13   |        tag: tap-1.1
  ...

  ? Really update the workload "spring-pet-clinic"? (y/N)
  ```

  - *filepath*(`-f`/`--file`): set a workload specification file to create the workload from, any other workload specification passed by flags to the command will set or override whatever is in the file. Another way to use this flag is using `-` in the command, to receive workload definition through standard input. Refer to [Working with Yaml Files](../../usage.md#a-idyaml-filesaworking-with-yaml-files) section to check an example.
  ```bash
  tanzu apps workload apply spring-pet-clinic -f pet-clinic.yaml --param-yaml server=$'port: 9090\nmanagement-port: 9190'
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  build:
       11 + |    env:
       12 + |    - name: JAVA_VERSION
       13 + |      value: "1.8"
       14 + |  params:
       15 + |  - name: server
       16 + |    value:
       17 + |      management-port: 9190
       18 + |      port: 9090
       19 + |  source:
       20 + |    git:
       21 + |      ref:
       22 + |        tag: tap-1.1
       23 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  - *git repo*(`--git-repo`): git repo from which the workload is going to be created. Along with this, `--git-tag`, `--git-commit` or `--git-branch` can be specified.

  - *git branch*(`--git-branch`): branch in git repo from where the workload is going to be created. This can be specified along with a commit or a tag.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload?
  ```

  - *git tag*(`--git-tag`): tag in git repo from which the workload is going to be created. Used with `--git-commit` or `--git-branch`

  - *git commit*(`--git-commit`): commit in git repo from where the workload is going to be resolved. Can be used with `--git-branch` or `git-tag`.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.2 --git-commit 207852f1e8ed239b6ec51a559c6e0f93a5cf54d1 --type web
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        commit: 207852f1e8ed239b6ec51a559c6e0f93a5cf54d1
     14 + |        tag: tap-1.2
     15 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload?
  ```

  - *image*(`--image`): sets the OSI image to be used as the workload application source instead of a git repository
  ```bash
  tanzu apps workload apply spring-pet-clinic --image private.repo.domain.com/spring-pet-clinic --type web
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  image: private.repo.domain.com/spring-pet-clinic

  ? Do you want to create this workload?
  ```

  - *label*(`--label`): set the label to be applied to the workload, to specify more than one label set the flag multiple times
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --label stage=production
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |    stage: production
        8 + |  name: spring-pet-clinic
        9 + |  namespace: default
       10 + |spec:
       11 + |  source:
       12 + |    git:
       13 + |      ref:
       14 + |        branch: main
       15 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  To unset labels, use `-` after their name.
  ```bash
  tanzu apps workload apply spring-pet-clinic --label stage-
  Update workload:
  ...
    3,  3   |kind: Workload
    4,  4   |metadata:
    5,  5   |  labels:
    6,  6   |    apps.tanzu.vmware.com/workload-type: web
    7     - |    stage: production
    8,  7   |  name: spring-pet-clinic
    9,  8   |  namespace: default
   10,  9   |spec:
   11, 10   |  source:
  ...

  ? Really update the workload "spring-pet-clinic"? (y/N)
  ```

  - *limit cpu*(`--limit-cpu`): refers to the maximum CPU the workload pods are allowed to use.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --limit-cpu .2
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  resources:
     11 + |    limits:
     12 + |      cpu: 200m
     13 + |  source:
     14 + |    git:
     15 + |      ref:
     16 + |        branch: main
     17 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  - *limit memory*(`--limit-memory`): refers to the maximum memory the workload pods are allowed to use.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --limit-memory 200Mi
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  resources:
     11 + |    limits:
     12 + |      memory: 200Mi
     13 + |  source:
     14 + |    git:
     15 + |      ref:
     16 + |        branch: main
     17 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  - *live update*(`--live-update`): enables to deploy a workload once, save changes to the code and see those changes reflected within seconds in the workload running on the cluster.
    - A usage example with a spring boot application.
      - Clone repo in https://github.com/sample-accelerators/tanzu-java-web-app
      - In `Tiltfile`, first change the `SOURCE_IMAGE` variable to use your registry and project. After that, at the very end of the file add
      ```bash
      allow_k8s_contexts('your-cluster-name')
      ```
      - Then, inside folder, run:
      ```bash
      tanzu apps workload apply tanzu-java-web-app --live-update --local-path . -s gcr.io/my-project/tanzu-java-web-app-live-update -y

      The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
      Publishing source in "." to "gcr.io/my-project/tanzu-java-web-app-live-update"...
      Published source
      Create workload:
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

      Created workload "tanzu-java-web-app"

      To see logs:   "tanzu apps workload tail tanzu-java-web-app"
      To get status: "tanzu apps workload get tanzu-java-web-app"

      ```
      - Run Tilt to deploy the workload.
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
      tanzu-java-w… │ 	Cannot extract live updates on this build graph structure
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
 
  - *local path*(`--local-path`): set the path to a source in the local machine from where the workload will create an image to use as application source. The local path can be a folder, a .jar, .zip or .war file and, so far, Java/Spring Boot compiled binaries are also supported. This flag must be used with `--source-image` flag.
  **Note**: If Java/Spring binary is passed, the command will take less time to apply the workload since buildpack will skip the compiling steps and will simply start uploading the image.
  
  When working with local source code, you can exclude files from the source code to be uploaded within the image by creating a file `.tanzuignore` at the root of the source code.
  The `.tanzuignore` file should contain a list of filepaths to exclude from the image including the file itself and the folders should not end with the system path separator (`/` or `\`). If the file contains files/folders that are not in the source code, they will be ignored as well as lines starting with `#` character.
  
  - *source image*(`-s`/`--source-image`): registry path where the local source code will be uploaded as an image.
  ```bash
  tanzu apps workload apply spring-pet-clinic --local-path /home/user/workspace/spring-pet-clinic --source-image gcr.io/spring-community/spring-pet-clinic --type web
  ? Publish source in "/home/user/workspace/spring-pet-clinic" to "gcr.io/spring-community/spring-pet-clinic"? It may be visible to others who can pull images from that repository Yes
  The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
  Publishing source in "/home/user/workspace/spring-pet-clinic" to "gcr.io/spring-community/spring-pet-clinic"...
  Published source
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  source:
       11 + |    image:gcr.io/spring-community/spring-pet-clinic:latest@sha256:5feb0d9daf3f639755d8683ca7b647027cfddc7012e80c61dcdac27f0d7856a7

  ? Do you want to create this workload? (y/N)
  ```

  - *namespace*(`-n`/`--namespace`): specifies the namespace in which the workload is to be created or updated.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --namespace my-namespace
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: my-namespace
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  - *parameters*(`--param`): additional parameters to be send to the supply chain, the value is send as a string, for complex yaml/json objects use `--param-yaml`
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --param port=9090 --param management-port=9190
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  params:
       11 + |  - name: port
       12 + |    value: "9090"
       13 + |  - name: management-port
       14 + |    value: "9190"
       15 + |  source:
       16 + |    git:
       17 + |      ref:
       18 + |        branch: main
       19 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  To unset parameters, use `-` after their name.
  ```bash
  tanzu apps workload apply spring-pet-clinic --param port-
  Update workload:
  ...
    7,  7   |  name: spring-pet-clinic
    8,  8   |  namespace: default
    9,  9   |spec:
   10, 10   |  params:
   11     - |  - name: port
   12     - |    value: "9090"
   13, 11   |  - name: management-port
   14, 12   |    value: "9190"
   15, 13   |  source:
   16, 14   |    git:
  ...

  ? Really update the workload "spring-pet-clinic"? (y/N)
  ```
 
  - *parameters in complex format*(`--param-yaml`):  additional parameters to be send to the supply chain, the value is send as complex object
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --param-yaml server=$'port: 9090\nmanagement-port: 9190'
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
       10 + |  params:
       11 + |  - name: server
       12 + |    value:
       13 + |      management-port: 9190
       14 + |      port: 9090
       15 + |  source:
       16 + |    git:
       17 + |      ref:
       18 + |        branch: main
       19 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  To unset parameters, use `-` after their name.
  ```bash
  tanzu apps workload apply spring-pet-clinic --param-yaml server-
  Update workload:
  ...
    6,  6   |    apps.tanzu.vmware.com/workload-type: web
    7,  7   |  name: spring-pet-clinic
    8,  8   |  namespace: default
    9,  9   |spec:
   10     - |  params:
   11     - |  - name: server
   12     - |    value:
   13     - |      management-port: 9190
   14     - |      port: 9090
   15, 10   |  source:
   16, 11   |    git:
   17, 12   |      ref:
   18, 13   |        branch: main
  ...

  ? Really update the workload "spring-pet-clinic"? (y/N)
  ```

  - *request cpu*(`--request-cpu`): refers to the minimum CPU the workload pods are requesting to use.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --request-cpu .3
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  resources:
     11 + |    requests:
     12 + |      cpu: 300m
     13 + |  source:
     14 + |    git:
     15 + |      ref:
     16 + |        branch: main
     17 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  - *request memory*(`--request-memory`): refers to the minimum memory the workload pods are requesting to use.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --request-memory 300Mi
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: spring-pet-clinic
      8 + |  namespace: default
      9 + |spec:
     10 + |  resources:
     11 + |    requests:
     12 + |      memory: 300Mi
     13 + |  source:
     14 + |    git:
     15 + |      ref:
     16 + |        branch: main
     17 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? (y/N)
  ```

  - *service account*(`--service-account`): <!--TODO Definition and example-->

  - *service bindings*(`--service-ref`): binds a service to a workload to provide the info from a service resource to an application.
  ```bash
  tanzu apps workload apply rmq-sample-app --git-repo https://github.com/jhvhs/rabbitmq-sample --git-branch main --service-ref "rmq=rabbitmq.com/v1beta1:RabbitmqCluster:example-rabbitmq-cluster-1"
  Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: rmq-sample-app
      6 + |  namespace: default
      7 + |spec:
      8 + |  serviceClaims:
      9 + |  - name: rmq
     10 + |    ref:
     11 + |      apiVersion: rabbitmq.com/v1beta1
     12 + |      kind: RabbitmqCluster
     13 + |      name: example-rabbitmq-cluster-1
     14 + |  source:
     15 + |    git:
     16 + |      ref:
     17 + |        branch: main
     18 + |      url: https://github.com/jhvhs/rabbitmq-sample

  ? Do you want to create this workload? (y/N)
  ```

  To delete service binding, use the service name followed by `-`.
  ```bash
  tanzu apps workload apply rmq-sample-app --service-ref rmq-
  Update workload:
  ...
    4,  4   |metadata:
    5,  5   |  name: rmq-sample-app
    6,  6   |  namespace: default
    7,  7   |spec:
    8     - |  serviceClaims:
    9     - |  - name: rmq
   10     - |    ref:
   11     - |      apiVersion: rabbitmq.com/v1beta1
   12     - |      kind: RabbitmqCluster
   13     - |      name: example-rabbitmq-cluster-1
   14,  8   |  source:
   15,  9   |    git:
   16, 10   |      ref:
   17, 11   |        branch: main
  ...

  ? Really update the workload "rmq-sample-app"? (y/N)
  ```

  - *subpath*(`--sub-path`): it's used to define which path is going to be used as root to create/update the workload.
    - Git repo
      ```bash
      tanzu apps workload apply subpathtester --git-repo https://github.com/tfynes-pivotal/subpathtester --git-branch main --type web --sub-path service1

      Create workload:
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
        14 + |      url: https://github.com/tfynes-pivotal/subpathtester
        15 + |    subPath: service1

      ? Do you want to create this workload? (y/N)
      ```

    - Local path
        - In the folder of the project you want to create the workload from
        ```bash
        tanzu apps workload apply my-workload --local-path . -s gcr.io/my-registry/my-workload-image --sub-path subpath_folder
        ? Publish source in "." to "gcr.io/my-registry/my-workload-image"? It may be visible to others who can pull images from that repository Yes
        Publishing source in "." to "gcr.io/my-registry/my-workload-image"...
        Published source
        Create workload:
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

        ? Do you want to create this workload? (y/N)

        ```

  - *tail*(`--tail`): prints the logs of the workload creation in every step.
   ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --tail
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
      10 + |  source:
      11 + |    git:
      12 + |      ref:
      13 + |        branch: main
      14 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? Yes
  Created workload "spring-pet-clinic"

  To see logs:   "tanzu apps workload tail spring-pet-clinic"
  To get status: "tanzu apps workload get spring-pet-clinic"

  Waiting for workload "spring-pet-clinic" to become ready...
  + spring-pet-clinic-build-1-build-pod › prepare
  spring-pet-clinic-build-1-build-pod[prepare] Build reason(s): CONFIG
  spring-pet-clinic-build-1-build-pod[prepare] CONFIG:
  spring-pet-clinic-build-1-build-pod[prepare] 	+ env:
  spring-pet-clinic-build-1-build-pod[prepare] 	+ - name: BP_OCI_SOURCE
  spring-pet-clinic-build-1-build-pod[prepare] 	+   value: main/d381fb658cb435a04e2271ca85bd3e8627a5e7e4
  spring-pet-clinic-build-1-build-pod[prepare] 	resources: {}
  spring-pet-clinic-build-1-build-pod[prepare] 	- source: {}
  spring-pet-clinic-build-1-build-pod[prepare] 	+ source:
  spring-pet-clinic-build-1-build-pod[prepare] 	+   blob:
  spring-pet-clinic-build-1-build-pod[prepare] 	+     url: http://source-controller.flux-system.svc.cluster.local./gitrepository/default/spring-pet-clinic/d381fb658cb435a04e2271ca85bd3e8627a5e7e4.tar.gz
  ...
  ...
  ...
  ```

  - *tail with timestamp*(`--tail-timestamp`): prints the logs of the workload creation in every step adding the time in which the log is occurring.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web --tail-timestamp
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
      10 + |  source:
      11 + |    git:
      12 + |      ref:
      13 + |        branch: main
      14 + |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Do you want to create this workload? Yes
  Created workload "spring-pet-clinic"

  To see logs:   "tanzu apps workload tail spring-pet-clinic"
  To get status: "tanzu apps workload get spring-pet-clinic"

  Waiting for workload "spring-pet-clinic" to become ready...
  + spring-pet-clinic-build-1-build-pod › prepare
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.348418803-05:00 Build reason(s): CONFIG
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.364719405-05:00 CONFIG:
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.364761781-05:00 	+ env:
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.364771861-05:00 	+ - name: BP_OCI_SOURCE
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.364781718-05:00 	+   value: main/d381fb658cb435a04e2271ca85bd3e8627a5e7e4
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.364788374-05:00 	resources: {}
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.364795451-05:00 	- source: {}
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.365344965-05:00 	+ source:
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.365364101-05:00 	+   blob:
  spring-pet-clinic-build-1-build-pod[prepare] 2022-06-15T11:28:01.365372427-05:00 	+     url: http://source-controller.flux-system.svc.cluster.local./gitrepository/default/spring-pet-clinic/d381fb658cb435a04e2271ca85bd3e8627a5e7e4.tar.gz
  ...
  ...
  ...
  ```

  - *type*(`--type`): sets the type of the workload by adding the label `apps.tanzu.vmware.com/workload-type`, which is very common to be used as a matcher by supply chains.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-branch main --type web
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
      10 + |  source:
      11 + |    git:
      12 + |      ref:
      13 + |        branch: main
      14 + |      url: https://github.com/sample-accelerators/spring-petclinic
  ```

  - *wait*(`--wait`): holds until workload is ready.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --wait
  Update workload:
  ...
  10, 10   |  source:
  11, 11   |    git:
  12, 12   |      ref:
  13, 13   |        branch: main
      14 + |        tag: tap-1.1
  14, 15   |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Really update the workload "spring-pet-clinic"? Yes
  Updated workload "spring-pet-clinic"

  To see logs:   "tanzu apps workload tail spring-pet-clinic"
  To get status: "tanzu apps workload get spring-pet-clinic"

  Waiting for workload "spring-pet-clinic" to become ready...
  Workload "spring-pet-clinic" is ready
  ```

  - *wait with timeout*(`--wait-timeout`): sets a timeout to wait for workload to become ready.
  ```bash
  tanzu apps workload apply spring-pet-clinic --git-repo https://github.com/sample-accelerators/spring-petclinic --git-tag tap-1.1 --type web --wait --wait-timeout 1m
  Update workload:
  ...
  10, 10   |  source:
  11, 11   |    git:
  12, 12   |      ref:
  13, 13   |        branch: main
  14     - |        tag: tap-1.2
      14 + |        tag: tap-1.1
  15, 15   |      url: https://github.com/sample-accelerators/spring-petclinic

  ? Really update the workload "spring-pet-clinic"? Yes
  Updated workload "spring-pet-clinic"

  To see logs:   "tanzu apps workload tail spring-pet-clinic"
  To get status: "tanzu apps workload get spring-pet-clinic"

  Waiting for workload "spring-pet-clinic" to become ready...
  Workload "spring-pet-clinic" is ready
  ```

  - *yes*(`-y`/`--yes`): assume yes on all the survey prompts
  ```bash
  tanzu apps workload apply spring-pet-clinic --local-path/home/user/workspace/spring-pet-clinic --source-image gcr.io/spring-community/spring-pet-clinic --type web -y
  The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
  Publishing source in "/Users/dalfonso/Documents/src/java/tanzu-java-web-app" to "gcr.io/spring-community/spring-pet-clinic"...
  Published source
  Create workload:
        1 + |---
        2 + |apiVersion: carto.run/v1alpha1
        3 + |kind: Workload
        4 + |metadata:
        5 + |  labels:
        6 + |    apps.tanzu.vmware.com/workload-type: web
        7 + |  name: spring-pet-clinic
        8 + |  namespace: default
        9 + |spec:
      10 + |  source:
      11 + |    image: gcr.io/spring-community/spring-pet-clinic:latest@sha256:5feb0d9daf3f639755d8683ca7b647027cfddc7012e80c61dcdac27f0d7856a7

  Created workload "spring-pet-clinic"

  To see logs:   "tanzu apps workload tail spring-pet-clinic"
  To get status: "tanzu apps workload get spring-pet-clinic"

  ```

  </details>
