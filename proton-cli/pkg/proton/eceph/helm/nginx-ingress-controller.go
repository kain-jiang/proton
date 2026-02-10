package helm

func ValuesForNGINXIngressController(registry string, port int, class string) *Values {
	return &Values{
		Namespace: NamespaceDefault,
		Image: ValuesImage{
			Registry: registry,
		},
		Service: &ValuesService{
			HTTPPort:     port,
			IngressClass: class,
		},
	}
}
