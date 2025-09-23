package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/A-pen-app/invitation-sdk/models"
	"github.com/A-pen-app/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/thanhpk/randstr"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type invitationStore struct {
	db *sqlx.DB
}

func NewInvitationStore(db *sqlx.DB) Invitation {
	return &invitationStore{db: db}
}

func (us *invitationStore) Create(ctx context.Context, p models.InvitationCreateParam) (*string, error) {

	id := uuid.New().String()
	var kv map[string]interface{} = map[string]interface{}{
		"id":          id,
		"user_id":     p.UserID,
		"name":        p.Name,
		"photo_url":   p.PhotoURL,
		"departments": pq.Array(p.Departments),
		"position":    p.Position,
		"seniority":   p.Seniority,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
		"code":        p.Code,
	}
	query := `
	INSERT INTO public.invitation (
		id,
		user_id,
		name,
		photo_url,
		departments,
		position,
		seniority,
		created_at,
		updated_at,
		code
	) VALUES (
		:id,
		:user_id,
		:name,
		:photo_url,
		:departments,
		:position,
		:seniority,
		:created_at,
		:updated_at,
		:code
	)`
	if _, err := us.db.NamedExec(query, kv); err != nil {
		return nil, err
	}
	return &id, nil
}

func (us *invitationStore) Update(ctx context.Context, id string, p *models.InvitationUpdateParam) error {

	columns := []string{"updated_at=:now"}
	params := map[string]interface{}{
		"id":  id,
		"now": time.Now(),
	}

	if p.UserID != nil {
		columns = append(columns, "user_id=:user_id", "bound_at=:now")
		params["user_id"] = *p.UserID
	}
	if p.Departments != nil {
		columns = append(columns, "departments=:departments")
		params["departments"] = pq.Array(p.Departments)
	}
	if p.Position != nil {
		columns = append(columns, "position=:position")
		params["position"] = *p.Position
	}
	if p.Seniority != nil {
		columns = append(columns, "seniority=:seniority")
		params["seniority"] = *p.Seniority
	}
	if p.Gender != nil {
		columns = append(columns, "gender=:gender")
		params["gender"] = *p.Gender
	}
	if p.PassedAt != nil {
		columns = append(columns, "passed_at=:passed_at")
		params["passed_at"] = *p.PassedAt
	}

	if len(columns) == 1 {
		return fmt.Errorf("no fields to be updated")
	}

	query := fmt.Sprintf(`UPDATE public.invitation SET %s WHERE id=:id`, strings.Join(columns, ", "))

	if p.UserID != nil {
		query += " AND user_id IS NULL"
	}

	result, err := us.db.NamedExec(query, params)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already used")
	}

	logging.Infow(ctx, "update invitation", "query", query, "params", params, "invitation_id", id)
	return nil
}

func (us *invitationStore) Get(ctx context.Context, opts models.IDOptions) (*models.Invitation, error) {

	query := `
	SELECT
		id,
		user_id,
		name,
		photo_url,
		departments,
		position,
		seniority,
		created_at,
		updated_at,
		bound_at,
		gender,
		code,
		passed_at
	FROM public.invitation`

	conditions := []string{}
	params := []interface{}{}
	if opts.ID != nil {
		conditions = append(conditions, "id=?")
		params = append(params, *opts.ID)
	}
	if opts.UserID != nil {
		conditions = append(conditions, "user_id=?")
		params = append(params, *opts.UserID)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no options provided")
	}

	query += " WHERE " + strings.Join(conditions, " AND ")

	query = us.db.Rebind(query)

	invitation := &models.Invitation{}
	if err := us.db.QueryRow(query, params...).Scan(
		&invitation.ID,
		&invitation.UserID,
		&invitation.Name,
		&invitation.PhotoURL,
		pq.Array(&invitation.Departments),
		&invitation.Position,
		&invitation.Seniority,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&invitation.BoundAt,
		&invitation.Gender,
		&invitation.Code,
	); err != nil {
		logging.Errorw(ctx, "get invitation failed", "err", err)
		return nil, err
	}

	return invitation, nil
}

func (us *invitationStore) CreateCode(ctx context.Context, userID string) (string, error) {
	tx, err := us.db.BeginTx(ctx, nil)
	if err != nil {
		logging.Errorw(ctx, "begin tx failed", "err", err)
		return "", err
	}
	defer tx.Rollback()

	var code string

	query := `
		SELECT
			code
		FROM public.invitation_code
		WHERE user_id=?
	`
	query = us.db.Rebind(query)

	if err := tx.QueryRow(query, userID).Scan(&code); err == sql.ErrNoRows {

		code = randstr.String(6, charset)

		query2 := `
			INSERT INTO public.invitation_code (
				code,
				user_id,
				created_at
			) VALUES (
				?,
				?,
				?
			)`
		query2 = us.db.Rebind(query2)
		if _, err := tx.Exec(query2, code, userID, time.Now()); err != nil {
			logging.Errorw(ctx, "insert invitation code failed", "err", err)
			return "", err
		}
	} else if err != nil {
		logging.Errorw(ctx, "get invitation code failed", "err", err)
		return "", err
	}

	if err := tx.Commit(); err != nil {
		logging.Errorw(ctx, "commit tx failed", "err", err)
		return "", err
	}

	return code, nil
}

func (us *invitationStore) GetCode(ctx context.Context, opts models.CodeOptions) (*models.Code, error) {
	query := `	
	SELECT
		code,
		user_id,
		created_at
	FROM public.invitation_code`

	conditions := []string{}
	params := []interface{}{}
	if opts.UserID != nil {
		conditions = append(conditions, "user_id=?")
		params = append(params, *opts.UserID)
	}
	if opts.Code != nil {
		conditions = append(conditions, "code=?")
		params = append(params, *opts.Code)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no options provided")
	}

	query += " WHERE " + strings.Join(conditions, " AND ")

	query = us.db.Rebind(query)

	code := &models.Code{}
	if err := us.db.QueryRowx(query, params...).Scan(
		&code.Code,
		&code.UserID,
		&code.CreatedAt,
	); err != nil {
		logging.Errorw(ctx, "get invitation failed", "err", err)
		return nil, err
	}

	return code, nil
}
