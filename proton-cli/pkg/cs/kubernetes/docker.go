package kubernetes

type Ulimit struct {
	Hard int    `json:"Hard"`
	Name string `json:"Name"`
	Soft int    `json:"Soft"`
}

// Define the struct for default ulimits.
type DefaultUlimits map[string]Ulimit

// Define the struct for log options.
type LogOpts struct {
	MaxFile string `json:"max-file"`
	MaxSize string `json:"max-size"`
}

// Define the main struct that holds all the configuration data.
type DockerConfig struct {
	Bip                    string         `json:"bip"`
	DataRoot               string         `json:"data-root"`
	DefaultUlimits         DefaultUlimits `json:"default-ulimits"`
	ExecOpts               []string       `json:"exec-opts"`
	InsecureRegistries     []string       `json:"insecure-registries"`
	LogDriver              string         `json:"log-driver"`
	LogOpts                LogOpts        `json:"log-opts"`
	MaxConcurrentDownloads int            `json:"max-concurrent-downloads"`
	MaxConcurrentUploads   int            `json:"max-concurrent-uploads"`
	Runtimes               map[string]any `json:"runtimes,omitempty"`
	DefaultRuntime         string         `json:"default-runtime,omitempty"`
}
