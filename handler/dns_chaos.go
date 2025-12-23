package handler

import (
	"context"
	"fmt"
	"strconv"

	controllers "github.com/LGU-SE-Internal/chaos-experiment/controllers"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

// DNSErrorSpec defines the DNS error chaos injection parameters
type DNSErrorSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	System         int `range:"0-0" dynamic:"true" description:"String"`
	DNSEndpointIdx int `range:"0-0" dynamic:"true" description:"DNS Endpoint Index"`
}

func (s *DNSErrorSpec) Create(cli cli.Client, opts ...Option) (string, error) {
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

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllDNSEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS endpoints: %w", err)
	}

	if s.DNSEndpointIdx < 0 || s.DNSEndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.DNSEndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.DNSEndpointIdx]
	serviceName := endpointPair.AppName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.ErrorAction

	return controllers.CreateDnsChaos(cli, ctx, ns, serviceName, action, []string{endpointPair.Domain}, duration, annotations, labels)
}

// DNSRandomSpec defines the DNS random chaos injection parameters
type DNSRandomSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	System         int `range:"0-0" dynamic:"true" description:"String"`
	DNSEndpointIdx int `range:"0-0" dynamic:"true" description:"DNS Endpoint Index"`
}

func (s *DNSRandomSpec) Create(cli cli.Client, opts ...Option) (string, error) {
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

	endpoints, err := resourcelookup.GetSystemCache(system).GetAllDNSEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS endpoints: %w", err)
	}

	if s.DNSEndpointIdx < 0 || s.DNSEndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.DNSEndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.DNSEndpointIdx]
	serviceName := endpointPair.AppName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.RandomAction

	return controllers.CreateDnsChaos(cli, ctx, ns, serviceName, action, []string{endpointPair.Domain}, duration, annotations, labels)
}
