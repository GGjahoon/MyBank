package api

import (
	"database/sql"
	"fmt"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccoundID   int64  `json:"to_accound_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var request createTransferRequest

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}
	if !server.validAccount(ctx, request.FromAccountID, request.Currency) {
		return
	}

	if !server.validAccount(ctx, request.ToAccoundID, request.Currency) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: request.FromAccountID,
		ToAccountID:   request.ToAccoundID,
		Amount:        request.Amount,
	}
	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
	}
	ctx.JSON(http.StatusOK, result)
}

// validAccount to valid the account's currency is same as the request's currency or not
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return false
	}
	if account.Currency != currency {
		err = fmt.Errorf("account [%d] currency mismatch %s vs %s ", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return false
	}
	return true
}
