package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	es "github.com/ONSdigital/dp-census-dataset-search-api/internal/elasticsearch"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

const (
	defaultDatasetIndex        = "dataset-test"
	defaultElasticsearchAPIURL = "http://localhost:9200"
	defaultFilename            = "datasets.csv"
	defaultTaxonomyFile        = "../taxonomy/taxonomy.json"
	mappingsFile               = "dataset-mappings.json"
)

var (
	datasetIndex, elasticsearchAPIURL, filename, taxonomyFilename string
	taxonomy                                                      Taxonomy
	topicLevels                                                   = make(map[string]TopicLevels)
)

// Dataset represents the data stored against a resource in elasticsearch index
type Dataset struct {
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	Topic1      string `json:"topic1,omitempty"`
	Topic2      string `json:"topic2,omitempty"`
	Topic3      string `json:"topic3,omitempty"`
}

// Taxonomy ...
type Taxonomy struct {
	Topics []Topic `json:"topics"`
}

type Topic struct {
	Title          string  `json:"title"`
	FormattedTitle string  `json:"formatted_title"`
	ChildTopics    []Topic `json:"child_topics,omitempty"`
}

type TopicLevels struct {
	TopicLevel1 string
	TopicLevel2 string
	TopicLevel3 string
}

func main() {
	ctx := context.Background()
	flag.StringVar(&datasetIndex, "dataset-index", defaultDatasetIndex, "the elasticsearch index that datasets will be uploaded to")
	flag.StringVar(&elasticsearchAPIURL, "elasticsearch-url", defaultElasticsearchAPIURL, "the elasticsearch url")
	flag.StringVar(&filename, "filename", defaultFilename, "the csv filename that contains data to upload to elasticsearch")
	flag.StringVar(&taxonomyFilename, "taxonomy-filename", defaultTaxonomyFile, "the file locataion and name that contains the taxonomy hierarchy")
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

	if taxonomyFilename == "" {
		taxonomyFilename = defaultTaxonomyFile
	}

	log.Event(ctx, "script variables", log.INFO, log.Data{"dataset_index": datasetIndex, "elasticsearch_api_url": elasticsearchAPIURL, "filename": filename, "taxonomy-file": taxonomyFilename})

	cli := dphttp.NewClient()
	esAPI := es.NewElasticSearchAPI(cli, elasticsearchAPIURL)

	// Read in Taxonomy into memory
	taxonomyFile, err := ioutil.ReadFile(taxonomyFilename)
	if err != nil {
		log.Event(ctx, "failed to read taxonomy file", log.ERROR, log.Error(err), log.Data{"taxonomy_filename": taxonomyFilename})
		os.Exit(1)
	}

	if err = json.Unmarshal([]byte(taxonomyFile), &taxonomy); err != nil {
		log.Event(ctx, "unable to unmarshal taxonomy into struct", log.ERROR, log.Error(err), log.Data{"taxonomy_filename": taxonomyFilename})
		os.Exit(1)
	}

	// Invert taxonomy so each topic has a list of parent topics and store in map
	for _, topic := range taxonomy.Topics {
		topicLevels[topic.FormattedTitle] = TopicLevels{
			TopicLevel1: topic.FormattedTitle,
		}

		for _, topic2 := range topic.ChildTopics {
			topicLevels[topic2.FormattedTitle] = TopicLevels{
				TopicLevel1: topic.FormattedTitle,
				TopicLevel2: topic2.FormattedTitle,
			}

			for _, topic3 := range topic2.ChildTopics {
				topicLevels[topic3.FormattedTitle] = TopicLevels{
					TopicLevel1: topic.FormattedTitle,
					TopicLevel2: topic2.FormattedTitle,
					TopicLevel3: topic3.FormattedTitle,
				}
			}
		}
	}

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

		datasetDoc := &Dataset{
			Alias:       row[headerIndex["alias"]],
			Description: row[headerIndex["description"]],
			Link:        row[headerIndex["ons-link"]],
			Title:       row[headerIndex["title"]],
		}

		topic := row[headerIndex["topic"]]
		if topic != "" {
			log.Event(ctx, "topic?", log.Data{"topic": topic})
			// find topic hierarchy - using taxonomy map
			taxonomy := topicLevels[topic]

			datasetDoc.Topic1 = taxonomy.TopicLevel1
			datasetDoc.Topic2 = taxonomy.TopicLevel2
			datasetDoc.Topic3 = taxonomy.TopicLevel3
			log.Event(ctx, "dataset?", log.Data{"datasets": datasetDoc})
		}

		bytes, err := json.Marshal(datasetDoc)
		if err != nil {
			log.Event(ctx, "failed to marshal dataset document to bytes", log.ERROR, log.Error(err), log.Data{"count": count})
			return err
		}

		// Add document to elasticsearch index
		if _, err = esAPI.AddDocument(ctx, indexName, bytes); err != nil {
			log.Event(ctx, "failed to upload dataset document to index", log.ERROR, log.Error(err), log.Data{"count": count})
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
	"topic":       true,
}

func check(headerRow []string) (map[string]int, error) {
	hasHeaders := map[string]bool{
		"alias":       false,
		"description": false,
		"ons-link":    false,
		"title":       false,
		"topic":       false,
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
