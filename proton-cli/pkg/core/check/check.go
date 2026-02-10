package check

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/check/detectors"
)

type HandlerFunc func() [][]string

func Check() {
	var handlers []HandlerFunc
	handlers = append(handlers, detectors.NewK8SDetector)
	handlers = append(handlers, detectors.NewOSDetector)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Category", "Item", "Result", "Status", "Remarks"})
	table.SetRowLine(true)
	table.SetColWidth(50)
	table.SetAutoWrapText(true)

	for _, handler := range handlers {
		result := handler()
		for _, v := range result {
			table.Append(v)
		}
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()

}
