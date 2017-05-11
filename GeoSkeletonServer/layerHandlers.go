package geo_skeleton_server

import (
	"fmt"
	"net/http"

	"github.com/paulmach/go.geojson"
)

// ViewLayersHandler returns json containing customer layers
// @param apikey customer id
// @return json
func ViewLayersHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	customer, err := job.GetCustomer()
	var js []byte
	if nil == err {
		js = job.MarshalJsonFromStruct(customer)
	} else {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}
	job.SendJsonResponse(js)
}

// NewLayerHandler creates a new geojson layer. Saves layer to database and adds layer to customer
// @param apikey
// @return json
func NewLayerHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	customer, err := job.GetCustomer()
	var js []byte
	if nil == err {
		datasource_id, err := GeoDB.NewLayer()
		if nil == err {
			customer.addDatasource(datasource_id)
			data := HttpMessageResponse{Status: "success", Datasource: datasource_id}
			js = job.MarshalJsonFromStruct(data)
		}
	}
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}
	job.SendJsonResponse(js)
}

// ViewLayerHandler returns geojson of requested layer. Apikey/customer is checked for permissions to requested layer.
// @param ds
// @param apikey
// @return geojson
func ViewLayerHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	js, err := func() ([]byte, error) {
		customer, err := job.GetCustomer()
		if nil != err {
			return []byte{}, err
		}
		datasource_id, err := job.GetDatasource()
		if nil != err {
			return []byte{}, err
		}
		if customer.hasDatasource(datasource_id) {
			lyr, err := GeoDB.GetLayer(datasource_id)
			if nil != err {
				return []byte{}, fmt.Errorf(`Not found`)
			}
			js, err := lyr.MarshalJSON()
			return js, err
		}
		return []byte{}, fmt.Errorf(`Unauthorized`)
	}()
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}
	job.SendJsonResponse(js)
}

// DeleteLayerHandler deletes layer from database and removes it from customer list.
// @param ds
// @param apikey
// @return json
func DeleteLayerHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	js, err := func() ([]byte, error) {
		customer, err := job.GetCustomer()
		if nil != err {
			return []byte{}, err
		}
		datasource_id, err := job.GetDatasource()
		if nil != err {
			return []byte{}, err
		}
		if customer.hasDatasource(datasource_id) {
			customer.removeDatasource(datasource_id)
			err = GeoDB.DeleteLayer(datasource_id)
			if nil != err {
				return []byte{}, err
			}
			data := HttpMessageResponse{Status: "success", Datasource: datasource_id, Data: "datasource deleted"}
			js := job.MarshalJsonFromStruct(data)
			return js, err
		}
		return []byte{}, fmt.Errorf(`Unauthorized`)
	}()
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}
	job.SendJsonResponse(js)
}

// ViewLayerTimestampsHandler returns list of requested layer timestamps. Apikey/customer is checked for permissions to requested layer.
// @param ds
// @param apikey
// @return array
func ViewLayerTimestampsHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	js, err := func() ([]byte, error) {
		customer, err := job.GetCustomer()
		if nil != err {
			return []byte{}, err
		}
		datasource_id, err := job.GetDatasource()
		if nil != err {
			return []byte{}, err
		}
		if customer.hasDatasource(datasource_id) {
			lyr_ts, err := GeoDB.SelectTimeseriesDatasource(datasource_id)
			if nil != err {
				return []byte{}, err
			}
			timestamps := lyr_ts.GetSnapshots()
			var resp HttpMessageResponse
			resp.Status = "success"
			var snapshots []string
			for i := range timestamps {
				snapshots = append(snapshots, fmt.Sprintf("%v", timestamps[i]))
			}
			data := make(map[string]interface{})
			data["snapshots"] = snapshots
			resp.Data = data
			js := job.MarshalJsonFromStruct(resp)
			return js, err
		}
		return []byte{}, fmt.Errorf(`Unauthorized`)
	}()
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}
	job.SendJsonResponse(js)
}

// ViewLayerPerviousTimestampHandler returns geojson of requested layer for given timestamps. Apikey/customer is checked for permissions to requested layer.
// @param ds
// @param apikey
// @return array
func ViewLayerPerviousTimestampHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	js, err := func() ([]byte, error) {
		customer, err := job.GetCustomer()
		if nil != err {
			return []byte{}, err
		}
		datasource_id, err := job.GetDatasource()
		if nil != err {
			return []byte{}, err
		}
		if customer.hasDatasource(datasource_id) {
			ts, err := job.GetTimestamp()
			lyr_ts, err := GeoDB.SelectTimeseriesDatasource(datasource_id)
			if nil != err {
				return []byte{}, err
			}
			val, err := lyr_ts.GetPreviousByTimestamp(ts)
			if nil != err {
				return []byte{}, err
			}
			lyr, err := geojson.UnmarshalFeatureCollection([]byte(val))
			if err != nil {
				return []byte{}, err
			}
			js, err := lyr.MarshalJSON()
			return js, err
		}
		return []byte{}, fmt.Errorf(`Unauthorized`)
	}()
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}
	job.SendJsonResponse(js)
}
