package geo_skeleton_server

import (
	"fmt"
	"net/http"

	"github.com/paulmach/go.geojson"
)

// NewFeatureHandler creates a new feature and adds it to a layer.
// Layer is then saved to database. All active clients viewing layer
// are notified of update via websocket hub.
// @param apikey customer id
// @oaram ds datasource uuid
// @return json
func NewFeatureHandler(w http.ResponseWriter, r *http.Request) {
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

			body, err := job.GetRequestBody()
			if nil != err {
				return []byte{}, err
			}

			feat, err := geojson.UnmarshalFeature(body)
			if err != nil {
				return []byte{}, err
			}

			err = GeoDB.InsertFeature(datasource_id, feat)
			if err != nil {
				return []byte{}, err
			}

			data := HttpMessageResponse{Status: "success", Datasource: datasource_id, Data: "feature added"}
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

// ViewFeatureHandler finds feature in layer via array index. Returns feature geojson.
// @param apikey customer id
// @oaram ds datasource uuid
// @return feature geojson
func ViewFeatureHandler(w http.ResponseWriter, r *http.Request) {
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

			data, err := GeoDB.GetLayer(datasource_id)
			if err != nil {
				return []byte{}, err
			}

			feat_id, err := job.GetFeatureId()
			if err != nil {
				return []byte{}, err
			}

			for _, v := range data.Features {
				geo_id := fmt.Sprintf("%v", v.Properties["geo_id"])
				if geo_id == feat_id {
					js, err := v.MarshalJSON()
					return js, err
				}
			}

			// Feature not found
			err = fmt.Errorf("Not found")
			return []byte{}, err
		}
		return []byte{}, fmt.Errorf(`Unauthorized`)
	}()

	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js = job.MarshalJsonFromStruct(data)
	}

	job.SendJsonResponse(js)
}

// EditFeatureHandler finds feature in layer via array index. Edits feature.
// @param apikey customer id
// @oaram ds datasource uuid
func EditFeatureHandler(w http.ResponseWriter, r *http.Request) {
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

			body, err := job.GetRequestBody()
			if nil != err {
				return []byte{}, err
			}

			feat, err := geojson.UnmarshalFeature(body)
			if err != nil {
				return []byte{}, err
			}

			geo_id, err := job.GetFeatureId()
			if err != nil {
				return []byte{}, err
			}

			err = GeoDB.EditFeature(datasource_id, geo_id, feat)
			if err != nil {
				return []byte{}, err
			}

			data := HttpMessageResponse{Status: "success", Datasource: datasource_id, Data: "feature edited"}
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
