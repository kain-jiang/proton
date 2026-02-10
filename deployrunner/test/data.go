package test

import (
	"embed"
)

// TestCharts chart data
//
//go:embed testdata/charts
var TestCharts embed.FS

// TestAPP test application bytes
//
//go:embed testdata/testapp.json
var TestAPP []byte
