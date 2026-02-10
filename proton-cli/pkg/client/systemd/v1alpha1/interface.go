package v1alpha1

type Interface interface {
	Start(name string) error
	Enabled(name string, now bool) error

	IsActive(name string) (bool, error)
	IsEnabled(name string) (bool, error)
}
