package settlement

import (
	"fmt"
	"math"
)

// SettlementType defines the settlement model.
type SettlementType string

const (
	SettlementCPS    SettlementType = "cps"    // Cost Per Sale
	SettlementCPE    SettlementType = "cpe"    // Cost Per Engagement
	SettlementCPM    SettlementType = "cpm"    // Cost Per Mille (1000 impressions)
	SettlementHybrid SettlementType = "hybrid" // Base CPM + incentive CPS
	SettlementFixed  SettlementType = "fixed"  // Fixed price
)

type SettlementRequest struct {
	Type            SettlementType
	Budget          float64
	SalesAmount     float64            // for CPS
	EngagementCount map[string]int64   // for CPE: likes, comments, shares, saves
	Impressions     int64              // for CPM
	CPMRate         float64            // CPM rate per 1000 impressions
	CPSRate         float64            // CPS commission rate (0-1)
	CPEWeights      map[string]float64 // Engagement type weights
	CommissionRate  float64            // Platform commission rate (0-1)
	TotalEarnings   float64            // Creator's cumulative earnings for tiered commission
}

type SettlementResult struct {
	SettlementType     SettlementType
	GrossAmount        float64
	PlatformCommission float64
	CreatorAmount      float64
	Breakdown          map[string]float64
}

type Engine struct {
	tieredCommission []CommissionTier
}

type CommissionTier struct {
	MinEarnings float64
	Rate        float64
}

func NewEngine() *Engine {
	return &Engine{
		tieredCommission: []CommissionTier{
			{MinEarnings: 0, Rate: 0.10},
			{MinEarnings: 10000, Rate: 0.08},
			{MinEarnings: 50000, Rate: 0.06},
			{MinEarnings: 100000, Rate: 0.04},
			{MinEarnings: 500000, Rate: 0.02},
		},
	}
}

func (e *Engine) Calculate(req *SettlementRequest) (*SettlementResult, error) {
	switch req.Type {
	case SettlementCPS:
		return e.calcCPS(req)
	case SettlementCPE:
		return e.calcCPE(req)
	case SettlementCPM:
		return e.calcCPM(req)
	case SettlementHybrid:
		return e.calcHybrid(req)
	case SettlementFixed:
		return e.calcFixed(req)
	default:
		return nil, fmt.Errorf("unsupported settlement type: %s", req.Type)
	}
}

func (e *Engine) calcCPS(req *SettlementRequest) (*SettlementResult, error) {
	gross := req.SalesAmount * req.CPSRate
	commission := gross * e.getCommissionRate(req.TotalEarnings)
	return &SettlementResult{
		SettlementType:     SettlementCPS,
		GrossAmount:        gross,
		PlatformCommission: commission,
		CreatorAmount:      gross - commission,
		Breakdown:          map[string]float64{"sales_amount": req.SalesAmount, "cps_rate": req.CPSRate},
	}, nil
}

func (e *Engine) calcCPE(req *SettlementRequest) (*SettlementResult, error) {
	var totalWeighted float64
	weights := req.CPEWeights
	if weights == nil {
		weights = map[string]float64{"like": 1, "comment": 3, "share": 5, "save": 4}
	}
	breakdown := make(map[string]float64)
	for engType, count := range req.EngagementCount {
		weight := weights[engType]
		if weight == 0 {
			weight = 1
		}
		amount := float64(count) * weight
		totalWeighted += amount
		breakdown[engType] = amount
	}
	gross := math.Min(totalWeighted, req.Budget)
	commission := gross * e.getCommissionRate(req.TotalEarnings)
	return &SettlementResult{
		SettlementType:     SettlementCPE,
		GrossAmount:        gross,
		PlatformCommission: commission,
		CreatorAmount:      gross - commission,
		Breakdown:          breakdown,
	}, nil
}

func (e *Engine) calcCPM(req *SettlementRequest) (*SettlementResult, error) {
	gross := (float64(req.Impressions) / 1000) * req.CPMRate
	commission := gross * e.getCommissionRate(req.TotalEarnings)
	return &SettlementResult{
		SettlementType:     SettlementCPM,
		GrossAmount:        gross,
		PlatformCommission: commission,
		CreatorAmount:      gross - commission,
		Breakdown:          map[string]float64{"impressions": float64(req.Impressions), "cpm_rate": req.CPMRate},
	}, nil
}

func (e *Engine) calcHybrid(req *SettlementRequest) (*SettlementResult, error) {
	cpmAmount := (float64(req.Impressions) / 1000) * req.CPMRate
	cpsAmount := req.SalesAmount * req.CPSRate
	gross := cpmAmount + cpsAmount
	commission := gross * e.getCommissionRate(req.TotalEarnings)
	return &SettlementResult{
		SettlementType:     SettlementHybrid,
		GrossAmount:        gross,
		PlatformCommission: commission,
		CreatorAmount:      gross - commission,
		Breakdown:          map[string]float64{"cpm_amount": cpmAmount, "cps_amount": cpsAmount},
	}, nil
}

func (e *Engine) calcFixed(req *SettlementRequest) (*SettlementResult, error) {
	commission := req.Budget * e.getCommissionRate(req.TotalEarnings)
	return &SettlementResult{
		SettlementType:     SettlementFixed,
		GrossAmount:        req.Budget,
		PlatformCommission: commission,
		CreatorAmount:      req.Budget - commission,
		Breakdown:          map[string]float64{"fixed_amount": req.Budget},
	}, nil
}

func (e *Engine) getCommissionRate(totalEarnings float64) float64 {
	rate := 0.10 // default 10%
	for _, tier := range e.tieredCommission {
		if totalEarnings >= tier.MinEarnings {
			rate = tier.Rate
		}
	}
	return rate
}
