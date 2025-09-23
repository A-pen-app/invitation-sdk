package service

import (
	"context"
	"database/sql"

	"github.com/A-pen-app/deeplink"
	"github.com/A-pen-app/errors"
	"github.com/A-pen-app/invitation-sdk/models"
	"github.com/A-pen-app/invitation-sdk/store"
	"github.com/A-pen-app/logging"
)

type InvitationService struct {
	i store.Invitation
}

func NewInvitationService(i store.Invitation) Invitation {
	return &InvitationService{i: i}
}

func (is *InvitationService) GetValidation(ctx context.Context, p models.ValidationParam) (*models.InvitationValidationStatus, error) {
	switch p.Type {
	case models.InvitationTypeID:
		invitation, err := is.i.Get(ctx, models.IDOptions{ID: &p.ReferenceCode})
		if err == sql.ErrNoRows {
			return nil, errors.ErrorNotFound
		} else if err != nil {
			logging.Errorw(ctx, "get invitation failed", "err", err, "invitation_id", p.ReferenceCode)
			return nil, err
		}
		status := models.InvitationValidationStatusValid

		if invitation.UserID != nil {
			status = models.InvitationValidationStatusUsed
		}

		return &status, nil
	case models.InvitationTypeCode:

		_, err := is.i.GetCode(ctx, models.CodeOptions{Code: &p.ReferenceCode})
		if err == sql.ErrNoRows {
			return nil, errors.ErrorNotFound
		} else if err != nil {
			logging.Errorw(ctx, "get invitation code failed", "err", err, "invitation_code", p.ReferenceCode)
			return nil, err
		}
		status := models.InvitationValidationStatusValid

		return &status, nil
	default:
		logging.Errorw(ctx, "invitation type not supported", "invitation_type", p.Type)
		return nil, errors.ErrorWrongParams
	}
}

func (is *InvitationService) CreateCode(ctx context.Context, userID string, platform deeplink.Platform) (*models.InvitationResponse, *models.Reward, error) {

	code, err := is.i.CreateCode(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	link, err := deeplink.NewReferralLink(platform, code)
	if err != nil {
		logging.Errorw(ctx, "failed generating link to invite", "err", err)
		return nil, nil, err
	}

	deeplink, err := link.Build()
	if err != nil {
		logging.Errorw(ctx, "failed generating link to invite", "err", err)
		return nil, nil, err
	}

	return &models.InvitationResponse{
		Type:          models.InvitationTypeCode,
		ReferenceCode: code,
		Deeplink:      &deeplink,
	}, &models.InvitationReward, nil
}

func (is *InvitationService) GetRewardInfo(ctx context.Context, t models.InvitationType) (*models.Reward, error) {

	switch t {
	case models.InvitationTypeCode:
		return &models.InvitationReward, nil
	default:
		logging.Errorw(ctx, "invitation type not supported", "invitation_type", t)
		return nil, errors.ErrorWrongParams
	}

}
