package cms

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.SkipStrings = []string{
		//"CMS Get .*",
		//"CMS Set .*",
		//"CMS Del .*",
	}
	reporterConfig.FullTrace = true
	RunSpecs(t, "CMS Suite", suiteConfig, reporterConfig)
}

var _ = Describe("CMS", func() {
	var (
		rdsUrl  = "http://localhost.test/api/cms/v1/configuration/service/rds?namespace=anyshare"
		rdsFile = httpmock.File("testdata/rds.json")
		mockErr = fmt.Errorf("error")
		testCli = New("localhost.test", "anyshare")
	)
	{
		BeforeEach(func() {
			httpmock.ActivateNonDefault((testCli).(*cli).cli.GetClient())
		})
		AfterEach(func() {
			httpmock.DeactivateAndReset()
		})
	}

	Describe("Get", func() {
		Context("获取成功", func() {
			It("获取成功", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusOK, rdsFile),
				)
				{
					data, err := testCli.Get("rds")
					Ω(data).Should(And(HaveKey("host"), HaveKey("port"), HaveKey("user"), HaveKey("password")))
					Ω(err).Should(Succeed())
				}
			})
		})
		Context("获取失败", func() {
			It("获取出错", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					rdsUrl,
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.Get("rds")).Error().Should(MatchError(mockErr))
				}
			})
			It("获取失败", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				{
					Ω(testCli.Get("rds")).Error().Should(MatchError("code: 404, resp: null"))
				}
			})
			It("数据为空", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/rds.empty.json")),
				)
				{
					Ω(testCli.Get("rds")).Error().Should(MatchError("data not found in cms rds"))
				}
			})
		})
	})

	Describe("Set", func() {
		mockDel := func() *gomonkey.Patches {
			return gomonkey.ApplyMethod(
				reflect.TypeOf(testCli),
				"Del",
				func(_ *cli, _ string) error { return nil },
			)
		}

		Context("设置成功", func() {
			It("设置成功", func() {
				httpmock.RegisterResponder(
					http.MethodPatch,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusOK, rdsFile),
				)
				defer mockDel().Reset()
				{
					Ω(testCli.Set("rds", nil)).Should(Succeed())
				}
			})
		})

		Context("设置成功", func() {
			It("设置成功", func() {
				httpmock.RegisterResponder(
					http.MethodPatch,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, rdsFile),
				)
				httpmock.RegisterResponder(
					http.MethodPost,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusOK, rdsFile),
				)
				defer mockDel().Reset()
				{
					Ω(testCli.Set("rds", nil)).Should(Succeed())
				}
			})
		})

		Context("设置失败", func() {
			It("设置出错", func() {
				httpmock.RegisterResponder(
					http.MethodPatch,
					rdsUrl,
					httpmock.NewErrorResponder(mockErr),
				)
				defer mockDel().Reset()
				{
					Ω(testCli.Set("rds", nil)).Should(MatchError(mockErr))
				}
			})
			It("设置失败", func() {
				httpmock.RegisterResponder(
					http.MethodPatch,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)

				httpmock.RegisterResponder(
					http.MethodPost,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				defer mockDel().Reset()
				{
					Ω(testCli.Set("rds", nil)).Should(MatchError("code: 404, resp: null"))
				}
			})
		})
	})

	Describe("Del", func() {
		Context("删除成功", func() {
			It("删除成功", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusOK, nil),
				)
				{
					Ω(testCli.Del("rds")).Should(Succeed())
				}
			})
		})

		Context("删除失败", func() {
			It("删除出错", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					rdsUrl,
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.Del("rds")).Should(MatchError(mockErr))
				}
			})
			It("删除失败", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					rdsUrl,
					httpmock.NewJsonResponderOrPanic(http.StatusInternalServerError, nil),
				)
				{
					Ω(testCli.Del("rds")).Should(MatchError("code: 500, resp: null"))
				}
			})
		})
	})
})
