# FuseML Runnable Examples

```
runnable:
  id: example-runnable-1234
  description: |
    custom runnable example
  kind: custom
  implem:
    image: oci-registry.example.io/my-repo:tag
    env:
      ENV_VAR_ONE: "value-one"
      PARAM_ONE_VALUE: "{{inputs[example-parameter-one]}}"
      CODESET_ONE_PATH: "{{inputs[codeset-input-one].path}}"
      PARAM_THREE_OUTPATH: "{{outputs[example-parameter-three].path}}"
    entrypoint: "/usr/bin/entrypoint.sh"
    args:
      - --codeset-two-path
      - "{{inputs[codeset-input-two].path}}"
      - --param-two-path
      - "{{inputs[example-parameter-two].path}}"
      - --param-four-outpath
      - "{{inputs[example-parameter-four].path}}"
  defaultInputPath: /inputs
  defaultOutputPath: /outputs
  labels:
    type: custom
    purpose: example
    target: everyone
  inputs:
    example-parameter-one:
      type: parameter
      description: |
        example optional input parameter using the default value passing (default input path and other expressions)
      optional: true
      defaultValue: example-value
    example-parameter-two:
      type: parameter
      description: |
        example input parameter with custom composability labels and value passed through file
        (can be used e.g. to pass script and configuration contents inline as parameters)
      optional: false
      labels:
        field: name
        parameter-type: string
        length: "10"
      passByValue:
        toPath: /inputs/parameter-two.txt
    example-codeset-input-one:
      type: codeset
      description: |
        example input codeset artifact with a label matching selector and default value passing config
        (codeset contents are mounted as a dir under the default input path).
      optional: false
      labels:
        mlflow: project
        ml-library: "sklearn|pytorch"
        gpu-acceleration: "cuda"
    example-codeset-input-two:
      type: codeset
      description: |
        example input codeset artifact with a selector configured to match a very specific codeset 
        and contents mounted in custom path
      artifact:
        store: gitea
        project: my-project
        name: my-codeset
        version: v1.3
      passByValue:
        toPath: /codesets/two
    example-model-input-one:
      type: model
      description: |
        example input model artifact with a label selector matching several models and value passed by reference (URL) 
      optional: false
      artifact:
        store: mlflow
        project: my-project
        minCount: 1
        maxCount: -1 # unbounded
      labels:
        experiment: first
      passByReference:
    example-opaque-input-one:
      type: opaque
      description: |
        example input opaque artifact of type git passed by reference (URL)
      optional: false
      passByReference:
      storeType: git
    example-opaque-input-two:
      type: opaque
      description: |
        example input opaque artifact with custom composability labels passed by content in a custom path
      optional: false
      passByValue:
        toPath: /dataset/two
      labels:
        dataset: sample
        format: csv

  outputs:
    example-parameter-three:
      type: parameter
      description: |
        example optional output parameter using the default value passing (default output path)
      optional: true
      defaultValue: example-value
    example-parameter-four:
      type: parameter
      description: |
        example output parameter with custom composability labels and custom value passing path
      optional: false
      provides:
        field: name
        parameter-type: string
        length: "10"
      passByValue:
        fromPath: /custompath/parameter-four.txt
    example-codeset-output-three:
      type: codeset
      description: |
        example output codeset artifact with a label matching selector and default value passing config
        (codeset contents are generated as a dir under the default input path).
      optional: false
      labels:
        mlflow: project
        ml-library: "sklearn|pytorch"
        gpu-acceleration: "cuda"
    example-codeset-output-four:
      type: codeset
      description: |
        example output codeset artifact generating a very specific codeset 
        and contents passed in a custom output path
      artifact:
        store: gitea
        project: my-project
        name: my-codeset
        version: staging
      passByValue:
        toPath: /codesets/two
    example-model-output-two:
      type: model
      description: |
        example output model artifact generating several models and value passed by reference (list of URLs)
        in the default output path
      optional: false
      artifact:
        store: mlflow
        project: my-project
      labels:
        experiment: first
      passByReference:
    example-opaque-output-three:
      type: opaque
      description: |
        example output opaque artifact of type git passed by reference (URL)
      optional: false
      passByReference:
      artifact:
        storeType: git
    example-opaque-output-four:
      type: opaque
      description: |
        example output opaque artifact with custom composability labels passed by content in a custom path  
      optional: false
      passByValue:
        fromPath: /dataset/two
      labels:
        dataset: sample
        format: csv
```