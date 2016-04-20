package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/supergiant/supergiant/common"
)

type Locatable interface {
	locationKey() string
	parent() Locatable
	child(string) Locatable
}

type Collection interface {
	// locationKey() string
	// parent() Locatable
	// child(string) Locatable

	initializeResource(Resource)
}

type Resource interface {
	// locationKey() string
	// parent() Locatable
	// child(string) Locatable

	Action(string) *Action
	decorate() error
}

// // OrderedResource is similar to Resource, but provides a setID() method to
// // set an auto-generated ID from etcd on the Resource.
// type OrderedResource interface {
// 	setID(id common.ID)
// }

// NOTE Core will implement Locatable, but no top-level Collection (like Apps) will actually return Core as parent()
// So, Core can just implement parent() and locationKey() with nil and ""

func locationChain(location Locatable) (locs []Locatable) {
	for location != nil {
		locs = append([]Locatable{location}, locs...) // prepend (we're going up the tree; ascending; reversing)
		location = location.parent()
	}
	return
}

// ResourceLocation will take a Locatable and return a path just as "/apps/:name/components/:name/releases/:timestamp/instances/0"
// This will mirror API routes.
func ResourceLocation(location Locatable) (path string) {
	for _, loc := range locationChain(location) {
		path = path + "/" + loc.locationKey()
	}
	return
}

func LocateResource(coreLoc Locatable, path string) Resource {
	keys := strings.Split(path, "/")[1:] // element 0 is empty due to starting /
	location := coreLoc
	for _, key := range keys {
		location = location.child(key).(Locatable)
	}
	return location.(Resource)
}

func etcdKey(location Locatable) string {
	// In etcd, the top-most directory will be the name of the Collection of the
	// Location. For example, in etcd, a Release will be stored under:
	// /releases/:app_name/:component_name/:timestamp
	//
	// but in the API, it would be:
	// /apps/:name/components/:name/releases/:timestamp
	collection := location
	if _, ok := location.(Resource); ok {
		collection = location.parent()
	}
	path := "/" + collection.locationKey()

	for _, loc := range locationChain(location) {
		if _, ok := loc.(Resource); ok { // by filtering out Collections, this is a path of just IDs
			path = path + "/" + loc.locationKey()
		}
	}
	return path
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
