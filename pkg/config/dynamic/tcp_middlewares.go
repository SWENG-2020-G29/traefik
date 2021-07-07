package dynamic

// +k8s:deepcopy-gen=true

// TCPMiddleware holds the TCPMiddleware configuration.
type TCPMiddleware struct {
	IPWhiteList  *TCPIPWhiteList `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
	WebspaceBoot *WebspaceBoot   `json:"webspaceBoot,omitempty" toml:"webspaceBoot,omitempty" yaml:"webspaceBoot,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPIPWhiteList holds the TCP ip white list configuration.
type TCPIPWhiteList struct {
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
