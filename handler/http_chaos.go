package handler

import (
	"context"
	"fmt"
	"strconv"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/chaos"
	controllers "github.com/LGU-SE-Internal/chaos-experiment/controllers"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

// HTTPRequestAbortSpec defines HTTP request abort chaos
type HTTPRequestAbortSpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	System      int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
}

func (s *HTTPRequestAbortSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	abort := true

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
		chaos.WithAbort(&abort),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "request-abort", duration, annotations, labels, optss...)
}

// HTTPResponseAbortSpec defines HTTP response abort chaos
type HTTPResponseAbortSpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	System      int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
}

func (s *HTTPResponseAbortSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	abort := true

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
		chaos.WithAbort(&abort),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "response-abort", duration, annotations, labels, optss...)
}

// HTTPRequestDelaySpec defines HTTP request delay chaos injection
type HTTPRequestDelaySpec struct {
	Duration      int `range:"1-60" description:"Time Unit Minute"`
	System        int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx   int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
	DelayDuration int `range:"10-5000" description:"Delay in milliseconds"`
}

func (s *HTTPRequestDelaySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	delay := fmt.Sprintf("%dms", s.DelayDuration)

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
		chaos.WithDelay(&delay),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "request-delay", duration, annotations, labels, optss...)
}

// HTTPResponseDelaySpec defines HTTP response delay chaos injection
type HTTPResponseDelaySpec struct {
	Duration      int `range:"1-60" description:"Time Unit Minute"`
	System        int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx   int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
	DelayDuration int `range:"10-5000" description:"Delay in milliseconds"`
}

func (s *HTTPResponseDelaySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	delay := fmt.Sprintf("%dms", s.DelayDuration)

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
		chaos.WithDelay(&delay),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "response-delay", duration, annotations, labels, optss...)
}

// ReplaceBodyType for HTTP response body replacement
type ReplaceBodyType int

const (
	EmptyBody ReplaceBodyType = iota
	RandomBody
)

// HTTPResponseReplaceBodySpec defines HTTP response body replacement chaos
type HTTPResponseReplaceBodySpec struct {
	Duration    int             `range:"1-60" description:"Time Unit Minute"`
	System      int             `range:"0-0" dynamic:"true" description:"Namespace Index (0-based)"`
	EndpointIdx int             `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
	BodyType    ReplaceBodyType `range:"0-1" description:"Body Type (0=Empty, 1=Random)"`
}

func (s *HTTPResponseReplaceBodySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
	}

	// Add body replacement based on type
	if s.BodyType == EmptyBody {
		optss = append(optss, chaos.WithReplaceBody([]byte("")))
	} else {
		optss = append(optss, chaos.WithRandomReplaceBody())
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "response-replace-body", duration, annotations, labels, optss...)
}

// HTTPResponsePatchBodySpec defines HTTP response body patching chaos
type HTTPResponsePatchBodySpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	System      int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
}

func (s *HTTPResponsePatchBodySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
		chaos.WithPatchBody(`{"foo": "bar"}`),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "response-patch-body", duration, annotations, labels, optss...)
}

// HTTPRequestReplacePathSpec defines HTTP request path replacement chaos
type HTTPRequestReplacePathSpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	System      int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
}

func (s *HTTPRequestReplacePathSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	newPath := "/api/v2/"

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
		chaos.WithReplacePath(&newPath),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "request-replace-path", duration, annotations, labels, optss...)
}

// HTTPRequestReplaceMethodSpec defines HTTP request method replacement chaos
type HTTPRequestReplaceMethodSpec struct {
	Duration      int `range:"1-60" description:"Time Unit Minute"`
	System        int `range:"0-0" dynamic:"true" description:"Namespace Index (0-based)"`
	EndpointIdx   int `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
	ReplaceMethod int `range:"0-6" description:"HTTP Method index (filtered, excluding original method)"`
}

func (s *HTTPRequestReplaceMethodSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Get filtered method excluding the original endpoint method
	filteredMethod := GetFilteredHTTPMethodByIndex(endpoint.Method, s.ReplaceMethod)
	newMethod := GetHTTPMethodName(filteredMethod)

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
		chaos.WithReplaceMethod(&newMethod),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "request-replace-method", duration, annotations, labels, optss...)
}

// HTTPResponseReplaceCodeSpec defines HTTP response status code replacement chaos
type HTTPResponseReplaceCodeSpec struct {
	Duration    int            `range:"1-60" description:"Time Unit Minute"`
	System      int            `range:"0-0" dynamic:"true" description:"Namespace Index (0-based)"`
	EndpointIdx int            `range:"0-0" dynamic:"true" description:"Flattened HTTP Endpoint Index"`
	StatusCode  HTTPStatusCode `range:"0-9" description:"HTTP Status Code to replace with"`
}

func (s *HTTPResponseReplaceCodeSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllHTTPEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	endpoint := &HTTPEndpoint{
		ServiceName:   serviceName,
		Route:         endpointPair.Route,
		Method:        endpointPair.Method,
		TargetService: endpointPair.ServerAddress,
		Port:          endpointPair.ServerPort,
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	code := GetHTTPStatusCode(s.StatusCode)

	// Create options with endpoint-specific values
	optss := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
		chaos.WithReplaceCode(&code),
	}

	// Add common HTTP options (port, path and method)
	optss = AddCommonHTTPOptions(endpoint, optss)

	return controllers.CreateHTTPChaos(cli, ctx, ns, serviceName, "response-replace-code", duration, annotations, labels, optss...)
}
