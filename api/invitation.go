package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/A-pen-app/errors"
	"github.com/A-pen-app/invitation-sdk/models"
	"github.com/A-pen-app/invitation-sdk/service"
	"github.com/A-pen-app/logging"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

type invitationAPI struct {
	i service.Invitation
}

func Routes(root *gin.Engine, i service.Invitation) {
	invitation := &invitationAPI{
		i: i,
	}
	invitationGroup := root.Group("invitation")
	invitationGroup.GET("validation", invitation.getValidation)
	invitationGroup.GET("reward", invitation.getRewardInfo)
}

func (invitation *invitationAPI) getValidation(ctx *gin.Context) {
	c := ctx.Request.Context()

	p := struct {
		Type          models.InvitationType `form:"type" binding:"oneof=0 1"`
		ReferenceCode string                `form:"code" binding:"required,min=1"`
	}{}

	if err := ctx.BindQuery(&p); err != nil {
		respondWithErrorMessage(ctx, http.StatusBadRequest, err.Error())
		return
	}

	validation, err := invitation.i.GetValidation(c, models.ValidationParam{
		Type:          p.Type,
		ReferenceCode: p.ReferenceCode,
	})
	if err != nil {
		switch err {
		case errors.ErrorWrongParams:
			respondWithErrorMessage(ctx, http.StatusBadRequest, err.Error())
		case errors.ErrorNotFound:
			respondWithErrorMessage(ctx, http.StatusNotFound, err.Error())
		default:
			respondWithErrorMessage(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": validation})
}

func (invitation *invitationAPI) getRewardInfo(ctx *gin.Context) {
	c := ctx.Request.Context()

	t := struct {
		Type models.InvitationType `form:"type" binding:"oneof=0 1"`
	}{}

	if err := ctx.BindQuery(&t); err != nil {
		respondWithErrorMessage(ctx, http.StatusBadRequest, err.Error())
		return
	}

	reward, err := invitation.i.GetRewardInfo(c, t.Type)
	if err != nil {
		switch err {
		case errors.ErrorWrongParams:
			respondWithErrorMessage(ctx, http.StatusBadRequest, err.Error())
		default:
			respondWithErrorMessage(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"reward": reward})
}

func respondWithErrorMessage(ctx *gin.Context, status int,
	format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	requestID := trace.SpanContextFromContext(ctx.Request.Context()).TraceID().String()
	logging.Error(ctx, message)

	if strings.HasPrefix(message, "pq") {
		message = "database error"
	}
	ctx.AbortWithStatusJSON(status, struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}{
		Error:     message,
		RequestID: requestID,
	})
}
