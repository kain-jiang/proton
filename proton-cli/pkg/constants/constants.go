package constants

const (
	// CRISocketContainerd is the containerd CRI endpoint
	CRISocketContainerd = "unix:///var/run/containerd/containerd.sock"
	// CRISocketCRIO is the cri-o CRI endpoint
	CRISocketCRIO = "unix:///var/run/crio/crio.sock"
	// CRISocketDocker is the cri-dockerd CRI endpoint
	CRISocketDocker = "unix:///var/run/cri-dockerd.sock"
	// CRISocketDockerShim is the docker shim cri endpoint
	CRISocketDockerShim = "unix:///var/run/dockershim.sock"

	// DefaultCRISocket defines the default CRI socket
	DefaultCRISocket = CRISocketContainerd
)
