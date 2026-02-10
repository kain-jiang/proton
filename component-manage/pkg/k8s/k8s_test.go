package k8s

import (
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// 必须加上 -gcflags="all=-N -l"
var _ = Describe("K8S", func() {
	mockReadFile := func(content []byte, err error) *gomonkey.Patches {
		return gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return content, err
		})
	}

	Describe("Client", func() {
		testCli := &cli{
			clientSet: fake.NewSimpleClientset(
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: "test",
					},
					Data: map[string]string{
						"test": "test-data",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"test": []byte("test-data"),
					},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node",
						Labels: map[string]string{
							"node-role.kubernetes.io/master": "",
						},
					},
				},
			),
		}
		It("GetMasterNodes", func() {
			nodes, err := testCli.GetMasterNodes()
			Expect(nodes).To(HaveLen(1))
			Expect(err).To(BeNil())
		})

		It("ConfigMapGet", func() {
			cm, err := testCli.ConfigMapGet("test-cm", "test")
			Expect(cm).NotTo(BeNil())
			Expect(err).To(BeNil())
			cm, err = testCli.ConfigMapGet("test-cm-not-found", "test")
			Expect(cm).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("ConfigMapDel", func() {
			err := testCli.ConfigMapDel("test-cm-not-found", "test")
			Expect(err).To(BeNil())
		})

		It("SecretDel", func() {
			err := testCli.ConfigMapDel("test-cm-not-found", "test")
			Expect(err).To(BeNil())
		})

		It("ConfigMapSet", func() {
			err := testCli.ConfigMapSet("test-cm-not-found", "test", map[string]string{"test-data": "test-data"})
			Expect(err).To(BeNil())
			cm, err := testCli.ConfigMapGet("test-cm-not-found", "test")
			Expect(cm).NotTo(BeNil())
			Expect(err).To(BeNil())
		})

		It("SelfNameSpace", func() {
			defer mockReadFile([]byte("hello"), nil).Reset()
			Expect(testCli.SelfNameSpace()).To(Equal("hello"))
		})

		It("SecretExist", func() {
			exist, err := testCli.SecretExist("test-secret-not-found", "test")
			Expect(exist).To(BeFalse())
			Expect(err).To(BeNil())
			exist, err = testCli.SecretExist("test-secret", "test")
			Expect(exist).To(BeTrue())
			Expect(err).To(BeNil())
		})

		It("SecretGet", func() {
			secret, err := testCli.SecretGet("test-secret-not-found", "test")
			Expect(secret).To(BeEmpty())
			Expect(err).To(BeNil())
			secret, err = testCli.SecretGet("test-secret", "test")
			Expect(secret).NotTo(BeEmpty())
			Expect(err).To(BeNil())
		})
	})

	Describe("Utils", func() {
		Describe("SelfNameSpace", func() {
			It("正确读取", func() {
				defer mockReadFile([]byte("hello"), nil).Reset()
				Ω(SelfNameSpace()).Should(Equal("hello"))
			})
			It("正确失败", func() {
				defer mockReadFile(nil, fmt.Errorf("ReadFailed")).Reset()
				Ω(SelfNameSpace()).Should(Equal("default"))
			})
		})

		Describe("ClusterDomain", func() {
			It("正确读取", func() {
				defer mockReadFile([]byte(`nameserver 10.96.0.10
search resource.svc.cluster.local2 svc.cluster.local2 cluster.local2
options ndots:5`), nil).Reset()
				Ω(ClusterDomain()).Should(Equal("cluster.local2"))
			})
		})
	})
})

func TestK8s(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8S Suite")
}
