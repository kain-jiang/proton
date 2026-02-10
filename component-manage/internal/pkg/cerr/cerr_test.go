package cerr_test

import (
	"errors"
	"net/http"
	"testing"

	"component-manage/internal/pkg/cerr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.SkipStrings = []string{}
	reporterConfig.FullTrace = true
	RunSpecs(t, "Cerr Suite", suiteConfig, reporterConfig)
}

var _ = Describe("Code", func() {
	It("CodeValue", func() {
		Expect(cerr.ServerProduceError).Should(Equal(500024000))
		Expect(cerr.ParamsInvalidError).Should(Equal(400024001))
		Expect(cerr.PluginNotFoundError).Should(Equal(404024002))
		Expect(cerr.ComponentAlreadyExistsError).Should(Equal(409024003))
		Expect(cerr.ComponentNotFoundError).Should(Equal(404024004))
	})
})

var _ = Describe("Err", func() {
	It("NewError", func() {
		Expect(cerr.NewError(cerr.ServerProduceError, "test", "test")).
			Should(BeEquivalentTo(cerr.E{
				Code:    cerr.ServerProduceError,
				Message: "test",
				Cause:   "test",
			}))
	})
	It("AsError", func() {
		Expect(cerr.AsError(errors.New("test"))).
			Should(BeEquivalentTo(cerr.E{
				Code:    cerr.ServerProduceError,
				Message: "server produce internal error",
				Cause:   "test",
			}))
		Expect(cerr.AsError(cerr.NewError(cerr.ServerProduceError, "test", "test"))).
			Should(BeEquivalentTo(cerr.E{
				Code:    cerr.ServerProduceError,
				Message: "test",
				Cause:   "test",
			}))
	})
	It("E Functions", func() {
		err := cerr.NewError(cerr.ComponentNotFoundError, "test", "test")
		Expect(err.HCode()).Should(Equal(http.StatusNotFound))
		Expect(err.Error()).Should(Equal("code=404024004, message='test', cause='test'"))
	})
})
