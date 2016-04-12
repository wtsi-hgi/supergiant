package mock

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// FakeKeysAPI is used to mock etcd operations in tests
type FakeKeysAPI struct {
	GetFn           func() (*etcd.Response, error)
	SetFn           func() (*etcd.Response, error)
	DeleteFn        func() (*etcd.Response, error)
	CreateFn        func(val string) (*etcd.Response, error)
	CreateInOrderFn func() (*etcd.Response, error)
	UpdateFn        func(val string) (*etcd.Response, error)
}

// Get implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) Get(ctx context.Context, key string, opts *etcd.GetOptions) (*etcd.Response, error) {
	return f.GetFn()
}

// Set implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) Set(ctx context.Context, key, value string, opts *etcd.SetOptions) (*etcd.Response, error) {
	return f.SetFn()
}

// Delete implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) Delete(ctx context.Context, key string, opts *etcd.DeleteOptions) (*etcd.Response, error) {
	return f.DeleteFn()
}

// Create implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) Create(ctx context.Context, key, value string) (*etcd.Response, error) {
	return f.CreateFn(value)
}

// CreateInOrder implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) CreateInOrder(ctx context.Context, dir, value string, opts *etcd.CreateInOrderOptions) (*etcd.Response, error) {
	return f.CreateInOrderFn()
}

// Update implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) Update(ctx context.Context, key, value string) (*etcd.Response, error) {
	return f.UpdateFn(value)
}

// Watcher implements the etcd.KeysAPI interface
func (f *FakeKeysAPI) Watcher(key string, opts *etcd.WatcherOptions) etcd.Watcher {
	return nil
}
