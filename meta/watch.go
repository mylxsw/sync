package meta

type SyncRelay struct {
	From  string   `json:"from,omitempty" yaml:"from,omitempty"`
	Token string   `json:"token,omitempty" yaml:"token,omitempty"`
	Names []string `json:"names" yaml:"names"`
}
