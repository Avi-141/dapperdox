/*
Copyright (C) 2016-2017 dapperdox.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/

// Package reference provides handler for registered resources.
package reference

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/kenjones-cisco/dapperdox/render"
	"github.com/kenjones-cisco/dapperdox/spec"
)

type versionedMethod map[string]spec.Method      // key is version
type versionedResource map[string]*spec.Resource // key is version

var pathVersionMethod map[string]versionedMethod     // Key is path
var pathVersionResource map[string]versionedResource // Key is path

// Register creates routes for specification resource
func Register(r *mux.Router) {
	log().Info("Registering reference documentation")

	pathVersionMethod = make(map[string]versionedMethod)
	pathVersionResource = make(map[string]versionedResource)

	// Loop for all APISpecification's in the APISuite
	for _, specification := range spec.APISuite {

		specID := "/" + specification.ID

		log().Debugf("Registering reference for OpenAPI specification %q", specification.APIInfo.Title)

		for _, api := range specification.APIs {
			log().Debugf("  - Scanning API [%s] %s", api.ID, api.Name)
			r.Path(specID + "/reference/" + api.ID).Methods(http.MethodGet).HandlerFunc(apiHandler(specification, api))

			version := api.CurrentVersion

			for _, method := range api.Methods {
				path := specID + "/reference/" + api.ID + "/" + method.ID

				log().Debugf("    + method %s [%s]", path, method.Name)

				// Add version->method to pathVersionMethod
				if _, ok := pathVersionMethod[path]; !ok {
					pathVersionMethod[path] = make(versionedMethod)
					r.Path(path).Methods(http.MethodGet).HandlerFunc(methodHandler(specification, api, path))
				}
				pathVersionMethod[path][version] = method
			}
			for version, methods := range api.Versions {
				for _, method := range methods {
					log().Debugf("    + %s %s", method.ID, method.Name)
					path := specID + "/reference/" + api.ID + "/" + method.ID
					// Add version->resource to pathVersionResource
					if _, ok := pathVersionMethod[path]; !ok {
						pathVersionMethod[path] = make(versionedMethod)
						r.Path(path).Methods(http.MethodGet).HandlerFunc(methodHandler(specification, api, path))
					}
					pathVersionMethod[path][version] = method
				}
			}
		}

		log().Debug("  - Registering resources")
		for version, resources := range specification.ResourceList {
			log().Debugf("    - Version %s", version)
			for id, resource := range resources {
				path := specID + "/resources/" + id
				log().Debugf("      + resource %s", id)
				if _, ok := pathVersionResource[path]; !ok {
					pathVersionResource[path] = make(versionedResource)
					r.Path(path).Methods(http.MethodGet).HandlerFunc(globalResourceHandler(specification, path))
				}
				pathVersionResource[path][version] = resource
			}
		}
	}
}

func getVersionMethod(api spec.APIGroup, version string) []spec.Method {

	var methods []spec.Method
	var ok bool

	if methods, ok = api.Versions[version]; !ok {
		methods = api.Methods
	}
	return methods[0:]
}

func getMethodVersions(api spec.APIGroup, versions versionedMethod) []string {
	// See how many versions there are across the whole API. If 1, then version selection is not required.
	if len(api.Versions) < 2 {
		return nil
	}
	keys := make([]string, len(versions))
	ix := 0
	for key := range versions {
		keys[ix] = key
		ix++
	}
	return keys
}

func getAPIVersions(api spec.APIGroup) []string {
	count := len(api.Versions)
	if count < 2 {
		return nil // There is only one version defined
	}
	keys := make([]string, count)
	ix := 0
	for key := range api.Versions {
		keys[ix] = key
		ix++
	}
	return keys
}

// apiHandler is a http.Handler for rendering API reference docs
func apiHandler(specification *spec.APISpecification, api spec.APIGroup) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		version := req.FormValue("v") // Get the resource version
		if version == "" {
			version = api.CurrentVersion
		}
		versions := getAPIVersions(api)
		methods := getVersionMethod(api, version)

		tmpl := "api"
		customTmpl := "reference/" + api.ID
		if render.TemplateLookup(customTmpl) != nil {
			tmpl = customTmpl
		}

		log().Tracef("-- template: %s  Version %s", tmpl, version)

		render.HTML(w, http.StatusOK, tmpl,
			render.DefaultVars(req, specification,
				render.Vars{
					"Title":         api.Name,
					"API":           api,
					"Methods":       methods,
					"Version":       version,
					"Versions":      versions,
					"LatestVersion": api.CurrentVersion,
				}))
	}
}

// methodHandler is a http.Handler for rendering API method reference docs
func methodHandler(specification *spec.APISpecification, api spec.APIGroup, path string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		version := req.FormValue("v") // Get the resource version
		if version == "" {
			version = api.CurrentVersion
		}
		versions := getMethodVersions(api, pathVersionMethod[path])
		method := pathVersionMethod[path][version]

		tmpl := "method"
		customTmpl := "reference/" + api.ID + "/" + method.ID
		if render.TemplateLookup(customTmpl) != nil {
			tmpl = customTmpl
		}

		log().Tracef("-- template: %s  Version %s", tmpl, version)

		// TODO default to latest if version not found, or 404 ?
		method = pathVersionMethod[path][version]

		render.HTML(w, http.StatusOK, tmpl,
			render.DefaultVars(req, specification,
				render.Vars{
					"Title":         method.Name,
					"API":           api,
					"Method":        method,
					"Version":       version,
					"Versions":      versions,
					"LatestVersion": api.CurrentVersion,
				}))
	}
}

// globalResourceHandler is a http.Handler for rendering API resource reference docs
func globalResourceHandler(specification *spec.APISpecification, path string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		version := req.FormValue("v") // Get the resource version - blank is the latest
		if version == "" {
			version = "latest"
		}

		// Get list of versions
		var versions []string
		ix := 0
		versionList := pathVersionResource[path]

		if len(versionList) > 1 {
			// There is more than one version (there is always a "latest"), so
			// compile list of those available for resource
			versions = make([]string, len(pathVersionResource[path]))
			for key := range versionList {
				versions[ix] = key
				ix++
			}
		}

		resource := pathVersionResource[path][version]

		log().Debugf("Render resource %s", resource.ID)
		tmpl := "resource"

		customTmpl := "resources/" + resource.ID

		if render.TemplateLookup(customTmpl) != nil {
			tmpl = customTmpl
		}

		log().Tracef("-- template: %s  Version %s", tmpl, version)

		render.HTML(w, http.StatusOK, tmpl, render.DefaultVars(req, specification, render.Vars{"Title": resource.Title, "Resource": resource, "Version": version, "Versions": versions}))
	}
}
