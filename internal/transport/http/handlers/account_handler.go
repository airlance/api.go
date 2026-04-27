package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/resoul/api/internal/domain"
	"github.com/resoul/api/internal/middleware"
	"github.com/resoul/api/internal/transport/http/utils"
)

// AccountResponse is the typed response payload for account endpoints.
// Merges auth identity fields with application-level account data.
type AccountResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	BucketName   string  `json:"bucket_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone,omitempty"`
	Role         string  `json:"role"`
	DisplayName  string  `json:"display_name"`
	AvatarURL    string  `json:"avatar_url,omitempty"`
	Bio          string  `json:"bio,omitempty"`
	LastSignInAt *string `json:"last_sign_in_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// AccountHandler handles account-related HTTP routes.
type AccountHandler struct {
	svc domain.AccountService
}

// NewAccountHandler returns an AccountHandler backed by the given service.
func NewAccountHandler(svc domain.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// GetMe returns the authenticated user's merged auth + account data.
// Creates an empty account on first call (idempotent).
//
// GET /api/v1/user/me
func (h *AccountHandler) GetMe(c *gin.Context) {
	authUser, ok := contextUser(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", "unauthenticated request")
		return
	}

	account, err := h.svc.GetOrCreate(c.Request.Context(), authUser.ID)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, toAccountResponse(authUser, account))
}

// UpdateAccount applies a partial update to the authenticated user's account.
// All fields are optional — only non-null fields in the JSON body are updated.
//
// PATCH /api/v1/user/account
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	authUser, ok := contextUser(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", "unauthenticated request")
		return
	}

	var inp domain.UpdateAccountInput
	if err := c.ShouldBindJSON(&inp); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	account, err := h.svc.Update(c.Request.Context(), authUser.ID, inp)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, toAccountResponse(authUser, account))
}

func contextUser(c *gin.Context) (*middleware.AuthUser, bool) {
	raw, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		return nil, false
	}
	user, ok := raw.(*middleware.AuthUser)
	return user, ok
}

func toAccountResponse(u *middleware.AuthUser, a *domain.Account) AccountResponse {
	return AccountResponse{
		ID:          a.ID,
		UserID:      a.UserID,
		BucketName:  a.BucketName,
		Email:       u.Email,
		Role:        u.Role,
		DisplayName: a.DisplayName,
		AvatarURL:   a.AvatarURL,
		Bio:         a.Bio,
		CreatedAt:   a.CreatedAt.String(),
	}
}
