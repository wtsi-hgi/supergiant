package core

import (
	"fmt"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/client"
	"github.com/supergiant/supergiant/common"
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

func retryableOp(fn func() (*etcd.Response, error)) (resp *etcd.Response, err error) {
	waitErr := common.WaitFor("etcd connection", 15*time.Minute, 1*time.Second, func() (bool, error) {
		// doesn't matter what we do here, as long as we hit etcd
		if resp, err = fn(); err == nil || !strings.Contains(err.Error(), "etcd cluster is unavailable or misconfigured") {
			return true, nil
		}
		Log.Error(err.Error())
		Log.Warn("Waiting for etcd connection")
		return false, nil
	})
	if waitErr != nil {
		return nil, waitErr
	}
	return
}

func (e *etcdClient) compareAndSwap(key string, prevValue string, value string) (*etcd.Response, error) {
	return retryableOp(func() (*etcd.Response, error) {
		return e.kapi.Set(context.Background(), fullKey(key), value, &etcd.SetOptions{PrevValue: prevValue})
	})
}

func (e *etcdClient) create(key string, value string) (*etcd.Response, error) {
	return retryableOp(func() (*etcd.Response, error) {
		return e.kapi.Create(context.Background(), fullKey(key), value)
	})
}

func (e *etcdClient) get(key string) (*etcd.Response, error) {
	return retryableOp(func() (*etcd.Response, error) {
		return e.kapi.Get(context.Background(), fullKey(key), nil)
	})
}

func (e *etcdClient) update(key string, value string) (*etcd.Response, error) {
	return retryableOp(func() (*etcd.Response, error) {
		return e.kapi.Update(context.Background(), fullKey(key), value)
	})
}

func (e *etcdClient) delete(key string) (*etcd.Response, error) {
	return retryableOp(func() (*etcd.Response, error) {
		return e.kapi.Delete(context.Background(), fullKey(key), nil)
	})
}

func (e *etcdClient) createDir(key string) (*etcd.Response, error) {
	return retryableOp(func() (*etcd.Response, error) {
		return e.kapi.Set(context.Background(), key, "", &etcd.SetOptions{Dir: true})
	})
}
