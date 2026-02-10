package main

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// 测试：静态路由匹配
func TestStaticRoute(t *testing.T) {
	router := NewRouter()
	router.AddRoute("/home", func(ctx *gin.Context) {
		t.Log("home page")
	})

	handler := router.Match("/home")
	if handler == nil {
		t.Fatal("Expected handler for /home, but got nil")
	}
}

// 测试：最长前缀匹配
func TestLongestPrefixMatch(t *testing.T) {
	router := NewRouter()

	testcase := []struct {
		h    HandlerFunc
		path string
		rs   []string
	}{
		{
			h:    func(*gin.Context) {},
			path: "/abcd/qwe",
			rs: []string{
				"/abcd/qwe",
				"/abcd/qwe/qwe",
				"/abcd/qwe/",
			},
		},
		{
			h:    func(*gin.Context) {},
			path: "/abcd/qwe/ads",
			rs: []string{
				"/abcd/qwe/ads",
				"/abcd/qwe/ads/qwe",
				"/abcd/qwe/ads/",
			},
		},
		{
			h:    func(*gin.Context) {},
			path: "/a/b/",
			rs: []string{
				"/a/b/",
			},
		},
	}

	for _, i := range testcase {
		router.AddRoute(i.path, i.h)
	}

	for _, i := range testcase {
		for _, r := range i.rs {
			handler := router.Match(r)
			if handler == nil {
				t.Fatalf("Expected handler %s for %s, but got nil", r, i.path)
			}
		}
	}
}

// 测试：未匹配情况
func TestNotFound(t *testing.T) {
	router := NewRouter()

	handler := router.Match("/not/exist")
	if handler != nil {
		t.Fatal("Expected nil for /not/exist, but got a handler")
	}

	router.AddRoute("/a/", func(ctx *gin.Context) {})
	if router.Match("/a") != nil {
		t.Fatal("Expected nil for /not/exist, but got a handler")
	}
}
