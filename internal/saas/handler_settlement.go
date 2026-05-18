package saas

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/wjames2000/opc-macs/internal/settlement"
)

// SettlementHandler handles settlement calculation and management.
type SettlementHandler struct {
	mu     sync.RWMutex
	engine *settlement.Engine
	orders map[string]*settlementRecord // in-memory settlement store
}

type settlementRecord struct {
	ID        string
	OrderID   string
	Type      settlement.SettlementType
	Gross     float64
	Commiss   float64
	Creator   float64
	Breakdown map[string]float64
}

func NewSettlementHandler() *SettlementHandler {
	return &SettlementHandler{
		engine: settlement.NewEngine(),
		orders: make(map[string]*settlementRecord),
	}
}

type settleCalReq struct {
	OrderID         string           `json:"order_id"`
	Type            string           `json:"type"`
	Budget          float64          `json:"budget,omitempty"`
	SalesAmount     float64          `json:"sales_amount,omitempty"`
	Impressions     int64            `json:"impressions,omitempty"`
	CPMRate         float64          `json:"cpm_rate,omitempty"`
	CPSRate         float64          `json:"cps_rate,omitempty"`
	EngagementCount map[string]int64 `json:"engagements,omitempty"`
	TotalEarnings   float64          `json:"total_earnings,omitempty"`
}

type settleCalResp struct {
	SettlementID       string             `json:"settlement_id"`
	OrderID            string             `json:"order_id"`
	Type               string             `json:"type"`
	GrossAmount        float64            `json:"gross_amount"`
	PlatformCommission float64            `json:"platform_commission"`
	CreatorAmount      float64            `json:"creator_amount"`
	Breakdown          map[string]float64 `json:"breakdown"`
}

type settleListResp struct {
	Settlements []settleCalResp `json:"settlements"`
	Total       int             `json:"total"`
}

// HandleCalculate handles POST /api/v1/settlement/calculate — manual settlement calculation.
func (h *SettlementHandler) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	var req settleCalReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.OrderID == "" {
		writeError(w, http.StatusBadRequest, "order_id required")
		return
	}

	st, err := parseSettlementType(req.Type)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid settlement type: "+err.Error())
		return
	}

	sr := &settlement.SettlementRequest{
		Type:            st,
		Budget:          req.Budget,
		SalesAmount:     req.SalesAmount,
		Impressions:     req.Impressions,
		CPMRate:         req.CPMRate,
		CPSRate:         req.CPSRate,
		EngagementCount: req.EngagementCount,
		TotalEarnings:   req.TotalEarnings,
	}

	result, err := h.engine.Calculate(sr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "calculation failed: "+err.Error())
		return
	}

	rec := &settlementRecord{
		ID:        settlementID(),
		OrderID:   req.OrderID,
		Type:      st,
		Gross:     result.GrossAmount,
		Commiss:   result.PlatformCommission,
		Creator:   result.CreatorAmount,
		Breakdown: result.Breakdown,
	}

	h.mu.Lock()
	h.orders[rec.ID] = rec
	h.mu.Unlock()

	writeJSON(w, http.StatusOK, settleCalResp{
		SettlementID:       rec.ID,
		OrderID:            rec.OrderID,
		Type:               string(rec.Type),
		GrossAmount:        rec.Gross,
		PlatformCommission: rec.Commiss,
		CreatorAmount:      rec.Creator,
		Breakdown:          rec.Breakdown,
	})
}

// HandleList handles GET /api/v1/settlement/orders — list settlements.
func (h *SettlementHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := make([]settleCalResp, 0, len(h.orders))
	for _, rec := range h.orders {
		list = append(list, settleCalResp{
			SettlementID:       rec.ID,
			OrderID:            rec.OrderID,
			Type:               string(rec.Type),
			GrossAmount:        rec.Gross,
			PlatformCommission: rec.Commiss,
			CreatorAmount:      rec.Creator,
			Breakdown:          rec.Breakdown,
		})
	}

	writeJSON(w, http.StatusOK, settleListResp{Settlements: list, Total: len(list)})
}

// HandleGet handles GET /api/v1/settlement/orders/{id} — settlement detail.
func (h *SettlementHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "settlement_id required")
		return
	}

	h.mu.RLock()
	rec, ok := h.orders[id]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "settlement not found")
		return
	}

	writeJSON(w, http.StatusOK, settleCalResp{
		SettlementID:       rec.ID,
		OrderID:            rec.OrderID,
		Type:               string(rec.Type),
		GrossAmount:        rec.Gross,
		PlatformCommission: rec.Commiss,
		CreatorAmount:      rec.Creator,
		Breakdown:          rec.Breakdown,
	})
}

// HandleInvoiceDownload handles GET /api/v1/settlement/invoices/{id} — download invoice stub.
func (h *SettlementHandler) HandleInvoiceDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "settlement_id required")
		return
	}

	h.mu.RLock()
	rec, ok := h.orders[id]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "settlement not found")
		return
	}

	// Generate a simple JSON invoice (real implementation would generate PDF)
	invoice := map[string]interface{}{
		"invoice_id":          id,
		"order_id":            rec.OrderID,
		"settlement_type":     string(rec.Type),
		"gross_amount":        rec.Gross,
		"platform_commission": rec.Commiss,
		"creator_amount":      rec.Creator,
		"breakdown":           rec.Breakdown,
		"format":              "json",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=settlement_"+id+".json")
	json.NewEncoder(w).Encode(invoice)
}

func (h *SettlementHandler) HandleSettleOrder(w http.ResponseWriter, r *http.Request) {
	h.HandleCalculate(w, r)
}

func parseSettlementType(t string) (settlement.SettlementType, error) {
	switch t {
	case "cps":
		return settlement.SettlementCPS, nil
	case "cpe":
		return settlement.SettlementCPE, nil
	case "cpm":
		return settlement.SettlementCPM, nil
	case "hybrid":
		return settlement.SettlementHybrid, nil
	case "fixed":
		return settlement.SettlementFixed, nil
	case "":
		return settlement.SettlementFixed, nil
	default:
		return "", errors.New("unknown type: " + t)
	}
}

func settlementID() string {
	return "stl_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
