package geo_skeleton_server

import "./utils"
import "github.com/paulmach/go.geojson"

// Customer structure for database
type Customer struct {
	Apikey      string      `json:"apikey"`
	Datasources []string    `json:"datasources"`
	TileLayers  []TileLayer `json:"tilelayers"`
}

func (self *Customer) hasDatasource(datasource_id string) bool {
	return !utils.StringInSlice(datasource_id, self.Datasources)
}

func (self *Customer) addDatasource(datasource_id string) {
	self.Datasources = append(self.Datasources, datasource_id)
}

func (self *Customer) removeDatasource(datasource_id string) {
	i := utils.SliceIndex(datasource_id, self.Datasources)
	self.Datasources = append(self.Datasources[:i], self.Datasources[i+1:]...)

}

type TileLayer struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

// MapData for html templates
type PageViewData struct {
	Apikey  string
	Version string
}

type TcpData struct {
	Apikey      string                     `json:"apikey"`
	Datasources []string                   `json:"datasources"`
	Datasource  string                     `json:"datasource"`
	Layer       *geojson.FeatureCollection `json:"layer"`
	Feature     *geojson.Feature           `json:"feature"`
}

type TcpMessage struct {
	Apikey     string                     `json:"apikey"`
	Method     string                     `json:"method"`
	Datasource string                     `json:"datasource"`
	File       string                     `json:"file"`
	GeoId      string                     `json:"geo_id"`
	Layer      *geojson.FeatureCollection `json:"layer"`
	Feature    *geojson.Feature           `json:"feature"`
	Data       TcpData                    `json:"data"`
}

type HttpMessageResponse struct {
	Status     string      `json:"status"`
	Datasource string      `json:"datasource,omitempty"`
	Apikey     string      `json:"apikey,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}
