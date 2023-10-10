package api

import (
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for banking service
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer create a new http server and setup routing
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	//add router to server
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	server.router = router
	return server
}

// Start runs the http server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
func errResponse(err error) gin.H {
	return gin.H{"err": err.Error()}
}
