package baseresource

import (
	"fmt"
	mongodbv1 "proton-mongodb-operator/api/v1"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
)

func NewMongoConfigmap(instance *mongodbv1.MongodbOperator) *corev1.ConfigMap {
	var i int32
	var logrotateconf [][]string
	mgmtconf := []string{
		"lang: zh_CN",
		"port: 28001",
		"mongodbPort: 28000",
		"defaultCollextionName: test",
		"logLevel: " + instance.Spec.MgmtSpec.LogLevel,
		"useEncryption: " + strconv.FormatBool(instance.Spec.MgmtSpec.UseEncryption),
		"cluster: mongod",
		"mongoDataDir: /data/mongodb_data/",
		"backupDataDir: /var/lib/proton-mongodb/",
		"agentPort: 8899",
		"releaseName: " + instance.Name,
		"replSet: " + instance.Spec.MongoDBSpec.Replset.Name}
	var tls string
	if !instance.Spec.MongoDBSpec.Mongodconf.TLS.Enabled {
		tls = "    mode: disabled"
	} else if instance.Spec.MongoDBSpec.Mongodconf.TLS.Enabled {
		tls = "    mode: allowTLS\n" +
			"    certificateKeyFile: /mongodb/tls/server-cert-key.pem\n" +
			"    CAFile: /mongodb/tls/ca.pem"
	}
	mongodconf := []string{
		"net:",
		"  ipv6: true",
		"  bindIpAll: true",
		"  maxIncomingConnections: 10000",
		"  tls:", tls,
		"  unixDomainSocket:",
		"    pathPrefix: /data/mongodb_data",
		"  serviceExecutor: adaptive",
		"operationProfiling:",
		"  mode: slowOp",
		"  slowOpThresholdMs: 1000",
		"processManagement:",
		"  pidFilePath: /data/mongodb_data/mongodb.pid",
		"storage:",
		"  dbPath: /data/mongodb_data",
		"  directoryPerDB: true",
		"  engine: wiredTiger",
		"  wiredTiger:",
		"    engineConfig:",
		"      cacheSizeGB: " + strconv.Itoa(int(instance.Spec.MongoDBSpec.Mongodconf.WiredTigerCacheSizeGB)),
		"      directoryForIndexes: true",
		"security:",
		"  javascriptEnabled: false",
		"systemLog:",
		"  destination: file",
		"  logAppend: true",
		"  logRotate: reopen",
		"  path: /data/mongodb_data/mongodb.log",
		"  quiet: true",
		"  traceAllExceptions: false",
	}

	if instance.Spec.MongoDBSpec.Mongodconf.AuditLog != nil {
		auditlog := []string{"auditLog:"}
		for k, v := range instance.Spec.MongoDBSpec.Mongodconf.AuditLog {
			switch v.Type {
			case intstr.Int:
				auditlog = append(auditlog, fmt.Sprintf("  %s: %d", k, v.IntVal))
			case intstr.String:
				auditlog = append(auditlog, fmt.Sprintf("  %s: %s", k, v.StrVal))
			}
		}
		mongodconf = append(mongodconf, auditlog...)
	}

	data := map[string]string{
		"mongodb.conf":      strings.Join(mongodconf, "\n"),
		"mongodb-mgmt.yaml": strings.Join(mgmtconf, "\n"),
	}
	if instance.Spec.MongoDBSpec.Storage.StorageClassName == "" {
		for i = 0; i < instance.Spec.MongoDBSpec.Replicas; i++ {
			logrotateconf = append(logrotateconf, []string{
				instance.Spec.MongoDBSpec.Storage.VolumeSpec[i].Path + "/*log {",
				"  missingok",
				"  notifempty",
				"  size " + instance.Spec.LogrotateSpec.Logsize,
				"  rotate " + strconv.Itoa(int(instance.Spec.LogrotateSpec.Logcount)),
				"  dateext",
				"  dateformat -%Y%m%d%H%s",
				"  postrotate",
				"    pwd=`echo -n ${MONGO_PASSWORD} | base64 -d`",
				"    adminKey=`echo -n ${MONGO_USERNAME}:${pwd} | base64`",
				"    curl -X POST http://" + fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt") + "-" + strconv.Itoa(int(i)) + "." + fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt") + "." + instance.Namespace + ".svc.cluster.local:28001/api/proton-mongodb-mgmt/v2/logrotate  --header \"admin-key:${adminKey}\"",
				"  endscript",
				"}"})
			data[fmt.Sprintf("mongodb-logrotate-%d", i)] = strings.Join(logrotateconf[i], "\n")
		}
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
			Namespace: instance.Namespace,
		},
		Data: data,
	}
	return cm
}
