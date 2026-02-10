package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

type Client struct {
	executor exec.Executor

	log logrus.FieldLogger
}

func New(e exec.Executor) *Client {
	return &Client{
		executor: e,
		log:      logger.NewLogger().WithField("client", "helm/v2"),
	}
}

func (c *Client) UpdateRepoCache(ctx context.Context, repo string) error {
	var args = []string{"repo", "update"}
	if repo != "" {
		args = append(args, repo)
	}
	return c.executor.Command("helm", args...).Run()
}

// Reconcile implements Interface.
//
// TODO: compare chart version and values
func (c *Client) Reconcile(ctx context.Context, release string, chart string, values map[string]any) error {
	out, err := c.executor.Command("helm", "list", "--output=json").Output()
	if err != nil {
		return err
	}

	list := &outputReleaseList{}

	// stdout is "" not "[]", when there is not any helm release.
	if len(out) != 0 {
		if err := json.Unmarshal(out, list); err != nil {
			return err
		}
	}

	if !slices.ContainsFunc(list.Releases, func(or outputRelease) bool { return or.Name == release }) {
		c.log.WithFields(logrus.Fields{"release": release, "chart": chart}).Info("install release")
		return c.install(ctx, release, chart, values)
	}

	c.log.WithFields(logrus.Fields{"release": release, "chart": chart}).Info("upgrade release")
	return c.upgrade(ctx, release, chart, values)
}

func (c *Client) install(ctx context.Context, release string, chart string, values map[string]any) error {
	f, err := os.CreateTemp("", "helm-v2-values-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	b, err := json.Marshal(values)
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	f.Close()

	c.log.WithFields(logrus.Fields{"release": release, "chart": chart, "values": string(b)}).Debug("install release")

	var args = []string{
		"install", chart,
		"--name=" + release,
		"--values=" + f.Name(),
	}

	result := c.executor.Command("helm", args...).Run()
	if result != nil {
		if strings.Contains(fmt.Sprintf("%v", result), "lost connection to pod") {
			c.log.Debug("helm2 command returned an error that indicates pod connection lost, will retry once. err:%v", result)
			time.Sleep(time.Duration(5) * time.Second)
			return c.executor.Command("helm", args...).Run()
		}
	}
	return result
}

func (c *Client) upgrade(ctx context.Context, release string, chart string, values map[string]any) error {
	f, err := os.CreateTemp("", "helm-v2-values-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	b, err := json.Marshal(values)
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	f.Close()

	c.log.WithFields(logrus.Fields{"release": release, "chart": chart, "values": string(b)}).Debug("upgrade release")

	var args = []string{
		"upgrade", release, chart,
		"--values=" + f.Name(),
	}

	result := c.executor.Command("helm", args...).Run()
	if result != nil {
		if strings.Contains(fmt.Sprintf("%v", result), "lost connection to pod") {
			c.log.Debug("helm2 command returned an error that indicates pod connection lost, will retry once. err:%v", result)
			time.Sleep(time.Duration(5) * time.Second)
			return c.executor.Command("helm", args...).Run()
		}
	}
	return result
}

var _ Interface = &Client{}

type outputRelease struct {
	Name string `json:"name,omitempty"`

	Revision int `json:"revision,omitempty"`

	Updated string `json:"updated,omitempty"`

	Status string `json:"status,omitempty"`

	AppVersion string `json:"appVersion,omitempty"`

	Namespace string `json:"namespace,omitempty"`
}

type outputReleaseList struct {
	Releases []outputRelease `json:"releases,omitempty"`
}
