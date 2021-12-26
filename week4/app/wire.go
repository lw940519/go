//+build wireinject

package app

import (
	"ceshi/week4/app/internal/biz"
	"ceshi/week4/app/internal/data"
	"ceshi/week4/app/internal/server"
	"ceshi/week4/app/internal/service"

	"github.com/google/wire"
)

func InitService() (*server.GRPCServer, error) {
	panic(wire.Build(data.ProviderSet, biz.ProviderSet, service.ProviderSet, server.ProviderSet))
}
