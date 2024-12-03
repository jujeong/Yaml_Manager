package ystruct

type ReqResource struct {
	Version string  `json:"version,omitempty" yaml:"version,omitempty"` // 여기에서 json 태그 변경
	Request Request `json:"request,omitempty" yaml:"request,omitempty"`
}

type Request struct {
	Name       string      `json:"name,omitempty" yaml:"name,omitempty"`
	ID         string      `json:"id,omitempty" yaml:"id,omitempty"`
	Date       string      `json:"date,omitempty" yaml:"date,omitempty"`
	Containers []Container `json:"containers,omitempty" yaml:"containers,omitempty"`
	Attribute  Attribute   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
}

type Attribute struct {
	WorkloadType     string  `json:"workloadType,omitempty" yaml:"workloadType,omitempty"`
	IsCronJob        bool    `json:"isCronJob,omitempty" yaml:"isCronJob,omitempty"`
	DevOpsType       string  `json:"devOpsType,omitempty" yaml:"devOpsType,omitempty"`
	CudaVersion      float64 `json:"cudaVersion,omitempty" yaml:"cudaVersion,omitempty"`
	GPUDriverVersion float64 `json:"gpuDriverVersion,omitempty" yaml:"gpuDriverVersion,omitempty"`
	WorkloadFeature  string  `json:"workloadFeature,omitempty" yaml:"workloadFeature,omitempty"`
	UserID           string  `json:"userId,omitempty" yaml:"userId,omitempty"`
	Yaml             string  `json:"yaml,omitempty" yaml:"yaml,omitempty"`
}

type RespResource struct {
	Response Response `json:"result,omitempty" yaml:"result,omitempty"`
}

type Response struct {
	ID               string      `json:"id,omitempty" yaml:"id,omitempty"`
	Date             string      `json:"date,omitempty" yaml:"date,omitempty"`
	Containers       []Container `json:"containers,omitempty" yaml:"containers,omitempty"`
	Cluster          string      `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	PriorityClass    string      `json:"priorityClass,omitempty" yaml:"priorityClass,omitempty"`
	Priority         string      `json:"priority,omitempty" yaml:"priority,omitempty"`
	PreemptionPolicy string      `json:"preemptionPolicy,omitempty" yaml:"preemptionPolicy,omitempty"`
}

type Result struct {
	Cluster          string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Node             string `json:"node,omitempty" yaml:"node,omitempty"`
	PriorityClass    string `json:"priorityClass,omitempty" yaml:"priorityClass,omitempty"`
	Priority         string `json:"priority,omitempty" yaml:"priority,omitempty"`
	PreemptionPolicy string `json:"preemptionPolicy,omitempty" yaml:"preemptionPolicy,omitempty"`
}
