package web

import (
	"crypto/tls"
	"crypto/x509"
	"cryptz/crypto"
	"github.com/gorilla/websocket"
	"net/http"
)

type Client interface {
	Run() error
}

type client struct {
	URL      string
	Origin   string
	CertPool *x509.CertPool
}

func (c *client) Run() error {
	dialer := &websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            c.CertPool,
		},
	}

	header := http.Header{
		"Origin": {c.Origin},
	}

	conn, _, err := dialer.Dial(c.URL, header)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if messageType == websocket.TextMessage {
			// decrypt the message before displaying
			result, err := crypto.DecryptMessage(string(p))
			if err != nil {
				return err
			}
			println(">>> " + result)
			println("")
		}
	}
}

func NewWSClient(url, origin string, pool *x509.CertPool) Client {
	return &client{
		URL:      url,
		Origin:   origin,
		CertPool: pool,
	}
}
