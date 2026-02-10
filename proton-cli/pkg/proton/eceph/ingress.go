package eceph

import (
	"context"
	"embed"
	"io"
	"io/fs"

	"github.com/go-test/deep"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

var (
	//go:embed manifests
	manifests embed.FS

	ingress8001 = must(loadIngress(manifests, "manifests/networking.k8s.io/v1/ingresses/ingress-8001.yaml"))
	ingress8002 = must(loadIngress(manifests, "manifests/networking.k8s.io/v1/ingresses/ingress-8002.yaml"))

	ingresses = []networking.Ingress{
		ingress8001,
		ingress8002,
	}
)

func (m *Manager) reconcileIngresses() error {
	m.Logger.Info("reconcile kubernetes ingresses")

	for _, ing := range ingresses {
		name := types.NamespacedName{Name: ing.Name, Namespace: ing.Namespace}
		log := m.Logger.WithField("name", name)

		got := &networking.Ingress{}
		if err := m.Kube.Get(context.TODO(), name, got); errors.IsNotFound(err) {
			log.Info("create kubernetes ingress")
			if err := m.Kube.Create(context.TODO(), &ing); err != nil {
				return err
			}
			continue
		} else if err != nil {
			return err
		}

		// Not only spec but also annotations. Because annotations also affects
		// the configuration of ingress controller.
		var differences []string
		differences = append(differences, deep.Equal(got.Annotations, ing.Annotations)...)
		differences = append(differences, deep.Equal(got.Spec, ing.Spec)...)
		if differences == nil {
			log.Debug("skip updating kubernetes ingress")
			return nil
		}

		log.Info("update kubernetes ingress")
		return m.Kube.Update(context.TODO(), &ing)
	}

	return nil
}

// must is useful for initialization.
func must[E any](in E, err error) E {
	if err != nil {
		panic(err)
	}
	return in
}

func loadIngress(fs fs.FS, name string) (ing networking.Ingress, err error) {
	r, err := fs.Open(name)
	if err != nil {
		return
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(b, &ing)
	return
}
