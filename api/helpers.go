package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/supergiant/supergiant/core"

	"github.com/gorilla/mux"
)

type errorMessage struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func renderError(w http.ResponseWriter, err error, status int) {
	msg := &errorMessage{status, err.Error()}
	body, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		panic(err)
	}
	http.Error(w, string(body), status)
}

// loadApp loads an App resource from URL params, or renders an HTTP Not Found
// error.
func loadApp(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.AppResource, error) {
	name := mux.Vars(r)["app_name"]
	app, err := core.Apps().Get(&name)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return app, nil
}

// loadComponent loads an Component resource from URL params, or renders an HTTP
// Not Found error.
func loadComponent(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.ComponentResource, error) {
	app, err := loadApp(core, w, r)
	if err != nil {
		return nil, err
	}

	name := mux.Vars(r)["comp_name"]
	component, err := app.Components().Get(&name)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return component, nil
}

// loadRelease loads an Release resource from URL params, or renders an HTTP
// Not Found error.
func loadRelease(c *core.Core, w http.ResponseWriter, r *http.Request) (*core.ReleaseResource, error) {
	component, err := loadComponent(c, w, r)
	if err != nil {
		return nil, err
	}

	// TODO
	var release *core.ReleaseResource
	releaseIdentifier := mux.Vars(r)["release_timestamp"]

	switch releaseIdentifier {
	case "current":
		if component.CurrentReleaseTimestamp == nil {
			err = errors.New("No current release")
			renderError(w, err, http.StatusNotFound)
			return nil, err
		}
		release, err = component.CurrentRelease()
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return nil, err
		}

	case "target":
		if component.TargetReleaseTimestamp == nil {
			err = errors.New("No target release")
			renderError(w, err, http.StatusNotFound)
			return nil, err
		}
		release, err = component.TargetRelease()
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return nil, err
		}

	default:
		release, err = component.Releases().Get(&releaseIdentifier)
		if err != nil {
			renderError(w, err, http.StatusNotFound)
			return nil, err
		}
	}

	return release, nil
}

// loadInstance loads an Instance resource from URL params, or renders an HTTP
// Not Found error.
func loadInstance(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.InstanceResource, error) {
	release, err := loadRelease(core, w, r)
	if err != nil {
		return nil, err
	}

	id := mux.Vars(r)["instance_id"]
	if err != nil {
		return nil, err
	}
	instance, err := release.Instances().Get(&id)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return instance, nil
}

// loadNode loads an Node resource from URL params, or renders an HTTP Not Found
// error.
func loadNode(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.NodeResource, error) {
	id := mux.Vars(r)["node_id"]
	node, err := core.Nodes().Get(&id)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return node, nil
}

// loadImageRepo loads an ImageRepo resource from URL params, or renders an HTTP
// Not Found error.
func loadImageRepo(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.ImageRepoResource, error) {
	name := mux.Vars(r)["name"]
	repo, err := core.ImageRepos().Get(&name)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return repo, nil
}

// loadEntrypoint loads an Entrypoint resource from URL params, or renders an HTTP
// Not Found error.
func loadEntrypoint(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.EntrypointResource, error) {
	domain := mux.Vars(r)["domain"]
	entrypoint, err := core.Entrypoints().Get(&domain)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return entrypoint, nil
}

// loadTask loads an Task resource from URL params, or renders an HTTP
// Not Found error.
func loadTask(core *core.Core, w http.ResponseWriter, r *http.Request) (*core.TaskResource, error) {
	id := mux.Vars(r)["id"]
	task, err := core.Tasks().Get(&id)
	if err != nil {
		renderError(w, err, http.StatusNotFound)
		return nil, err
	}

	return task, nil
}

// unmarshalBodyInto decodes a JSON body into an interface or renders an HTTP
// Not Found error.
func unmarshalBodyInto(w http.ResponseWriter, r *http.Request, out interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(out)
	if err != nil {
		renderError(w, err, http.StatusBadRequest)
		return err
	}
	return nil
}

// marshalBody marshals an interface into a JSON body or renders an HTTP Bad
// Request error.
func marshalBody(w http.ResponseWriter, in interface{}) (string, error) {
	out, err := json.MarshalIndent(in, "", "  ")
	// out, err := json.Marshal(in)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return "", err
	}
	return string(out) + "\n", nil
}

// renderWithStatusAccepted renders a response with HTTP status 202.
func renderWithStatusAccepted(w http.ResponseWriter, body string) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, body)
}

// renderWithStatusCreated renders a response with HTTP status 201.
func renderWithStatusCreated(w http.ResponseWriter, body string) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, body)
}

// renderWithStatusOK renders a response with HTTP status 200.
func renderWithStatusOK(w http.ResponseWriter, body string) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, body)
}
