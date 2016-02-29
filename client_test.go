package core

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	etcd "github.com/coreos/etcd/client"
)

type FakeKeysAPI struct {
	getResult    *etcd.Response
	getError     error
	createResult *etcd.Response
	createError  error
}

func (f FakeKeysAPI) Get(ctx context.Context, key string, opts *etcd.GetOptions) (*etcd.Response, error) {
	return f.getResult, f.getError
}

func (f FakeKeysAPI) Set(ctx context.Context, key, value string, opts *etcd.SetOptions) (*etcd.Response, error) {
	return f.createResult, f.createError
}

func (f FakeKeysAPI) Delete(ctx context.Context, key string, opts *etcd.DeleteOptions) (*etcd.Response, error) {
	return nil, nil
}

func (f FakeKeysAPI) Create(ctx context.Context, key, value string) (*etcd.Response, error) {
	return nil, nil
}

func (f FakeKeysAPI) CreateInOrder(ctx context.Context, dir, value string, opts *etcd.CreateInOrderOptions) (*etcd.Response, error) {
	return nil, nil
}

func (f FakeKeysAPI) Update(ctx context.Context, key, value string) (*etcd.Response, error) {
	return nil, nil
}

func (f FakeKeysAPI) Watcher(key string, opts *etcd.WatcherOptions) etcd.Watcher {
	return nil
}

// Incoming query (assert result)
func TestGet(t *testing.T) {
	cases := []struct {
		getResult     *etcd.Response
		getError      error
		expectedValue string
		expectedError error
	}{
		// When key exists
		{
			getResult: &etcd.Response{
				Action: "get",
				Node: &etcd.Node{
					Key:   "/test/foo",
					Value: "bar",
				},
			},
			getError:      nil,
			expectedValue: "bar",
			expectedError: nil,
		},
		// When key does not exist
		{
			getResult:     nil,
			getError:      errors.New("100: Key not found /test/foo [1]"),
			expectedValue: "",
			expectedError: errors.New("100: Key not found /test/foo [1]"),
		},
	}

	for _, c := range cases {
		client := &Client{
			KAPI: FakeKeysAPI{
				getResult: c.getResult,
				getError:  c.getError,
			},
		}
		resp := client.Get("test", "foo")

		if resp.raw != nil && resp.raw.Node.Value != c.expectedValue {
			t.Errorf("Expected %q but got %q", c.expectedValue, resp.raw.Node.Value)
		}

		if !reflect.DeepEqual(resp.err, c.expectedError) {
			t.Errorf("Expected err to be %q but it was %q", c.expectedError, resp.err)
		}
	}
}

// Incoming query (assert result)
func TestList(t *testing.T) {
	cases := []struct {
		getResult      *etcd.Response
		getError       error
		expectedValues []string
		expectedError  error
	}{
		// When key exists
		{
			getResult: &etcd.Response{
				Action: "get",
				Node: &etcd.Node{
					Key: "/foods",
					Nodes: []*etcd.Node{
						&etcd.Node{
							Key:   "/one",
							Value: "thos_beans",
						},
						&etcd.Node{
							Key:   "/two",
							Value: "taters",
						},
					},
				},
			},
			getError:       nil,
			expectedValues: []string{"thos_beans", "taters"},
			expectedError:  nil,
		},
		// When key does not exist
		{
			getResult:      nil,
			getError:       errors.New("100: Key not found /foods [1]"),
			expectedValues: nil,
			expectedError:  errors.New("100: Key not found /foods [1]"),
		},
	}

	for _, c := range cases {
		client := &Client{
			KAPI: FakeKeysAPI{
				getResult: c.getResult,
				getError:  c.getError,
			},
		}
		resp := client.List("test")

		if resp.raw != nil {
			for i, node := range resp.raw.Node.Nodes {
				expVal := c.expectedValues[i]
				if node.Value != expVal {
					t.Errorf("Expected %q but got %q", expVal, node.Value)
				}
			}
		}

		if !reflect.DeepEqual(resp.err, c.expectedError) {
			t.Errorf("Expected err to be %q but it was %q", c.expectedError, resp.err)
		}
	}
}

func TestCreate(t *testing.T) {
	cases := []struct {
		createResult  *etcd.Response
		createError   error
		expectedValue string
		expectedError error
	}{
		{
			createResult: &etcd.Response{
				Action: "set",
				Node: &etcd.Node{
					Key:   "/test/a",
					Value: "b",
				},
			},
			createError:   nil,
			expectedValue: "b",
			expectedError: nil,
		},
	}

	for _, c := range cases {
		client := &Client{
			KAPI: FakeKeysAPI{
				createResult: c.createResult,
				createError:  c.createError,
			},
		}
		resp := client.Create("test", "a", "b")

		fmt.Println(resp)

		val := resp.raw.Node.Value

		if val != c.expectedValue {
			t.Errorf("Expected %q but got %q", c.expectedValue, val)
		}

		if !reflect.DeepEqual(resp.err, c.expectedError) {
			t.Errorf("Expected err to be %q but it was %q", c.expectedError, resp.err)
		}
	}
}
