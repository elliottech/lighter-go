package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	core "github.com/elliottech/lighter-go/client"
	"golang.org/x/net/proxy"
)

var _ core.MinimalHTTPClient = (*client)(nil)

type client struct {
	endpoint   string
	dialer     *net.Dialer
	transport  *http.Transport
	httpClient *http.Client
}

func NewClient(baseUrl, proxyUrl, localAddress string) core.MinimalHTTPClient {
	if baseUrl == "" {
		return nil
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxConnsPerHost:     1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}
	if proxyUrl != "" || localAddress != "" {
		var err error
		transport, err = createTransport(proxyUrl, localAddress)
		if err != nil {
			return nil
		}
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}

	return &client{
		endpoint:   baseUrl,
		dialer:     dialer,
		transport:  transport,
		httpClient: httpClient,
	}
}

// createTransport 创建自定义的 HTTP Transport，支持 proxy 和 localAddr
// proxy 优先，如果 proxy 为空，则使用 localAddr
func createTransport(proxyURL, localAddr string) (*http.Transport, error) {
	// 创建基础 transport
	transport := &http.Transport{
		MaxConnsPerHost:     1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	// 优先判断代理
	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}

		// 支持 http/https/socks5 代理
		if proxyURLParsed.Scheme == "socks5" {
			// 使用 golang.org/x/net/proxy 处理 socks5
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}
			proxyDialer, err := proxy.FromURL(proxyURLParsed, dialer)
			if err != nil {
				return nil, err
			}
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return proxyDialer.Dial(network, addr)
			}
		} else {
			// http/https 代理
			transport.Proxy = http.ProxyURL(proxyURLParsed)
		}
	} else if len(localAddr) > 0 {
		// proxy 为空时，才使用 localAddr
		tcpAddr, err := net.ResolveTCPAddr("tcp", localAddr+":0")
		if err != nil {
			return nil, err
		}

		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
			LocalAddr: tcpAddr,
		}
		transport.DialContext = dialer.DialContext
	} else {
		// 两者都为空，使用默认 dialer
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}
		transport.DialContext = dialer.DialContext
	}

	return transport, nil
}
