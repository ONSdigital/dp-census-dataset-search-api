package api

import (
	"context"

	"github.com/ONSdigital/dp-census-dataset-search-api/models"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

// SearchAPI manages searches across indices
type SearchAPI struct {
	defaultMaxResults int
	elasticsearch     Elasticsearcher
	router            *mux.Router
	datasetIndex      string
	taxonomy          models.Taxonomy
}

// CreateAndInitialiseSearchAPI manages all the routes configured to API
func CreateAndInitialiseSearchAPI(ctx context.Context, bindAddr string, esAPI Elasticsearcher, defaultMaxResults int, datasetIndex string, taxonomy models.Taxonomy, errorChan chan error) {

	router := mux.NewRouter()
	routes(ctx,
		router,
		esAPI,
		defaultMaxResults,
		datasetIndex,
		taxonomy,
	)

	httpServer = server.New(bindAddr, router)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "api http server returned error", log.ERROR, log.Error(err))
			errorChan <- err
		}
	}()
}

func routes(ctx context.Context,
	router *mux.Router,
	elasticsearch Elasticsearcher,
	defaultMaxResults int,
	datasetIndex string,
	taxonomy models.Taxonomy) *SearchAPI {

	api := SearchAPI{
		defaultMaxResults: defaultMaxResults,
		elasticsearch:     elasticsearch,
		router:            router,
		datasetIndex:      datasetIndex,
		taxonomy:          taxonomy,
	}

	api.router.HandleFunc("/datasets", api.getDatasets).Methods("GET", "OPTIONS")
	api.router.HandleFunc("/taxonomy", api.getTaxonomy).Methods("GET", "OPTIONS")
	api.router.HandleFunc("/taxonomy/{topic}", api.getTopic).Methods("GET", "OPTIONS")

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Event(ctx, "graceful shutdown of http server complete", log.INFO)
	return nil
}
