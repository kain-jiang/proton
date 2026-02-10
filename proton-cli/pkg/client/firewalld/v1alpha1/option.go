package v1alpha1

import "fmt"

type option string

const (
	optState  option = "--state"
	optReload option = "--reload"

	optPermanent option = "--permanent"
	optZone      option = "--zone"
	optIPSet     option = "--ipset"

	optGetZones   option = "--get-zones"
	optNewZone    option = "--new-zone"
	optDeleteZone option = "--delete-zone"

	optGetTarget option = "--get-target"
	optSetTarget option = "--set-target"

	optListSources  option = "--list-sources"
	optAddSource    option = "--add-source"
	optRemoveSource option = "--remove-source"

	optListInterfaces  option = "--list-interfaces"
	optAddInterface    option = "--add-interface"
	optRemoveInterface option = "--remove-interface"

	optGetIPSets   option = "--get-ipsets"
	optNewIPSet    option = "--new-ipset"
	optDeleteIPSet option = "--delete-ipset"
	optType        option = "--type"
	optFamily      option = "--family"

	optGetEntries  option = "--get-entries"
	optAddEntry    option = "--add-entry"
	optRemoveEntry option = "--remove-entry"
)

func appendOption(opts []string, opt option) []string {
	return append(opts, string(opt))
}

func appendOptionCondition(opts []string, opt option, condition bool) []string {
	if condition {
		opts = append(opts, string(opt))
	}
	return opts
}

func appendOptionWithValue(opts []string, opt option, values ...string) []string {
	for _, v := range values {
		opts = append(opts, fmt.Sprintf("%s=%s", opt, v))
	}
	return opts
}
