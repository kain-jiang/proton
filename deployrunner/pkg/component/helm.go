package component

import (
	"encoding/json"

	"taskrunner/trait"
)

// HelmComponent helm component
type HelmComponent struct {
	trait.ComponentMeta `json:",inline"`
	HelmComponentSpec   `json:",inline"`
}

// HelmComponentSpec helm component special defiend
type HelmComponentSpec struct {
	Images              []string `json:"images"`
	Repository          string   `json:"repository"`
	HelmChartAPIVersion string   `json:"helmChartAPIVersion"`
}

// Decode decode from special
func (spec *HelmComponentSpec) Decode(bs []byte) error {
	return json.Unmarshal(bs, spec)
}

// Encode encode into bytes
func (spec *HelmComponentSpec) Encode() ([]byte, error) {
	return json.Marshal(spec)
}
