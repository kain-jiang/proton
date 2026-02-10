package util

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	It("StringElse", func() {
		Ω(StringElse("1", "2")).Should(Equal("1"))
		Ω(StringElse("", "2")).Should(Equal("2"))
	})

	It("Min", func() {
		Ω(Min(2.1, 3.2)).Should(Equal(2.1))
		Ω(Min(6, 3)).Should(Equal(3))
	})

	It("Max", func() {
		Ω(Max(2, 3)).Should(Equal(3))
		Ω(Max(6.2, 3.1)).Should(Equal(6.2))
	})

	It("InSlice", func() {
		Ω(InSlice(2, []int{2, 3})).Should(BeTrue())
		Ω(InSlice(6.2, []float32{6.2, 7.8})).Should(BeTrue())
		Ω(InSlice("OK", []string{"yes", "no"})).Should(BeFalse())
		Ω(InSlice("OK", []string{"OK", "yes"})).Should(BeTrue())
	})

	It("VersionOrdinal", func() {
		Ω(VersionOrdinal("7.0.2.1") > VersionOrdinal("7.0.2")).Should(BeTrue())
		Ω(VersionOrdinal("7.0.3.1") > VersionOrdinal("7.1.2")).Should(BeFalse())
		Ω(VersionOrdinal("7.0.3.1") > VersionOrdinal("")).Should(BeTrue())
		Ω(VersionOrdinal("7.0.3.10") > VersionOrdinal("7.0.3.9")).Should(BeTrue())
		Ω(VersionOrdinal("2.0.6") > VersionOrdinal("2.0.7")).Should(BeFalse())
		// Ω(VersionOrdinal("2.0.6") > VersionOrdinal("2.0.6")).Should(BeFalse()) //nolint:staticcheck
	})

	It("IsIPv6", func() {
		Ω(IsIPv6("::1")).Should(BeTrue())
		Ω(IsIPv6("127.0.0.1")).Should(BeFalse())
		Ω(IsIPv6("localhost")).Should(BeFalse())
		Ω(IsIPv6("")).Should(BeFalse())
	})

	It("ErrorString", func() {
		Ω(ErrorString(nil, "default")).Should(Equal("default"))
		Ω(ErrorString(fmt.Errorf("error"), "default")).Should(Equal("error"))
	})

	It("Map", func() {
		Ω(Map([]int{1, 2, 3}, func(i int) int { return i * 2 })).
			Should(Equal([]int{2, 4, 6}))
	})

	Context("Struct", func() {
		type data struct {
			A string `yaml:"a"`
			B int    `yaml:"b"`
		}

		It("Map", func() {
			Ω(ToMap(&data{A: "a", B: 1})).Should(Equal(map[string]any{"a": "a", "b": uint64(1)}))
			Ω(ToMap(data{A: "a", B: 1})).Should(Equal(map[string]any{"a": "a", "b": uint64(1)}))
		})

		It("FromMap", func() {
			Ω(FromMap[data](map[string]any{"a": "a", "b": 1})).Should(Equal(&data{A: "a", B: 1}))
		})

		It("ToYamlBytes", func() {
			Ω(ToYamlBytes(data{A: "a", B: 1})).To(Equal([]byte("a: a\nb: 1\n")))
		})

		It("FromYamlBytes", func() {
			Ω(FromYamlBytes[data]([]byte("a: a\nb: 1\n"))).To(Equal(data{A: "a", B: 1}))
		})
	})

	It("AnyMapToMapAny", func() {
		Ω(AnyMapToMapAny(map[string]string{"a": "a", "b": "b"})).Should(Equal(map[string]any{"a": "a", "b": "b"}))
	})
})

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Util Suite")
}
