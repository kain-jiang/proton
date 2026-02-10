package utils_test

import (
	"testing"

	"taskrunner/pkg/utils"

	"github.com/sirupsen/logrus"
)

func TestTopoSort(t *testing.T) {
	type l = []string
	type m = map[string]l
	result, err := utils.TopoSort(m{
		"a": l{"b"},
		"b": l{"c", "d"},
	}, logrus.New())
	t.Log(len(result), result, err)
}
