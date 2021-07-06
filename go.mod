module github.com/triggermesh/routing

go 1.15

require (
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/google/cel-go v0.7.2
	github.com/google/go-cmp v0.5.6
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	go.uber.org/zap v1.17.0
	google.golang.org/genproto v0.0.0-20210416161957-9910b6c460de
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/code-generator v0.20.7
	knative.dev/eventing v0.24.0
	knative.dev/networking v0.0.0-20210512050647-ace2d3306f0b
	knative.dev/pkg v0.0.0-20210701025203-30f9568e894e
	knative.dev/serving v0.23.0
)
