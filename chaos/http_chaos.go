package chaos

import (
	"errors"
	"fmt"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
)

func NewHttpChaos(opts ...OptChaos) (*chaosmeshv1alpha1.HTTPChaos, error) {
	config := ConfigChaos{}
	for _, opt := range opts {
		if opt != nil {
			opt(&config)
		}
	}

	if config.Name == "" {
		return nil, errors.New("the resource name is required")
	}
	if config.Namespace == "" {
		return nil, errors.New("the namespace is required")
	}
	if config.HttpChaos == nil {
		return nil, errors.New("httpChaos is required")
	}

	httpChaos := chaosmeshv1alpha1.HTTPChaos{}
	httpChaos.Name = config.Name
	httpChaos.Namespace = config.Namespace
	config.HttpChaos.DeepCopyInto(&httpChaos.Spec)

	if config.Labels != nil {
		httpChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		httpChaos.Annotations = config.Annotations
	}

	return &httpChaos, nil
}


type OptHTTPChaos func(opt *chaosmeshv1alpha1.HTTPChaosSpec)

func WithTarget(target chaosmeshv1alpha1.PodHttpChaosTarget) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.Target = target
	}
}

func WithPort(port int32) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.Port = port
	}
}

func WithPath(path *string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.Path = path
	}
}

func WithMethod(method *string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.Method = method
	}
}

func WithCode(code *int32) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.Code = code
	}
}

func WithRequestHeaders(headers map[string]string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.RequestHeaders = headers
	}
}

func WithResponseHeaders(headers map[string]string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.ResponseHeaders = headers
	}
}

func WithDuration(duration *string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.Duration = duration
	}
}

func WithAbort(abort *bool) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.PodHttpChaosActions.Abort = abort
	}
}

func WithDelay(delay *string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		opt.PodHttpChaosActions.Delay = delay
	}
}

func WithReplace(replace *chaosmeshv1alpha1.PodHttpChaosReplaceActions) OptHTTPChaos {
    return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
        if opt.PodHttpChaosActions.Replace == nil {
            opt.PodHttpChaosActions.Replace = &chaosmeshv1alpha1.PodHttpChaosReplaceActions{}
        }
        if replace != nil {
            opt.PodHttpChaosActions.Replace = replace
        }
    }
}


func WithPatch(patch *chaosmeshv1alpha1.PodHttpChaosPatchActions) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		if opt.PodHttpChaosActions.Patch == nil {
			opt.PodHttpChaosActions.Patch = &chaosmeshv1alpha1.PodHttpChaosPatchActions{}
		}
		if patch != nil {
			opt.PodHttpChaosActions.Patch = patch
		}
	}
}

func WithPatchBody(body string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithPatch(nil)(opt)
		opt.PodHttpChaosActions.Patch.Body = &chaosmeshv1alpha1.PodHttpChaosPatchBodyAction{
			Type: "JSON",
			Value: body,
		}
	}
}

func WithPatchQueries(queries [][]string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithPatch(nil)(opt)
		opt.PodHttpChaosActions.Patch.Queries = queries
	}
}

func WithPatchHeaders(headers [][]string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithPatch(nil)(opt)
		opt.PodHttpChaosActions.Patch.Headers = headers
	}
}


func WithReplacePath(path *string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithReplace(nil)(opt)
		opt.PodHttpChaosActions.Replace.Path = path
	}
}

func WithReplaceMethod(method *string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithReplace(nil)(opt)
		opt.PodHttpChaosActions.Replace.Method = method
	}
}

func WithReplaceCode(code *int32) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithReplace(nil)(opt)
		opt.PodHttpChaosActions.Replace.Code = code
	}
}

func WithReplaceBody(body []byte) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithReplace(nil)(opt)
		opt.PodHttpChaosActions.Replace.Body = body
	}
}

func WithReplaceQueries(queries map[string]string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithReplace(nil)(opt)
		opt.PodHttpChaosActions.Replace.Queries = queries
	}
}

func WithReplaceHeaders(headers map[string]string) OptHTTPChaos {
	return func(opt *chaosmeshv1alpha1.HTTPChaosSpec) {
		WithReplace(nil)(opt)
		opt.PodHttpChaosActions.Replace.Headers = headers
	}
}

func GenerateHttpChaosSpec(namespace string, appName string, opts ...OptHTTPChaos) *chaosmeshv1alpha1.HTTPChaosSpec {
	spec := &chaosmeshv1alpha1.HTTPChaosSpec{
		PodSelector: chaosmeshv1alpha1.PodSelector{
			Selector: chaosmeshv1alpha1.PodSelectorSpec{
				GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
					Namespaces:     []string{namespace},
					LabelSelectors: map[string]string{"app": appName},
				},
			},
			Mode: chaosmeshv1alpha1.AllMode,
		},
		Target:              chaosmeshv1alpha1.PodHttpRequest,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(spec)
		}
	}
	return spec
}
func GenerateSetsOfHttpChaosSpec(namespace string, podName string) []chaosmeshv1alpha1.HTTPChaosSpec {
	specs := make([]chaosmeshv1alpha1.HTTPChaosSpec, 0)

	basicSpec := &chaosmeshv1alpha1.HTTPChaosSpec{
		PodSelector: chaosmeshv1alpha1.PodSelector{
			Selector: chaosmeshv1alpha1.PodSelectorSpec{
				GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
					Namespaces: []string{namespace},
				},
				Pods: map[string][]string{
					namespace: {
						podName,
					},
				},
			},
			Mode: chaosmeshv1alpha1.OneMode,
		},
		Target:              chaosmeshv1alpha1.PodHttpRequest, // can change
		PodHttpChaosActions: chaosmeshv1alpha1.PodHttpChaosActions{
			//Abort:   nil,
			//Delay: nil,
			//Replace: nil,
			//Patch:   nil,
		},
		Port:            8080,
		Path:            nil, // for filtering the request
		Method:          nil, // for filtering the request
		Code:            nil, // for filtering the request
		RequestHeaders:  nil, // for filtering the request
		ResponseHeaders: nil, // for filtering the request
		//Duration:        nil,
	}
	for _, target := range []chaosmeshv1alpha1.PodHttpChaosTarget{chaosmeshv1alpha1.PodHttpRequest, chaosmeshv1alpha1.PodHttpResponse} {
		cur := chaosmeshv1alpha1.HTTPChaosSpec{}
		basicSpec.DeepCopyInto(&cur)
		cur.Target = target

		for _, i := range []int{0, 1} {
			switch i {
			case 0:
				cur.PodHttpChaosActions.Abort = pointer.Bool(true)
				specs = append(specs, cur)
			case 1:
				for interval := 1; interval < 10; interval++ {
					cur.PodHttpChaosActions.Abort = nil
					cur.PodHttpChaosActions.Delay = pointer.String(fmt.Sprintf("%ds", interval))
					specs = append(specs, cur)
				}
			case 2:
				//cur.PodHttpChaosActions.Replace =
			case 3:
				//cur.PodHttpChaosActions.Patch =
			}
		}

	}
	return specs
}
