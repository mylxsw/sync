package meta

type SyncWatch struct {
	From  string   `json:"from" yaml:"from"`
	Token string   `json:"token,omitempty" yaml:"token,omitempty"`
	Names []string `json:"names" yaml:"names"`
}
