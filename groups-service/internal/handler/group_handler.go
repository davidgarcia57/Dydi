package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/dydi/groups-service/internal/db"
	"github.com/dydi/groups-service/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupHandler struct {
	pool         *pgxpool.Pool
	maxGroupSize int
}

func NewGroupHandler(pool *pgxpool.Pool) *GroupHandler {
	max := 8
	if v := os.Getenv("MAX_GROUP_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			max = n
		}
	}
	return &GroupHandler{pool: pool, maxGroupSize: max}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	code, err := db.GenerateInviteCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate invite code")
		return
	}

	group, err := db.CreateGroupWithOwner(r.Context(), h.pool, body.Name, code, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create group")
		return
	}

	writeJSON(w, http.StatusCreated, group)
}

func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	group, err := db.GetGroupByID(r.Context(), h.pool, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	members, err := db.GetMembers(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusOK, model.GroupWithMembers{Group: *group, Members: members})
}

func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	var body struct {
		InviteCode string `json:"invite_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.InviteCode == "" {
		writeError(w, http.StatusBadRequest, "invite_code is required")
		return
	}

	group, err := db.GetGroupByID(r.Context(), h.pool, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// Respond with 404 instead of 403 to avoid leaking group existence
	if group.InviteCode != body.InviteCode {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}

	h.addMemberToGroup(w, r, group, userID)
}

// JoinByCode resuelve el grupo desde el código de invitación solo, sin necesitar
// su id en la ruta. Existe porque la UI ya solo muestra la parte corta del código
// (`J4VZF3YT`, dictable por teléfono) en vez del `uuid:CODIGO` completo: sin este
// endpoint, quien teclea a mano lo que ve no podría entrar.
func (h *GroupHandler) JoinByCode(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		InviteCode string `json:"invite_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.InviteCode == "" {
		writeError(w, http.StatusBadRequest, "invite_code is required")
		return
	}

	group, err := db.GetGroupByInviteCode(r.Context(), h.pool, body.InviteCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	h.addMemberToGroup(w, r, group, userID)
}

// RotateInviteCode regenera el código de invitación del grupo. Es el mecanismo de
// revocación: sin esto, un código filtrado seguía sirviendo para siempre.
//
// Lo puede hacer cualquier miembro activo, no solo quien creó el grupo, y no se
// vota. Es una acción de seguridad: si alguien nota que el código se filtró, hacer
// que espere una votación de 24 h deja la puerta abierta justo cuando urge
// cerrarla. El costo de un abuso es bajo — se vuelve a rotar y se comparte el
// nuevo.
func (h *GroupHandler) RotateInviteCode(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	// 404 y no 403, igual que en JoinGroup: no se filtra si el grupo existe.
	if !member {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}

	// invite_code es UNIQUE. Con 32^8 combinaciones una colisión es casi
	// imposible, pero reintentar sale gratis y evita un 500 absurdo.
	var group *model.Group
	for attempt := 0; attempt < 3; attempt++ {
		code, genErr := db.GenerateInviteCode()
		if genErr != nil {
			writeError(w, http.StatusInternalServerError, "could not generate invite code")
			return
		}

		group, err = db.RotateInviteCode(r.Context(), h.pool, groupID, code)
		if err == nil {
			writeJSON(w, http.StatusOK, group)
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue // colisión de código: otro intento con uno nuevo
		}
		break
	}

	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "could not rotate invite code")
}

// addMemberToGroup es la cola compartida de JoinGroup y JoinByCode: valida que no
// sea ya miembro, respeta el tope del grupo y agrega la membresía.
func (h *GroupHandler) addMemberToGroup(
	w http.ResponseWriter, r *http.Request, group *model.Group, userID string,
) {
	already, err := db.IsMember(r.Context(), h.pool, group.ID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if already {
		writeError(w, http.StatusConflict, "already a member")
		return
	}

	count, err := db.CountMembers(r.Context(), h.pool, group.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if count >= h.maxGroupSize {
		writeError(w, http.StatusConflict, "group is full")
		return
	}

	if err := db.AddMember(r.Context(), h.pool, group.ID, userID, "member"); err != nil {
		writeError(w, http.StatusInternalServerError, "could not join group")
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func (h *GroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	members, err := db.GetMembers(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (h *GroupHandler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groups, err := db.GetGroupsByUserID(r.Context(), h.pool, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	if err := db.SetMembershipStatus(r.Context(), h.pool, groupID, userID, "left"); err != nil {
		writeError(w, http.StatusInternalServerError, "could not leave group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckMembership is an internal endpoint (called by realtime-service before it
// accepts a WebSocket). 204 = active member, 403 = not. Identity here comes from
// the URL, not X-User-ID, because the caller is a trusted service, not the user.
func (h *GroupHandler) CheckMembership(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	userID := chi.URLParam(r, "userID")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
