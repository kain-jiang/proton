package file

import (
	"testing"
)

func TestGetManagedContent(t *testing.T) {
	type args struct {
		orig string
		head string
		tail string
	}
	tests := []struct {
		name     string
		args     args
		wantCont string
	}{
		{
			name: "sample",
			args: args{
				orig: "0123456789",
				head: "2",
				tail: "7",
			},
			wantCont: "3456",
		},
		{
			name: "from-empty",
			args: args{
				head: "2",
				tail: "7",
			},
		},
		{
			name: "missing-head",
			args: args{
				orig: "0123456789",
				head: "x",
				tail: "7",
			},
		},
		{
			name: "missing-tail",
			args: args{
				orig: "0123456789",
				head: "2",
				tail: "x",
			},
		},
		{
			name: "missing-both",
			args: args{
				orig: "0123456789",
				head: "a",
				tail: "b",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCont := GetManagedContent([]byte(tt.args.orig), []byte(tt.args.head), []byte(tt.args.tail)); string(gotCont) != tt.wantCont {
				t.Errorf("GetManagedContent() = %v, want %v", gotCont, tt.wantCont)
			}
		})
	}
}

func TestSetManagedContent(t *testing.T) {
	type args struct {
		orig string
		head string
		tail string
		cont string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sample",
			args: args{
				orig: "0123456789",
				head: "2",
				tail: "7",
				cont: "x",
			},
			want: "012x789",
		},
		{
			name: "to-empty",
			args: args{
				head: "2",
				tail: "7",
				cont: "x",
			},
			want: "2x7",
		},
		{
			name: "missing-head",
			args: args{
				orig: "0123456789",
				head: "a",
				tail: "7",
				cont: "x",
			},
			want: "0123456ax789",
		},
		{
			name: "missing-tail",
			args: args{
				orig: "0123456789",
				head: "2",
				tail: "b",
				cont: "x",
			},
			want: "012xb3456789",
		},
		{
			name: "missing-both",
			args: args{
				orig: "0123456789",
				head: "a",
				tail: "b",
				cont: "x",
			},
			want: "0123456789axb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetManagedContent([]byte(tt.args.orig), []byte(tt.args.head), []byte(tt.args.tail), []byte(tt.args.cont)); string(got) != tt.want {
				t.Errorf("SetManagedContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetManagedIndexes(t *testing.T) {
	type args struct {
		orig string
		head string
		tail string
	}
	tests := []struct {
		name   string
		args   args
		wantHH int
		wantHT int
		wantTH int
		wantTT int
	}{
		{
			name: "sample",
			args: args{
				orig: "0123456789",
				head: "2",
				tail: "7",
			},
			wantHH: 2,
			wantHT: 3,
			wantTH: 7,
			wantTT: 8,
		},
		{
			name: "from-empty",
			args: args{
				head: "2",
				tail: "7",
			},
		},
		{
			name: "missing-head",
			args: args{
				orig: "0123456789",
				head: "x",
				tail: "7",
			},
			wantHH: 7,
			wantHT: 7,
			wantTH: 7,
			wantTT: 8,
		},
		{
			name: "missing-tail",
			args: args{
				orig: "0123456789",
				head: "2",
				tail: "x",
			},
			wantHH: 2,
			wantHT: 3,
			wantTH: 3,
			wantTT: 3,
		},
		{
			name: "missing-both",
			args: args{
				orig: "0123456789",
				head: "a",
				tail: "b",
			},
			wantHH: 10,
			wantHT: 10,
			wantTH: 10,
			wantTT: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHh, gotH, gotT, gotTt := GetManagedIndexes([]byte(tt.args.orig), []byte(tt.args.head), []byte(tt.args.tail))
			if gotHh != tt.wantHH {
				t.Errorf("GetManagedIndexes() gotHH = %v, want %v", gotHh, tt.wantHH)
			}
			if gotH != tt.wantHT {
				t.Errorf("GetManagedIndexes() gotHT = %v, want %v", gotH, tt.wantHT)
			}
			if gotT != tt.wantTH {
				t.Errorf("GetManagedIndexes() gotTH = %v, want %v", gotT, tt.wantTH)
			}
			if gotTt != tt.wantTT {
				t.Errorf("GetManagedIndexes() gotTT = %v, want %v", gotTt, tt.wantTT)
			}
		})
	}
}
