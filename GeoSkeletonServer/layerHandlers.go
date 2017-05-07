package geo_skeleton_server

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/paulmach/go.geojson"
)

// ViewLayersHandler returns json containing customer layers
// @param apikey customer id
// @return json
func ViewLayersHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		customer, err := GetCustomerFromDatabase(w, r, apikey)
		if nil == err {
			js, err := MarshalJsonFromStruct(w, r, customer)
			if nil == err {
				SendJsonResponse(w, r, js)
			}
		}
	}
}

// NewLayerHandler creates a new geojson layer. Saves layer to database and adds layer to customer
// @param apikey
// @return json
func NewLayerHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		customer, err := GetCustomerFromDatabase(w, r, apikey)
		if nil == err {
			// Create datasource
			ds, err := GeoDB.NewLayer()
			if nil != err {
				InternalServerErrorHandler(err, w, r)
				return
			}
			// Add datasource uuid to customer
			// customer.Datasources = append(customer.Datasources, ds)
			customer.addDatasource(ds)
			DB.InsertCustomer(customer)
			// Generate message
			data := HttpMessageResponse{Status: "success", Datasource: ds}
			js, err := MarshalJsonFromStruct(w, r, data)
			if nil == err {
				SendJsonResponse(w, r, js)
			}
		}
	}
}

// ViewLayerHandler returns geojson of requested layer. Apikey/customer is checked for permissions to requested layer.
// @param ds
// @param apikey
// @return geojson
func ViewLayerHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)
	vars := mux.Vars(r)
	ds := vars["ds"]
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		if CheckCustomerForDatasource(w, r, apikey, ds) {
			// Get layer from database
			lyr, err := GeoDB.GetLayer(ds)
			if nil != err {
				NotFoundHandler(err, w, r)
				return
			}
			// Marshal datasource layer to json
			js, err := lyr.MarshalJSON()
			if nil != err {
				InternalServerErrorHandler(err, w, r)
				return
			}
			// Return layer json
			SendJsonResponse(w, r, js)
		}
	}
}

// DeleteLayerHandler deletes layer from database and removes it from customer list.
// @param ds
// @param apikey
// @return json
func DeleteLayerHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)
	vars := mux.Vars(r)
	ds := vars["ds"]
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		customer, err := GetCustomerFromDatabase(w, r, apikey)
		if nil == err {
			if CheckCustomerForDatasource(w, r, apikey, ds) {
				customer.removeDatasource(ds)
				DB.InsertCustomer(customer)
				// Delete layer from database
				err = GeoDB.DeleteLayer(ds)
				if nil != err {
					InternalServerErrorHandler(err, w, r)
					return
				}
				// Generate message
				data := HttpMessageResponse{Status: "success", Datasource: ds, Data: "datasource deleted"}
				js, err := MarshalJsonFromStruct(w, r, data)
				if nil == err {
					SendJsonResponse(w, r, js)
				}
			}
		}
	}
}

// ViewLayerTimestampsHandler returns list of requested layer timestamps. Apikey/customer is checked for permissions to requested layer.
// @param ds
// @param apikey
// @return array
func ViewLayerTimestampsHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)
	vars := mux.Vars(r)
	ds := vars["ds"]
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		if CheckCustomerForDatasource(w, r, apikey, ds) {
			lyr_ts, err := GeoDB.SelectTimeseriesDatasource(ds)
			if nil != err {
				NotFoundHandler(err, w, r)
				return
			}
			timestmaps := lyr_ts.GetSnapshots()
			js, err := MarshalJsonFromStruct(w, r, timestmaps)
			if nil == err {
				SendJsonResponse(w, r, js)
			}
		}
	}
}

// ViewLayerPerviousTimestampHandler returns geojson of requested layer for given timestamps. Apikey/customer is checked for permissions to requested layer.
// @param ds
// @param apikey
// @return array
func ViewLayerPerviousTimestampHandler(w http.ResponseWriter, r *http.Request) {
	NetworkLogger.Debug("[In] ", r)
	vars := mux.Vars(r)
	ds := vars["ds"]
	ts, err := strconv.ParseInt(vars["ts"], 10, 64)
	if nil != err {
		BadRequestHandler(err, w, r)
		return
	}
	apikey := GetApikeyFromRequest(w, r)
	if "" != apikey {
		if CheckCustomerForDatasource(w, r, apikey, ds) {
			lyr_ts, err := GeoDB.SelectTimeseriesDatasource(ds)
			if nil != err {
				NotFoundHandler(err, w, r)
				return
			}
			val, err := lyr_ts.GetPreviousByTimestamp(ts)
			if nil != err {
				InternalServerErrorHandler(err, w, r)
				return
			}
			// Unmarshal feature
			lyr, err := geojson.UnmarshalFeatureCollection([]byte(val))
			if err != nil {
				InternalServerErrorHandler(err, w, r)
				return
			}
			// Marshal datasource layer to json
			js, err := lyr.MarshalJSON()
			if nil != err {
				InternalServerErrorHandler(err, w, r)
				return
			}
			SendJsonResponse(w, r, js)
		}
	}
}
