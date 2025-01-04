package handler

type Level string

const (
	Span    Level = "span"
	Service Level = "service"
	Metric  Level = "metric"
	Pod     Level = "pod"
)

type Groudtruth struct {
	Level Level
	Name  string
}

func (s *ContainerKillSpec) GetGroudtruth() []Groudtruth {
	return nil
}

func (s *PodFailureSpec) GetGroudtruth() []Groudtruth {
	return nil
}

func (s *PodKillSpec) GetGroudtruth() []Groudtruth {
	return nil
}

func (s *CPUStressChaosSpec) GetGroudtruth() []Groudtruth {
	return nil
}

func (s *MemoryStressChaosSpec) GetGroudtruth() []Groudtruth {
	return nil
}

func (s *HTTPChaosReplaceSpec) GetGroudtruth() []Groudtruth {
	return nil
}

func (s *HTTPChaosAbortSpec) GetGroudtruth() []Groudtruth {
	return nil
}
