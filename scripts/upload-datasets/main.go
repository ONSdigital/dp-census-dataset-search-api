package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"

	es "github.com/ONSdigital/dp-census-dataset-search-api/internal/elasticsearch"
	"github.com/ONSdigital/dp-census-dataset-search-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

const (
	defaultDatasetIndex        = "dataset-test"
	defaultElasticsearchAPIURL = "http://localhost:9200"
	defaultFilename            = "datasets.csv"
	mappingsFile               = "dataset-mappings.json"
)

var datasetIndex, elasticsearchAPIURL, filename string

func main() {
	ctx := context.Background()
	flag.StringVar(&datasetIndex, "dataset-index", defaultDatasetIndex, "the elasticsearch index that datasets will be uploaded to")
	flag.StringVar(&elasticsearchAPIURL, "elasticsearch-url", defaultElasticsearchAPIURL, "the elasticsearch url")
	flag.StringVar(&filename, "filename", defaultFilename, "the csv filename that contains data to upload to elasticsearch")
	flag.Parse()

	if datasetIndex == "" {
		datasetIndex = defaultDatasetIndex
	}

	if elasticsearchAPIURL == "" {
		elasticsearchAPIURL = defaultElasticsearchAPIURL
	}

	if filename == "" {
		filename = defaultFilename
	}

	log.Event(ctx, "script variables", log.INFO, log.Data{"dataset_index": datasetIndex, "elasticsearch_api_url": elasticsearchAPIURL, "filename": filename})

	cli := dphttp.NewClient()
	esAPI := es.NewElasticSearchAPI(cli, elasticsearchAPIURL)

	// delete existing elasticsearch index if already exists
	status, err := esAPI.DeleteSearchIndex(ctx, datasetIndex)
	if err != nil {
		if status != http.StatusNotFound {
			log.Event(ctx, "failed to delete index", log.ERROR, log.Error(err), log.Data{"status": status})
			os.Exit(1)
		}

		log.Event(ctx, "failed to delete index as index cannot be found, continuing", log.WARN, log.Error(err), log.Data{"status": status})
	}

	// create elasticsearch index with settings/mapping
	status, err = esAPI.CreateSearchIndex(ctx, datasetIndex, mappingsFile)
	if err != nil {
		log.Event(ctx, "failed to create index", log.ERROR, log.Error(err), log.Data{"status": status})
		os.Exit(1)
	}

	// upload geo locations from data/datasets-test.csv and manipulate data into models.GeoDoc
	if err = uploadDocs(ctx, esAPI, datasetIndex, filename); err != nil {
		log.Event(ctx, "failed to retrieve dataset docs", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "successfully loaded in dataset docs", log.INFO)
}

func uploadDocs(ctx context.Context, esAPI *es.API, indexName, filename string) error {
	csvfile, err := os.Open(filename)
	if err != nil {
		log.Event(ctx, "failed to open the csv file", log.ERROR, log.Error(err))
		return err
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	headerRow, err := r.Read()
	if err != nil {
		log.Event(ctx, "failed to read header row", log.ERROR, log.Error(err))
		return err
	}

	headerIndex, err := check(headerRow)
	if err != nil {
		log.Event(ctx, "header row missing expected headers", log.ERROR, log.Error(err))
		return err
	}

	count := 0

	// Iterate through the records
	for {
		count++
		// Read each record from csv
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Event(ctx, "failed to read row", log.ERROR, log.Error(err))
		}

		datasetDoc := &models.Dataset{
			Alias:       row[headerIndex["alias"]],
			Description: row[headerIndex["description"]],
			Link:        row[headerIndex["ons-link"]],
			Title:       row[headerIndex["title"]],
		}

		// Add document to elasticsearch index
		if _, err = esAPI.AddDataset(ctx, indexName, datasetDoc); err != nil {
			log.Event(ctx, "failed to upload document to index", log.ERROR, log.Error(err), log.Data{"count": count})
			return err
		}
	}

	return nil
}

var validHeaders = map[string]bool{
	"alias":       true,
	"description": true,
	"ons-link":    true,
	"title":       true,
}

func check(headerRow []string) (map[string]int, error) {
	hasHeaders := map[string]bool{
		"alias":       false,
		"description": false,
		"ons-link":    false,
		"title":       false,
	}

	if len(headerRow) < 1 {
		return nil, errors.New("empty header row")
	}

	var indexHeader = make(map[string]int)
	for i, header := range headerRow {
		if !validHeaders[header] {
			return nil, errors.New("invalid header: " + header)
		}

		hasHeaders[header] = true
		indexHeader[header] = i
	}

	var hasHeadersMissing bool
	var missingHeaders string
	for key, value := range hasHeaders {
		if !value {
			hasHeadersMissing = true
			missingHeaders = missingHeaders + key + " "
		}
	}

	if hasHeadersMissing {
		return nil, errors.New("missing header in row: " + missingHeaders)
	}

	return indexHeader, nil
}
