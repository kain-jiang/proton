package helm

func ValuesForConfigWeb(registry string) *Values4ECeph {
	return &Values4ECeph{
		Namespace: NamespaceDefault,
		Image: ValuesImage{
			Registry: registry,
		},
	}
}
