package geo_skeleton_server

import (
	"net/http"
	"runtime"
	"time"
)

//
// import (
// 	"./utils"
// )

// PingHandler provides an api route for server health check
func PingHandler(w http.ResponseWriter, r *http.Request) {
	job := HttpRequest{w: w, r: r}
	var data map[string]interface{}
	data = make(map[string]interface{})
	data["status"] = "success"
	result := make(map[string]interface{})
	result["result"] = "pong"
	result["registered"] = startTime.UTC()
	result["uptime"] = time.Since(startTime).Seconds()
	result["num_cores"] = runtime.NumCPU()
	data["data"] = result
	js := job.MarshalJsonFromStruct(data)
	job.SendJsonResponse(js)
}

// NewCustomerHandler superuser route to create new api customers/apikeys
// func NewCustomerHandler(w http.ResponseWriter, r *http.Request) {
// 	// NetworkLogger.Trace("[In] ", r)
// 	job := HttpRequest{w: w, r: r}
// 	if CheckAuthKey(w, r) {
// 		// new customer
// 		apikey := utils.NewAPIKey(12)
// 		customer := Customer{Apikey: apikey}
// 		err := DB.InsertCustomer(customer)
// 		if err != nil {
// 			ServerLogger.Error(err)
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		data := HttpMessageResponse{Status: "success", Apikey: apikey, Data: "customer created"}
// 		js, err := MarshalJsonFromStruct(w, r, data)
// 		if nil == err {
// 			SendJsonResponse(w, r, js)
// 		}
// 	}
// }
