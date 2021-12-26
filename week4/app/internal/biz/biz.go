package biz

import "errors"
import "github.com/google/wire"

var (
	ErrorBizInternal = errors.New("biz internal")
)

var ProviderSet = wire.NewSet(NewUserBaseUseCase, NewLevelInfoUseCase)
