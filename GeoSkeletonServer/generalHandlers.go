package geo_skeleton_server

import (
	"fmt"
	"html/template"
	"net/http"
)

// IndexHandler returns html page containing api docs
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://sjsafranek.github.io/gospatial/", 200)
	return
}

// MapHandler returns leaflet map view for customer layers
// @param apikey customer id
// @return map template
func MapHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	customer, err := job.GetCustomer()
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js := job.MarshalJsonFromStruct(data)
		job.SendJsonResponse(js)
		return
	}

	htmlFile := "./templates/map.html"
	tmpl, _ := template.ParseFiles(htmlFile)
	message := fmt.Sprintf(" %v %v [200]", r.Method, r.URL.Path)
	NetworkLogger.Info(r.RemoteAddr, message)
	tmpl.Execute(w, PageViewData{Apikey: customer.Apikey, Version: VERSION})
}

// DashboardHandler returns customer management gui.
// Allows customers to create and delete both geojson layers and tile baselayers.
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	customer, err := job.GetCustomer()
	if nil != err {
		data := HttpMessageResponse{Status: "error", Message: err.Error()}
		js := job.MarshalJsonFromStruct(data)
		job.SendJsonResponse(js)
		return
	}

	htmlFile := "./templates/management.html"
	tmpl, _ := template.ParseFiles(htmlFile)
	message := fmt.Sprintf(" %v %v [200]", r.Method, r.URL.Path)
	NetworkLogger.Info(r.RemoteAddr, message)
	tmpl.Execute(w, PageViewData{Apikey: customer.Apikey, Version: VERSION})
}
