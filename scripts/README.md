# Scripts

A list of helpful scripts to load data for use in the Search API.

## A list of scripts

- [retrieve cmd datasets](#retrieve-cmd-datasets)
- [load parent docs](#load-datasets)
- [retrieve dataset taxonomy](#retrieve-dataset-taxonomy)

### Retrieve CMD Datasets

This script retrieves a list of datasets stored in mongodb instance and will check that the url to dataset resource on the ons website exists before storing the data in a csv file.

You can run either of the following commands:

- Use Makefile
    - Set `mongodb_bind_addr` and/or `filename` environment variable with:
    ```
    export mongodb_bind_addr=<mongodb bind address>
    export filename=<file name and loaction>
    ```
    - Run `make cmd-datasets-csv`
- Use go run command with or without flags `-mongodb-bind-addr` and/or `-filename` being set
    - `go run retrieve-cmd-datasets/main.go -mongodb-bind-addr=<mongodb bind address> -filename=<file name and loaction>`
    
if you do not set the flags or environment variables for mongodb bind address and filename, the script will use a default value set to `localhost:27017` and `cmd-datasets.csv` respectively.

### Load Datasets

This script reads a csv file defined by flag/environment variable or default value and stores the dataset data into elasticsearch. The csv must contain particular headers (but not in any necessary order).

One can use the Retrieve cmd datasets script to generate a new csv file or use the pre-generated one stored as `datasets.csv`.

- Use Makefile
    - Set `dataset_index`, `filename` and/or `elasticsearch_url` environment variable with:
    ```
    export dataset_index=<elasticsearch index>
    export filename=<file name and loaction>
    export elasticsearch_url=<elasticsearch bind address>
    ```
    - Run `make upload-datasets`
- Use go run command with or without flags `-dataset-index`, `-filename` and/or `elasticsearch_url` being set
    - `go run upload-datasets/main.go -dataset-index=<elasticsearch index> -filename=<file name and loaction> -elasticsearch_url=<elasticsearch bin address>`

### Retrieve Dataset Taxonomy

This script scrapes the ons website to pull out taxonomy hierarchy by iterating through pages.

You can run either of the following commands:

- Use Makefile
    - Set `taxonomy_filename` environment variable with, should end with `.json`:
    ```
    export taxonomy_filename=<filename and location>
    ```
    - Run `make taxonomy-json`
- Use go run command with or without flags `-filename` being set
    - `go run retrieve-dataset-taxonomy/main.go -filename=<file name and loaction>`
    
if you do not set the flag or environment variable for filename, then the script will use a default value set to `../taxonomy/taxonomy.json`.