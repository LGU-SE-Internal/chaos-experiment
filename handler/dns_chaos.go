package handler

import (
	"fmt"
	"strconv"

	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

// DNSErrorSpec defines the DNS error chaos injection parameters
type DNSErrorSpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	Namespace   int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx int `range:"0-0" dynamic:"true" description:"DNS Endpoint Index"`
}

func (s *DNSErrorSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := TargetNamespace
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	endpoints, err := resourcelookup.GetAllDNSEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.ErrorAction

	return controllers.CreateDnsChaos(cli, ns, serviceName, action, []string{endpointPair.Domain}, duration)
}

// DNSRandomSpec defines the DNS random chaos injection parameters
type DNSRandomSpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	Namespace   int `range:"0-0" dynamic:"true" description:"String"`
	EndpointIdx int `range:"0-0" dynamic:"true" description:"DNS Endpoint Index"`
}

func (s *DNSRandomSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := TargetNamespace
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	endpoints, err := resourcelookup.GetAllDNSEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS endpoints: %w", err)
	}

	if s.EndpointIdx < 0 || s.EndpointIdx >= len(endpoints) {
		return "", fmt.Errorf("endpoint index out of range: %d (max: %d)", s.EndpointIdx, len(endpoints)-1)
	}

	endpointPair := endpoints[s.EndpointIdx]
	serviceName := endpointPair.AppName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.RandomAction

	return controllers.CreateDnsChaos(cli, ns, serviceName, action, []string{endpointPair.Domain}, duration)
}
