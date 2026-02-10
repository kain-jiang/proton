package helm

// Image defines image of proton package store's helm values.
type Image struct {
	Registry string `json:"registry"`
}

func imageFor(registry string) Image { return Image{Registry: registry} }
