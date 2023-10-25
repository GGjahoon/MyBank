package api

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type renewAccessTokenResponse struct {
	AccessToken         string    `json:"access_token"`
	AccessTokenExpireAt time.Time `json:"access_token_expire_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}
	//get the payload of refreshtoken
	refreshTokenPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}

	//query the session in db with payload id
	session, err := server.store.GetSession(ctx, refreshTokenPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	//the session is blocked or not
	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}
	// check the session time duration
	if time.Now().After(session.ExpireAt) {
		err := fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}
	//check the name of payload and session
	if session.Username != refreshTokenPayload.Username {
		err := fmt.Errorf("incorrect user")
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}
	//check the refreshToken in session is same as refreshToken in request or not
	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched refresh token")
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}

	//create a new access token
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(session.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	res := &renewAccessTokenResponse{
		AccessToken:         accessToken,
		AccessTokenExpireAt: accessPayload.ExpireAt,
	}
	ctx.JSON(http.StatusOK, res)
}
