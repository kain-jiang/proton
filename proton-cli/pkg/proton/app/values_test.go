package app

import (
	"testing"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestBuildHelmValuesKeepsKafkaHeadlessAddress(t *testing.T) {
	cfg := &configuration.ClusterConfig{
		ResourceConnectInfo: &configuration.ResourceConnectInfo{
			Mq: &configuration.MqInfo{
				SourceType: configuration.Internal,
				MqType:     configuration.KafkaType,
				MqHosts:    "kafka-headless.resource.svc.cluster.local.",
				MqPort:     9097,
				Auth: &configuration.Auth{
					Username:  "kweaver",
					Password:  "kweaver",
					Mechanism: configuration.Plain,
				},
			},
		},
	}

	values := BuildHelmValues(cfg, "kweaver")
	depServices, ok := values["depServices"].(map[string]interface{})
	if !ok {
		t.Fatalf("depServices has unexpected type %T", values["depServices"])
	}
	mq, ok := depServices["mq"].(map[string]interface{})
	if !ok {
		t.Fatalf("depServices.mq has unexpected type %T", depServices["mq"])
	}

	if got, want := mq["mqHost"], "kafka-headless.resource.svc.cluster.local."; got != want {
		t.Fatalf("mqHost = %v, want %v", got, want)
	}
	if got, want := mq["mqPort"], 9097; got != want {
		t.Fatalf("mqPort = %v, want %v", got, want)
	}
}
