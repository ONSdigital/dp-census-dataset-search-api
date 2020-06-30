package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-census-dataset-search-api/apierrors"
	"github.com/ONSdigital/dp-census-dataset-search-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

const (
	defaultLimit  = 50
	defaultOffset = 0

	internalError         = "internal server error"
	exceedsDefaultMaximum = "the maximum offset has been reached, the offset cannot be more than"
)

func (api *SearchAPI) getDatasets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	var err error

	id := vars["id"]

	requestedLimit := r.FormValue("limit")
	requestedOffset := r.FormValue("offset")

	logData := log.Data{
		"id":               id,
		"requested_limit":  requestedLimit,
		"requested_offset": requestedOffset,
	}

	log.Event(ctx, "getDatasets endpoint: incoming request", log.INFO, logData)

	limit := defaultLimit
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			log.Event(ctx, "getDatasets endpoint: request limit parameter error", log.ERROR, log.Error(err), logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	offset := defaultOffset
	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			log.Event(ctx, "getDatasets endpoint: request offset parameter error", log.ERROR, log.Error(err), logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	page := &models.PageVariables{
		DefaultMaxResults: api.defaultMaxResults,
		Limit:             limit,
		Offset:            offset,
	}

	if err = page.Validate(); err != nil {
		log.Event(ctx, "getDatasets endpoint: validate pagination", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	logData["limit"] = page.Limit
	logData["offset"] = page.Offset

	log.Event(ctx, "getDatasets endpoint: just before querying search index", log.INFO, logData)

	var query interface{}
	// TODO build dataset search query
	// TODO do search by calling QueryGeoLocation(ctx, api.datasetIndex, query interface{}, limit, offset int)
	searchResults, status, err := api.elasticsearch.QueryGeoLocation(ctx, api.datasetIndex, query, limit, offset)
	if err != nil {
		logData["es_status"] = status
		log.Event(ctx, "getDatasets endpoint: failed to get search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
	}

	b, err := json.Marshal(searchResults)
	if err != nil {
		log.Event(ctx, "getDatasets endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
	}

	setJSONContentType(w)
	setAccessControl(w)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getDatasets endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getDatasets endpoint: successfully searched index", log.INFO, logData)
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
func setAccessControl(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func setErrorCode(w http.ResponseWriter, err error) {

	switch {
	case errs.NotFoundMap[err]:
		http.Error(w, err.Error(), http.StatusNotFound)
	case errs.BadRequestMap[err]:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case strings.Contains(err.Error(), exceedsDefaultMaximum):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
	}
}
