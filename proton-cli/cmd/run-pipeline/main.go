package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2/textlogger"
	"k8s.io/utils/ptr"
)

const (
	DefaultHost = "devops.aishu.cn"

	DefaultOrganization = "AISHUDevOps"

	DefaultProject = "ICT"

	DefaultPipeline = "proton-cli"

	DefaultReference = "refs/heads/master"

	DefaultTimeout = time.Hour
)

type Config struct {
	Host string `json:"host,omitempty"`

	Organization string `json:"organization,omitempty"`

	PersonalAccessToken string `json:"-"`

	Project string `json:"project,omitempty"`

	Pipeline string `json:"pipeline,omitempty"`

	Reference string `json:"reference,omitempty"`

	StagesToSkip []string `json:"stagesToSkip,omitempty"`

	Timeout time.Duration `json:"timeout,omitempty"`
}

func (c *Config) OrganizationUrl() string {
	u := url.URL{Scheme: "https", Host: c.Host, Path: c.Organization}
	return u.String()
}

func main() {
	config := new(Config)
	pflag.StringVar(&config.Host, "host", DefaultHost, "host[:port]")
	pflag.StringVar(&config.Organization, "org", DefaultOrganization, "organization name")
	pflag.StringVar(&config.PersonalAccessToken, "pat", "", "personal access token")
	pflag.StringVar(&config.Project, "project", DefaultProject, "project name")
	pflag.StringVar(&config.Pipeline, "pipeline", DefaultPipeline, "pipeline name")
	pflag.StringVar(&config.Reference, "ref", DefaultReference, "git repo branch")
	pflag.StringArrayVar(&config.StagesToSkip, "skip", nil, "skip these stages of the pipeline")
	pflag.DurationVar(&config.Timeout, "timeout", DefaultTimeout, "timeout, 0: no timeout")
	pflag.Parse()

	// log := klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
	log := textlogger.NewLogger(textlogger.NewConfig())

	log.Info("display config", "config", string(toJSON(config)))

	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(config.OrganizationUrl(), config.PersonalAccessToken)

	// Create a client to interact with the pipelines area
	client := pipelines.NewClient(ctx, connection)

	// Run the pipeline
	r, err := RunPipeline(ctx, client, config.Project, config.Pipeline, config.Reference, config.StagesToSkip, log)
	if err != nil {
		log.Error(err, "run pipeline fail")
		panic(err)
	}

	// Wait the pipeline run
	r, err = Wait(ctx, client, config.Project, r, log)
	if err != nil {
		log.Error(err, "wait pipeline run fail")
		panic(err)
	}

	if r.Result == nil {
		log.Info("run.Result == nil")
	} else {
		log.Info("pipeline completed", "result", *r.Result)
	}
}

func RunPipeline(ctx context.Context, c pipelines.Client, project, name, ref string, stagesToSkip []string, log logr.Logger) (*pipelines.Run, error) {
	log = log.WithValues("project", project)

	log.Info("list pipelines")
	pipelineList, err := c.ListPipelines(ctx, pipelines.ListPipelinesArgs{Project: &project})
	if err != nil {
		return nil, fmt.Errorf("list pipelines fail: %w", err)
	}
	if pipelineList == nil {
		return nil, errors.New("pipelineList == nil")
	}

	var pipelineId *int
	for _, p := range *pipelineList {
		if p.Name != nil && name == *p.Name {
			pipelineId = p.Id
			break
		}
	}
	if pipelineId == nil {
		return nil, fmt.Errorf("pipeline %q not found", name)
	}

	resources := &pipelines.RunResourcesParameters{
		Repositories: &map[string]pipelines.RepositoryResourceParameters{
			"self": {
				RefName: ptr.To(ref),
			},
		},
	}

	parameters := &pipelines.RunPipelineParameters{
		Resources:    resources,
		StagesToSkip: &stagesToSkip,
	}

	log.Info("run pipeline", "name", name, "args", string(toJSON(parameters)))
	r, err := c.RunPipeline(ctx, pipelines.RunPipelineArgs{RunParameters: parameters, Project: &project, PipelineId: pipelineId})
	if err != nil {
		return nil, fmt.Errorf("run pipeline fail: %w", err)
	}

	return r, nil
}

func Wait(ctx context.Context, c pipelines.Client, project string, r *pipelines.Run, log logr.Logger) (rr *pipelines.Run, err error) {
	log = log.WithValues("project", project, "pipeline", *r.Pipeline.Name, "name", *r.Name)
	log.Info("wait pipeline run")
	for {
		time.Sleep(10 * time.Second)

		rr, err = c.GetRun(ctx, pipelines.GetRunArgs{Project: &project, PipelineId: r.Pipeline.Id, RunId: r.Id})
		if err != nil {
			log.Error(err, "get pipeline run fail")
			continue
		}

		if rr.State == nil {
			log.Info("run.State == nil")
			continue
		}

		if *rr.State == pipelines.RunStateValues.Completed {
			log.Info("pipeline run is completed")
			return
		}

		log.Info("pipeline is running", "state", *r.State)
	}
}

// this function seems to be useless
func PrintRun(ctx context.Context, c pipelines.Client, project string, pipelineId *int, runId *int) error {
	// log := klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
	log := textlogger.NewLogger(textlogger.NewConfig())
	r, err := c.GetRun(ctx, pipelines.GetRunArgs{
		Project:    &project,
		PipelineId: pipelineId,
		RunId:      runId,
	})
	if err != nil {
		log.Error(err, "get pipeline run fail")
		return err
	}

	_ = json.NewEncoder(os.Stdout).Encode(r)
	return nil
}

func toJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
