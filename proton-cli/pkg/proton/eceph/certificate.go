package eceph

import (
	"bytes"
	"context"

	"github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *Manager) reconcileCertificate() error {
	m.Logger.Info("reconcile certificate")

	name := types.NamespacedName{Namespace: KubernetesNamespace, Name: m.Spec.TLS.Secret}

	data := generateSecretData(m.Spec.TLS.CertificateData, m.Spec.TLS.KeyData)

	return reconcileKubernetesSecret(m.Kube, name, data, m.Logger)
}

func generateSecretData(cert, key []byte) map[string][]byte {
	return map[string][]byte{
		core_v1.TLSCertKey:       cert,
		core_v1.TLSPrivateKeyKey: key,
	}
}

func reconcileKubernetesSecret(c client.Client, name types.NamespacedName, data map[string][]byte, log logrus.FieldLogger) error {
	secret := &core_v1.Secret{}
	if err := c.Get(context.TODO(), name, secret); errors.IsNotFound(err) {
		log.WithField("name", name).Info("create kubernetes secret")
		return c.Create(context.TODO(), &core_v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{Name: name.Name, Namespace: name.Namespace},
			Data:       data,
			Type:       core_v1.SecretTypeTLS,
		})
	} else if err != nil {
		return err
	}

	var changed bool
	for k, v := range data {
		if !bytes.Equal(secret.Data[k], v) {
			changed = true
			log.WithFields(logrus.Fields{"name": name, "key": k}).Debug("kubernetes secret data has changed")
			secret.Data[k] = v
		}
	}

	if !changed {
		log.WithField("name", name).Debug("skip update kubernetes secret")
		return nil
	}

	log.WithField("name", name).Info("update kubernetes secret")
	return c.Update(context.TODO(), secret)
}
