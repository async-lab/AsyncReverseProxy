package session

import (
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

type MetaSession struct {
	Name string
}

func NewMetaSession(name string) *MetaSession {
	return &MetaSession{Name: name}
}

type MetaSessionForConnection struct {
	MetaSession
	EndConns *structure.SyncMap[string, *comm.Conn]
}

func NewMetaSessionForConnection(name string) *MetaSessionForConnection {
	return &MetaSessionForConnection{
		MetaSession: *NewMetaSession(name),
		EndConns:    structure.NewSyncMap[string, *comm.Conn](),
	}
}
