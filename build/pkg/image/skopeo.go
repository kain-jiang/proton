package image

import (
	"fmt"
	"net/url"

	"k8s.io/utils/exec"
)

type skopeoTransport string

const (
	skopeoTransportDocker skopeoTransport = "docker"
	skopeoTransportOCI    skopeoTransport = "oci"
)

type skopeoCopyImageName struct {
	transport skopeoTransport
	address   string
	name      string
}

func (n skopeoCopyImageName) String() string {
	var transport string
	switch n.transport {
	case skopeoTransportDocker:
		u := url.URL{Scheme: "docker", Host: n.address}
		transport = u.String()
	case skopeoTransportOCI:
		transport = "oci:"
	default:
		transport = ""
	}

	return transport + n.name
}

func formatImageNameOCI(directory string, name string) string {
	return fmt.Sprintf("oci:%s:%s", directory, name)
}

func formatImageNameDocker(host string, ref *Reference) string {
	return fmt.Sprintf("docker://%s/%s", host, ref)
}

func skopeoCopyFromDockerToOCI(executor exec.Interface, host, directory string, ref *Reference, arch string) error {
	src := formatImageNameDocker(host, ref)
	dst := formatImageNameOCI(directory, host+"/"+ref.String())
	return executor.Command("skopeo", "--insecure-policy", "--override-arch", arch, "copy", src, dst).Run()
}
