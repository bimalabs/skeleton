// Don't change anything in this file, this file used by Skeleton Module Manager
package dics

import (
    //@modules:import
	"app/dics"

	core "github.com/bimalabs/framework/v4/dics"
	"github.com/sarulabs/dingo/v4"
)

type (
	Engine struct {
		dingo.BaseProvider
	}
)

func (p *Engine) Load() error {
	if err := p.AddDefSlice(core.Application); err != nil {
		return err
	}

	if err := p.AddDefSlice(dics.Container); err != nil {
		return err
	}
    //@modules:register

	return nil
}
