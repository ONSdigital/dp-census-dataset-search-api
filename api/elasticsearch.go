package api

import (
	"context"

	"github.com/ONSdigital/dp-census-dataset-search-api/models"
)

// Elasticsearcher - An interface used to access elasticsearch
type Elasticsearcher interface {
	QueryGeoLocation(ctx context.Context, indexName string, query interface{}, limit, offset int) (*models.Datasets, int, error)
}
