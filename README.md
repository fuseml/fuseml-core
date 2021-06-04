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

* Run the server locally

  After successful installation using [fuseml-installer](https://github.com/fuseml/fuseml), `fuseml-core` server is already running in the Kubernetes cluster. However, for development and testing purposes it is also possible to run it locally by setting the following environment variables:
  
  - `GITEA_URL`
  - `GITEA_ADMIN_USERNAME`
  - `GITEA_ADMIN_PASSWORD`
  
  and executing `bin/fuseml_core`.

  The variables contain the values describing connection to the gitea server and credentials for the admin user.
  If you have used [fuseml-installer](https://github.com/fuseml/fuseml) to set up the environment, there is already default gitea instance installed in your Kubernetes cluster. In such case set the values this way:
  ```bash
  export GITEA_URL=http://$(kubectl get VirtualService -n gitea gitea -o jsonpath="{.spec.hosts[0]}")
  export GITEA_ADMIN_USERNAME=$(kubectl get Secret -n fuseml-workloads gitea-creds -o jsonpath="{.data.username}" | base64 -d)
  export GITEA_ADMIN_PASSWORD=$(kubectl get Secret -n fuseml-workloads gitea-creds -o jsonpath="{.data.password}" | base64 -d)
  ```
  It is possible to use external gitea server instead. Make sure to provide correct environment variables.
  
  `TEKTON_DASHBOARD_URL` is the path to the Tekton server. As with other components, Tekton is installed by `fuseml-installer` into your cluster. To get the right URL, call
  ```bash
  export TEKTON_DASHBOARD_URL=http://$(kubectl get VirtualService -n tekton-pipelines tekton -o jsonpath="{.spec.hosts[0]}")
  ```

  Now it's possible to execute `bin/fuseml_core`.
  Use the `--help` flag to get the command line options that you can supply. By default the server listens on the follwing ports: 8000 (http) and 8080 (grpc)

* Run the client

  Executing the client with `--help` option will show the usage instructions
  ```bash
  > bin/fuseml --help
  FuseML command line client

  Usage:
    bin/fuseml [command]

  Available Commands:
    application application management
    codeset     codeset management
    help        Help about any command
    runnable    runnable management
    version     display version information
    workflow    Workflow management

  Flags:
    -h, --help          help for bin/fuseml
        --timeout int   (FUSEML_HTTP_TIMEOUT) maximum number of seconds to wait for response (default 30)
    -u, --url string    (FUSEML_SERVER_URL) URL where the FuseML service is running
    -v, --verbose       (FUSEML_VERBOSE) print verbose information, such as HTTP request and response details

  Use "bin/fuseml [command] --help" for more information about a command.
  ```

  Instead of providing `--url` argument to each command call, you can save the service URL into `FUSEML_SERVER_URL` environment variable and export it. Use this command to fill the variable with the URL of the `fuseml-core` server service that is installed in your Kubernetes cluster:

  ```bash
  export FUSEML_SERVER_URL=http://$(kubectl get VirtualService -n fuseml-core fuseml-core -o jsonpath="{.spec.hosts[0]}")
  ```

  The FuseML client allows you to manage the various supported artifacts (application, codeset, runnable and workflow). Use the `--help` on each available command to get a more detailed description the command and instructions on how to use it.

  * Codesets contain the code of your ML application, for example MLflow project. They are currently implemented as git repositories.

    Create new codeset by calling the `bin/fuseml codeset register` command. 

    Example:
    ```bash
    bin/fuseml codeset register --name "test" --project "mlflow-project-01" "/tmp/mlflow/mlflow-01"
    ```

    Last argument points to the directory on your machine where your ML application code is located.

    After registering, use
    ```
    bin/fuseml codeset list
    ```
    command to list available codesets and check the value of "URL" in the output of your newly registered codeset. Use this value to git clone the code into another directory. Now you can work in this directory just like with any other project saved in git. Assuming that you assigned a workflow to this codeset (see workflows section), every time you push new code changes a new workflow run will be created and an updated application will be created.

    Note: the `codeset list` command allows filtering the output by project or user defined labels.


  * Workflows define the full AI/ML workflow. In short, this could be described as a way to process the input (the Codeset) and turn it into the output application (e.g. ML predictor).

    To create a new workflow, use
    ```bash
    bin/fuseml workflow create workflow.yaml
    ``` command.

    After the workflow is created it can be assigned to a codeset. By doing that, the workflow will be automatically executed every time you push a new change to the codeset that the workflow has been assigned to. The first time the workflow is assigned to a codeset a workflow run is also created, which will execute the workflow with its default inputs and the codeset it has been assigned to.

    ```bash
    bin/fuseml workflow assign --name "workflow-name" --codeset-name "test" --codeset-project "mlflow-project-01"
    ```

    To see the progress of running workflow, check the `list-runs` command:
    ```bash
    bin/fuseml workflow list-runs --workflow-name mlflow-sklearn-e2e
    ```

    To get even more details follow, use the `yaml` format for the output (`--format yaml`) which will provide a `url` to the Tekton Pipeline status on the Tekton Dashboard. Alternatively go to Tekton dashboard in your browser (remember `TEKTON_DASHBOARD_URL` extracted earlier) and select the correspondent pipeline run under the PipelineRuns menu.

  * Applications are basically the output services of AI/ML workflow. So if your workflow describes the way from the code, to the trained model, to the serving, the application being served as the last step is considered the FuseML application.

    Applications are registered automatically by workflows. Use

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
  export FUSEML_SERVER_URL=http://$(kubectl get VirtualService -n fuseml-core fuseml-core -o jsonpath="{.spec.hosts[0]}")
  ```
  

* Get the example code

  Check out the [examples](https://github.com/fuseml/examples) project:

  ```bash
  git clone git@github.com:fuseml/examples.git fuseml-examples
  ```

  Under `codesets/mlflow-wines` directory there is the example MLflow project. It is a slightly modified example based on [upstream MLflow](https://github.com/mlflow/mlflow/tree/master/examples).

  Under `workflows` directory there is an example definition of a FuseML workflow.

* Register the codeset

  Register the example MLflow project as a codeset:
  ```bash
  fuseml codeset register --name "mlflow-wines" --project "mlflow-project-01" "fuseml-examples/codesets/mlflow-wines"
  ```

* Update the example workflow to fit your setup

  The [workflow definition example](https://github.com/fuseml/examples/blob/main/workflows/mlflow-sklearn-e2e.yaml) has some hardcoded values that need to be changed for your specific environment. Namely, see the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` values, these are the credentials to the S3 based minio store that was installed into your kubernetes cluster by `fuseml-installer`.

  To get these values from your setup, run:
  ```bash
  export ACCESS=$(kubectl get secret -n fuseml-workloads mlflow-minio -o json | jq -r '.["data"]["accesskey"]' | base64 -d)
  export SECRET=$(kubectl get secret -n fuseml-workloads mlflow-minio -o json | jq -r '.["data"]["secretkey"]' | base64 -d)
  ```

  Now replace the original values in the `mlflow-sklearn-e2e.yaml` example. You can do it by editing the file manually or by running following commands:
  ```bash
  sed -i -e "/AWS_ACCESS_KEY_ID/{N;s/value: [^ \t]*/value: $ACCESS/}" fuseml-examples/workflows/mlflow-sklearn-e2e.yaml
  sed -i -e "/AWS_SECRET_ACCESS_KEY/{N;s/value: [^ \t]*/value: $SECRET/}" fuseml-examples/workflows/mlflow-sklearn-e2e.yaml
  ```

* Create a workflow

  Use the modified example workflow definition:

  ```bash
  fuseml workflow create fuseml-examples/workflows/mlflow-sklearn-e2e.yaml
  ```

* Assign the workflow to the codeset

  ```bash
  fuseml workflow assign --name mlflow-sklearn-e2e --codeset-name mlflow-wines --codeset-project mlflow-project-01
  ```

  Now that the Workflow is assigned to the Codeset, a new workflow run was created. To check the workflow run status, execute the following command:

  ```bash
  fuseml workflow list-runs --name mlflow-sklearn-e2e
  ```

  This command shows you detailed information about workflows runs. You can also access the `TEKTON_DASHBOARD_URL` to check all available workflow runs.
  
  Once the running workflow reaches the `Succeeded` state, a new FuseML application has been be created serving the trained model from the Codeset.

* Use the prediction service

  Check the available applications by running the following command:

  ```bash
  fuseml application list
  ```

  This should produce output similar to the following (notice the fake "example.io" domain here):

  ```bash
  +--------------------------------+-----------+------------------------------------------------------+------------------------------------------------------------------------------------------------------------------+--------------------+
  | NAME                           | TYPE      | DESCRIPTION                                          | URL                                                                                                              | WORKFLOW           |
  +--------------------------------+-----------+------------------------------------------------------+------------------------------------------------------------------------------------------------------------------+--------------------+
  | mlflow-project-01-mlflow-wines | predictor | Application generated by mlflow-sklearn-e2e workflow | http://mlflow-project-01-mlflow-wines.fuseml-workloads.example.io/v2/models/mlflow-project-01-mlflow-wines/infer | mlflow-sklearn-e2e |
  +--------------------------------+-----------+------------------------------------------------------+------------------------------------------------------------------------------------------------------------------+--------------------+
  ```
  
  Save the application URL in a variable for using in a later step:

  ```bash
  PREDICTION_URL=$(fuseml application get -n mlflow-project-01-mlflow-wines --format json | jq -r ".url")
  ```

  Take a look at the example data file from the examples repository, the prediction will be made according to the submitted `data`:

  ```bash
  > cat fuseml-examples/prediction/data-wines-kfserving.json
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

  Now perform a request to `$PREDICTION_URL` with the example data:

  ```bash
  curl -d @fuseml-examples/prediction/data-wines-kfserving.json $PREDICTION_URL
  ```

  The output should look as follows, the prediction result is under `outputs/data`:

  ```json
  {
    "model_name":"mlflow-project-01-mlflow-wines",
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
  * `openapi.go` - definition fo the openapi service, which exposes a HTTP file server endpoint serving the generated OpenAPI specification
* `gen/` - contains the boilerplate code generated by Goa (output of `$ goa gen github.com/fuseml/fuseml-core/design`)
  * `runnable/` - houses the transport-independent runnable service code
  * `grpc/` - contains the protocol buffer descriptions for the runnable gRPC service as well as the server and client code which hooks up the protoc-generated gRPC server and client code along with the logic to encode and decode requests and responses. The cli subdirectory contains the CLI code to build gRPC requests from the command line.
  * `http/` - describes the HTTP transport which defines server and client code with the logic to encode and decode requests and responses and the CLI code to build HTTP requests from the command line. It also contains the Open API 2.0/3.0 specification files in both json and yaml formats
* `cmd/` - a basic implementation of the service along with buildable server files that spins up goroutines to start a HTTP and a gRPC server and client files that can make requests to the server (output of `$ goa example github.com/fuseml/fuseml-core/design`)
* `runnable.go` - contains a dummy implementation of the methods described in the design (`design/runnable.go`) for the runnable service, the actual implementation goes here.

## NOTES

* The code generated by `goa gen` cannot be edited. This directory is re-generated entirely from scratch each time the command is run (e.g. after the design has changed). This is by design to keep the interface between generated and non generated code clean and using standard Go constructs (i.e. function calls). The code generated by `goa example however` is your code. You should modify it, add tests to it etc. This command generates a starting point for the service to help bootstrap development - in particular it is **NOT** meant to be re-run when the design changes. Instead simply edit the files accordingly.
