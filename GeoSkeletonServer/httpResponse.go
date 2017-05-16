package geo_skeleton_server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"./utils"
	"github.com/gorilla/mux"
)

type HttpRequest struct {
	w            http.ResponseWriter
	r            *http.Request
	rid          string
	wroteHeaders bool
}

func (self *HttpRequest) GetRId() string {
	if "" == self.rid {
		self.rid = utils.NewAPIKey(8)
	}
	return self.rid
}

func (self *HttpRequest) GetRequestBody() ([]byte, error) {
	body, err := ioutil.ReadAll(self.r.Body)
	self.r.Body.Close()
	return body, err
}

func (self *HttpRequest) WriteHeaders(statusCode int) {
	self.wroteHeaders = true
	self.w.WriteHeader(statusCode)
	message := fmt.Sprintf("[%v] %v %v %v [%v]", self.GetRId(), self.r.RemoteAddr, statusCode, self.r.Method, self.r.URL.Path)
	NetworkLogger.Info(message)
}

func (self *HttpRequest) GetApikey() (string, error) {
	apikey := self.r.FormValue("apikey")
	if "" == apikey {
		self.WriteHeaders(http.StatusUnauthorized)
		err := fmt.Errorf(`Unauthorized`)
		return apikey, err
	}
	return apikey, nil
}

func (self *HttpRequest) GetAuthkey() (string, error) {
	apikey := self.r.FormValue("apikey")
	if "" == apikey {
		self.WriteHeaders(http.StatusUnauthorized)
		err := fmt.Errorf(`Unauthorized`)
		return apikey, err
	}
	return apikey, nil
}

func (self *HttpRequest) GetDatasource() (string, error) {
	vars := mux.Vars(self.r)
	if "" == vars["ds"] {
		self.WriteHeaders(http.StatusBadRequest)
		err := fmt.Errorf(`Missing parameter`)
		return vars["ds"], err
	}
	return vars["ds"], nil
}

func (self *HttpRequest) GetFeatureId() (string, error) {
	vars := mux.Vars(self.r)
	if "" == vars["k"] {
		self.WriteHeaders(http.StatusBadRequest)
		err := fmt.Errorf(`Missing parameter`)
		return vars["k"], err
	}
	return vars["k"], nil
}

func (self *HttpRequest) GetTimestamp() (int64, error) {
	vars := mux.Vars(self.r)
	ts, err := strconv.ParseInt(vars["ts"], 10, 64)
	if nil != err {
		self.WriteHeaders(http.StatusBadRequest)
	}
	return ts, err
}

func (self *HttpRequest) GetCustomer() (Customer, error) {
	apikey, err := self.GetApikey()
	if nil != err {
		self.WriteHeaders(http.StatusNotFound)
		return Customer{}, err
	}
	customer, err := DB.GetCustomer(apikey)
	if nil != err {
		self.WriteHeaders(http.StatusNotFound)
	}
	return customer, err
}

func (self *HttpRequest) Success() {
	self.w.WriteHeader(http.StatusOK)
}

func (self *HttpRequest) MarshalJsonFromString(data string) []byte {
	js, err := json.Marshal(data)
	if err != nil {
		self.WriteHeaders(http.StatusInternalServerError)
		return self.MarshalJsonFromString(`{"status": "error", "message": "` + err.Error() + `"}`)
	}
	return js
}

func (self *HttpRequest) MarshalJsonFromStruct(data interface{}) []byte {
	js, err := json.Marshal(data)
	if err != nil {
		self.WriteHeaders(http.StatusInternalServerError)
		return self.MarshalJsonFromString(`{"status": "error", "message": "` + err.Error() + `"}`)
	}
	return js
}

func (self *HttpRequest) InternalServerErrorHandler(err error) {
	message := fmt.Sprintf("[%v] %v %v [500]", self.GetRId(), self.r.Method, self.r.URL.Path)
	NetworkLogger.Critical(self.r.RemoteAddr, message)
	http.Error(self.w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusInternalServerError)
}

func (self *HttpRequest) BadRequestHandler(err error) {
	message := fmt.Sprintf("[%v] %v %v [400]", self.GetRId(), self.r.Method, self.r.URL.Path)
	NetworkLogger.Critical(self.r.RemoteAddr, message)
	http.Error(self.w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusBadRequest)
}

func (self *HttpRequest) NotFoundHandler(err error) {
	message := fmt.Sprintf("[%v] %v %v [404]", self.GetRId(), self.r.Method, self.r.URL.Path)
	NetworkLogger.Critical(self.r.RemoteAddr, message)
	http.Error(self.w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusNotFound)
}

func (self *HttpRequest) UnauthorizedHandler(err error) {
	message := fmt.Sprintf("[%v] %v %v [401]", self.GetRId(), self.r.Method, self.r.URL.Path)
	NetworkLogger.Critical(self.r.RemoteAddr, message)
	http.Error(self.w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusUnauthorized)
}

// Sends http response
func (self *HttpRequest) SendJsonResponse(js []byte) {
	if !self.wroteHeaders {
		self.WriteHeaders(http.StatusOK)
	}
	NetworkLogger.Trace(fmt.Sprintf("[%v] [In]  %v %v", self.GetRId(), self.r.RemoteAddr, self.r))
	NetworkLogger.Trace(fmt.Sprintf("[%v] [Out] %v %v", self.GetRId(), self.r.RemoteAddr, string(js)))
	// set response headers
	self.w.Header().Set("Content-Type", "application/json")
	// allow cross domain AJAX requests
	self.w.Header().Set("Access-Control-Allow-Origin", "*")
	self.w.Write(js)
}
