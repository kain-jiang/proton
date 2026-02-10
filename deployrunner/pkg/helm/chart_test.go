package helm

import (
	"bytes"
	"fmt"
	"math"
	"runtime/debug"
	"testing"

	"taskrunner/test"
)

func testWantError(t *testing.T, want, get error) {
	if want != get {
		debug.PrintStack()
		t.Fatalf("want:%#v, get %#v", want, get)
	}
}

func TestPasrseChartV1(t *testing.T) {
	tt := test.TestingT{T: t}
	fsc := test.TestCharts
	chpath := "testdata/charts/python3.1.tgz"
	fin, err := fsc.Open(chpath)
	tt.AssertNil(err)
	ch, err := ParseChartFromTGZ(fin, "v1")
	tt.AssertNil(err)
	img, err := ch.Images()
	tt.AssertNil(err)
	tt.Assert(img[0], "nginx:1.16.0")

	fmt.Println(int(math.Floor(float64(25) * (float64(1)) / 100)))
}

func TestChart(t *testing.T) {
	fsc := test.TestCharts
	chartsPath := []string{
		"testdata/charts/python3-0.1.0.tgz",
		"testdata/charts/python4-0.1.0.tgz",
	}
	want := map[string]int{
		"nginx:1.16.0": 0,
	}
	tt := test.TestingT{T: t}

	for _, p := range chartsPath {
		bs, err := fsc.ReadFile(p)
		testWantError(t, nil, err)
		br := bytes.NewReader(bs)
		c, err := ParseChartFromTGZ(br, "v2")
		tt.AssertNil(err)
		imgs, err := c.Images()
		tt.AssertNil(err)

		for _, img := range imgs {
			if _, ok := want[img]; !ok {
				t.Fatal(img)
			}
			want[img]++
		}
		for k, count := range want {
			if count == 0 {
				t.Fatal(k)
			}
		}
	}
}
