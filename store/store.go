package store

import (
	"context"

	"github.com/A-pen-app/invitation-sdk/models"
)

type Invitation interface {
	Create(ctx context.Context, p models.InvitationCreateParam) (*string, error)
	Update(ctx context.Context, id string, p *models.InvitationUpdateParam) error
	Get(ctx context.Context, opts models.InvitationOptions) (*models.Invitation, error)
	CreateCode(ctx context.Context, userID string) (string, error)
	GetCode(ctx context.Context, opts models.CodeOptions) (*models.Code, error)
	UpdateCodeShareLink(ctx context.Context, code string, shareLink string) error
}
