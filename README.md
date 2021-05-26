# FuseML Core

This repository contains the FuseML APIs definitions and core service. For the general information about FuseML project, check the [main repository page](https://github.com/fuseml/fuseml).

## Installation

* If you are using [fuseml installer](https://github.com/fuseml/fuseml), the core service is installed into your Kubernetes cluster along with other components and the command line client is downloaded to your working directory. It is recommended to copy the client to the location with other executables, e.g. `/usr/local/bin/` on the Linux systems.


* To download pre-built client and server components, check the [releases page](https://github.com/fuseml/fuseml-core/releases).


* To build the latest version directly from sources:

  * Make sure you have `go` installed, at least version 1.16.

  * Download [protocol buffers](https://github.com/protocolbuffers/protobuf). Use the release page for [version 3.15.7](https://github.com/protocolbuffers/protobuf/releases/tag/v3.15.7) and install only the `protoc` binary.

    On Linux, you can proceed this way:
    ```bash
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.7/protoc-3.15.7-linux-x86_64.zip
    unzip protoc-3.15.7-linux-x86_64.zip
    sudo cp -a bin/protoc /usr/local/bin/
    ```

  * Clone `fuseml-core` repository
    ```bash
    git clone git@github.com:fuseml/fuseml-core.git
    cd fuseml-core
    ```

  * Install go dependencies
    ```bash
    make deps
    ```

  * Generate server, client and CLI code and build the binaries.
    ```bash
    make all
    ```
    This will produce server (`bin/fuseml_core`) and command line client (`bin/fuseml`) binaries. The client binary is compiled for your current architecture.


## Usage

* Run the server localy

  After successful installation using [fuseml-installer](https://github.com/fuseml/fuseml), `fuseml-core` server is already running in the Kubernetes cluster. However, for developmnent and testing purposes it is possible to run it locally locally by setting the env variables
  
  `GITEA_URL`, `GITEA_USERNAME`, `GITEA_PASSWORD`
  
  and executing `bin/fuseml_core`.

  The variables contain the values describing connection to the gitea server and credentials for the admin user.
  If you have used [fuseml-installer](https://github.com/fuseml/fuseml) to set up the environment, there is already default gitea instance installed in your Kubernetes cluster. In such case set the values this way:
  ```bash
  export GITEA_URL=http://$(kubectl get VirtualService -n gitea gitea -o jsonpath="{.spec.hosts[0]}")
  export GITEA_USERNAME=$(kubectl get Secret -n fuseml-workloads gitea-creds -o jsonpath="{.data.username}" | base64 -d)
  export GITEA_PASSWORD=$(kubectl get Secret -n fuseml-workloads gitea-creds -o jsonpath="{.data.password}" | base64 -d)
  ```
  It is possible to use external gitea server instead. Make sure to provide correct environment variables.
  
  `TEKTON_DASHBOARD_URL` is the path to the Tekton server. As with other components, Tekton is installed by `fuseml-installer` into your cluster. To get the right URL, call
  ```bash
  export TEKTON_DASHBOARD_URL=http://$(kubectl get VirtualService -n tekton-pipelines tekton -o jsonpath="{.spec.hosts[0]}")
  ```

  Now it's possible to execute `bin/fuseml_core`.
  Use the `--help` flag to get the command line options that you can supply. By default the server listens on the follwing ports: 8000 (http) and 8080 (grpc)

* Run the client

  Executing the client with `--help` option will show the usage instuctions
```bash
    > bin/fuseml --help
bin/fuseml is a command line client for the FuseML API.

Usage:
    fuseml [-host HOST][-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -host HOST:  server host (dev). valid values: dev, prod
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
    application (list|register|get|delete)
    codeset (list|register|get|delete)
    runnable (list|register|get)
    workflow (list|register|get|assign|list-runs)

Additional help:
    bin/fuseml SERVICE [ENDPOINT] --help

Example:
    bin/fuseml application list --type "predictor" --workflow "mlflow-sklearn-e2e"
    bin/fuseml codeset list --project "mlflow-project-01" --label "mlflow"
    bin/fuseml runnable list --id "ml-trainer-123" --kind "trainer" --labels '{
          "function": "predict|train",
          "library": "pytorch"
       }'
    bin/fuseml workflow list --name "workflow"
```

  Instead of providing `-url` argument to each command call, save the service URL into `FUSEML_SERVER_URL` environment variable and export it. Use this command to fill the variable with the URL of the `fuseml-core` server service that is installed in your Kubernetes cluster:

  ```bash
  export FUSEML_SERVER_URL=http://$(kubectl get VirtualService -n fuseml-core fuseml-core -o jsonpath="{.spec.hosts[0]}")
  ```

  If neither `FUSEML_SERVER_URL` nor `-url` value is set, client tries to connect to the server running locally at your machine.


  The FuseML client allows you to manage the various supported artefacts (application, codeset, runnable and workflow). It offers multiple commands for managing these artefact. Use the `--help` option to get the description of any command usage.

  * Codesets contain the code of your ML application, for example MLflow project. They are currently implemented as git repositories.

    Create new codeset by calling the `bin/fuseml codeset register` command. 

    Example:
    ```bash
    bin/fuseml codeset register --name "test" --project "mlflow-project-01" --location "/tmp/mlflow/mlflow-01"
    ```

    `--location` argument points to the directory on your machine where your ML application code is located.

    After registering, use
    ```
    bin/fuseml codeset list
    ```
    command to list available codesets and check the value of "URL" in the output of your newly registered codeset. Use this value to git clone the code into another directory. Now you can work in this directory just like with any other project saved in git. Assuming that you assigned a workflow to this codeset (see workflows section), every time you push new code changes a new workflow run will be created and an updated application will be created.

    Note: the `codeset list` command allows filtering the output by project or user defined labels.


  * Workflows define the full AI/ML workflow. In short, this could be described as a way to process the input (the Codeset) and turn it into the output application (e.g. ML predictor).

    For registering new workflow, use
    ```bash
    bin/fuseml workflow register
    ``` command.

    After the workflow is registered, you should assign a codeset to it. That way the workflow will be automatically executed every time you push a new change to your git repository that is represented by a codeset. For the first time the workflow is executed right after `workflow assign` command, even without any code changes.

    ```bash
    bin/fuseml workflow assign --name "workflow-name" --codeset-name "test" --codeset-project "mlflow-project-01"
    ```

    To see the progress of running workflow, check the `list-runs` command:
    ```bash
    bin/fuseml workflow list-runs --workflow-name mlflow-sklearn-e2e
    ```

    To get even more details follow the `url` value that is part of the output section from the `list-runs`. This will get you to the Tekton Pipeline status on the Tekton Dashboard. Alternativly go to Tekton dashboard in your browser (remember `TEKTON_DASHBOARD_URL` extracted earlier) and select among available Tekton Pipelines in the menu.

  * Applications are basically the output services of AI/ML workflow. So if your workflow describes the way from the code, to the trained model, to the serving, the application being server as the last step is considered the FuseML application.

    Applications are registed automatically by workflows. Use

    ```bash
    bin/fuseml application list
    ```
    to list existing applications. The output contains the URL where the application can be accessed, e.g. the URL of the prediction service.


## Example

Let's look at the example for MLflow model, being trained by MLflow and served with KFServing.

* Install FuseML and fuseml-core client

  Follow the [FuseML installation guide](https://github.com/fuseml/fuseml#usage) to install all necessary services, including `fuseml-core`.

  Install `fuseml` binary to some place within your PATH so you do not need to execute it with the full path.

  Set the value of `FUSEML_SERVER_URL`, to point to the server URL:

  ```bash
  export FUSEML_SERVER_URL=http//$(kubectl get VirtualService -n fuseml-core fuseml-core -o jsonpath="{.spec.hosts[0]}")
  ```
  

* Get the example code

  Check out the [examples](https://github.com/fuseml/examples) project:

  ```bash
  git clone git@github.com:fuseml/examples.git fuseml-examples
  cd fuseml-examples
  ```

  Under `models/mlflow-wines` directory there is the example MLflow project. It's only slightly modified example based on the [upstream MLflow](https://github.com/mlflow/mlflow/tree/master/examples) one.

  Under `pipelines` directory there is an example of FuseML workflow definition.

* Register the codeset

  Register the example MLflow model as a codeset:
  ```bash
  fuseml codeset register --name "mlflow-test" --project "mlflow-project-01" --location "models/mlflow-wines"
  ```

* Update the example to fit your setup

  The [workflow definition example](https://github.com/fuseml/examples/blob/main/pipelines/pipeline-01.yaml) has some hardcoded values that need to be changed for your specific environment. Namely, see the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` values: these are the credentials to the S3 based minio store that was installed to your cluster by `fuseml-installer`.

  To get these values from your cluster setup, run
  ```bash
  export ACCESS=$(kubectl get secret -n fuseml-workloads mlflow-minio -o json| jq -r '.["data"]["accesskey"]' | base64 -d)
  export SECRET=$(kubectl get secret -n fuseml-workloads mlflow-minio -o json| jq -r '.["data"]["secretkey"]' | base64 -d)
  ```

  Now replace the original values in the pipeline-01.yaml example. You can do it by editing the file manually or by running following command:
  ```bash
  sed -i -e "/AWS_ACCESS_KEY_ID/{N;s/value: [^ \t]*/value: $ACCESS/}" pipelines/pipeline-01.yaml
  sed -i -e "/AWS_SECRET_ACCESS_KEY/{N;s/value: [^ \t]*/value: $SECRET/}" pipelines/pipeline-01.yaml
  ```

* Create a workflow

  Use the modified example workflow definition:

  ```bash
  workflow=$(cat pipelines/pipeline-01.json)
  fuseml workflow register --body "$(cat pipelines/pipeline-01.yaml)"
  ```

* Assign the codeset to workflow

  ```bash
  fuseml workflow assign --name mlflow-sklearn-e2e --codeset-name mlflow-test --codeset-project mlflow-project-01
  ```

  Now that the Workflow is assigned to the Codeset, a new workflow run was created. To watch the workflow progress, check "workflow run" with

  ```bash
  fuseml workflow list-runs --workflow-name mlflow-sklearn-e2e
  ```

  This command shows you detailed information about running workflow. Follow the `url` value under `output` section to see relevant Tekton PipelineRun which implements the workflow run.

  Or browse to `TEKTON_DASHBOARD_URL` to check all available PipelineRuns. Once the run is succeeded, new FuseML application will be created.

* Use the prediction service

  Once the application is created, check the applications list with

  ```bash
  fuseml application list
  ```

  This should produce output similar to this one (notice the fake "example.io" domain here):

  ```bash
  - name: mlflow-project-01-mlflow-test
    type: predictor
    description: Application generated by mlflow-sklearn-e2e workflow
    url: http://mlflow-project-01-mlflow-test.fuseml-workloads.example.io/v2/models/mlflow-project-01-mlflow-test/infer
    workflow: mlflow-sklearn-e2e
  ```
  
  Use the URL from the new application to run the prediction. First, prepare the data

  ```bash
  > cat data.json
  {
    "inputs": [
      {
        "name": "input-0",
        "shape": [1, 11],
        "datatype": "FP32",
        "data": [
          [12.8, 0.029, 0.48, 0.98, 6.2, 29, 7.33, 1.2, 0.39, 90, 0.86]
        ]
      }
    ]
  }
  ```

  and pass the data to the the prediction service. Assuming the service URL was saved to `PREDICTOR_URL`, call

  ```bash
  curl -d @data.json http://$PREDICTOR_URL
  ```

  The output should look like

  ```json
  {
    "model_name":"mlflow-project-01-mlflow-test",
    "model_version":null,
    "id":"44d5d037-052b-49b6-aace-1c5346a35004",
    "parameters":null,
    "outputs": [
      {
          "name":"predict",
          "shape":[1],
          "datatype":"FP32",
          "parameters":null,
          "data": [ 6.486344809506676 ]
      }
    ]
  }
  ```



## Feedback

If you find a problem or have a suggestion for an enhancement, use the [https://github.com/fuseml/fuseml-core/issues](page).

## Code structure

* `design/` - contains specification consumed by Goa out of which REST API server and cli code are generated (HTTP and gRPC)
  * `api.go` - defines the http server and a list of services that the server will host
  * `runnable.go` - definition of the runnable service
  * `openapi.go` - defintion fo the openapi service, which exposes a HTTP file server endpoint serving the generated OpenAPI specification
* `gen/` - contains the boilerplate code generated by Goa (output of `$ goa gen github.com/fuseml/fuseml-core/design`)
  * `runnable/` - houses the transport-independent runnable service code
  * `grpc/` - contains the protocol buffer descriptions for the runnable gRPC service as well as the server and client code which hooks up the protoc-generated gRPC server and client code along with the logic to encode and decode requests and responses. The cli subdirectory contains the CLI code to build gRPC requests from the command line.
  * `http/` - describes the HTTP transport which defines server and client code with the logic to encode and decode requests and responses and the CLI code to build HTTP requests from the command line. It also contains the Open API 2.0/3.0 specification files in both json and yaml formats
* `cmd/` - a basic implementation of the service along with buildable server files that spins up goroutines to start a HTTP and a gRPC server and client files that can make requests to the server (outpug of `$ goa example github.com/fuseml/fuseml-core/design`)
* `runnable.go` - contains a dummy implementation of the methods described in the design (`design/runnable.go`) for the runnable service, the actual implementation goes here.

## NOTES

* The code generated by `goa gen` cannot be edited. This directory is re-generated entirely from scratch each time the command is run (e.g. after the design has changed). This is by design to keep the interface between generated and non generated code clean and using standard Go constructs (i.e. function calls). The code generated by `goa example however` is your code. You should modify it, add tests to it etc. This command generates a starting point for the service to help bootstrap development - in particular it is **NOT** meant to be re-run when the design changes. Instead simply edit the files accordingly.
