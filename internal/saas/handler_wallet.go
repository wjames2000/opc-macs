package saas

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type WalletHandler struct {
	wallet     *WalletStore
	withdrawal *WithdrawalStore
}

func NewWalletHandler(wallet *WalletStore, withdrawal *WithdrawalStore) *WalletHandler {
	return &WalletHandler{wallet: wallet, withdrawal: withdrawal}
}

func (h *WalletHandler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := getTenantID(r), getUserID(r)
	wal, err := h.wallet.EnsureWallet(r.Context(), tenantID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get wallet: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wal)
}

func (h *WalletHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	_, userID := r.Context().Value(CtxTenantID).(string), r.Context().Value(CtxUserID).(string)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	wallet, err := h.wallet.EnsureWallet(r.Context(), r.Context().Value(CtxTenantID).(string), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get wallet: "+err.Error())
		return
	}
	txs, err := h.wallet.GetTransactions(r.Context(), wallet.ID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get txs: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, txs)
}

type createWithdrawalRequest struct {
	Amount      float64 `json:"amount"`
	Method      string  `json:"method"`
	AccountInfo string  `json:"account_info"`
}

func (h *WalletHandler) HandleCreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := r.Context().Value(CtxTenantID).(string), r.Context().Value(CtxUserID).(string)

	var req createWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Method == "" {
		writeError(w, http.StatusBadRequest, "method is required")
		return
	}
	if req.AccountInfo == "" {
		writeError(w, http.StatusBadRequest, "account_info is required")
		return
	}

	wallet, err := h.wallet.EnsureWallet(r.Context(), tenantID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get wallet: "+err.Error())
		return
	}
	if wallet.Balance < req.Amount {
		writeError(w, http.StatusBadRequest, "insufficient balance")
		return
	}

	wdr := &Withdrawal{
		ID:          GenerateID(),
		TenantID:    tenantID,
		UserID:      userID,
		WalletID:    wallet.ID,
		Amount:      req.Amount,
		Method:      WithdrawalMethod(req.Method),
		AccountInfo: req.AccountInfo,
	}

	if err := h.withdrawal.Create(r.Context(), wdr); err != nil {
		writeError(w, http.StatusInternalServerError, "create withdrawal: "+err.Error())
		return
	}

	if err := h.wallet.Freeze(r.Context(), wallet.ID, req.Amount, wdr.ID, "withdrawal freeze"); err != nil {
		writeError(w, http.StatusInternalServerError, "freeze: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, wdr)
}

func (h *WalletHandler) HandleListWithdrawals(w http.ResponseWriter, r *http.Request) {
	_, userID := r.Context().Value(CtxTenantID).(string), r.Context().Value(CtxUserID).(string)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	wdrs, err := h.withdrawal.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list withdrawals: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wdrs)
}

func (h *WalletHandler) HandleAdminListPendingWithdrawals(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(CtxTenantID).(string)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	wdrs, err := h.withdrawal.ListPending(r.Context(), tenantID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list pending: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wdrs)
}

type approveWithdrawalRequest struct {
	Action string `json:"action"`
	Note   string `json:"note,omitempty"`
}

func (h *WalletHandler) HandleAdminProcessWithdrawal(w http.ResponseWriter, r *http.Request) {
	withdrawalID := r.PathValue("id")
	adminID := r.Context().Value(CtxUserID).(string)

	var req approveWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	switch req.Action {
	case "approve":
		if err := h.withdrawal.Approve(r.Context(), withdrawalID, adminID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "reject":
		if err := h.withdrawal.Reject(r.Context(), withdrawalID, req.Note, adminID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		writeError(w, http.StatusBadRequest, "action must be 'approve' or 'reject'")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *WalletHandler) HandleAdminListWithdrawals(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(CtxTenantID).(string)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	wdrs, err := h.withdrawal.ListPending(r.Context(), tenantID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wdrs)
}
