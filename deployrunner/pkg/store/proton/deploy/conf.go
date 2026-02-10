package deploy

// CoreConfig deployCore信息
type CoreConfig struct {
	Namespace string `json:"namespace"`
	Database  string `json:"database"`
}

func (c *CoreConfig) ToMapValues() map[string]any {
	return map[string]any{
		"namespace": c.Namespace,
		"database":  c.Database,
	}
}
