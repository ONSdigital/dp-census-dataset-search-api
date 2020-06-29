# dp-census-dataset-search-api

TODO description

### Requirements

In order to run the service locally you will need the following:

- Go
- Git
- Java 8 or greater for elasticsearch
- ElasticSearch (version 6.7 or 6.8)

### Getting started

- Clone the repo go get github.com/ONSdigital/dp-census-dataset-search-api
- Run elasticsearch e.g. ./elasticsearch<version>/bin/elasticsearch
- Follow setting up data
- Run `make debug`

#### Setting up data

TODO

### Configuration

| Environment variable        | Default               | Description
| --------------------------- | --------------------- | -----------
| BIND_ADDR                   | :10200                | The host and port to bind to |
| DATASET_INDEX               | dataset-test          | The index in which the search datasets are stored against in elasticsearch |
| ELASTIC_SEARCH_URL          | http://localhost:9200 | The host name for elasticsearch |
| MAX_SEARCH_RESULTS_OFFSET   | 1000                  | The maximum offset for the number of results returned by search query |
| SIGN_ELASTICSEARCH_REQUESTS | false                 | Boolean flag to identify whether elasticsearch requests via elastic API need to be signed if elasticsearch cluster is running in aws |

### Notes

See [command list](COMMANDS.md) for a list of helpful commands to run alongside setting up data, useful to check what search indexes exist and their individual mappings and number of documents etc..

One can run the unit tests with `make test`
