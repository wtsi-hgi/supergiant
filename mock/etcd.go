package mock

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func (f *FakeEtcd) ReturnOnList(vals []string, err error) *FakeEtcd {
	f.GetFn = func() (*etcd.Response, error) {
		var nodes []*etcd.Node
		for _, val := range vals {
			nodes = append(nodes, &etcd.Node{Value: val})
		}
		return &etcd.Response{
			Node: &etcd.Node{
				Nodes: nodes,
			},
		}, nil
	}
	return f
}

func (f *FakeEtcd) ReturnOnGet(val string, err error) *FakeEtcd {
	f.GetFn = func() (*etcd.Response, error) {
		return &etcd.Response{
			Node: &etcd.Node{
				Value: val,
			},
		}, err
	}
	return f
}

func (f *FakeEtcd) OnCreate(clbk func(string, string) error) *FakeEtcd {
	f.CreateFn = func(key string, val string) (*etcd.Response, error) {
		if err := clbk(key, val); err != nil {
			return nil, err
		}
		return &etcd.Response{
			Node: &etcd.Node{
				Key:   key,
				Value: val,
			},
		}, nil
	}
	return f
}

func (f *FakeEtcd) OnUpdate(clbk func(string, string) error) *FakeEtcd {
	f.UpdateFn = func(key string, val string) (*etcd.Response, error) {
		if err := clbk(key, val); err != nil {
			return nil, err
		}
		return &etcd.Response{
			Node: &etcd.Node{
				Key:   key,
				Value: val,
			},
		}, nil
	}
	return f
}

func (f *FakeEtcd) OnDelete(clbk func(string) error) *FakeEtcd {
	f.DeleteFn = func(key string) (*etcd.Response, error) {
		if err := clbk(key); err != nil {
			return nil, err
		}
		return new(etcd.Response), nil
	}
	return f
}

// FakeEtcd is used to mock etcd operations in tests
type FakeEtcd struct {
	GetFn           func() (*etcd.Response, error)
	SetFn           func() (*etcd.Response, error)
	DeleteFn        func(key string) (*etcd.Response, error)
	CreateFn        func(key string, val string) (*etcd.Response, error)
	CreateInOrderFn func() (*etcd.Response, error)
	UpdateFn        func(key string, val string) (*etcd.Response, error)
}

// Get implements the etcd.KeysAPI interface
func (f *FakeEtcd) Get(ctx context.Context, key string, opts *etcd.GetOptions) (*etcd.Response, error) {
	return f.GetFn()
}

// Set implements the etcd.KeysAPI interface
func (f *FakeEtcd) Set(ctx context.Context, key, value string, opts *etcd.SetOptions) (*etcd.Response, error) {
	return f.SetFn()
}

// Delete implements the etcd.KeysAPI interface
func (f *FakeEtcd) Delete(ctx context.Context, key string, opts *etcd.DeleteOptions) (*etcd.Response, error) {
	return f.DeleteFn(key)
}

// Create implements the etcd.KeysAPI interface
func (f *FakeEtcd) Create(ctx context.Context, key, value string) (*etcd.Response, error) {
	return f.CreateFn(key, value)
}

// CreateInOrder implements the etcd.KeysAPI interface
func (f *FakeEtcd) CreateInOrder(ctx context.Context, dir, value string, opts *etcd.CreateInOrderOptions) (*etcd.Response, error) {
	return f.CreateInOrderFn()
}

// Update implements the etcd.KeysAPI interface
func (f *FakeEtcd) Update(ctx context.Context, key, value string) (*etcd.Response, error) {
	return f.UpdateFn(key, value)
}

// Watcher implements the etcd.KeysAPI interface
func (f *FakeEtcd) Watcher(key string, opts *etcd.WatcherOptions) etcd.Watcher {
	return nil
}
