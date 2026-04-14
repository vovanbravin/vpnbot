package fsm

import (
	"tgbot/internal/database"

	"github.com/vitaliy-ukiru/fsm-telebot/v2"
	"github.com/vitaliy-ukiru/fsm-telebot/v2/pkg/storage/memory"
	"github.com/vitaliy-ukiru/telebot-filter/v2/dispatcher"
)

type FSM struct {
	Manager *fsm.Manager
	mongoDb *database.MongoDB
}

func NewFSM(mongoDb *database.MongoDB) *FSM {
	return &FSM{
		Manager: fsm.New(memory.NewStorage()),
		mongoDb: mongoDb,
	}
}

func (f *FSM) Setup(dispatcher *dispatcher.Dispatcher) {
	f.SetupSupport(dispatcher)
}
