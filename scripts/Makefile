SHELL=bash

BUILD=build
BIN_DIR?=.

MONGODB_BIND_ADDR=${mongodb_bind_addr}
FILENAME=${filename}
DATASET_INDEX=${dataset_index}
ELASTICSEARCH_URL=${elasticsearch_url}
DIMENSIONS_JSON=${dimensions_filename}
TAXONOMY_JSON=${taxonomy_filename}

RETRIEVE_CMD_DATASETS=retrieve-cmd-datasets
RETRIEVE_DATASET_TAXONOMY=retrieve-dataset-taxonomy
UPLOAD_DATASETS=upload-datasets

build:
	go generate ../...
	@mkdir -p ../$(BUILD)/$(BIN_DIR)
	
cmd-datasets-csv: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(RETRIEVE_CMD_DATASETS) $(RETRIEVE_CMD_DATASETS)/main.go
	HUMAN_LOG=1 go run -race $(RETRIEVE_CMD_DATASETS)/main.go -mongodb-bind-addr=$(MONGODB_BIND_ADDR) -filename=$(FILENAME)

taxonomy-json: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(RETRIEVE_DATASET_TAXONOMY) $(RETRIEVE_DATASET_TAXONOMY)/main.go
	HUMAN_LOG=1 go run -race $(RETRIEVE_DATASET_TAXONOMY)/main.go -filename=$(TAXONOMY_JSON)

upload-datasets: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(UPLOAD_DATASETS) $(UPLOAD_DATASETS)/main.go
	HUMAN_LOG=1 go run -race $(UPLOAD_DATASETS)/main.go -filename=$(FILENAME) -dimensions-filename=$(DIMENSIONS_JSON) -taxonomy-filename=$(TAXONOMY_JSON) -dataset-index=$(DATASET_INDEX) -elasticsearch-url=$(ELASTICSEARCH_URL)

test:
	go test -cover -race ./...

.PHONY: build cmd-datasets-csv upload-datasets