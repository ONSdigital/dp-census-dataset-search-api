package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-census-dataset-search-api/apierrors"
	"github.com/ONSdigital/dp-census-dataset-search-api/models"
	"github.com/ONSdigital/log.go/log"
)

const (
	defaultLimit  = 50
	defaultOffset = 0

	internalError         = "internal server error"
	exceedsDefaultMaximum = "the maximum offset has been reached, the offset cannot be more than"
)

func (api *SearchAPI) getDatasets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var err error

	q := r.FormValue("q")
	requestedLimit := r.FormValue("limit")
	requestedOffset := r.FormValue("offset")

	logData := log.Data{
		"query_term":       q,
		"requested_limit":  requestedLimit,
		"requested_offset": requestedOffset,
	}

	log.Event(ctx, "getDatasets endpoint: incoming request", log.INFO, logData)

	// Remove leading and/or trailing whitespace
	term := strings.TrimSpace(q)

	if term == "" {
		log.Event(ctx, "getDatasets endpoint: query parameter \"q\" empty", log.ERROR, log.Error(errs.ErrEmptySearchTerm), logData)
		setErrorCode(w, errs.ErrEmptySearchTerm)
		return
	}

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

	// build dataset search query
	query := buildSearchQuery(term, limit, offset)

	response, status, err := api.elasticsearch.QueryGeoLocation(ctx, api.datasetIndex, query, limit, offset)
	if err != nil {
		logData["elasticsearch_status"] = status
		log.Event(ctx, "getDatasets endpoint: failed to get search results", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
	}

	searchResults := &models.SearchResults{
		Limit:      page.Limit,
		Offset:     page.Offset,
		TotalCount: response.Hits.Total,
	}

	for _, result := range response.Hits.HitList {

		doc := result.Source
		doc.Matches = result.Matches
		searchResults.Items = append(searchResults.Items, doc)
	}

	searchResults.Count = len(searchResults.Items)

	b, err := json.Marshal(searchResults)
	if err != nil {
		log.Event(ctx, "getDatasets endpoint: failed to marshal search resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getDatasets endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getDatasets endpoint: successfully searched index", log.INFO, logData)
}

func setAccessControl(w http.ResponseWriter, method string) {
	w.Header().Set("Access-Control-Allow-Methods", method+",OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Content-Type", "application/json")
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

// Body represents the request body to elasticsearch
type Body struct {
	From      int        `json:"from"`
	Size      int        `json:"size"`
	Highlight *Highlight `json:"highlight,omitempty"`
	Query     Query      `json:"query"`
	Sort      []Scores   `json:"sort"`
	TotalHits bool       `json:"track_total_hits"`
}

// Highlight represents parts of the fields that matched
type Highlight struct {
	PreTags  []string          `json:"pre_tags,omitempty"`
	PostTags []string          `json:"post_tags,omitempty"`
	Fields   map[string]Object `json:"fields,omitempty"`
	Order    string            `json:"score,omitempty"`
}

// Object represents an empty object (as expected by elasticsearch)
type Object struct{}

// Query represents the request query details
type Query struct {
	Bool Bool `json:"bool"`
}

// Bool represents the desirable goals for query
type Bool struct {
	Must   []Match `json:"must,omitempty"`
	Should []Match `json:"should,omitempty"`
}

// Match represents the fields that the term should or must match within query
type Match struct {
	Match map[string]string `json:"match,omitempty"`
}

// Scores represents a list of scoring, e.g. scoring on relevance, but can add in secondary
// score such as alphabetical order if relevance is the same for two search results
type Scores struct {
	Score Score `json:"_score"`
}

// Score contains the ordering of the score (ascending or descending)
type Score struct {
	Order string `json:"order"`
}

func buildSearchQuery(term string, limit, offset int) interface{} {
	var object Object
	highlight := make(map[string]Object)

	highlight["alias"] = object
	highlight["description"] = object
	highlight["title"] = object

	alias := make(map[string]string)
	description := make(map[string]string)
	title := make(map[string]string)
	alias["alias"] = term
	description["description"] = term
	title["title"] = term

	aliasMatch := Match{
		Match: alias,
	}

	descriptionMatch := Match{
		Match: description,
	}

	titleMatch := Match{
		Match: title,
	}

	scores := Scores{
		Score: Score{
			Order: "desc",
		},
	}

	listOfScores := []Scores{}
	listOfScores = append(listOfScores, scores)

	query := &Body{
		From: offset,
		Size: limit,
		Highlight: &Highlight{
			PreTags:  []string{"<b><em>"},
			PostTags: []string{"</em></b>"},
			Fields:   highlight,
		},
		Query: Query{
			Bool: Bool{
				Should: []Match{
					aliasMatch,
					descriptionMatch,
					titleMatch,
				},
			},
		},
		Sort:      listOfScores,
		TotalHits: true,
	}

	return query
}
