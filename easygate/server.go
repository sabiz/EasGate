package easygate

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/elazarl/goproxy"
)

type Server struct {
	service    *http.Server
}

func NewServer(config *Config) *Server {
	server := new(Server)
	handler := goproxy.NewProxyHttpServer()
	handler.Tr.Proxy = proxy()
	handler.Verbose = true
	connectReqHandler := func(req *http.Request) {
		addBasicAuth(req, config.Proxy.UserName, config.Proxy.Password)
	}
	handler.ConnectDial = goproxy.NewProxyHttpServer().NewConnectDialToProxyWithHandler(config.Proxy.Url, connectReqHandler)
	handler.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		addBasicAuth(req, config.Proxy.UserName, config.Proxy.Password)
		return req, nil
	}))
	server.service = &http.Server{Addr: ":"+config.Serve.ListenPort, Handler: handler}
	return server
}

func (server Server) Start() {
	// go func() {
		if err := server.service.ListenAndServe(); err != nil {
			panic(err)
		}
	// }()
}

func (server Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	if err := server.service.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func addBasicAuth(req *http.Request, user string, pass string) {
	req.Header.Set("Proxy-Authorization",
		fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString([]byte(user+":"+pass))))
}

func isDirect(url *url.URL) bool {
	return false // TODO Support pac file
}

func proxy() func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		if isDirect(req.URL) {
			return nil, nil
		} else {
			return http.ProxyFromEnvironment(req)
		}
	}
}
