//go:generate go get github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -pkg elasticsearch ./postcode-mappings.json ./geography-mappings.json

package elasticsearch
