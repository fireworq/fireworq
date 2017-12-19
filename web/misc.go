package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fireworq/fireworq/config"
)

func (app *Application) serveVersion(w http.ResponseWriter, req *http.Request) error {
	fmt.Fprintf(w, "%s\n", app.Version)
	return nil
}

func (app *Application) serveSettings(w http.ResponseWriter, req *http.Request) error {
	keys := config.Keys()
	settings := make(map[string]string)

	for _, k := range keys {
		settings[k] = config.Get(k)
	}

	j, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	writeJSON(w, j)
	return nil
}
