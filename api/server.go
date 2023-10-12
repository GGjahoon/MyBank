package api

import (
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for banking service
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer create a new http server and setup routing
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	//binding.Validator.Engine() to get the current validator engin that gin is using,return any.
	// is a pointer to the validator object
	//convert the output to *validator.Validate to register our own validator(断言过程)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		//call *validator.Validate.RegisterValidation to register our custom validate function
		//the first argument is the name of the validation tag
		//the second is validCurrency function
		v.RegisterValidation("currency", validCurrency)
	}

	//add router to server
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	router.POST("/transfers", server.createTransfer)
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
