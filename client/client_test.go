package client

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	// Create a client
	Convey("When creating a new Supergiant client.", t, func() {
		client := New("test")

		Convey("We would expect the resulting client to look like our expected Client object.", func() {
			// Our expected output.
			expected := &Client{
				baseURL: "test",
				http:    &http.Client{},
			}
			So(client, ShouldResemble, expected)

		})
	})
}
