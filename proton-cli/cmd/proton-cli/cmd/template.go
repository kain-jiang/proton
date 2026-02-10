/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed templateInternal.yaml
var tmplInternal string

//go:embed templateExternal.yaml
var tmplExternal string

//go:embed templatePerfRec.yaml
var tmplPerfRec string

var templateType string

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template --type ?",
	Short: "show proton cluster template conf.",
	Long: `show proton cluster template conf. For example:
    proton-cli get template --type internal
You can edit this the template create proton cluster .
There are 3 templates:
internal --- for use with local cluster deployment
external --- for use with deployment in existing clusters, like a managed K8S platform
perfrec  --- is not really a template and is used to show various recommended configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		switch templateType {
		case "internal":
			fmt.Println(tmplInternal)
		case "external":
			fmt.Println(tmplExternal)
		case "perfrec":
			fmt.Println(tmplPerfRec)
		default:
			fmt.Println("invalid template type, please specify one of internal, external or perfrec with the parameter --type ?")
		}
	},
}

func init() {
	getCmd.AddCommand(templateCmd)
	templateCmd.PersistentFlags().StringVar(&templateType, "type", "internal", `choose a template to show`)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
