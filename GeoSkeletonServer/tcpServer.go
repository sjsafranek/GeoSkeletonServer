package geo_skeleton_server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/textproto"
	"os"
	"os/exec"
	"strings"

	"./utils"
	"github.com/paulmach/go.geojson"
)

const (
	TCP_DEFAULT_CONN_HOST = "localhost"
	TCP_DEFAULT_CONN_PORT = "3333"
	TCP_DEFAULT_CONN_TYPE = "tcp"
)

type TcpServer struct {
	Host             string
	Port             string
	ConnType         string
	ActiveTcpClients int
}

func (self *TcpServer) getHost() string {
	if self.Host == "" {
		self.Host = TCP_DEFAULT_CONN_HOST
		return TCP_DEFAULT_CONN_HOST
	}
	return self.Host
}

func (self *TcpServer) getPort() string {
	if self.Port == "" {
		self.Port = TCP_DEFAULT_CONN_PORT
		return TCP_DEFAULT_CONN_PORT
	}
	return self.Port
}

func (self *TcpServer) getConnType() string {
	if self.ConnType == "" {
		self.ConnType = TCP_DEFAULT_CONN_TYPE
		return TCP_DEFAULT_CONN_TYPE
	}
	return self.ConnType
}

func (self TcpServer) Start() {
	go func() {
		// Check settings and apply defaults
		serv := fmt.Sprintf("%v:%v", self.getHost(), self.getPort())

		// Listen for incoming connections.
		l, err := net.Listen(self.getConnType(), serv)
		if err != nil {
			ServerLogger.Error("Error listening:", err.Error())
			panic(err)
		}
		ServerLogger.Info("Tcp Listening on " + serv)

		// Close the listener when the application closes.
		defer l.Close()

		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				NetworkLogger.Error("Error accepting connection: ", err.Error())
				return
				// conn.Close()
			}

			NetworkLogger.Info("Connection open ", conn.RemoteAddr().String(), " [TCP]")

			// check for local connection
			if strings.Contains(conn.RemoteAddr().String(), "127.0.0.1") {
				// Handle connections in a new goroutine.
				go self.tcpClientHandler(conn)
			} else {
				// don't accept not local connections
				conn.Close()
			}

		}
	}()
}

// close tcp client
func (self *TcpServer) closeClient(conn net.Conn) {
	self.ActiveTcpClients--
	conn.Close()
}

// Handles incoming requests.
func (self *TcpServer) tcpClientHandler(conn net.Conn) {

	self.ActiveTcpClients++
	defer self.closeClient(conn)

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	for {
		// will listen for message to process ending in newline (\n)
		message, _ := tp.ReadLine()

		// output message received
		NetworkLogger.Info("[TCP] Message Received: ", string([]byte(message)))

		// json parse message
		req := TcpMessage{}
		err := json.Unmarshal([]byte(message), &req)
		if err != nil {
			// invalid message
			// close connection
			// '\x04' end of transmittion character
			NetworkLogger.Warn("error:", err)
			resp := `{"status": "error", "error": "` + fmt.Sprintf("%v", err) + `",""}`
			conn.Write([]byte(resp + "\n"))
			NetworkLogger.Info("Connection closed", " [TCP]")
			return
		}

		switch {

		case req.Method == "ping":
			self.handleSuccess(`{"message": "pong", "version": "`+VERSION+`"}`, conn)

		case req.Method == "help":
			conn.Write([]byte("Methods:\n"))
			conn.Write([]byte("\t ping\n"))
			conn.Write([]byte("\t assign_datasource\n"))
			conn.Write([]byte("\t create_apikey\n"))
			conn.Write([]byte("\t insert_apikey\n"))
			conn.Write([]byte("\t insert_feature\n"))
			conn.Write([]byte("\t edit_feature\n"))
			conn.Write([]byte("\t create_datasource\n"))
			conn.Write([]byte("\t export_apikeys\n"))
			conn.Write([]byte("\t export_apikey\n"))
			conn.Write([]byte("\t export_datasources\n"))
			conn.Write([]byte("\t export_datasource\n"))
			conn.Write([]byte("\t import_file\n"))

		// APIKEYS
		case req.Method == "create_apikey":
			self.create_apikey(req, conn)

		case req.Method == "insert_apikey":
			self.insert_apikey(req, conn)

		case req.Method == "export_apikeys":
			self.export_apikeys(req, conn)

		case req.Method == "export_apikey":
			self.export_apikey(req, conn)

		// FEATURE
		case req.Method == "insert_feature":
			self.insert_feature(req, conn)

		case req.Method == "edit_feature":
			self.edit_feature(req, conn)

		// DATASOURCES
		case req.Method == "assign_datasource":
			self.assign_datasource(req, conn)

		case req.Method == "create_datasource":
			self.create_datasource(req, conn)

		case req.Method == "insert_layer":
			self.create_datasource(req, conn)

		case req.Method == "delete_layer":
			self.delete_datasource(req, conn)

		case req.Method == "delete_datasource":
			self.delete_datasource(req, conn)

		// TODO: ERROR HANDLING
		case req.Method == "export_datasources":
			self.export_datasources(req, conn)

		case req.Method == "export_datasource":
			self.export_datasource(req, conn)

		case req.Method == "export_layer":
			self.export_datasource(req, conn)

		case req.Method == "import_file":
			self.import_file(req, conn)

			/*

			   "export_datasource_snapshots
			   "export_datasource_by_snapshot"
			   "export_datasource_by_range"

			*/

		default:
			err := errors.New("Method not found")
			self.handleError(err, conn)
		}
	}
}

func (self TcpServer) handleError(err error, conn net.Conn) {
	conn.Write([]byte("{\"status\": \"error\", \"error\": \"" + err.Error() + "\"}\n"))
}

func (self TcpServer) handleSuccess(data string, conn net.Conn) {
	conn.Write([]byte("{\"status\": \"ok\", \"data\": " + data + "}\n"))
}

func (self TcpServer) missingParams(conn net.Conn) {
	err := errors.New("Missing required parameters")
	self.handleError(err, conn)
}

func (self TcpServer) mashalJsonFromStructResponse(data interface{}, conn net.Conn) {
	js, err := json.Marshal(data)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(string(js), conn)
}

// APIKEYS
func (self TcpServer) create_apikey(req TcpMessage, conn net.Conn) {
	// {"method":"create_apikey"}
	apikey := utils.NewAPIKey(12)
	customer := Customer{Apikey: apikey}
	err := DB.InsertCustomer(customer)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(`{"apikey": "`+apikey+`"}`, conn)
}

func (self TcpServer) insert_apikey(req TcpMessage, conn net.Conn) {
	// {"method": "insert_apikey"}
	if "" == req.Data.Apikey {
		self.missingParams(conn)
		return
	}
	customer := Customer{Apikey: req.Data.Apikey, Datasources: req.Data.Datasources}
	err := DB.InsertCustomer(customer)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(`{"apikey": "`+req.Data.Apikey+`"}`, conn)
}

func (self TcpServer) export_apikeys(req TcpMessage, conn net.Conn) {
	// {"method":"export_apikeys"}
	apikeys, err := DB.GetCustomers()
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.mashalJsonFromStructResponse(apikeys, conn)
}

func (self TcpServer) export_apikey(req TcpMessage, conn net.Conn) {
	// {"method":"export_apikey","apikey":"12dB6BlenIeB"}
	apikey, err := DB.GetCustomer(req.Apikey)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.mashalJsonFromStructResponse(apikey, conn)
}

// DATASOURCES
func (self TcpServer) assign_datasource(req TcpMessage, conn net.Conn) {
	// {"method":"assign_datasource"}
	datasource_id := req.Datasource
	apikey := req.Apikey

	if "" == datasource_id || "" == apikey {
		self.missingParams(conn)
		return
	}

	customer, err := DB.GetCustomer(apikey)
	if err != nil {
		self.handleError(err, conn)
		return
	}

	_, err = GeoDB.GetLayer(datasource_id)
	if err != nil {
		self.handleError(err, conn)
		return
	}

	if !utils.StringInSlice(datasource_id, customer.Datasources) {
		customer.Datasources = append(customer.Datasources, datasource_id)
		DB.InsertCustomer(customer)
	}

	self.handleSuccess(`{}`, conn)
}

func (self TcpServer) create_datasource(req TcpMessage, conn net.Conn) {
	// {"method":"create_datasource"}
	datasource_id := req.Datasource
	var err error

	fmt.Println(req.Datasource, req.Layer)

	if "" != req.Datasource {
		err = GeoDB.InsertLayer(req.Datasource, req.Layer)
	} else {
		datasource_id, err = GeoDB.NewLayer()
	}

	if err != nil {
		self.handleError(err, conn)
		return
	}

	self.handleSuccess(`{"datasource_id":"`+datasource_id+`"}`, conn)
}

func (self TcpServer) export_datasources(req TcpMessage, conn net.Conn) {
	// {"method":"export_datasources"}
	layers, err := GeoDB.GetLayers()
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.mashalJsonFromStructResponse(layers, conn)
}

func (self TcpServer) export_datasource(req TcpMessage, conn net.Conn) {
	// {"method":"export_datasource","datasource":"20f3332781ea4d7b8d509d12517ac5fa"}
	layer, err := GeoDB.GetLayer(req.Datasource)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.mashalJsonFromStructResponse(layer, conn)
}

func (self TcpServer) delete_datasource(req TcpMessage, conn net.Conn) {
	// {"method":"delete_layer", "datasource":"f79aac397a484998b94b56d345287096"}
	if "" == req.Datasource {
		self.missingParams(conn)
		return
	}
	err := GeoDB.DeleteLayer(req.Datasource)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(`{"datasource_id":"`+req.Data.Datasource+`", "message":"layer deleted"}`, conn)
}

// FEATURES
func (self TcpServer) insert_feature(req TcpMessage, conn net.Conn) {
	// {"method":"insert_feature"}
	if "" == req.Datasource {
		self.missingParams(conn)
		return
	}
	err := GeoDB.InsertFeature(req.Datasource, req.Feature)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(`{"datasource_id":"`+req.Data.Datasource+`", "message":"feature added"}`, conn)
}

func (self TcpServer) edit_feature(req TcpMessage, conn net.Conn) {
	// {"method":"edit_feature"}
	if "" == req.Datasource || "" == req.GeoId {
		self.missingParams(conn)
		return
	}
	err := GeoDB.EditFeature(req.Datasource, req.GeoId, req.Feature)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(`{"datasource_id":"`+req.Datasource+`", "message":"edited added"}`, conn)
}

// FILE
func (self TcpServer) import_file(req TcpMessage, conn net.Conn) {
	// {"method":"import_file","file":"springfield_projects_edit.geojson"}
	result, err := importDatasource(req.File)
	if err != nil {
		self.handleError(err, conn)
		return
	}
	self.handleSuccess(`{"datasource": "`+result+`"}`, conn)
}

func importDatasource(importFile string) (string, error) {
	//fmt.Println("Importing", importFile)
	// get geojson file
	var geojsonFile string
	// check if file exists
	if _, err := os.Stat(importFile); os.IsNotExist(err) {
		return "", err
	}
	// ERROR
	// CRASHES IF NO "." character FOUND
	ext := strings.Split(importFile, ".")[1]
	// convert shapefile
	if ext == "shp" {
		// Convert .shp to .geojson
		geojsonFile := strings.Replace(importFile, ".shp", ".geojson", -1)
		fmt.Println("ogr2ogr", "-f", "GeoJSON", "-t_srs", "crs:84", geojsonFile, importFile)
		out, err := exec.Command("ogr2ogr", "-f", "GeoJSON", "-t_srs", "crs:84", geojsonFile, importFile).Output()
		if err != nil {
			return fmt.Sprintf("%v", out), err
		}
	} else if ext == "geojson" {
		geojsonFile = importFile
	} else {
		return fmt.Sprintf("Unsupported file type: %v", ext), errors.New(fmt.Sprintf("Unsupported file type: %v", ext))
	}
	// Read .geojson file
	file, err := ioutil.ReadFile(geojsonFile)
	if err != nil {
		return "", err
	}
	// Unmarshal to geojson struct
	geojs, err := geojson.UnmarshalFeatureCollection(file)
	if err != nil {
		return "", err
	}
	// Create datasource
	ds, _ := utils.NewUUID()
	GeoDB.InsertLayer(ds, geojs)
	// Cleanup artifacts
	if geojsonFile != importFile {
		os.Remove(geojsonFile)
	}
	return ds, nil
}
