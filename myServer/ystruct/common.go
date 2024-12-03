package ystruct

type Container struct {
	Name      string    `json:"name,omitempty" yaml:"name,omitempty"`
	Resources Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	Cluster   string    `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Node      string    `json:"node,omitempty" yaml:"node,omitempty"`
}

type Resources struct {
	Requests ResourceDetails `json:"requests,omitempty" yaml:"requests,omitempty"`
	Limits   ResourceDetails `json:"limits,omitempty" yaml:"limits,omitempty"`
}

type ResourceDetails struct {
	NvidiaGPU        string `json:"nvidia.com/gpu,omitempty" yaml:"nvidia.com/gpu,omitempty"`
	CPU              string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory           string `json:"memory,omitempty" yaml:"memory,omitempty"`
	GPU              string `json:"gpu,omitempty" yaml:"gpu,omitempty"`
	EphemeralStorage string `json:"ephemeral-storage,omitempty" yaml:"ephemeral-storage,omitempty"`
}
