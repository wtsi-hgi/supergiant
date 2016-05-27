package core

import (
	"fmt"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type etcdClient struct {
	kapi etcd.KeysAPI
}

const (
	baseDir = "/supergiant"
)

func newetcdClient(endpoints []string) *etcdClient {
	client, err := etcd.New(etcd.Config{Endpoints: endpoints})
	if err != nil {
		panic(err)
	}
	etcdClient := etcdClient{etcd.NewKeysAPI(client)}
	etcdClient.createDir(baseDir)
	return &etcdClient
}

func fullKey(key string) string {
	return fmt.Sprintf("%s%s", baseDir, key)
}

func (e *etcdClient) compareAndSwap(key string, prevValue string, value string) (*etcd.Response, error) {
	return e.kapi.Set(context.Background(), fullKey(key), value, &etcd.SetOptions{PrevValue: prevValue})
}

func (e *etcdClient) create(key string, value string) (*etcd.Response, error) {
	return e.kapi.Create(context.Background(), fullKey(key), value)
}

func (e *etcdClient) get(key string) (*etcd.Response, error) {
	return e.kapi.Get(context.Background(), fullKey(key), nil)
}

func (e *etcdClient) update(key string, value string) (*etcd.Response, error) {
	return e.kapi.Update(context.Background(), fullKey(key), value)
}

func (e *etcdClient) delete(key string) (*etcd.Response, error) {
	return e.kapi.Delete(context.Background(), fullKey(key), nil)
}

// func (e *etcdClient) createInOrder(key string, value string) (*etcd.Response, error) {
// 	return e.kapi.CreateInOrder(context.Background(), fullKey(key), value, nil)
// }
//
// func (e *etcdClient) getInOrder(key string) (*etcd.Response, error) {
// 	return e.kapi.Get(context.Background(), fullKey(key), &etcd.GetOptions{Sort: true})
// }

func (e *etcdClient) createDir(key string) (*etcd.Response, error) {
	return e.kapi.Set(context.Background(), key, "", &etcd.SetOptions{Dir: true})
}
