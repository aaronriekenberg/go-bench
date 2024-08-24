package clients

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/aaronriekenberg/go-bench/config"
	"golang.org/x/net/http2"
)

type HttpClientPool struct {
	clients         []*http.Client
	mutex           sync.Mutex
	nextClientIndex int
}

func NewHttpClientPool(
	configuration config.HttpClientPoolConfiguration,
) *HttpClientPool {

	clients := make([]*http.Client, 0, configuration.NumHttpClients)

	for i := 0; i < configuration.NumHttpClients; i++ {

		if configuration.UseH2C {
			h2cClient := &http.Client{
				Transport: &http2.Transport{
					// So http2.Transport doesn't complain the URL scheme isn't 'https'
					AllowHTTP: true,
					// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
					DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
						var d net.Dialer
						return d.DialContext(ctx, network, addr)
					},
				},
			}
			clients = append(clients, h2cClient)
		} else {
			clients = append(clients, &http.Client{})
		}
	}

	slog.Info("newHttpClientPool",
		"numClients", len(clients),
		"useH2C", configuration.UseH2C,
	)

	return &HttpClientPool{
		clients: clients,
	}
}

func (hcp *HttpClientPool) GetClient() (client *http.Client, index int) {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	index = hcp.nextClientIndex
	hcp.nextClientIndex = (hcp.nextClientIndex + 1) % len(hcp.clients)

	client = hcp.clients[index]
	return
}
