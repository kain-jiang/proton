package componentmanage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
)

type cli struct {
	log logrus.FieldLogger

	restyCli *resty.Client
}

type Client interface {
	ComponentUpgradable(_type, name string, newVersion string) (bool, string, error)

	EnableKafka(chart, version string) error
	UpgradeKafka(name string, reqData map[string]any, zkName string) (map[string]any, error)
	CreateKafka(name string, reqData map[string]any, zkName string) (map[string]any, error)
	GetKafka(name string) (map[string]any, error)
	DeleteKafka(name string) error

	EnableZookeeper(chart, version string) error
	UpgradeZookeeper(name string, reqData map[string]any) (map[string]any, error)
	CreateZookeeper(name string, reqData map[string]any) (map[string]any, error)
	GetZookeeper(name string) (map[string]any, error)
	DeleteZookeeper(name string) error

	EnableOpensearch(chart, version string) error
	UpgradeOpensearch(name string, reqData map[string]any) (map[string]any, error)
	CreateOpensearch(name string, reqData map[string]any) (map[string]any, error)
	GetOpensearch(name string) (map[string]any, error)
	DeleteOpensearch(name string) error

	EnableRedis(chart, version string) error
	UpgradeRedis(name string, reqData map[string]any) (map[string]any, error)
	CreateRedis(name string, reqData map[string]any) (map[string]any, error)
	GetRedis(name string) (map[string]any, error)
	DeleteRedis(name string) error

	EnableETCD(chart, version string) error
	UpgradeETCD(name string, reqData map[string]any) (map[string]any, error)
	CreateETCD(name string, reqData map[string]any) (map[string]any, error)
	GetETCD(name string) (map[string]any, error)
	DeleteETCD(name string) error

	EnablePolicyEngine(chart, version string) error
	UpgradePolicyEngine(name string, reqData map[string]any, eName string) (map[string]any, error)
	CreatePolicyEngine(name string, reqData map[string]any, eName string) (map[string]any, error)
	GetPolicyEngine(name string) (map[string]any, error)
	DeletePolicyEngine(name string) error
	EnableMariaDB(info MariaDBPluginInfo) error
	CreateMariaDB(name string, reqData map[string]any) (map[string]any, error)
	UpgradeMariaDB(name string, reqData map[string]any) (map[string]any, error)
	GetMariaDB(name string) (map[string]any, error)
	DeleteMariaDB(name string) error

	EnableMongoDB(info MongoDBPluginInfo) error
	CreateMongoDB(name string, reqData map[string]any) (map[string]any, error)
	UpgradeMongoDB(name string, reqData map[string]any) (map[string]any, error)
	GetMongoDB(name string) (map[string]any, error)
	DeleteMongoDB(name string) error

	EnableNebula(info NebulaPluginInfo) error
	CreateNebula(name string, reqData map[string]any) (map[string]any, map[string]any, error)
	UpgradeNebula(name string, reqData map[string]any) (map[string]any, map[string]any, error)
	GetNebula(name string) (map[string]any, error)
	DeleteNebula(name string) error
}

func New(namespace, service string, port int, log logrus.FieldLogger, direct bool) (Client, error) {

	var restyCli *resty.Client
	khcli, err := client.NewK8sHTTPClient()
	if err != nil {
		return nil, err
	}
	_, k := client.NewK8sClientInterface()

	if direct {
		svc, err := k.CoreV1().Services(namespace).Get(context.TODO(), service, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		restyCli = resty.New().SetBaseURL(fmt.Sprintf("http://%s:%d", svc.Spec.ClusterIP, port))
	} else {
		restyCli = resty.NewWithClient(khcli).SetBaseURL(
			k.
				CoreV1().
				RESTClient().
				Get().
				Namespace(namespace).
				Resource("services").
				SubResource("proxy").
				Name(net.JoinSchemeNamePort("http", service, strconv.Itoa(port))).
				URL().
				String(),
		)
	}
	log.WithField("baseurl", restyCli.BaseURL).Info("componentmanage client client is ready")
	return &cli{
		log:      log,
		restyCli: restyCli,
	}, nil
}

func errorOf(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("code: %s, resp: %s", resp.Status(), string(resp.Body()))
	}
	return nil
}

func compareComponentVersion(a, b string) (int, error) {
	if strings.Contains(a, "+") && strings.Contains(b, "+") {
		_a := strings.SplitN(a, "+", 2)
		_b := strings.SplitN(b, "+", 2)

		va1, err1 := version.NewSemver(_a[0])
		va2, err2 := version.NewSemver(_a[1])
		vb1, err3 := version.NewSemver(_b[0])
		vb2, err4 := version.NewSemver(_b[1])

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return 0, fmt.Errorf("version format error: %w", errors.Join(err1, err2, err3, err4))
		}

		v1r := va1.Compare(vb1)
		if v1r != 0 {
			return v1r, nil
		}
		return va2.Compare(vb2), nil
	} else {
		va, err1 := version.NewSemver(a)
		vb, err2 := version.NewSemver(b)
		if err1 != nil || err2 != nil {
			return 0, fmt.Errorf("version format error: %w", errors.Join(err1, err2))
		}
		return va.Compare(vb), nil
	}
}

func (c *cli) ComponentUpgradable(_type, name string, newVersion string) (bool, string, error) {
	var result struct {
		Version string `json:"version"`
	}
	resp, err := c.restyCli.R().SetResult(&result).SetPathParams(map[string]string{
		"type": _type,
		"name": name,
	}).Get("/api/component-manage/v1/components/release/{type}/{name}")

	// 实例不存在，也需要升级
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return true, "", nil
	}

	err = errorOf(resp, err)
	if err != nil {
		return false, "", fmt.Errorf("get %s failed: %w", _type, err)
	}

	oldVersion := result.Version
	/*
		0.0.1 < 0.0.2
		0.0.2+0.0.3 > 0.0.2+0.0.1
	*/
	cr, err := compareComponentVersion(newVersion, oldVersion)
	if err != nil {
		return false, "", fmt.Errorf("compare %s failed: %w", _type, err)
	}
	c.log.Debugf("compare %s upgradable: %s -> %s", _type, oldVersion, newVersion)
	return cr >= 0, oldVersion, nil
}
