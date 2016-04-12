package core

import (
	"fmt"
	"reflect"

	"github.com/supergiant/supergiant/common"
)

// Collection is an interface for defining behavior of a collection of
// Resources.
type Collection interface {
	etcdKey(id common.ID) string

	// initializeResource is called when unmarshalling objects from etcd.
	// Primarily, it sets a pointer to the Collection on the Resource.
	initializeResource(r Resource) error
}

// Resource is an interface used mainly for generalized marshalling purposes for
// resource common.
type Resource interface {
	// // ToApiObj allows for each resource to return a stripped or embellished
	// // version of itself for rendering in API responses.
	// ToApiObj() Resource
}

// OrderedResource is similar to Resource, but provides a setID() method to
// set an auto-generated ID from etcd on the Resource.
type OrderedResource interface {
	setID(id common.ID)
}

// TODO should maybe move this to util or helper file
// getItemsPtrAndItemType takes a Resource, which must be of the List type, and
// returns a pointer to the Items slice of the List and the underlying item type
// of the slice.
func getItemsPtrAndItemType(r interface{}) (reflect.Value, reflect.Type) {
	// The concrete value of an interface is a pair of 32-bit words, one pointing
	// to information about the type implementing the interface, and the other
	// pointing to the underlying data in the type.
	interfaceValue := reflect.ValueOf(r)

	// In this case, we expect out to have been passed as a pointer, so that
	// interfaceValue's real value is actually:
	//
	// [ pointer ] --> [ AppList type ]
	// [ pointer ] --> [ pointer to instance of AppList ]
	//
	// So, calling this will dereference the pointer, providing the underlying
	// value of AppList. It makes AppList addressable AND settable.
	// NOTE it will also panic if out was not passed as a pointer.
	modelValue := interfaceValue.Elem()

	// Items field on any ModelList should be a slice of the relevant Model.
	itemsField := modelValue.FieldByName("Items")
	if !itemsField.IsValid() {
		panic(fmt.Errorf("no Items field in %#v", r))
	}

	// Items field is a slice here... (not a pointer)

	// Must first get the pointer of the slice with Addr(), so we can then call
	// Elem() to make it settable.
	itemsPtr := itemsField
	// Type() returns the underlying element type of the slice, and Elem()
	// allows us to utilize the type with reflect.New().
	itemType := itemsPtr.Type().Elem().Elem()

	// This initializes the empty items slice, so that we don't return null in API
	itemPtrType := reflect.PtrTo(itemType)
	emptyItems := reflect.MakeSlice(reflect.SliceOf(itemPtrType), 0, 0)
	itemsPtr.Set(emptyItems)

	return itemsPtr, itemType
}

func getFieldValue(r Resource, f string) reflect.Value {
	field := reflect.ValueOf(r).Elem().FieldByName(f)
	if !field.IsValid() {
		panic(fmt.Errorf("No %s field in %#v", f, r))
	}
	return field
}

func newTimestampValue() reflect.Value {
	return reflect.ValueOf(common.NewTimestamp())
}

func setCreatedTimestamp(r Resource) {
	getFieldValue(r, "Created").Set(newTimestampValue())
}

func setUpdatedTimestamp(r Resource) {
	getFieldValue(r, "Updated").Set(newTimestampValue())
}
