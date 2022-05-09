package web

import (
	"encoding/json"
	"net/http"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"

	"github.com/gorilla/mux"
)

func (app *Application) serveRoutingList(w http.ResponseWriter, req *http.Request) error {
	routings, err := app.RoutingRepository.FindAll()
	if err != nil {
		return err
	}

	json, err := json.Marshal(routings)
	if err != nil {
		return err
	}
	writeJSON(w, json)

	return nil
}

func (app *Application) serveRouting(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	jobCategory := vars["category"]
	var definition model.Routing

	if req.Method == "PUT" {
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&definition); err != nil {
			return errBadRequest.WithDetail(err.Error())
		}
		definition.JobCategory = jobCategory

		if _, err := app.RoutingRepository.Add(jobCategory, definition.QueueName); err != nil {
			if _, ok := err.(*repository.QueueNotFoundError); ok {
				return errNotFound.WithDetail(err.Error())
			}
			return err
		}
	} else {
		qn := app.RoutingRepository.FindQueueNameByJobCategory(jobCategory)
		if qn == "" {
			return errNotFound
		}

		definition = model.Routing{
			JobCategory: jobCategory,
			QueueName:   qn,
		}

		if req.Method == "DELETE" {
			if err := app.RoutingRepository.DeleteByJobCategory(jobCategory); err != nil {
				return err
			}
		}
	}

	j, err := json.Marshal(&definition)
	if err != nil {
		return err
	}

	writeJSON(w, j)
	return nil
}
