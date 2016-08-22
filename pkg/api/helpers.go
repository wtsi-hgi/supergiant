package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

type Error struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

type Response struct {
	Status int
	Object interface{}
}

//------------------------------------------------------------------------------

type bodyDecodingError struct { // ------- status bad request
	err error
}

func (e *bodyDecodingError) Error() string {
	return "Error decoding JSON body: " + e.err.Error()
}

//------------------------------------------------------------------------------

func errorHttpStatus(err error) int {
	// 400
	if _, ok := err.(*bodyDecodingError); ok {
		return http.StatusBadRequest
	}
	// 404
	if err == gorm.ErrRecordNotFound {
		return http.StatusNotFound
	}
	// 500
	return http.StatusInternalServerError
}

func logHandler(core *core.Core) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validBasicAuth(core, w, r) {
			return
		}

		if core.LogPath == "" {
			w.Write([]byte("No log file configured\n"))
			return
		}

		file, err := os.Open(core.LogPath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// if _, err = file.Seek(0, 1); err != nil {
		// 	panic(err)
		// }

		stat, err := os.Stat(core.LogPath)
		// start := stat.Size() - 62

		// fmt.Println("start", start)

		buf := make([]byte, 1024)
		if _, err := file.ReadAt(buf, stat.Size()-int64(len(buf))); err != nil {
			panic(err)
		}
		w.Write(buf)
	}
}

func validBasicAuth(core *core.Core, w http.ResponseWriter, r *http.Request) bool {
	// TODO repeated in UI, move to core helpers
	w.Header().Set("WWW-Authenticate", `Basic realm="supergiant"`)
	username, password, _ := r.BasicAuth()
	if username != core.HTTPBasicUser || password != core.HTTPBasicPass {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
		return false
	}
	return true
}

func handlerFunc(core *core.Core, fn func(*core.Core, *http.Request) (*Response, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if !validBasicAuth(core, w, r) {
			return
		}

		resp, err := fn(core, r)

		if err != nil {
			status := errorHttpStatus(err)
			resp = &Response{
				Status: status,
				Object: &Error{status, err.Error()},
			}
		}

		body, marshalErr := json.MarshalIndent(resp.Object, "", "  ")
		if marshalErr != nil {
			panic(marshalErr)
		}

		w.WriteHeader(resp.Status)
		w.Write(append(body, []byte{10}...)) // add line break (without string conversion)
	}
}

//------------------------------------------------------------------------------

func parseID(r *http.Request) (*int64, error) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return nil, err
	}
	id64 := int64(id)
	return &id64, nil
}

func parseIncludes(r *http.Request) (includes []string) {
	if includesVal := r.URL.Query().Get("includes"); includesVal != "" {
		includes = strings.Split(includesVal, " ")
	}
	return
}

func decodeBodyInto(r *http.Request, item models.Model) error {
	if err := json.NewDecoder(r.Body).Decode(item); err != nil {
		return &bodyDecodingError{err}
	}
	models.ZeroReadonlyFields(item)
	return nil
}

func itemResponse(core *core.Core, item models.Model, status int) (*Response, error) {
	core.SetResourceActionStatus(item)
	return &Response{status, item}, nil
}

func handleList(core *core.Core, r *http.Request, model models.Model) (*Response, error) {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(model)), 0, 0)
	itemsPtr := reflect.New(slice.Type())
	items := itemsPtr.Elem()
	items.Set(slice)

	qstr := r.URL.Query()

	var andQueries []string
	for _, field := range models.IndexedFields(model) {
		if val := qstr.Get(field.JSONName); val != "" {
			if field.Kind == reflect.Int64 {
				andQueries = append(andQueries, fmt.Sprintf("%s = %s", field.JSONName, val))
			} else { // string
				andQueries = append(andQueries, fmt.Sprintf("%s = '%s'", field.JSONName, val))
			}
		}
	}
	andQuery := strings.Join(andQueries, " AND ")

	scope := core.DB
	if andQuery != "" {
		scope = scope.Where(andQuery)
	}

	if err := scope.Find(itemsPtr.Interface()); err != nil {
		return nil, err
	}

	for i := 0; i < items.Len(); i++ {
		core.SetResourceActionStatus(items.Index(i).Interface().(models.Model))
	}

	return &Response{
		http.StatusOK,
		items.Interface(),
	}, nil
}
