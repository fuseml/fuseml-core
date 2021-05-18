# Runnable definition examples

```
id: example-01_bare-minimum
description: |
  Example of a runnable with minimal definition
author: fuseml-examples
container:
  image: oci-registry.example.io/my-repo:tag
labels:
  function: example
  example: bare-minimum
```


```
id: example-02_parameter_use
description: |
  Example of using input and output parameters
author: fuseml-examples
container:
  image: oci-registry.example.io/my-repo:tag
  env:
    FOO_IN: "{{inputs.foo_in}}"
    FOO_OUT_PATH: "{{outputs.foo_out}}"
  entrypoint: "/usr/bin/run.sh"
  args:
    - execute
    - --bar-in-path
    - "{{inputs.bar_in}}"
    - --bar-out-path
    - "{{outputs.bar_out}}"
input:
  parameters:
    foo_in:
      description: |
        Input parameter foo
    bar_in:
      description: |
        Optional input parameter with custom composability labels and value passed through file
      optional: true
      defaultValue: baz
      path: /inputs/bar.txt
      labels:
        variable: bar
output:
  parameters:
    foo_out:
      description: |
        Output parameter foo
    bar_out:
      description: |
        Optional output parameter with custom composability labels and custom output file location
      optional: true
      defaultValue: baz
      path: /outputs/bar.txt
      labels:
        variable: bar
labels:
  function: example
  example: parameter-sample
```


```
id: example-03_local_artifacts
description: |
  Example of using input and output artifacts passed with the default local provider
author: fuseml-examples
container:
  image: oci-registry.example.io/my-repo:tag
  env:
    FOO_IN: "{{inputs.foo_in}}"
    FOO_OUT_PATH: "{{outputs.foo_out}}"
  entrypoint: "/usr/bin/run.sh"
  args:
    - execute
    - --bar-in-path
    - "{{inputs.bar_in}}"
    - --bar-out-path
    - "{{outputs.bar_out}}"
input:
  artifacts:
    trainer:
      description: |
        Input codeset with training logic
      kind:
        codeset:
          type:
            - code
          function:
            - model-training
          format:
            - MLProject
            - conda 
      path: /workspace/trainer
output:
  artifacts:
    model:
      description: |
        Output machine learning model
      kind:
        model:
          format:
            - SKLearn/PKL
          pretrained: True
          method: unsupervised
          class: regression
      path: /workspace/model
labels:
  function: example
  example: local-artifact-sample
```

