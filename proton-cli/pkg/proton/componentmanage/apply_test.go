package componentmanage

import (
	"reflect"
	"testing"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestMustToMap(t *testing.T) {
	type args struct {
		val configuration.MqInfo
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		// Test cases
		{
			name: "Test with mqInfo",
			args: args{val: configuration.MqInfo{SourceType: configuration.External}},
			want: map[string]interface{}{"source_type": "external"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustToMap(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mustToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustFromMap(t *testing.T) {
	type args struct {
		val map[string]any
	}
	tests := []struct {
		name string
		args args
		want *configuration.MqInfo
	}{
		// Test cases
		{
			name: "Test with mqInfo",
			args: args{val: map[string]interface{}{"source_type": "external"}},
			want: &configuration.MqInfo{SourceType: configuration.External},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustFromMap[configuration.MqInfo](tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mustToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
