package api

import (
	db "dnsbin/db/sqlc"
	"dnsbin/notify"
	"dnsbin/util"

	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config   util.Config
	store    db.Store
	router   *gin.Engine
	telegram *notify.Telegram
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store, telegram *notify.Telegram) (*Server, error) {
	server := &Server{
		config:   config,
		store:    store,
		telegram: telegram,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.Any("/httplog/*any", server.httpRequestLog)
	router.GET("/users/login", server.loginUser)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
