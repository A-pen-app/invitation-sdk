package models

import "time"

type InvitationType int

const (
	InvitationTypeID   InvitationType = iota //快速加入
	InvitationTypeCode                       //邀請碼
)

type InvitationStatus int

const (
	//快速加入
	InvitationStatusNewUserBound InvitationStatus = iota
	InvitationStatusOldUserBound
)

type InvitationValidationStatus int

const (
	InvitationValidationStatusValid InvitationValidationStatus = iota
	InvitationValidationStatusUsed
	InvitationValidationStatusExpired
)

type Invitation struct {
	ID        string    `json:"-" db:"id"`
	UserID    *string   `json:"-" db:"user_id"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`

	//快速加入
	Name        *string     `json:"-" db:"name"`
	Gender      *UserGender `json:"-" db:"gender"`
	PhotoURL    string      `json:"-" db:"photo_url"`
	Departments []string    `json:"-" db:"departments"`
	Position    *string     `json:"-" db:"position"`
	Seniority   *int        `json:"-" db:"seniority"`
	BoundAt     *time.Time  `json:"-" db:"bound_at"`

	//邀請碼
	Code     *string    `json:"-" db:"code"`
	PassedAt *time.Time `json:"-" db:"passed_at"`
}

type Code struct {
	Code      string    `json:"-" db:"code"`
	UserID    string    `json:"-" db:"user_id"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}

type CodeOptions struct {
	UserID *string
	Code   *string
}

type InvitationOptions struct {
	ID     *string
	UserID *string
}

type InvitationCreateParam struct {
	UserID      *string
	Name        *string
	PhotoURL    *string
	Departments []string
	Position    *string
	Seniority   *int
	Code        *string
}

type InvitationUpdateParam struct {
	UserID      *string     `json:"-"`
	Departments []string    `json:"departments"`
	Position    *string     `json:"position"`
	Seniority   *int        `json:"seniority"`
	Gender      *UserGender `json:"gender"`
	PassedAt    *time.Time  `json:"passed_at"`
}

type InvitationParam struct {
	Type          InvitationType `json:"type"`
	ReferenceCode string         `json:"code"`
}

type InvitationResponse struct {
	Type          InvitationType    `json:"type"`
	ReferenceCode *string           `json:"code,omitempty"`
	Status        *InvitationStatus `json:"status,omitempty"`
	Deeplink      *string           `json:"deeplink,omitempty"`
}

type Reward struct {
	Currency string `json:"currency"`
	Quantity string `json:"quantity"`
}

var InvitationReward = Reward{
	Currency: "coins",
	Quantity: "30",
}

type UserGender int

const (
	GenderMale UserGender = iota
	GenderFemale
)

func (ug UserGender) English() string {

	switch ug {
	case GenderMale:
		return "Male"
	case GenderFemale:
		return "Female"
	default:
		return "Male"
	}
}
