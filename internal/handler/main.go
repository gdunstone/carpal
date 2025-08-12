package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
)

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type resourceHandler struct {
	Driver driver.Driver
}

func NewResourceHandler(driver driver.Driver) Handler {
	return resourceHandler{driver}
}

func (handler resourceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
	
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Printf("invalid query params")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		return
	}
	resourceParamList, present := query["resource"]
	if !present {
		log.Printf("received blank resource request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		return
	}
	resourceParam := resourceParamList[0]
	if resourceParam == "" {
		log.Printf("received blank resource request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		return
	}
	// todo: properly handle the rel= params
	// the spec says that we should only return links specified by rel, and that we should 
	// be able to handle multiple rel params
	// https://datatracker.ietf.org/doc/html/rfc7033#section-4.3

	log.Printf("received request for resource %v", resourceParam)
	resourceStruct, err := handler.Driver.GetResource(resourceParam)
	if err != nil {
		if errors.As(err, &driver.ResourceNotFound{}) {
			log.Printf("resource %v not found: %v", resourceParam, err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		} else {
			log.Printf("error retrieving resource %v: %v", resourceParam, err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("bad gateway"))
			return
		}
	}

	JRD, err := resource.MarshalResource(*resourceStruct)
	if err != nil {
		log.Printf("unable to marshal resource: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/jrd+json")
	w.Write(JRD)
}
