package componentmanage

import (
	"fmt"
	"strings"
)

func (m *Applier) SearchImage(registry, repository, defaultTag string) string {
	searchResult := make([]string, 0)
	prefix := fmt.Sprintf("%s/%s:", registry, repository)
	if len(m.extraImages) > 0 {
		for _, image := range m.extraImages {
			if strings.HasPrefix(image, prefix) {
				searchResult = append(searchResult, strings.Replace(image, registry, m.registry, 1))
			}
		}
	}
	defaultImage := fmt.Sprintf("%s/%s:%s", m.registry, repository, defaultTag)
	switch len(searchResult) {
	case 0:
		return defaultImage
	case 1:
		return searchResult[0]
	default:
		log.Warnf("found multiple images for %s/%s, using default %s", registry, repository, defaultImage)
		return defaultImage
	}

}

func (m *Applier) SearchTag(registry, repository, defaultTag string) string {
	searchResult := make([]string, 0)
	prefix := fmt.Sprintf("%s/%s:", registry, repository)
	if len(m.extraImages) > 0 {
		for _, image := range m.extraImages {
			if strings.HasPrefix(image, prefix) {
				searchResult = append(searchResult, strings.TrimPrefix(image, prefix))
			}
		}
	}
	switch len(searchResult) {
	case 0:
		return defaultTag
	case 1:
		return searchResult[0]
	default:
		log.Warnf("found multiple tags for %s/%s, using default tag: %s", registry, repository, defaultTag)
		return defaultTag
	}
}
