package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-validator/validator"

	"github.com/supergiant/supergiant/common"
)

type Locatable interface {
	locationKey() string
	parent() Locatable
	child(string) Locatable
}

type Collection interface {
	initializeResource(Resource)
}

type Resource interface {
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

// // This zeros out field values with db:"-" tag, and omitsempty with JSON.
// func stripNonDbFields(m Resource) interface{} { // we return a copy here so we don't strip fields on the actual object
// 	rv := reflect.ValueOf(m).Elem()
//
// 	rxp, _ := regexp.Compile("(.+)Resource")
// 	typeName := rxp.FindStringSubmatch(rv.Type().Name())[1]
//
// 	oldT := rv.FieldByName(typeName).Elem()
// 	newT := reflect.New(oldT.Type())
// 	newT.Elem().Set(oldT)
//
// 	out := newT.Interface()
//
// 	val := reflect.ValueOf(out).Elem()
// 	val.Set(newT.Elem())
//
// 	t := val.Type()
//
// 	for i := 0; i < val.NumField(); i++ {
// 		tag := string(t.Field(i).Tag)
//
// 		if strings.Contains(tag, "db:\"-\"") {
// 			field := val.Field(i)
// 			field.Set(reflect.Zero(field.Type()))
// 		}
// 	}
//
// 	return out
// }

func taggedResourceFieldOf(field reflect.StructField, fieldValue reflect.Value) *taggedResourceField {
	tag := field.Tag.Get("sg")
	parts := strings.Split(tag, ",")

	out := new(taggedResourceField)
	out.Field = fieldValue

	for _, part := range parts {
		subparts := strings.Split(part, "=")
		switch len(subparts) {
		case 1:

			switch subparts[0] {
			case "readonly":
				out.Readonly = true
			case "nostore":
				out.NoStore = true
			case "private":
				out.Private = true
			default:
				panic("Do not recognize tag key " + subparts[0])
			}

		case 2: // e.g. default=10

			switch kind := fieldValue.Kind(); kind {
			case reflect.String:
				out.Default = subparts[1] // already a string
			case reflect.Int:
				integer, err := strconv.Atoi(subparts[1])
				if err != nil {
					panic(err)
				}
				out.Default = integer
			default:
				panic("Cannot parse tag default with value " + subparts[1])
			}

		default:
			panic("Could not parse Resource tag " + tag)
		}
	}

	return out
}

func gatherTaggedResourceFieldsInto(obj reflect.Value, taggedFields *[]*taggedResourceField) {
	objType := obj.Type() // *common.App on first iteration

	for i := 0; i < obj.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := obj.Field(i)

		// 1. if we see an SG tag, pass it to the tag parsing func, and continue
		// 2. if no SG tag, AND it's a struct (or ptr to), then we have to call recursively
		// 3. if no SG tag, and it's NOT a struct, we don't care

		if tag := field.Tag.Get("sg"); tag != "" {
			taggedField := taggedResourceFieldOf(field, fieldValue)
			*taggedFields = append(*taggedFields, taggedField)
			continue
		}

		if fieldValue.Kind() == reflect.Ptr { // && fieldValue.Elem().Kind() == reflect.Struct {
			fieldValue = fieldValue.Elem()
		}

		if fieldValue.Kind() == reflect.Struct {
			gatherTaggedResourceFieldsInto(fieldValue, taggedFields)
		}
	}
}

type taggedResourceField struct {
	Field    reflect.Value
	Readonly bool
	NoStore  bool
	Private  bool
	Default  interface{}
}

func taggedResourceFieldsOf(r Resource) (taggedFields []*taggedResourceField) {
	// &AppResource{    	<-- resourceValue
	// 	App: &App{				<-- commonValue
	// 		Name: "test",
	// 	},
	// }
	resourceValue := reflect.ValueOf(r).Elem()
	rxp, _ := regexp.Compile("(.+)Resource")
	commonName := rxp.FindStringSubmatch(resourceValue.Type().Name())[1]
	commonValue := resourceValue.FieldByName(commonName).Elem()

	gatherTaggedResourceFieldsInto(commonValue, &taggedFields)
	return
}

// ZeroReadonlyFields takes a Resource with pointer, and zeroes any fields with
// the tag sg:"readonly".
func ZeroReadonlyFields(r Resource) {
	for _, tf := range taggedResourceFieldsOf(r) {
		if tf.Readonly {
			tf.Field.Set(reflect.Zero(tf.Field.Type()))
		}
	}
}

// ZeroPrivateFields takes a Resource with pointer, and zeroes any fields with
// the tag sg:"private".
func ZeroPrivateFields(r Resource) {
	for _, tf := range taggedResourceFieldsOf(r) {
		if tf.Private {
			tf.Field.Set(reflect.Zero(tf.Field.Type()))
		}
	}
}

// copyWithoutNoStoreFields takes a Resource with pointer, and returns a copy
// with zero values for any fields with the tag sg:"nostore".
//
// NOTE nostore has to be used in conjunction with json:"omitempty" in order to
// prevent an empty value being stored in the DB. The alternative is to return
// a copy of the Resource in the form of a map, but that seems kinda difficult.
func copyWithoutNoStoreFields(r Resource) Resource {
	origR := reflect.ValueOf(r).Elem()
	newR := reflect.New(origR.Type())

	in := origR.Interface()
	out := newR.Interface()

	marshalled, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(marshalled, out); err != nil {
		panic(err)
	}

	resource := out.(Resource)

	for _, tf := range taggedResourceFieldsOf(resource) {
		if tf.NoStore {
			tf.Field.Set(reflect.Zero(tf.Field.Type()))
		}
	}

	return resource
}

// setDefaultFields takes a Resource with a pointer and sets the default value
// on all fields with the tag sg:"default=something".
func setDefaultFields(r Resource) {
	for _, tf := range taggedResourceFieldsOf(r) {
		if d := tf.Default; d != nil {
			tf.Field.Set(reflect.ValueOf(d).Elem())
		}
	}
}

// validateFields takes a Resource with a pointer and runs a validation on every
// field with the validate:"..." tag.
func validateFields(r Resource) error {
	return validator.Validate(r)
}
