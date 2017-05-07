package geo_skeleton_server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func MarshalJsonFromString(w http.ResponseWriter, r *http.Request, data string) ([]byte, error) {
	js, err := json.Marshal(data)
	if err != nil {
		InternalServerErrorHandler(err, w, r)
		return js, err
	}
	return js, nil
}

func MarshalJsonFromStruct(w http.ResponseWriter, r *http.Request, data interface{}) ([]byte, error) {
	js, err := json.Marshal(data)
	if err != nil {
		InternalServerErrorHandler(err, w, r)
		return js, err
	}
	return js, nil
}

// Sends http response
func SendJsonResponse(w http.ResponseWriter, r *http.Request, js []byte) {
	// Log result
	message := fmt.Sprintf(" %v %v [200]", r.Method, r.URL.Path)
	NetworkLogger.Info(r.RemoteAddr, message)
	NetworkLogger.Trace("[Out] ", string(js))
	// set response headers
	w.Header().Set("Content-Type", "application/json")
	// allow cross domain AJAX requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// write response content
	w.Write(js)
}

// check request for valid authkey
func CheckAuthKey(w http.ResponseWriter, r *http.Request) bool {
	if SuperuserKey != r.FormValue("authkey") {
		err := fmt.Errorf(`{"status": "error", "message": "unauthorized"}`)
		UnauthorizedHandler(err, w, r)
		return false
	}
	return true
}

// Check for apikey in request
func GetApikeyFromRequest(w http.ResponseWriter, r *http.Request) string {
	// Get params
	apikey := r.FormValue("apikey")
	// Check for apikey in request
	if apikey == "" {
		err := fmt.Errorf(`{"status": "error", "message": "unauthorized"}`)
		UnauthorizedHandler(err, w, r)
	}
	// return apikey
	return apikey
}

// Get customer from database
func GetCustomerFromDatabase(w http.ResponseWriter, r *http.Request, apikey string) (Customer, error) {
	customer, err := DB.GetCustomer(apikey)
	if err != nil {
		NotFoundHandler(err, w, r)
		return customer, err
	}
	return customer, err
}

// Check customer datasource list
func CheckCustomerForDatasource(w http.ResponseWriter, r *http.Request, apikey string, ds string) bool {
	customer, err := GetCustomerFromDatabase(w, r, apikey)
	if nil != err {
		return false
	}
	// if !utils.StringInSlice(ds, customer.Datasources) {
	if !customer.hasDatasource(ds) {
		err := fmt.Errorf(`{"status": "error", "message": "unauthorized"}`)
		UnauthorizedHandler(err, w, r)
		return false
	}
	return true
}

//
func InternalServerErrorHandler(err error, w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf(" %v %v [500]", r.Method, r.URL.Path)
	NetworkLogger.Critical(r.RemoteAddr, message)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func BadRequestHandler(err error, w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf(" %v %v [400]", r.Method, r.URL.Path)
	NetworkLogger.Critical(r.RemoteAddr, message)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func NotFoundHandler(err error, w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf(" %v %v [404]", r.Method, r.URL.Path)
	NetworkLogger.Critical(r.RemoteAddr, message)
	http.Error(w, err.Error(), http.StatusNotFound)
}

func UnauthorizedHandler(err error, w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf(" %v %v [401]", r.Method, r.URL.Path)
	NetworkLogger.Critical(r.RemoteAddr, message)
	http.Error(w, err.Error(), http.StatusUnauthorized)
}
