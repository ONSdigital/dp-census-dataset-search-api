SHELL=bash

BUILD=build
BIN_DIR?=.

MONGODB_BIND_ADDR=${mongodb_bind_addr}
FILENAME=${filename}
DATASET_INDEX=${dataset_index}
ELASTICSEARCH_URL=${elasticsearch_url}

RETRIEVE_CMD_DATASETS=retrieve-cmd-datasets
UPLOAD_DATASETS=upload-datasets

build:
	go generate ../...
	@mkdir -p ../$(BUILD)/$(BIN_DIR)
	
cmd-datasets-csv: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(RETRIEVE_CMD_DATASETS) $(RETRIEVE_CMD_DATASETS)/main.go
	HUMAN_LOG=1 go run -race $(RETRIEVE_CMD_DATASETS)/main.go -mongodb-bind-addr=$(MONGODB_BIND_ADDR) -filename=$(FILENAME)

upload-datasets: build
	go build -o ../$(BUILD)/$(BIN_DIR)/$(UPLOAD_DATASETS) $(UPLOAD_DATASETS)/main.go
	HUMAN_LOG=1 go run -race $(UPLOAD_DATASETS)/main.go -filename=$(FILENAME) -dataset-index=$(DATASET_INDEX) -elasticsearch-url=$(ELASTICSEARCH_URL)

test:
	go test -cover -race ./...

.PHONY: build cmd-datasets-csv upload-datasets