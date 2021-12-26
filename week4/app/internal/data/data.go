package data

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, NewUserRepo, NewLevelInfoRepo)

// Data .
type Data struct {
}

// NewData .
func NewData() (*Data, error) {
	return &Data{}, nil
}
