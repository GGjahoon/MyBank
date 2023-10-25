package api

import (
	"database/sql"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"net/http"
	"time"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	res := userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	return res
}

func (server *Server) createUser(ctx *gin.Context) {
	var request createUserRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(request.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	arg := db.CreateUserParams{
		Username:       request.Username,
		HashedPassword: hashedPassword,
		FullName:       request.FullName,
		Email:          request.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	res := newUserResponse(user)
	ctx.JSON(http.StatusOK, res)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}
type loginUserResponse struct {
	SessionID            uuid.UUID    `json:"session_id"`
	AccessToken          string       `json:"access_token"`
	AccessTokenExpireAt  time.Time    `json:"access_token_expire_at"`
	RefreshToken         string       `json:"refresh_token"`
	RefreshTokenExpireAt time.Time    `json:"refresh_token_expire_at"`
	User                 userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
	}
	//check password is correct or not
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     refreshPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpireAt:     refreshPayload.ExpireAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	res := loginUserResponse{
		SessionID:            session.ID,
		AccessToken:          accessToken,
		AccessTokenExpireAt:  accessPayload.ExpireAt,
		RefreshToken:         refreshToken,
		RefreshTokenExpireAt: refreshPayload.ExpireAt,
		User:                 newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, res)
}
