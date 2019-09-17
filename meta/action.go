package meta

// SyncAction 文件同步前置后置任务
type SyncAction struct {
	Action string `json:"action,omitempty" yaml:"action,omitempty"`
	When   string `json:"when,omitempty" yaml:"when,omitempty"`

	// Match         string `json:"match,omitempty" yaml:"match,omitempty"`
	// Replace       string `json:"replace,omitempty" yaml:"replace,omitempty"`

	// --- command ---
	Command       string `json:"command,omitempty" yaml:"command,omitempty"`
	ParseTemplate bool   `json:"parse_template,omitempty" yaml:"parse_template,omitempty"`
	Timeout       string `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// --- dingding ---
	Body  string `json:"body,omitempty" yaml:"body,omitempty"`
	Token string `json:"token,omitempty" yaml:"token,omitempty"`
}
