package ystruct

type WorkloadInfo struct {
	WorkloadName     string `json:"workloadname"`
	YAML             string `json:"yaml"`
	Metadata         string `json:"metadata"`
	CreatedTimestamp string `json:"createdtimestamp"`
}

// API 요청에 사용할 데이터 구조체
type RequestData struct {
	Yaml      string                 `json:"yaml"`
	Metadata  map[string]interface{} `json:"metadata"` // 동적 JSON 필드를 처리
	Timestamp string                 `json:"timestamp"`
}

type Workflow struct {
	APIVersion string   `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string   `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec       Spec     `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type Metadata struct {
	GenerateName string            `json:"generateName,omitempty" yaml:"generateName,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type Spec struct {
	Entrypoint         string     `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Templates          []Template `json:"templates,omitempty" yaml:"templates,omitempty"`
	Arguments          Arguments  `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	ServiceAccountName string     `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
}

type Template struct {
	Name         string     `json:"name,omitempty" yaml:"name,omitempty"`
	Container    *Container `json:"container,omitempty" yaml:"container,omitempty"`
	Metadata     *Metadata  `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	NodeSelector NodeSelect `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	DAG          *DAG       `json:"dag,omitempty" yaml:"dag,omitempty"`
}

type ContainerResources struct {
	Limits   map[string]string `json:"limits,omitempty" yaml:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty" yaml:"requests,omitempty"`
}

type DAG struct {
	Tasks []Task `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

type Task struct {
	Name         string   `json:"name,omitempty" yaml:"name,omitempty"`
	Template     string   `json:"template,omitempty" yaml:"template,omitempty"`
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

type Arguments struct {
	Parameters []interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type NodeSelect struct {
	Node string `json:"kubernetes.io/hostname,omitempty" yaml:"kubernetes.io/hostname,omitempty"`
}
