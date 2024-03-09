package service

import (
	"log"
	"net/http"

	"github.com/warpdl/warpdl/internal/server"
	"github.com/warpdl/warpdl/pkg/warplib"
)

type Api struct {
	log     *log.Logger
	manager *warplib.Manager
	client  *http.Client
}

func NewApi(l *log.Logger) (*Api, error) {
	m, err := warplib.InitManager()
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	return &Api{
		log:     l,
		manager: m,
		client:  client,
	}, nil
}

func (s *Api) RegisterHandlers(server *server.Server) {
	server.RegisterHandler(UPDATE_DOWNLOAD, s.downloadHandler)
	server.RegisterHandler(UPDATE_RESUME, s.resumeHandler)
	server.RegisterHandler(UPDATE_ATTACH, s.attachHandler)
	server.RegisterHandler(UPDATE_FLUSH, s.flushHandler)
	server.RegisterHandler(UPDATE_STOP, s.stopHandler)
	server.RegisterHandler(UPDATE_LIST, s.listHandler)
}

func (s *Api) Close() error {
	return s.manager.Close()
}