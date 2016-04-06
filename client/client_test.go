package client

import (
	"crypto/tls"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	// Create a client
	Convey("When creating a new Kubernetes client.", t, func() {
		client := New("test", "test", "test", true)

		Convey("We would expect the resulting client to look like our expected Client object.", func() {
			// Our expected output.
			expected := &Client{
				baseURL:  "test",
				Username: "test",
				Password: "test",
				http: &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				},
			}
			So(client, ShouldResemble, expected)

		})
	})
}
