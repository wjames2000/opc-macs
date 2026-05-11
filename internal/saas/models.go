package saas

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never expose
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type TenantModel struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Provider   string    `json:"provider"`
	ModelName  string    `json:"model_name"`
	APIKey     string    `json:"api_key,omitempty"`
	APIBaseURL string    `json:"api_base_url,omitempty"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
}

type TokenUsageRecord struct {
	ID           int64     `json:"id"`
	TenantID     string    `json:"tenant_id"`
	UserID       string    `json:"user_id"`
	AgentName    string    `json:"agent_name"`
	ModelName    string    `json:"model_name"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
	CreatedAt    time.Time `json:"created_at"`
}

func GenerateAPIKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "tk_" + hex.EncodeToString(b)
}

func GenerateID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CostForTokens(model string, inTokens, outTokens int) float64 {
	rates := map[string]float64{
		"deepseek-v4-flash": 0.00027,
		"deepseek-chat":     0.00027,
		"gpt-4o-mini":       0.00015,
		"gemini-2.0-flash":  0.00010,
	}
	rate, ok := rates[model]
	if !ok {
		rate = 0.001
	}
	costIn := float64(inTokens) / 1000.0 * rate
	costOut := float64(outTokens) / 1000.0 * rate * 2
	return costIn + costOut
}

var _ = fmt.Sprintf
