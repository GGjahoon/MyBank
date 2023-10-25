package api

import (
	"fmt"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/token"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for banking service
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer create a new http server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
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
	server.setUpRouter()
	return server, nil
}

func (server *Server) setUpRouter() {
	router := gin.Default()
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.POST("/transfers", server.createTransfer)
	server.router = router
}

// Start runs the http server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
func errResponse(err error) gin.H {
	return gin.H{"err": err.Error()}
}
