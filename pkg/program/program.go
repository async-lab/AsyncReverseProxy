package program

import (
	"context"

	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

type MetaProgram struct {
	Ctx       context.Context
	CtxCancel context.CancelFunc
	EventBus  *event.EventBus
}

func NewMetaProgram(parentCtx context.Context) *MetaProgram {
	ctx, cancel := context.WithCancel(parentCtx)
	return &MetaProgram{
		Ctx:       ctx,
		CtxCancel: cancel,
		EventBus:  event.NewEventBus(),
	}
}

type IProgram interface {
	Run()
	ToMeta() *MetaProgram
}

func (meta *MetaProgram) ToMeta() *MetaProgram { return meta }

var Program IProgram = nil
