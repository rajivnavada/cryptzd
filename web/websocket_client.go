package web

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/gorilla/websocket"
	"github.com/rajivnavada/gpgme"
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
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			logError(err, "Error in websocket client")
			return err
		}
		// Handle text messages
		if messageType == websocket.TextMessage {
			// decrypt the message before displaying
			if result, err := gpgme.DecryptMessage(string(p)); err != nil {
				println("------------------------------------------------------------")
				println("An error occured when trying to decrypt message")
				println("Ignoring this message")
				println("------------------------------------------------------------")
			} else {
				println("------------------------------------------------------------")
				println(result)
				println("------------------------------------------------------------")
			}
		}
	}
	return nil
}

func NewWSClient(url, origin string, pool *x509.CertPool) Client {
	return &client{
		URL:      url,
		Origin:   origin,
		CertPool: pool,
	}
}
