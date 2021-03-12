# Triggermesh Content Filter

Triggermesh Content Filter is an addressable Kubernetes object that filters incoming 
CloudEvents according to the provided expression. Events with the content that makes 
the expression result "true", are forwarded to the Sink. Filter's expression is a 
combination of Google [CEL](https://github.com/google/cel-spec) with the inline types 
assertion required by [GJSON](https://github.com/tidwall/gjson). Here is an example of
JSON payload and the Filter expression:

CE data
```
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

[config/samples](./config/samples) directory contains an example of the Filter with the Sockeye service and Pingsources with the different payloads - deploy the manifest and play with the Filter's expression.


# README is WIP

## Support

We would love your feedback and help on these sources, so don't hesitate to let us know what is wrong and how we could improve them, just file an [issue](https://github.com/triggermesh/filter/issues/new) or join those of use who are maintaining them and submit a [PR](https://github.com/triggermesh/filter/compare)

## Commercial Support

TriggerMesh Inc supports this project commercially, email info@triggermesh.com to get more details.

## Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)

