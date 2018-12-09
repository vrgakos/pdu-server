package main

import (
	"net/url"
	"net/http"
	"net/http/httputil"
	"crypto/tls"
	"pdu-server/app"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/static"
)

func httpMain(mainApp *app.App) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	r := gin.Default()
	bgw := mainApp.GetBrowserGw()

	r.GET("/bgw", func(c *gin.Context) {
		bgw.GetMelodyRouter().HandleRequest(c.Writer, c.Request)
	})

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	r.NoRoute(func(c *gin.Context) {
		c.File("public/index.html")
	})

/*	r.GET("/api/v1/:any", asdProxy())
	r.GET("/api/v1/:any/:any", asdProxy())
	r.GET("/api/v1/:any/:any/:any", asdProxy())
	r.GET("/api/v1/:any/:any/:any/:any", asdProxy())
	r.Use(ReverseProxy1("http://127.0.0.1:3000")) */

	r.Run(":3001")
}


func ReverseProxy1(target string) gin.HandlerFunc {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func NewKubernetesReverseProxy() gin.HandlerFunc {
	target, _ := url.Parse("https://152.66.247.126:6443/")

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			//			r := c.Request
			req = c.Request
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = req.URL.Path
			req.Header["Authorization"] = []string{"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJuYW1lc3BhY2UtY29udHJvbGxlci10b2tlbi1rbGQ2eCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJuYW1lc3BhY2UtY29udHJvbGxlciIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjIzNWYwNmVlLTJjNGQtMTFlOC05NTQxLTAwNTA1NjhmNzYwNSIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDprdWJlLXN5c3RlbTpuYW1lc3BhY2UtY29udHJvbGxlciJ9.uP_5dmMOek7-_CKAkXYi0TNezNH56AXGpjVI7wy-SoQ_wDu8JIWRPjz570nYI1T64mPdVe8RdCiCGjwlK_qG_qG_kFw1Kk3w2sRZnCfjj3TAJ-zHppBkDheicA7ugZBKBkUmLN3b1qf8rmnIDv50GaIho2eYe7hJRtVslvK-_PQjzLKdfZNnRJBz80KyAzu5Jo5umRcE0J8nTN_jP-KdxCHmByJRW6HhqZCslaBzAwihpIgYOeiWZa9Sy0YP7W7WD7LQj_X27om19dEIBUwas9pGs5YkD0lZd_jiY3erA3MXzoEUYBnKUgO49m809pFITNbSNFJMabpN2h7mDgz-Mg"}
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}


func asdProxy() gin.HandlerFunc {
	target, _ := url.Parse("https://152.66.247.126:6443/")

	return func(c *gin.Context) {
		proxy := NewReverseProxy(target)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}