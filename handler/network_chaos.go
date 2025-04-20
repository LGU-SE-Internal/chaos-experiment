package handler

import (
	"fmt"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/chaos"
	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

// Map for Direction conversion from int to chaos-mesh Direction type
var directionMap = map[int]chaosmeshv1alpha1.Direction{
	1: chaosmeshv1alpha1.To,
	2: chaosmeshv1alpha1.From,
	3: chaosmeshv1alpha1.Both,
}

// Convert int direction code to chaos-mesh Direction
func getDirection(directionCode int) chaosmeshv1alpha1.Direction {
	if direction, ok := directionMap[directionCode]; ok {
		return direction
	}
	return chaosmeshv1alpha1.To // Default to "To" direction
}

// Helper function to validate and get network pair from index
func getNetworkPairByIndex(networkPairIdx int) (*resourcelookup.AppNetworkPair, error) {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		return nil, fmt.Errorf("failed to get network pairs: %w", err)
	}

	if networkPairIdx < 0 || networkPairIdx >= len(networkPairs) {
		return nil, fmt.Errorf("network pair index out of range: %d (max: %d)",
			networkPairIdx, len(networkPairs)-1)
	}

	return &networkPairs[networkPairIdx], nil
}

// NetworkPartitionSpec defines network partition chaos parameters
type NetworkPartitionSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	Namespace      int `range:"0-0" dynamic:"true" description:"String"`
	NetworkPairIdx int `range:"0-0" dynamic:"true" description:"Flattened network pair index"`
	Direction      int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkPartitionSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := GetTargetNamespace(s.Namespace)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	pair, err := getNetworkPairByIndex(s.NetworkPairIdx)
	if err != nil {
		return "", err
	}

	sourceName := pair.SourceService
	targetName := pair.TargetService

	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))
	direction := getDirection(s.Direction)

	// Create network partition between the source and target services
	optss := []chaos.OptNetworkChaos{
		chaos.WithNetworkTargetAndDirection(ns, targetName, direction),
	}

	return controllers.CreateNetworkChaos(cli, ns, sourceName,
		chaosmeshv1alpha1.PartitionAction, duration, labels, optss...)
}

// NetworkDelaySpec defines network delay chaos parameters
type NetworkDelaySpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	Namespace      int `range:"0-0" dynamic:"true" description:"String"`
	NetworkPairIdx int `range:"0-0" dynamic:"true" description:"Flattened network pair index"`
	Latency        int `range:"1-2000" description:"Latency in milliseconds"`
	Correlation    int `range:"0-100" description:"Correlation percentage"`
	Jitter         int `range:"0-1000" description:"Jitter in milliseconds"`
	Direction      int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkDelaySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := GetTargetNamespace(s.Namespace)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	pair, err := getNetworkPairByIndex(s.NetworkPairIdx)
	if err != nil {
		return "", err
	}

	sourceName := pair.SourceService
	targetName := pair.TargetService

	// Convert int values to appropriate string format
	latency := fmt.Sprintf("%dms", s.Latency)
	correlation := fmt.Sprintf("%d", s.Correlation)
	jitter := fmt.Sprintf("%dms", s.Jitter)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))
	direction := getDirection(s.Direction)

	// Create network delay between the source and target services
	optss := []chaos.OptNetworkChaos{
		chaos.WithNetworkTargetAndDirection(ns, targetName, direction),
		chaos.WithNetworkDelay(latency, correlation, jitter),
	}

	return controllers.CreateNetworkChaos(cli, ns, sourceName,
		chaosmeshv1alpha1.DelayAction, duration, labels, optss...)
}

// NetworkLossSpec defines network packet loss chaos parameters
type NetworkLossSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	Namespace      int `range:"0-0" dynamic:"true" description:"String"`
	NetworkPairIdx int `range:"0-0" dynamic:"true" description:"Flattened network pair index"`
	Loss           int `range:"1-100" description:"Packet loss percentage"`
	Correlation    int `range:"0-100" description:"Correlation percentage"`
	Direction      int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkLossSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := GetTargetNamespace(s.Namespace)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	pair, err := getNetworkPairByIndex(s.NetworkPairIdx)
	if err != nil {
		return "", err
	}

	sourceName := pair.SourceService
	targetName := pair.TargetService

	// Convert int values to appropriate string format
	loss := fmt.Sprintf("%d", s.Loss)
	correlation := fmt.Sprintf("%d", s.Correlation)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))
	direction := getDirection(s.Direction)

	// Create network loss between the source and target services
	optss := []chaos.OptNetworkChaos{
		chaos.WithNetworkTargetAndDirection(ns, targetName, direction),
		chaos.WithNetworkLoss(loss, correlation),
	}

	return controllers.CreateNetworkChaos(cli, ns, sourceName,
		chaosmeshv1alpha1.LossAction, duration, labels, optss...)
}

// NetworkDuplicateSpec defines network packet duplication chaos parameters
type NetworkDuplicateSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	Namespace      int `range:"0-0" dynamic:"true" description:"String"`
	NetworkPairIdx int `range:"0-0" dynamic:"true" description:"Flattened network pair index"`
	Duplicate      int `range:"1-100" description:"Packet duplication percentage"`
	Correlation    int `range:"0-100" description:"Correlation percentage"`
	Direction      int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkDuplicateSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := GetTargetNamespace(s.Namespace)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	pair, err := getNetworkPairByIndex(s.NetworkPairIdx)
	if err != nil {
		return "", err
	}

	sourceName := pair.SourceService
	targetName := pair.TargetService

	// Convert int values to appropriate string format
	duplicate := fmt.Sprintf("%d", s.Duplicate)
	correlation := fmt.Sprintf("%d", s.Correlation)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))
	direction := getDirection(s.Direction)

	// Create network duplicate between the source and target services
	optss := []chaos.OptNetworkChaos{
		chaos.WithNetworkTargetAndDirection(ns, targetName, direction),
		chaos.WithNetworkDuplicate(duplicate, correlation),
	}

	return controllers.CreateNetworkChaos(cli, ns, sourceName,
		chaosmeshv1alpha1.DuplicateAction, duration, labels, optss...)
}

// NetworkCorruptSpec defines network packet corruption chaos parameters
type NetworkCorruptSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	Namespace      int `range:"0-0" dynamic:"true" description:"String"`
	NetworkPairIdx int `range:"0-0" dynamic:"true" description:"Flattened network pair index"`
	Corrupt        int `range:"1-100" description:"Packet corruption percentage"`
	Correlation    int `range:"0-100" description:"Correlation percentage"`
	Direction      int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkCorruptSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := GetTargetNamespace(s.Namespace)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	pair, err := getNetworkPairByIndex(s.NetworkPairIdx)
	if err != nil {
		return "", err
	}

	sourceName := pair.SourceService
	targetName := pair.TargetService

	// Convert int values to appropriate string format
	corrupt := fmt.Sprintf("%d", s.Corrupt)
	correlation := fmt.Sprintf("%d", s.Correlation)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))
	direction := getDirection(s.Direction)

	// Create network corrupt between the source and target services
	optss := []chaos.OptNetworkChaos{
		chaos.WithNetworkTargetAndDirection(ns, targetName, direction),
		chaos.WithNetworkCorrupt(corrupt, correlation),
	}

	return controllers.CreateNetworkChaos(cli, ns, sourceName,
		chaosmeshv1alpha1.CorruptAction, duration, labels, optss...)
}

// NetworkBandwidthSpec defines network bandwidth limit chaos parameters
type NetworkBandwidthSpec struct {
	Duration       int `range:"1-60" description:"Time Unit Minute"`
	Namespace      int `range:"0-0" dynamic:"true" description:"String"`
	NetworkPairIdx int `range:"0-0" dynamic:"true" description:"Flattened network pair index"`
	Rate           int `range:"1-1000000" description:"Bandwidth rate in kbps"`
	Limit          int `range:"1-10000" description:"Number of bytes that can be queued"`
	Buffer         int `range:"1-10000" description:"Maximum amount of bytes available instantaneously"`
	Direction      int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkBandwidthSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}
	ns := GetTargetNamespace(s.Namespace)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	pair, err := getNetworkPairByIndex(s.NetworkPairIdx)
	if err != nil {
		return "", err
	}

	sourceName := pair.SourceService
	targetName := pair.TargetService

	// Convert rate from kbps to string with unit
	rate := fmt.Sprintf("%dkbps", s.Rate)
	limit := uint32(s.Limit)
	buffer := uint32(s.Buffer)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))
	direction := getDirection(s.Direction)

	// Create network bandwidth between the source and target services
	optss := []chaos.OptNetworkChaos{
		chaos.WithNetworkTargetAndDirection(ns, targetName, direction),
		chaos.WithNetworkBandwidth(rate, limit, buffer),
	}

	return controllers.CreateNetworkChaos(cli, ns, sourceName,
		chaosmeshv1alpha1.BandwidthAction, duration, labels, optss...)
}
