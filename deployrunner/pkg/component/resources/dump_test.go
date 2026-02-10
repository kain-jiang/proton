package resources

import (
	"fmt"
	"testing"

	"taskrunner/test"

	"github.com/ghodss/yaml"
)

func TestResourceInfoDump(t *testing.T) {
	objs := map[string]map[string]any{}
	tt := test.TestingT{T: t}
	{
		// 构造组件对象,后续输出
		{
			etcd := Etcd{}
			obj, err := etcd.ToDepMap("qwe")
			tt.AssertNil(err)
			objs["etcd"] = obj
		}
		{
			r := GraphDB{}
			obj, err := r.ToDepMap()
			tt.AssertNil(err)
			objs["graphdb"] = obj
		}
		{
			r := MongoDB{
				Options: map[string]interface{}{
					"qwe": "qweqwe",
				},
				Port:       123,
				AuthSource: "认证源",
				MgmtHost:   "mgntHost",
				MgmtPort:   123,
				AdminKey:   "qwe",
			}
			obj, err := r.ToDepMap()
			tt.AssertNil(err)
			objs["mongodb"] = obj
		}
		{
			r := Nsq{}
			obj, err := r.ToMap()
			tt.AssertNil(err)
			objs["nsq"] = obj
		}
		{
			r := MQ{
				Auth: &MQAuth{},
			}
			obj, err := r.ToMap()
			tt.AssertNil(err)
			objs["其他mq"] = obj
		}
		{
			r := Opensearch{}
			obj, err := r.ToDepMap()
			tt.AssertNil(err)
			objs["opensearch"] = obj
		}
		{
			r := Redis{
				ConnectType: "sentinel",
			}
			obj, err := r.ToMap()
			tt.AssertNil(err)
			objs["redis_sentinel"] = obj
		}
		{
			r := Redis{
				ConnectType: "master-slave",
			}
			obj, err := r.ToMap()
			tt.AssertNil(err)
			objs["redis_master-slave"] = obj
		}
		{
			r := Redis{
				ConnectType: "standalone",
			}
			obj, err := r.ToMap()
			tt.AssertNil(err)
			objs["redis_standalone"] = obj
		}
	}

	for k, v := range objs {
		ybs, rerr := yaml.Marshal(v)
		tt.AssertNil(rerr)
		fmt.Printf("---\n#%s\n\n%s\n\n", k, string(ybs))
	}
}
