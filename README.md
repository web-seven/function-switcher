# function-switcher
[![CI](https://github.com/kndpio/function-switcher/actions/workflows/ci.yml/badge.svg)](https://github.com/kndpio/function-switcher/actions/workflows/ci.yml)

Function Switcher is a Crossplane function that enables Composition users to enable or disable creation of resources, without update schema of XRD, just using annotations like in on this example:
```yaml
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
  annotations:
    ## Annotations for enable/disable resources of Composition.
    switcher.fn.kndp.io/disabled: "resourceTwo,resourceThree"
    switcher.fn.kndp.io/enabled: "resourceOne"
spec:
  resourceOne:
    field1: "one"
    field2: "two"
  resourceTwo:
    field1: "three"
    field2: "four"
  resourceThree:
    field1: "five"
    field2: "six"
```
## Installation:

1. Create function using function package from registry: 
```yaml
apiVersion: pkg.crossplane.io/v1beta1
kind: Function
metadata:
  name: function-switcher
spec:
  package: xpkg.upbound.io/kndp/function-switcher:v0.0.1
```

2. Setup function required to add it after resources generation function like `function-patch-and-transform`:
```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: function-switcher
spec:
  mode: Pipeline
  pipeline:
  - step: patch-and-transform
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:

      - name: resourceOne
        base:
          apiVersion: example.crossplane.io/v1
          kind: Resource
          spec:
            field1: ""
            field2: ""
        patches:
          - type: FromCompositeFieldPath
            fromFieldPath: "spec.resourceOne.field1"
            toFieldPath: "spec.field1"
          - type: FromCompositeFieldPath
            fromFieldPath: "spec.resourceOne.field2"
            toFieldPath: "spec.field2"

      - name: resourceTwo
        base:
          apiVersion: example.crossplane.io/v1
          kind: Resource
          spec:
            field1: ""
            field2: ""
        patches:
          - type: FromCompositeFieldPath
            fromFieldPath: "spec.resourceTwo.field1"
            toFieldPath: "spec.field1"
          - type: FromCompositeFieldPath
            fromFieldPath: "spec.resourceTwo.field2"
            toFieldPath: "spec.field2"

      - name: resourceThree
        base:
          apiVersion: example.crossplane.io/v1
          kind: Resource
          spec:
            field1: ""
            field2: ""
        patches:
          - type: FromCompositeFieldPath
            fromFieldPath: "spec.resourceThree.field1"
            toFieldPath: "spec.field1"
          - type: FromCompositeFieldPath
            fromFieldPath: "spec.resourceThree.field2"
            toFieldPath: "spec.field2"
  ## Enable or disable resources from previous step
  - step: enable-disable
    functionRef:
      name: function-switcher

  compositeTypeRef:
    apiVersion: example.crossplane.io/v1
    kind: XR
```
