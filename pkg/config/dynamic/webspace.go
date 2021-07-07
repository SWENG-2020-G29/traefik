package dynamic

// WebspaceBoot holds the WebspaceBoot configuration
type WebspaceBoot struct {
	URL      string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty" export:"true"`
	IAMToken string `json:"iamToken,omitempty" toml:"iamToken,omitempty" yaml:"iamToken,omitempty" export:"true"`
	UserID   int    `json:"userID,omitempty" toml:"userID,omitempty" yaml:"userID,omitempty" export:"true"`
}
