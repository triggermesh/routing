# Triggermesh Content Filter

Triggermesh Content Filter is an addressable Kubernetes object that filters
incoming CloudEvents according to the provided expression. Filter's
Specification consists of only two fields, but they are both required: the
expression and the sink. Events whose content makes the result of the expression
equal to "true" are dispatched to the sink. Filter's expression is a combination
of Google [CEL](https://github.com/google/cel-spec) with the inline types
assertion required by [GJSON](https://github.com/tidwall/gjson). Here is an
example of JSON payload and the expression that access different payload's
paths:

CE data
```json
{
    "id": {
        "first":5,
        "second":3
    },
    "foo": "bar",
    "options": [true, false]
}
```

Possible expression
```
($id.first.(int64) + $id.second.(int64) >= 8) || $foo.(string) == "bar" && $options.0.(bool) == true 
```

The expression variables are defined as `$.<json path>.(type)`, where "type" can be any of the following:
- bool
- int64
- uint64
- double (Go's `float64`)
- string


Example of Filter Object:

```yaml
apiVersion: routing.triggermesh.io/v1alpha1
kind: Filter
metadata:
  name: filter-test
spec:
  expression: $id.first.(int64) + $id.second.(int64) == 8
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: sockeye
```
<i>CloudEvent will be sent to Sockeye service only if event's payload contains
`id.first` and `id.second` paths and the sum of their values is equal to 8.</i>


## Installation

Filter can be compiled and deployed from source with [ko](https://github.com/google/ko):

```
ko apply -f ./config
```

You can verify that it's installed by checking that the controller is running:

```
$ kubectl -n filter get pods -l app=filter-controller
NAME                                 READY   STATUS    RESTARTS   AGE
filter-controller-6ff9bfb568-8w977   1/1     Running   0          2m
```

A custom resource of kind `Filter` can now be created, check a
[sample](config/samples/filter.yaml).


## Performance

TBD

## Support

We would love your feedback and help on these sources, so don't hesitate to let us know what is wrong and how we could improve them, just file an [issue](https://github.com/triggermesh/filter/issues/new) or join those of use who are maintaining them and submit a [PR](https://github.com/triggermesh/filter/compare)

## Commercial Support

TriggerMesh Inc supports this project commercially, email info@triggermesh.com to get more details.

## Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)

