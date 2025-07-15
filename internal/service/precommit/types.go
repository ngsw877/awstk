package precommit

// Status はpre-commitフックの状態を表す
type Status struct {
	Enabled      bool   `json:"enabled"`
	HooksPath    string `json:"hooks_path,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
