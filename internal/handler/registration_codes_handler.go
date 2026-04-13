package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/handler/dto"
	"github.com/111zxc/rss-communicator/internal/service"
)

type RegistrationCodeHandler struct {
	regCodeService *service.RegistrationCodeService
}

func NewRegistrationCodeHandler(regCodeService *service.RegistrationCodeService) *RegistrationCodeHandler {
	return &RegistrationCodeHandler{regCodeService: regCodeService}
}

func (h *RegistrationCodeHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseListParams(r)
	items, total, err := h.regCodeService.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to list registration codes", nil)
		return
	}
	out := make([]dto.RegistrationCodeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapRegistrationCodeResponse(item))
	}
	writeJSON(w, 200, dto.ListResponse[dto.RegistrationCodeResponse]{Items: out, Total: total})
}

func (h *RegistrationCodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.regCodeService.GetByID(r.Context(), chi.URLParam(r, "codeID"))
	if err != nil {
		writeServiceError(w, err, "failed to get registration code")
		return
	}
	writeJSON(w, 200, mapRegistrationCodeResponse(item))
}

func (h *RegistrationCodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.RegistrationCodeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	item, err := h.regCodeService.Create(r.Context(), service.RegistrationCodeInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		MaxUses:     req.MaxUses,
		ExpiresAt:   req.ExpiresAt,
	})
	if err != nil {
		writeServiceError(w, err, "failed to create registration code")
		return
	}
	writeJSON(w, 201, mapRegistrationCodeResponse(item))
}

func (h *RegistrationCodeHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.RegistrationCodeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	item, err := h.regCodeService.Update(r.Context(), chi.URLParam(r, "codeID"), service.RegistrationCodeInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		MaxUses:     req.MaxUses,
		ExpiresAt:   req.ExpiresAt,
	})
	if err != nil {
		writeServiceError(w, err, "failed to update registration code")
		return
	}
	writeJSON(w, 200, mapRegistrationCodeResponse(item))
}

func (h *RegistrationCodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.regCodeService.Delete(r.Context(), chi.URLParam(r, "codeID")); err != nil {
		writeServiceError(w, err, "failed to delete registration code")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RegistrationCodeHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	items, err := h.regCodeService.ListGroups(r.Context(), chi.URLParam(r, "codeID"))
	if err != nil {
		writeServiceError(w, err, "failed to list registration code groups")
		return
	}
	out := make([]dto.GroupResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapGroupResponse(item))
	}
	writeJSON(w, 200, dto.ListResponse[dto.GroupResponse]{Items: out, Total: len(out)})
}

func (h *RegistrationCodeHandler) AddGroup(w http.ResponseWriter, r *http.Request) {
	var req dto.RegistrationCodeGroupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	if err := h.regCodeService.AddGroup(r.Context(), chi.URLParam(r, "codeID"), req.GroupID); err != nil {
		writeServiceError(w, err, "failed to add registration code group")
		return
	}
	writeJSON(w, 201, map[string]any{"registration_code_id": chi.URLParam(r, "codeID"), "group_id": req.GroupID})
}

func (h *RegistrationCodeHandler) RemoveGroup(w http.ResponseWriter, r *http.Request) {
	if err := h.regCodeService.RemoveGroup(r.Context(), chi.URLParam(r, "codeID"), chi.URLParam(r, "groupID")); err != nil {
		writeServiceError(w, err, "failed to remove registration code group")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapRegistrationCodeResponse(rc domain.RegistrationCode) dto.RegistrationCodeResponse {
	return dto.RegistrationCodeResponse{
		ID:          rc.ID,
		Code:        rc.Code,
		Name:        rc.Name,
		Description: rc.Description,
		Enabled:     rc.Enabled,
		MaxUses:     rc.MaxUses,
		UseCount:    rc.UseCount,
		ExpiresAt:   rc.ExpiresAt,
		CreatedAt:   rc.CreatedAt,
		UpdatedAt:   rc.UpdatedAt,
	}
}
