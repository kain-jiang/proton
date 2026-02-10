package store_test

import (
	"context"
	"reflect"
	"testing"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestProton(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()

	objs := []trait.ProtonCompoent{
		{
			ProtonComponentMeta: trait.ProtonComponentMeta{
				Name:   "test0",
				Type:   "0",
				System: trait.System{},
			},
		},
		{
			ProtonComponentMeta: trait.ProtonComponentMeta{
				Name:   "test1",
				Type:   "0",
				System: trait.System{},
			},
		},
		{
			ProtonComponentMeta: trait.ProtonComponentMeta{
				Name: "test0",
				Type: "1",
				System: trait.System{
					SID: 1,
				},
			},
		},
		{
			ProtonComponentMeta: trait.ProtonComponentMeta{
				Name: "test1",
				Type: "1",
				System: trait.System{
					SID: 1,
				},
			},
		},
	}

	for _, obj := range objs {
		err := s.InsertProtonComponent(ctx, obj)
		tt.AssertNil(err)
	}

	for _, obj := range objs {
		err := s.UpdateProtonComponent(ctx, obj)
		tt.AssertNil(err)
	}

	for _, obj := range objs {
		o, err := s.GetProtonComponent(ctx, obj.Name, obj.Type, obj.SID)
		tt.AssertNil(err)
		if !reflect.DeepEqual(o, &obj) {
			t.Errorf("\nwant: %#v\n get: %#v", o, &obj)
			t.FailNow()
		}
	}

	conditoins := []struct {
		Name string
		Type string
		Sid  int
		want int
	}{
		{
			Name: "test0",
			want: 2,
			Sid:  -1,
		},
		{
			Type: "0",
			want: 2,
			Sid:  -1,
		},
		{
			want: 2,
			Sid:  0,
		},
		{
			Type: "0",
			Sid:  0,
			want: 2,
		},
		{
			Type: "1",
			Sid:  0,
			want: 0,
		},
		{
			Sid:  -1,
			want: 4,
		},
	}

	for _, obj := range conditoins {
		count, err := s.CountProtonConponent(ctx, obj.Name, obj.Type, obj.Sid)
		tt.AssertNil(err)
		tt.Assert(obj.want, count)
		os, err := s.ListProtonConponent(ctx, obj.Name, obj.Type, obj.Sid, 100, 0)
		tt.AssertNil(err)
		tt.Assert(obj.want, len(os))

		os, err = s.ListProtonConponentWithInternal(ctx, obj.Name, obj.Type, obj.Sid, 100, 0)
		tt.AssertNil(err)
		tt.Assert(obj.want, len(os))
	}
}
