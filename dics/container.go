package dics

import (
	bima "github.com/bimalabs/framework/v4"
	"github.com/bimalabs/framework/v4/configs"
	"github.com/bimalabs/framework/v4/events"
	"github.com/bimalabs/framework/v4/handlers"
	"github.com/bimalabs/framework/v4/paginations"
	"github.com/bimalabs/framework/v4/paginations/adapters"
	"github.com/bimalabs/framework/v4/repositories"
	"github.com/sarulabs/dingo/v4"
)

var Container = []dingo.Def{
	{
		Name:  "bima:repository:gorm",
		Scope: bima.Application,
		Build: (*repositories.GormRepository)(nil),
	},
	{
		Name:  "bima:pagination:adapter:gorm",
		Scope: bima.Application,
		Build: func(
			env *configs.Env,
			dispatcher *events.Dispatcher,
		) (*adapters.GormAdapter, error) {
			return &adapters.GormAdapter{
				Debug:      env.Debug,
				Dispatcher: dispatcher,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:event:dispatcher"),
		},
	},
	{
		Name:  "bima:handler",
		Scope: bima.Application,
		Build: func(
			env *configs.Env,
			dispatcher *events.Dispatcher,
			repository repositories.Repository,
			adapter paginations.Adapter,
		) (handlers.Handler, error) {
			return handlers.New(env.Debug, dispatcher, repository, adapter), nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:event:dispatcher"),
			"2": dingo.Service("bima:repository:gorm"),
			"3": dingo.Service("bima:pagination:adapter:gorm"),
		},
	},
}
