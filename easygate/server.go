package easygate

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/darren/gpac"
	"github.com/elazarl/goproxy"
)

type loggerAdapter struct {
	impl *Logger
	re   *regexp.Regexp
}

func (l *loggerAdapter) Printf(format string, v ...interface{}) {
	if l.impl == nil {
		l.impl = GetLogger()
		l.re = regexp.MustCompile("\n$")
	}
	l.impl.Debug(l.re.ReplaceAllString(format, ""), v...)
}

type Server struct {
	service   *http.Server
	handler   *goproxy.ProxyHttpServer
	isRunning bool
	pac       *gpac.Parser
}

func NewServer(config *Config) *Server {
	server := new(Server)
	server.handler = goproxy.NewProxyHttpServer()
	server.handler.Tr.Proxy = server.proxy()
	server.handler.Verbose = true
	server.handler.Logger = new(loggerAdapter)
	connectReqHandler := func(req *http.Request) {
		addBasicAuth(req, config.Proxy.UserName, config.Proxy.Password)
	}
	server.handler.ConnectDial = goproxy.NewProxyHttpServer().NewConnectDialToProxyWithHandler(config.Proxy.Url, connectReqHandler)
	server.handler.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		addBasicAuth(req, config.Proxy.UserName, config.Proxy.Password)
		return req, nil
	}))
	server.isRunning = false
	pac, _ := gpac.New(loadPac(config.Serve.PacFilePath))
	server.pac = pac
	return server
}

func (server *Server) Start(config *Serve) {
	if server.isRunning {
		GetLogger().Info("Already running")
		return
	}
	if server.pac == nil {
		GetLogger().Info("Pac ignore. Always use proxy.")
	}
	server.service = &http.Server{Addr: ":" + config.ListenPort, Handler: server.handler}
	server.isRunning = true
	go func() {
		GetLogger().Info("Start service...")
		if err := server.service.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()
}

func (server *Server) Stop() {
	if !server.isRunning {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.service.Shutdown(ctx); err != nil {
		GetLogger().Error("Service shutdown error : %s", err)
		server.service.Close()
	}
	server.isRunning = false
	GetLogger().Info("Stop service...")
}

func (server *Server) IsRunning() bool {
	return server.isRunning
}

func (server *Server) isDirect(url *url.URL) bool {
	if server.pac == nil {
		return false
	}
	result, err := server.pac.FindProxyForURL(url.String())
	if err != nil {
		return false
	}
	if strings.HasPrefix(result, "DIRECT") {
		GetLogger().Info("Bypass proxy: %s", url.String())
		return true
	}
	return false
}

func (server *Server) proxy() func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		if server.isDirect(req.URL) {
			return nil, nil
		} else {
			return http.ProxyFromEnvironment(req)
		}
	}
}

func addBasicAuth(req *http.Request, user string, pass string) {
	req.Header.Set("Proxy-Authorization",
		fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString([]byte(user+":"+pass))))
}

func loadPac(filePath string) string {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		GetLogger().Warn("Pac file read faild. %s", err)
		return ""
	}
	return string(data)
}
