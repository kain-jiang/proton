package helm

func ValuesForNodeAgent(registry string, bindAddr string) *Values4ECeph {
	return &Values4ECeph{
		Image: ValuesImage{
			Registry: registry,
		},
		Service: map[string]string{
			"bind-addr": bindAddr,
		},
	}
}
