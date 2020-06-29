# dp-census-dataset-search-api

## Requirements

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

#### Notes

See [command list](COMMANDS.md) for a list of helpful commands to run alongside setting up data, useful to check what search indexes exist and their individual mappings and number of documents etc..

One can run the unit tests with `make test`
