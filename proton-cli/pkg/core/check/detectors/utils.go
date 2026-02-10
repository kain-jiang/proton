package detectors

import "github.com/fatih/color"

func colorRedOutput(content string) string {
	return color.New(color.FgRed).SprintFunc()(content)
}

func colorGreenOutput(content string) string {
	return color.New(color.FgGreen).SprintFunc()(content)
}

func colorYellowOutput(content string) string {
	return color.New(color.FgYellow).SprintFunc()(content)
}
