package web

import (
	"strconv"

	"github.com/fireworq/fireworq/config"
)

var disableKeepAlives = true

// Init initializes global parameters of the Web server by
// configuration values.
func Init() {
	b, err := strconv.ParseBool(config.Get("keep_alive"))
	if err != nil {
		b, _ = strconv.ParseBool(config.GetDefault("keep_alive"))
	}
	disableKeepAlives = !b
}
