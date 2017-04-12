package geo_skeleton_server

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

import (
	"github.com/sjsafranek/GeoSkeletonDB"
	"github.com/sjsafranek/SkeletonDB"
)

// https://gist.github.com/DavidVaini/10308388
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func RoundToPrecision(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}

// DB application Database
var (
	DB              Database
	GeoDB           geoskeleton.Database
	COMMIT_LOG_FILE string = "api_commit.log"
)

// Database strust for application.
type Database struct {
	File             string
	commit_log_queue chan string
	DB               skeleton.Database
}

// Init creates bolt database if existing one not found.
// Creates layers and apikey tables. Starts database caching for layers
// @returns Error
func (self *Database) Init() error {

	// start commit log
	go self.StartCommitLog()

	// create database if not exists
	self.DB = skeleton.Database{File: "api.db"}
	self.DB.Init()

	// connect to db
	conn := self.DB.Connect()
	defer conn.Close()
	// datasources
	err := self.DB.CreateTable(conn, "layers")
	if err != nil {
		panic(err)
		return err
	}

	GeoDB = geoskeleton.NewGeoSkeletonDB("geo.db")
	go GeoDB.StartCommitLog()

	// Add table for datasource owner
	// permissions
	err = self.DB.CreateTable(conn, "apikeys")
	if err != nil {
		panic(err)
		return err
	}
	// close and return err
	return err
}

// Starts Database commit log
func (self *Database) StartCommitLog() {
	self.commit_log_queue = make(chan string, 10000)
	// open file to write database commit log
	COMMIT_LOG, err := os.OpenFile(COMMIT_LOG_FILE, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}
	defer COMMIT_LOG.Close()
	// read from chan and write to file
	for {
		if len(self.commit_log_queue) > 0 {
			line := <-self.commit_log_queue
			if _, err := COMMIT_LOG.WriteString(line + "\n"); err != nil {
				panic(err)
			}
		} else {
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

// CommitQueueLength returns length of database commit_log_queue
// @returns int
func (self *Database) CommitQueueLength() int {
	return len(self.commit_log_queue)
}

// InsertCustomer inserts customer into apikeys table
// @param customer {Customer}
// @returns Error
func (self *Database) InsertCustomer(customer Customer) error {

	fmt.Printf("\n%v\n\n", customer)

	value, err := json.Marshal(customer)
	if err != nil {
		return err
	}
	self.commit_log_queue <- `{"method": "insert_apikey", "data":` + string(value) + `}`
	// Insert customer into database
	err = self.DB.Insert("apikeys", customer.Apikey, value)
	if err != nil {
		panic(err)
	}
	return err
}

// GetCustomer returns customer from database
// @param apikey {string}
// @returns Customer
// @returns Error
func (self *Database) GetCustomer(apikey string) (Customer, error) {
	// If customer not found get from database
	val, err := self.DB.Select("apikeys", apikey)
	if err != nil {
		panic(err)
	}
	// datasource not found
	if "" == string(val) {
		return Customer{}, fmt.Errorf("Apikey not found")
	}
	// Read to struct
	customer := Customer{}
	err = json.Unmarshal(val, &customer)
	if err != nil {
		return Customer{}, err
	}
	// Close database connection
	return customer, nil
}

func (self *Database) GetCustomers() ([]string, error) {
	// If customer not found get from database
	val, err := self.DB.SelectAll("apikeys")
	if err != nil {
		panic(err)
	}
	return val, nil
}
