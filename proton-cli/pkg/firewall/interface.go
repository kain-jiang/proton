package firewall

type Interface interface {
	Apply() error
}
