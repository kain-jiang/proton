package chartmuseum

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.SkipStrings = []string{
		//`Chartmuseum GetNewest .*`,
		//`Chartmuseum Get .*`,
		//`Chartmuseum Push .*`,
		//`Chartmuseum PushAfterDel .*`,
		//`Chartmuseum Del .*`,
	}
	reporterConfig.FullTrace = true
	RunSpecs(t, "Chartmuseum Suite", suiteConfig, reporterConfig)
}

var _ = Describe("Chartmuseum", func() {
	var (
		mockErr       = fmt.Errorf("error")
		testCli       = New("http://localhost.test", "test", "test")
		testHarborCli = New("http://localhost.test/chartrepo/test", "test", "test")
	)
	{
		BeforeEach(func() {
			httpmock.ActivateNonDefault((testCli).(*cli).cli.GetClient())
			httpmock.ActivateNonDefault((testHarborCli).(*cli).cli.GetClient())
		})
		AfterEach(func() {
			httpmock.DeactivateAndReset()
		})
	}

	Describe("GetNewest", func() {
		Context("查询成功", func() {
			It("获取正常", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/deploy-services.json")),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/charts/deploy-service-1.2.6-mission.tgz",
					httpmock.NewBytesResponder(http.StatusOK, httpmock.File("testdata/deploy-service-1.2.6-mission.tgz").Bytes()),
				)
				{
					c, err := testCli.GetNewest("deploy-service")
					Ω(c.Metadata.Version).Should(Equal("1.2.6-mission"))
					Ω(err).Should(Succeed())
				}
			})
		})
		Context("查询失败", func() {
			It("搜索出错", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.GetNewest("deploy-service")).Error().Should(MatchError(mockErr))
				}
			})
			It("搜索失败", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				{
					Ω(testCli.GetNewest("deploy-service")).Error().Should(MatchError("code: 404, resp: null"))
				}
			})
			It("获取出错", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/deploy-services.json")),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/charts/deploy-service-1.2.6-mission.tgz",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.GetNewest("deploy-service")).Error().Should(MatchError(mockErr))
				}
			})
			It("获取失败", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/deploy-services.json")),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/charts/deploy-service-1.2.6-mission.tgz",
					httpmock.NewBytesResponder(http.StatusNotFound, nil),
				)
				{
					Ω(testCli.GetNewest("deploy-service")).Error().Should(MatchError("code: 404, resp: "))
				}
			})
		})
	})

	Describe("Get", func() {
		Context("查询成功", func() {
			It("获取正常", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/deploy-service.json")),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/charts/deploy-service-1.2.6-mission.tgz",
					httpmock.NewBytesResponder(http.StatusOK, httpmock.File("testdata/deploy-service-1.2.6-mission.tgz").Bytes()),
				)
				{
					c, err := testCli.Get("deploy-service", "1.2.6-mission")
					Ω(c.Metadata.Version).Should(Equal("1.2.6-mission"))
					Ω(err).Should(Succeed())
				}
			})
		})
		Context("查询失败", func() {
			It("查询出错", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.Get("deploy-service", "1.2.6-mission")).Error().Should(MatchError(mockErr))
				}
			})
			It("查询失败", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				{
					Ω(testCli.Get("deploy-service", "1.2.6-mission")).Error().Should(MatchError("code: 404, resp: null"))
				}
			})
			It("获取出错", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/deploy-service.json")),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/charts/deploy-service-1.2.6-mission.tgz",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.Get("deploy-service", "1.2.6-mission")).Error().Should(MatchError(mockErr))
				}
			})
			It("获取失败", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, httpmock.File("testdata/deploy-service.json")),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/charts/deploy-service-1.2.6-mission.tgz",
					httpmock.NewBytesResponder(http.StatusNotFound, nil),
				)
				{
					Ω(testCli.Get("deploy-service", "1.2.6-mission")).Error().Should(MatchError("code: 404, resp: "))
				}
			})
		})
	})

	Describe("Push", func() {
		Context("推送成功", func() {
			It("推送成功", func() {
				httpmock.RegisterResponder(
					http.MethodPost,
					"http://localhost.test/api/charts",
					httpmock.NewJsonResponderOrPanic(http.StatusCreated, nil),
				)
				{
					Ω(testCli.Push("testdata/deploy-service-1.2.6-mission.tgz")).Should(Succeed())
				}
			})
		})
		Context("推送失败", func() {
			It("推送出错", func() {
				httpmock.RegisterResponder(
					http.MethodPost,
					"http://localhost.test/api/charts",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.Push("testdata/deploy-service-1.2.6-mission.tgz")).Error().Should(MatchError(mockErr))
				}
			})
			It("推送失败", func() {
				httpmock.RegisterResponder(
					http.MethodPost,
					"http://localhost.test/api/charts",
					httpmock.NewJsonResponderOrPanic(http.StatusConflict, nil),
				)
				{
					Ω(testCli.Push("testdata/deploy-service-1.2.6-mission.tgz")).Error().Should(MatchError("code: 409, resp: null"))
				}
			})
		})
	})

	Describe("Del", func() {
		Context("删除成功", func() {
			It("推送成功", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, nil),
				)
				{
					Ω(testCli.Del("deploy-service", "1.2.6-mission")).Should(Succeed())
				}
			})
		})
		Context("删除失败", func() {
			It("删除出错", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					Ω(testCli.Del("deploy-service", "1.2.6-mission")).Error().Should(MatchError(mockErr))
				}
			})
			It("删除失败", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusInternalServerError, nil),
				)
				{
					Ω(testCli.Del("deploy-service", "1.2.6-mission")).Error().Should(MatchError("code: 500, resp: null"))
				}
			})
		})
	})

	Describe("PushAfterDel", func() {
		Context("推送成功", func() {
			It("推送成功", func() {
				httpmock.RegisterResponder(
					http.MethodPost,
					"http://localhost.test/api/charts",
					httpmock.NewJsonResponderOrPanic(http.StatusCreated, nil),
				) // Mock Push Success
				httpmock.RegisterResponder(
					http.MethodDelete,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, nil),
				) // Mock Del Success
				{
					Ω(testCli.PushAfterDel("testdata/deploy-service-1.2.6-mission.tgz")).Should(Succeed())
				}
			})
		})
		Context("推送失败", func() {
			It("读取失败", func() {
				{
					Ω(testCli.PushAfterDel("testdata/deploy-service-notexist-1.2.6-mission.tgz")).Error().Should(HaveOccurred())
				}
			})
			It("删除失败", func() {
				httpmock.RegisterResponder(
					http.MethodDelete,
					"http://localhost.test/api/charts/deploy-service/1.2.6-mission",
					httpmock.NewErrorResponder(mockErr),
				) // Mock Del Failed
				{
					Ω(testCli.PushAfterDel("testdata/deploy-service-1.2.6-mission.tgz")).Error().Should(MatchError(mockErr))
				}
			})
		})
	})

	Describe("SearchRepoUrl", func() {
		Context("内置仓库", func() {
			It("返回baseUrl", func() {
				repoUrl, err := testCli.SearchRepoUrl("nginx-ingress-controller", "1.2.0", []string{"testrepo"})
				Ω(repoUrl).Should(Equal("http://localhost.test"))
				Ω(err).Should(Succeed())
			})
		})
		Context("外置仓库", func() {
			It("查询出错", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/chartrepo/testrepo1/charts/nginx-ingress-controller/1.2.0",
					httpmock.NewErrorResponder(mockErr),
				)
				{
					_, err := testHarborCli.SearchRepoUrl("nginx-ingress-controller", "1.2.0", []string{"testrepo1", "testrepo2"})
					Ω(err).Error().Should(MatchError(mockErr))
				}
			})
			It("查询失败1", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/chartrepo/testrepo1/charts/nginx-ingress-controller/1.2.0",
					httpmock.NewJsonResponderOrPanic(http.StatusInternalServerError, nil),
				)
				{
					_, err := testHarborCli.SearchRepoUrl("nginx-ingress-controller", "1.2.0", []string{"testrepo1", "testrepo2"})
					Ω(err).Error().Should(MatchError("code: 500, resp: null"))
				}
			})
			It("查询失败2", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/chartrepo/testrepo1/charts/nginx-ingress-controller/1.2.0",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/chartrepo/testrepo2/charts/nginx-ingress-controller/1.2.0",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				{
					_, err := testHarborCli.SearchRepoUrl("nginx-ingress-controller", "1.2.0", []string{"testrepo1", "testrepo2"})
					Ω(err).Error().Should(HaveOccurred())
				}
			})
			It("查询成功", func() {
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/chartrepo/testrepo1/charts/nginx-ingress-controller/1.2.0",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, nil),
				)
				httpmock.RegisterResponder(
					http.MethodGet,
					"http://localhost.test/api/chartrepo/testrepo2/charts/nginx-ingress-controller/1.2.0",
					httpmock.NewJsonResponderOrPanic(http.StatusOK, nil),
				)
				{
					repoUrl, err := testHarborCli.SearchRepoUrl("nginx-ingress-controller", "1.2.0", []string{"testrepo1", "testrepo2"})
					Ω(repoUrl).Should(Equal("http://localhost.test/chartrepo/testrepo2"))
					Ω(err).Should(Succeed())
				}
			})
		})
	})
})
