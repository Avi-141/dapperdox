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

// Package static provides handler for static resources.
package static

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/pat"

	"github.com/kenjones-cisco/dapperdox/render"
	"github.com/kenjones-cisco/dapperdox/render/asset"
)

// Register creates routes for each static resource
func Register(r *pat.Router) {
	log().Debug("registering not found handler in static package")

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		render.HTML(w, http.StatusNotFound, "error", render.DefaultVars(req, nil, map[string]interface{}{"error": "Page not found", "code": 404}))
	})

	log().Debug("registering static content handlers for static package")

	var allow bool

	for _, file := range asset.Names() {
		mimeType := mime.TypeByExtension(filepath.Ext(file))

		if mimeType == "" {
			continue
		}

		log().Debugf("Got MIME type: %s", mimeType)

		switch {
		case strings.HasPrefix(mimeType, "image"),
			strings.HasPrefix(mimeType, "text/css"),
			strings.HasSuffix(mimeType, "javascript"):
			allow = true
		default:
			allow = false
		}

		if allow {
			// Drop assets/static prefix
			path := strings.TrimPrefix(file, "assets/static")

			log().Debugf("registering handler for static asset: %s", path)

			r.Path(path).Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if b, err := asset.Asset("assets/static" + path); err == nil {
					w.Header().Set("Content-Type", mimeType)
					w.Header().Set("Cache-control", "public, max-age=259200")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(b)
					return
				}
				// This should never happen!
				log().Errorf("it happened ¯\\_(ツ)_/¯ %s", path)
				r.NotFoundHandler.ServeHTTP(w, req)
			})
		}
	}
}
