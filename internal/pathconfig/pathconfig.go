package pathconfig

// PathConfig holds the canonical base path and an optional current path field to be used
// by the appropriate party i.e the templating package
type PathConfig struct {
	basePath    string
	currentPath string
}

// NewConfig creates and returns the address of a new config object
func NewConfig(basePath string) *PathConfig {
	return &PathConfig{
		basePath:    basePath,
		currentPath: "",
	}
}

// GetBasePath returns the specified base path
func (c *PathConfig) GetBasePath() string {
	return c.basePath
}

// GetCurrentPath returns the specified current path; "" if none is specified
func (c *PathConfig) GetCurrentPath() string {
	return c.currentPath
}

// SetCurrentPath sets the specified current path
func (c *PathConfig) SetCurrentPath(currentPath string) {
	c.currentPath = currentPath
}
