package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/imdario/mergo"
	"github.com/supergiant/supergiant/common"

	etcd "github.com/coreos/etcd/client"
)

type db struct {
	keys *etcdClient
}

func newdb(etcdEndpoints []string) *db {
	return &db{newetcdClient(etcdEndpoints)}
}

// TODO this is weird
func (db *db) setKeysAPI(kapi etcd.KeysAPI) {
	db.keys.kapi = kapi
}

func (db *db) list(r Collection, out interface{}) error {
	key := etcdKey(r.(Locatable))
	resp, err := db.keys.get(key)
	if err != nil && !isEtcdNotFoundErr(err) {
		// When listing, if it's key not found, it just means the dir has not been
		// created yet (which happens automatically when creating the first child
		// key). Here we return err ONLY if it's not that error
		return err
	}
	return decodeList(r, resp, out)
}

func (db *db) get(r Collection, id common.ID, out Resource) error {
	key := etcdKey(r.(Locatable)) + "/" + common.StringID(id) // TODO
	resp, err := db.keys.get(key)
	if err != nil {
		return err
	}
	return unmarshalNodeInto(r, resp.Node, out)
}

func (db *db) create(r Collection, id common.ID, m Resource) error {
	// NOTE we have to do this here to initialize Collection on the Resource
	r.initializeResource(m)

	setCreatedTimestamp(m)
	setDefaultFields(m)

	if err := validateFields(m); err != nil {
		return err
	}

	val, err := marshalResource(m)
	if err != nil {
		return err
	}

	key := etcdKey(m.(Locatable))

	_, err = db.keys.create(key, val)
	if err != nil {
		return err
	}

	// NOTE on create/update, we must explicitly call decorate() since we do not
	// unmarshal
	return m.decorate()
}

func (db *db) update(r Collection, id common.ID, m Resource) error {
	// NOTE we have to do this here to initialize Collection on the Resource
	r.initializeResource(m)

	setUpdatedTimestamp(m)

	if err := validateFields(m); err != nil {
		return err
	}

	val, err := marshalResource(m)
	if err != nil {
		return err
	}

	key := etcdKey(m.(Locatable))

	_, err = db.keys.update(key, val)
	if err != nil {
		return err
	}

	// NOTE on create/update, we must explicitly call decorate() since we do not
	// unmarshal
	return m.decorate()
}

// This works like a typical RESTful PATCH operation, a merge-update
func (db *db) patch(c Collection, id common.ID, r Resource) error {
	oldR := reflect.New(reflect.ValueOf(r).Elem().Type()).Interface().(Resource)
	if err := db.get(c, id, oldR); err != nil {
		return err
	}

	if err := mergo.Merge(r, oldR); err != nil {
		return err
	}

	return db.update(c, id, r)
}

func (db *db) delete(r Collection, id common.ID) error {
	key := etcdKey(r.(Locatable)) + "/" + common.StringID(id) // TODO
	_, err := db.keys.delete(key)
	return err
}

func (db *db) compareAndSwap(r Collection, id common.ID, old Resource, new Resource) error {
	key := etcdKey(old.(Locatable))

	oldVal, err := marshalResource(old)
	if err != nil {
		return err
	}
	newVal, err := marshalResource(new)
	if err != nil {
		return err
	}

	_, err = db.keys.compareAndSwap(key, oldVal, newVal)
	return err
}

func marshalResource(m Resource) (string, error) {
	t := copyWithoutNoStoreFields(m)

	out, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// To distinguish from all the other JSON unmarshalling errors
type dbUnmarshallingError struct {
	resource Resource
	jsonErr  error
}

func (e *dbUnmarshallingError) Error() string {
	return fmt.Sprintf("Error unmarshalling from etcd node into resource type %T: %s", e.resource, e.jsonErr)
}

func unmarshalNodeInto(r Collection, node *etcd.Node, m Resource) error {
	if err := json.Unmarshal([]byte(node.Value), m); err != nil {
		return &dbUnmarshallingError{m, err}
	}
	r.initializeResource(m)
	return m.decorate()
}

// CreateInOrder stuff...
// Was going to use "base" as a word here, like with file names. But it seems
// entirely weird to me that people inventing filesystems looked at:
//
// /home/dir/filename.txt
//
// and decided that "filename.txt" was the "base name".
func lastKeySegment(key string) common.ID {
	strs := strings.Split(key, "/")
	segment := strs[len(strs)-1]
	return &segment
}

func isEtcdNotFoundErr(err error) bool {
	etcdErr, ok := err.(etcd.Error)
	return ok && etcdErr.Code == etcd.ErrorCodeKeyNotFound
}

func decodeList(r Collection, resp *etcd.Response, out interface{}) error {
	itemsPtr, itemType := getItemsPtrAndItemType(out)

	// TODO we do this here, because the above method will initialize the Items
	// slice for us. Needs work.
	if resp == nil {
		return nil
	}

	for _, node := range resp.Node.Nodes {
		// Interface() is called to convert the new item Value into an interface
		// (that we can unmarshal to. The interface{} is then cast to ResourceList type.
		obj := reflect.New(itemType).Interface().(Resource)
		if err := unmarshalNodeInto(r, node, obj); err != nil {
			return err
		}

		// Get the Value of the unmarshalled object, and append it to the slice.
		newItem := reflect.ValueOf(obj).Elem().Addr()
		newItems := reflect.Append(itemsPtr, newItem)
		itemsPtr.Set(newItems)
	}
	return nil
}
