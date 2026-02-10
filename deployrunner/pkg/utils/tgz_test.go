package utils

import (
	"archive/tar"
	"bytes"
	"io"
	"reflect"
	"testing"

	"taskrunner/test"
)

func TestTgzReader(t *testing.T) {
	// tdata := "test/testdata/initc-0.1.0.tgz"
	// fin, err := os.Open(tdata)

	tcharts := test.TestCharts
	fpath := "testdata/charts/python3-0.1.0.tgz"
	fin, err := tcharts.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer fin.Close()
	bs0, err := io.ReadAll(fin)
	if err != nil {
		t.Fatal(err)
	}
	r0 := bytes.NewReader(bs0)
	tr, err := NewTGZReader(r0)
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	w := NewTGzWriter(buf)

	w.ziper.Header.OS = 0x3
	hs := []*tar.Header{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		hs = append(hs, h)
		err = w.WriteHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(w, tr); err != nil {
			t.Fatal(err)
		}
	}
	w.Close()

	bs1 := buf.Bytes()
	r1 := bytes.NewReader(bs1)

	tr1, err := NewTGZReader(r1)
	if err != nil {
		t.Fatal(err)
	}

	hs1 := []*tar.Header{}

	for {
		h1, err := tr1.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		hs1 = append(hs1, h1)

	}
	if !reflect.DeepEqual(hs, hs1) {
		t.Fatal(hs, hs1)
	}
}

func TestFlush(t *testing.T) {
	tt := test.TestingT{T: t}
	buf := bytes.NewBuffer(make([]byte, 0))
	tgz := NewTGzWriter(buf)
	defer tgz.Close()
	err := tgz.WriteHeader(&tar.Header{
		Name: "test",
		Size: 4,
	})
	tt.AssertNil(err)
	_, err = tgz.Write([]byte("test"))
	tt.AssertNil(err)
	err = tgz.Flush()
	tt.AssertNil(err)
	if len(buf.Bytes()) <= 0 {
		t.Fatalf("flush error")
	}
}
