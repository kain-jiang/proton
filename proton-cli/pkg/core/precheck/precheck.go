package precheck

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/precheck/cpu"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/precheck/disk"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/precheck/network"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/precheck/node"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/precheck/service"
)

type HandlerFunc func() [][]string

func PreCheck(password, ntpserver string) {
	var handlers []HandlerFunc
	handlers = append(handlers, node.NodeInfo)
	handlers = append(handlers, network.GatewayTest)
	handlers = append(handlers, network.PortAvaiable)
	handlers = append(handlers, service.ChronydAlived)
	handlers = append(handlers, cpu.CPUPerf)
	handlers = append(handlers, disk.DirectDD)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Item", "Result", "Status", "Remarks"})
	table.SetRowLine(true)

	for _, handler := range handlers {
		result := handler()
		for _, v := range result {
			table.Append(v)
		}
	}
	for _, v := range node.NTPCheck(ntpserver) {
		table.Append(v)
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()

}
