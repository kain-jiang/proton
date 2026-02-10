package helm

import (
	"context"
	"testing"

	"taskrunner/pkg/component"
	test "taskrunner/test"
	testcharts "taskrunner/test/charts"
	"taskrunner/trait"
)

func TestRepos(t *testing.T) {
	tt := test.TestingT{T: t}
	r := testcharts.MemoryHelmRepoMock{
		RepoName: "test",
		FS:       test.TestCharts,
	}

	repos := NewHelmIndexRepo(&r)
	if _, err := repos.Fetch(context.Background(), &component.HelmComponent{
		HelmComponentSpec: component.HelmComponentSpec{
			Repository: r.RepoName,
		},
		ComponentMeta: trait.ComponentMeta{
			ComponentNode: trait.ComponentNode{
				Name:    "python3",
				Version: "0.1.0",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := repos.Fetch(context.Background(), &component.HelmComponent{
		HelmComponentSpec: component.HelmComponentSpec{
			Repository: "qwe",
		},
		ComponentMeta: trait.ComponentMeta{
			ComponentNode: trait.ComponentNode{
				Name:    "python3",
				Version: "9999.0.0",
			},
		},
	})
	tt.AssertError(trait.ErrHelmRepoNoFound, err)
}
