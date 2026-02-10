package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 处理函数类型
type HandlerFunc = gin.HandlerFunc

// Trie 树节点
type TrieNode struct {
	children map[string]*TrieNode
	handler  HandlerFunc
}

// 路由器结构
type Router struct {
	root *TrieNode
}

// 创建一个新的 Router
func NewRouter() *Router {
	return &Router{root: &TrieNode{children: make(map[string]*TrieNode)}}
}

// 添加路由
// gpt生成的代码实现并不优雅，没能完成真正的前缀树，勉强用吧
func (r *Router) AddRoute(path string, handler HandlerFunc) {
	segments := strings.Split(strings.TrimLeft(path, "/"), "/")
	node := r.root

	for _, segment := range segments {
		if _, exists := node.children[segment]; !exists {
			node.children[segment] = &TrieNode{children: make(map[string]*TrieNode)}
		}
		node = node.children[segment]
	}
	node.handler = handler
}

// 查找最长匹配的路由
// gpt生成的代码实现并不优雅，没能完成真正的前缀树，勉强用吧
func (r *Router) Match(path string) HandlerFunc {
	segments := strings.Split(strings.TrimLeft(path, "/"), "/")
	node := r.root
	var lastMatchedHandler HandlerFunc

	// 遍历路径段
	for _, segment := range segments {
		if nextNode, exists := node.children[segment]; exists {
			node = nextNode
			if node.handler != nil {
				lastMatchedHandler = node.handler // 记录当前节点的 handler
			}
		} else {
			break
		}
	}

	return lastMatchedHandler
}

func (r *Router) Handler(ctx *gin.Context) {
	rpath := ctx.Request.URL.Path
	h := r.Match(rpath)
	if h != nil {
		h(ctx)
	} else {
		ctx.HTML(http.StatusNotFound, `404.html`, gin.H{
			"route":   rpath,
			"repoter": "deploy-mproxy",
		})
	}
}
