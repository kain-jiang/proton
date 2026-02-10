package eceph

import (
	"log"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

func TestLog(t *testing.T) {
	name := types.NamespacedName{Name: "sss"}
	logger.NewLogger().WithField("name", name).Info("print data")
	klog.InfoS("print data", "name", name)
	log.Print(name)
}
