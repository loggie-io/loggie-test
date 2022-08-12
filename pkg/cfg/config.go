package cfg

type Config struct {
	Resources map[string]ResourceWrapper `yaml:"resources,omitempty"`
	Cases     map[string]CaseWrapper     `yaml:"cases,omitempty"`
}

type ResourceWrapper struct {
	ResourceIns map[string]interface{} `yaml:",inline"`
}

type CaseWrapper struct {
	CaseConfig map[string]interface{} `yaml:",inline"`
}
