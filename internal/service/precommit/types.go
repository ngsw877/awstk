package precommit

// status はpre-commitフックの状態を表す
type status struct {
	Enabled      bool   `json:"enabled"`
	HooksPath    string `json:"hooks_path,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
