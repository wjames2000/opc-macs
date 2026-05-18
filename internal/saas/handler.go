package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type Server struct {
	Tenants          *TenantStore
	Users            *UserStore
	Loader           *runtime.Loader
	ModelKey         string // fallback model API key
	ModelURL         string // fallback model base URL
	Marketplace      *MarketplaceHandlers
	Wallet           *WalletHandler
	Settlement       *SettlementHandler
	Credit           *CreditHandler
	Notify           *NotificationHandler
	VideoScript      *VideoScriptHandler
	Engage           *EngageHandler
	TagGenerator     *TagGeneratorHandler
	TrendRadar       *TrendRadarHandler
	Delivery         *DeliveryHandler
	Review           *ReviewHandler
	Invite           *InviteHandler
	Recommend        *RecommendationHandler
	Invoice          *InvoiceHandler
	PaymentMethod    *PaymentMethodHandler
	ReportExport     *ReportExportHandler
	FraudDetect      *FraudDetectionHandler
	FraudAccount     *FraudAccountHandler
	FraudContent     *FraudContentHandler
	FraudAlert       *FraudAlertHandler
	WithdrawPayment  *WithdrawPaymentHandler
	AdvertiserReport *AdvertiserReportHandler
	CreatorReport    *CreatorReportHandler
	walletStore      *WalletStore
	creditStore      *CreditStore
	withdrawlStr     *WithdrawalStore
	notifyStore      *NotificationStore
	invoiceStore     *InvoiceStore
	paymentMthdStr   *PaymentMethodStore
	reportExpStr     *ReportExportStore
	fraudDetStr      *FraudDetectionStore
	fraudAcctStr     *FraudAccountStore
	fraudContStr     *FraudContentStore
	fraudAlrtStr     *FraudAlertStore
}

func NewServer(db *sql.DB, loader *runtime.Loader) (*Server, error) {
	tenants := NewTenantStore(db)
	users := NewUserStore(db)
	marketplace := NewMarketplaceStore(db)
	walletStr := NewWalletStore(db)
	withdrStr := NewWithdrawalStore(db)
	creditStr := NewCreditStore(db)

	if err := tenants.InitSchema(); err != nil {
		return nil, err
	}
	if err := marketplace.InitSchema(); err != nil {
		return nil, err
	}
	if err := NewInviteHandler(marketplace).InitSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("invite schema: %w", err)
	}
	if err := walletStr.InitSchema(); err != nil {
		return nil, err
	}
	if err := withdrStr.InitSchema(); err != nil {
		return nil, err
	}
	if err := creditStr.InitSchema(); err != nil {
		return nil, err
	}
	notifStr := NewNotificationStore(db)
	if err := notifStr.InitSchema(); err != nil {
		return nil, err
	}
	invoiceStr := NewInvoiceStore(db)
	pmtMthdStr := NewPaymentMethodStore(db)
	reportExpStr := NewReportExportStore(db)
	fraudDetStr := NewFraudDetectionStore(db)
	fraudAcctStr := NewFraudAccountStore(db)
	fraudContStr := NewFraudContentStore(db)
	fraudAlrtStr := NewFraudAlertStore(db)
	withdrawPayStr := NewWithdrawPaymentStore(db)
	advReportStr := NewAdvertiserReportStore(db)
	creatorReportStr := NewCreatorReportStore(db)
	platformAccStr := NewPlatformAccountStore(db)
	if err := platformAccStr.InitSchema(); err != nil {
		return nil, fmt.Errorf("platform account schema: %w", err)
	}
	lifecycle := NewOrderLifecycle(marketplace, walletStr)

	return &Server{
		Tenants:          tenants,
		Users:            users,
		Loader:           loader,
		Marketplace:      NewMarketplaceHandlers(marketplace),
		Wallet:           NewWalletHandler(walletStr, withdrStr),
		Settlement:       NewSettlementHandler(),
		Credit:           NewCreditHandler(creditStr),
		Notify:           NewNotificationHandler(notifStr, lifecycle),
		VideoScript:      NewVideoScriptHandler(loader),
		Engage:           NewEngageHandler(loader),
		TagGenerator:     NewTagGeneratorHandler(loader),
		TrendRadar:       NewTrendRadarHandler(loader),
		Delivery:         NewDeliveryHandler(marketplace),
		Review:           NewReviewHandler(marketplace),
		Invite:           NewInviteHandler(marketplace),
		Recommend:        NewRecommendationHandler(marketplace),
		Invoice:          NewInvoiceHandler(invoiceStr),
		PaymentMethod:    NewPaymentMethodHandler(pmtMthdStr),
		ReportExport:     NewReportExportHandler(reportExpStr),
		FraudDetect:      NewFraudDetectionHandler(fraudDetStr),
		FraudAccount:     NewFraudAccountHandler(fraudAcctStr),
		FraudContent:     NewFraudContentHandler(fraudContStr),
		FraudAlert:       NewFraudAlertHandler(fraudAlrtStr),
		WithdrawPayment:  NewWithdrawPaymentHandler(withdrawPayStr),
		AdvertiserReport: NewAdvertiserReportHandler(advReportStr),
		CreatorReport:    NewCreatorReportHandler(creatorReportStr),
		walletStore:      walletStr,
		creditStore:      creditStr,
		withdrawlStr:     withdrStr,
		notifyStore:      notifStr,
		invoiceStore:     invoiceStr,
		paymentMthdStr:   pmtMthdStr,
		reportExpStr:     reportExpStr,
		fraudDetStr:      fraudDetStr,
		fraudAcctStr:     fraudAcctStr,
		fraudContStr:     fraudContStr,
		fraudAlrtStr:     fraudAlrtStr,
	}, nil
}

// POST /api/v1/auth/register
func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	var req struct {
		TenantName string `json:"tenant_name"`
		Slug       string `json:"slug"`
		Email      string `json:"email"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		jsonError(w, 400, "email and password required")
		return
	}

	// Create tenant and owner user
	tenant, err := s.Tenants.Create(r.Context(), req.TenantName, req.Slug)
	if err != nil {
		jsonError(w, 409, "tenant creation failed: "+err.Error())
		return
	}
	user, err := s.Users.Create(r.Context(), tenant.ID, req.Email, req.Password)
	if err != nil {
		jsonError(w, 409, "user creation failed: "+err.Error())
		return
	}

	jsonResponse(w, 201, map[string]interface{}{
		"tenant_id": tenant.ID,
		"user_id":   user.ID,
		"api_key":   user.ID, // simple: use user ID as API key
	})
}

// POST /api/v1/auth/login
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request")
		return
	}
	user, err := s.Users.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		jsonError(w, 401, "invalid credentials")
		return
	}
	token := GenerateToken(user.ID)
	jsonResponse(w, 200, map[string]interface{}{
		"token":     token,
		"user_id":   user.ID,
		"tenant_id": user.TenantID,
		"role":      user.Role,
	})
}

// GET /api/v1/agents
func (s *Server) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	plugins := s.Loader.List()
	jsonResponse(w, 200, plugins)
}

// GET /api/v1/tenants
func (s *Server) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := s.Tenants.GetAll(r.Context())
	if err != nil {
		jsonError(w, 500, "failed to list tenants")
		return
	}
	jsonResponse(w, 200, tenants)
}

// GET /api/v1/healthz
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, map[string]string{"status": "ok", "service": "opc-agent-saas"})
}

// POST /api/v1/usage/record
func (s *Server) HandleRecordUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	var req struct {
		TenantID     string `json:"tenant_id"`
		AgentName    string `json:"agent_name"`
		ModelName    string `json:"model_name"`
		InputTokens  int    `json:"input_tokens"`
		OutputTokens int    `json:"output_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request")
		return
	}
	if req.TenantID == "" {
		jsonError(w, 400, "tenant_id required")
		return
	}

	cost := CostForTokens(req.ModelName, req.InputTokens, req.OutputTokens)
	_, err := s.Tenants.db.ExecContext(r.Context(),
		`INSERT INTO token_usage (tenant_id, agent_name, model_name, input_tokens, output_tokens, cost, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		req.TenantID, req.AgentName, req.ModelName, req.InputTokens, req.OutputTokens, cost)
	if err != nil {
		jsonError(w, 500, "failed to record usage")
		return
	}

	jsonResponse(w, 200, map[string]interface{}{
		"recorded":      true,
		"cost":          cost,
		"input_tokens":  req.InputTokens,
		"output_tokens": req.OutputTokens,
	})
}

// GET /api/v1/usage/summary?tenant_id=xxx&period=month
func (s *Server) HandleUsageSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		jsonError(w, 400, "tenant_id required")
		return
	}

	var totalTokens int64
	var totalCost float64
	var callCount int

	err := s.Tenants.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(SUM(input_tokens+output_tokens),0), COALESCE(SUM(cost),0), COUNT(*)
		 FROM token_usage WHERE tenant_id=$1 AND created_at > NOW() - INTERVAL '30 days'`,
		tenantID).Scan(&totalTokens, &totalCost, &callCount)
	if err != nil {
		jsonError(w, 500, "query failed")
		return
	}

	// Quota limits by plan
	planLimits := map[string]int64{"free": 100000, "pro": 1000000, "enterprise": 10000000}
	limit := planLimits["free"]
	if t, _ := s.Tenants.GetByID(r.Context(), tenantID); t != nil {
		if l, ok := planLimits[t.Plan]; ok {
			limit = l
		}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"tenant_id":      tenantID,
		"total_tokens":   totalTokens,
		"total_cost":     totalCost,
		"call_count":     callCount,
		"quota_limit":    limit,
		"quota_used_pct": float64(totalTokens) / float64(limit) * 100,
	})
}

// POST /api/v1/agents/{name}/execute  (simplified - uses path hack)
func (s *Server) HandleExecuteAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	tenantID := r.Context().Value(CtxTenantID)
	if tenantID == nil {
		jsonError(w, 401, "tenant not identified")
		return
	}

	var req struct {
		Agent string `json:"agent"`
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request")
		return
	}

	plugin, ok := s.Loader.Get(req.Agent)
	if !ok {
		jsonError(w, 404, "agent not found")
		return
	}

	// Execute plugin directly (non-LLM plugins run inline)
	result, err := plugin.Execute(r.Context(), req.Input, map[string]interface{}{})
	if err != nil {
		jsonError(w, 500, "execution failed: "+err.Error())
		return
	}

	// Record usage
	cost := CostForTokens(result.TokenUsage.ModelName, result.TokenUsage.InputTokens, result.TokenUsage.OutputTokens)
	s.Tenants.db.ExecContext(r.Context(),
		`INSERT INTO token_usage (tenant_id, agent_name, model_name, input_tokens, output_tokens, cost, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		tenantID, req.Agent, result.TokenUsage.ModelName,
		result.TokenUsage.InputTokens, result.TokenUsage.OutputTokens, cost)

	jsonResponse(w, 200, map[string]interface{}{
		"result": result.Data,
		"tokens": result.TokenUsage,
		"cost":   cost,
	})
}
