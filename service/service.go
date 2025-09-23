package service

import (
	"context"

	"github.com/A-pen-app/deeplink"
	"github.com/A-pen-app/invitation-sdk/models"
)

type Invitation interface {
	GetValidation(ctx context.Context, p models.InvitationParam) (*models.InvitationValidationStatus, error)
	CreateCode(ctx context.Context, userID string, platform deeplink.Platform) (*models.InvitationResponse, *models.Reward, error)
	GetRewardInfo(ctx context.Context, t models.InvitationType) (*models.Reward, error)
}
