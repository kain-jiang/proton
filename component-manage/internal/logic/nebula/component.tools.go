package nebula

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
)

type allPathes struct {
	GraphDLog     string
	MetaDLog      string
	MetaDData     string
	StorageDLog   string
	StorageDData0 string
}

func (p *allPathes) Pathes() []string {
	return []string{
		p.GraphDLog,
		p.MetaDLog,
		p.MetaDData,
		p.StorageDLog,
		p.StorageDData0,
	}
}

// tools
func toList(l []string) []string {
	if l == nil {
		return make([]string, 0)
	}
	return l
}

func defaultTo(val, dft string) string {
	if val == "" {
		return dft
	}
	return val
}

func defaultMapTo(val, dft map[string]any) map[string]any {
	if val == nil {
		return dft
	}
	return val
}

// filepath.Join(dataPath, "data", component)

func getDataPathes(dataPath string) *allPathes {
	return &allPathes{
		GraphDLog:     filepath.Join(dataPath, "logs", "graphd"),
		MetaDLog:      filepath.Join(dataPath, "logs", "metad"),
		MetaDData:     filepath.Join(dataPath, "data", "metad"),
		StorageDLog:   filepath.Join(dataPath, "logs", "graphd"),
		StorageDData0: filepath.Join(dataPath, "data", "graphd", "0"),
	}
}

func createRootPassword() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
