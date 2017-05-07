package geo_skeleton_server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/paulmach/go.geojson"
)

// NewFeatureHandler creates a new feature and adds it to a layer.
// Layer is then saved to database. All active clients viewing layer
// are notified of update via websocket hub.
// @param apikey customer id
// @oaram ds datasource uuid
// @return json
func NewFeatureHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)

	// Get request body
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		InternalServerErrorHandler(err, w, r)
		return
	}

	// Get ds from url path
	vars := mux.Vars(r)
	ds := vars["ds"]

	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {

		if CheckCustomerForDatasource(w, r, apikey, ds) {
			feat, err := geojson.UnmarshalFeature(body)
			if err != nil {
				BadRequestHandler(err, w, r)
				return
			}

			// Save feature to database
			err = GeoDB.InsertFeature(ds, feat)
			if err != nil {
				InternalServerErrorHandler(err, w, r)
				return
			}

			// Generate message
			data := HttpMessageResponse{Status: "success", Datasource: ds, Data: "feature added"}
			js, err := MarshalJsonFromStruct(w, r, data)
			if nil == err {
				// Update websockets
				conn := connection{ds: ds, ip: r.RemoteAddr}
				Hub.broadcast(true, &conn)
				// Return results
				SendJsonResponse(w, r, js)
			}
		}
	}
}

// ViewFeatureHandler finds feature in layer via array index. Returns feature geojson.
// @param apikey customer id
// @oaram ds datasource uuid
// @return feature geojson
func ViewFeatureHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)

	// Get ds from url path
	vars := mux.Vars(r)
	ds := vars["ds"]

	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		if CheckCustomerForDatasource(w, r, apikey, ds) {
			// Get layer from database
			data, err := GeoDB.GetLayer(ds)
			if err != nil {
				NotFoundHandler(err, w, r)
				return
			}

			// Check for feature
			var js []byte
			for _, v := range data.Features {
				geo_id := fmt.Sprintf("%v", v.Properties["geo_id"])
				if geo_id == vars["k"] {
					js, err = v.MarshalJSON()
					if err != nil {
						InternalServerErrorHandler(err, w, r)
						return
					}
					// Return results
					SendJsonResponse(w, r, js)
					return
				}
			}

			// Feature not found
			err = fmt.Errorf("Not found")
			NotFoundHandler(err, w, r)
		}
	}
}

// EditFeatureHandler finds feature in layer via array index. Edits feature.
// @param apikey customer id
// @oaram ds datasource uuid
func EditFeatureHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)

	// Get request body
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		InternalServerErrorHandler(err, w, r)
		return
	}

	// Get ds from url path
	vars := mux.Vars(r)
	ds := vars["ds"]
	geo_id := vars["k"]
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		if CheckCustomerForDatasource(w, r, apikey, ds) {

			// Unmarshal feature
			feat, err := geojson.UnmarshalFeature(body)
			if err != nil {
				BadRequestHandler(err, w, r)
				return
			}

			err = GeoDB.EditFeature(ds, geo_id, feat)
			if err != nil {
				NotFoundHandler(err, w, r)
			}

			// Generate message
			data := HttpMessageResponse{Status: "success", Datasource: ds, Data: "feature edited"}
			js, err := MarshalJsonFromStruct(w, r, data)
			if nil == err {

				// Update websockets
				conn := connection{ds: ds, ip: r.RemoteAddr}
				Hub.broadcast(true, &conn)

				// Feature not found
				SendJsonResponse(w, r, js)
			}
		}
	}
}
