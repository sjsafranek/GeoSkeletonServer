package geo_skeleton_server

import (
	"time"

	"./utils"
)

const (
	VERSION string = "1.11.5"
)

var (
	startTime           = time.Now()
	SuperuserKey string = utils.NewAPIKey(12)
)
