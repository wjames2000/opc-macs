package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wjames2000/opc-macs/internal/saas"
)

type MarketplaceToolSet struct {
	store     *saas.MarketplaceStore
	lifecycle *saas.OrderLifecycle
	wallet    *saas.WalletStore
}

func NewMarketplaceToolSet(store *saas.MarketplaceStore, lifecycle *saas.OrderLifecycle, wallet *saas.WalletStore) *MarketplaceToolSet {
	return &MarketplaceToolSet{
		store:     store,
		lifecycle: lifecycle,
		wallet:    wallet,
	}
}

func newInputSchema(s string) any {
	var v any
	json.Unmarshal([]byte(s), &v)
	return v
}

func (m *MarketplaceToolSet) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "listOpenTasks",
			Description: "列出所有开放的内容营销任务",
			InputSchema: newInputSchema(`{"type":"object","properties":{"platform":{"type":"string"},"limit":{"type":"integer"},"offset":{"type":"integer"}}}`),
		},
		{
			Name:        "createTask",
			Description: "创建内容营销任务",
			InputSchema: newInputSchema(`{"type":"object","properties":{"title":{"type":"string"},"description":{"type":"string"},"platform":{"type":"string"},"content_type":{"type":"string"},"budget":{"type":"number"},"deadline":{"type":"string"},"requirements":{"type":"string"}},"required":["title","platform","budget"]}`),
		},
		{
			Name:        "applyTask",
			Description: "创作者申请接受任务",
			InputSchema: newInputSchema(`{"type":"object","properties":{"task_id":{"type":"string"}},"required":["task_id"]}`),
		},
		{
			Name:        "completeOrder",
			Description: "完成订单并触发结算",
			InputSchema: newInputSchema(`{"type":"object","properties":{"order_id":{"type":"string"},"tenant_id":{"type":"string"}},"required":["order_id","tenant_id"]}`),
		},
		{
			Name:        "getWalletBalance",
			Description: "查询创作者钱包余额",
			InputSchema: newInputSchema(`{"type":"object","properties":{"user_id":{"type":"string"},"tenant_id":{"type":"string"}},"required":["user_id","tenant_id"]}`),
		},
		{
			Name:        "myOrders",
			Description: "查看我的任务订单列表",
			InputSchema: newInputSchema(`{"type":"object","properties":{"status":{"type":"string"}}}`),
		},
	}
}

func (m *MarketplaceToolSet) ExecuteTool(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
	switch name {
	case "listOpenTasks":
		return m.execListOpenTasks(ctx, args)
	case "createTask":
		return m.execCreateTask(ctx, args)
	case "applyTask":
		return m.execApplyTask(ctx, args)
	case "completeOrder":
		return m.execCompleteOrder(ctx, args)
	case "getWalletBalance":
		return m.execGetWalletBalance(ctx, args)
	case "myOrders":
		return m.execMyOrders(ctx, args)
	default:
		return nil, fmt.Errorf("unknown marketplace tool: %s", name)
	}
}

func (m *MarketplaceToolSet) execListOpenTasks(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var params struct {
		Platform string `json:"platform"`
		Limit    int    `json:"limit"`
		Offset   int    `json:"offset"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	tasks, err := m.store.ListOpenTasks(ctx, params.Platform, "", params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (m *MarketplaceToolSet) execCreateTask(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var params struct {
		TenantID       string  `json:"tenant_id"`
		CreatorID      string  `json:"creator_id"`
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		Platform       string  `json:"platform"`
		ContentType    string  `json:"content_type"`
		SettlementType string  `json:"settlement_type"`
		Budget         float64 `json:"budget"`
		Deadline       string  `json:"deadline"`
		Requirements   string  `json:"requirements"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	deadline, err := time.Parse("2006-01-02", params.Deadline)
	if err != nil {
		deadline = time.Now().Add(7 * 24 * time.Hour)
	}
	if params.SettlementType == "" {
		params.SettlementType = "fixed"
	}
	task := &saas.MarketplaceTask{
		ID:             fmt.Sprintf("task_%d", time.Now().UnixNano()),
		TenantID:       params.TenantID,
		CreatorID:      params.CreatorID,
		Title:          params.Title,
		Description:    params.Description,
		Platform:       params.Platform,
		ContentType:    params.ContentType,
		SettlementType: params.SettlementType,
		Budget:         params.Budget,
		Requirements:   params.Requirements,
		Status:         "open",
		Deadline:       deadline,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := m.store.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (m *MarketplaceToolSet) execApplyTask(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var params struct {
		TaskID       string `json:"task_id"`
		CreatorID    string `json:"creator_id"`
		AdvertiserID string `json:"advertiser_id"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	order := &saas.TaskOrder{
		ID:           fmt.Sprintf("order_%d", time.Now().UnixNano()),
		TaskID:       params.TaskID,
		CreatorID:    params.CreatorID,
		AdvertiserID: params.AdvertiserID,
		Status:       "applied",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := m.store.CreateOrder(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (m *MarketplaceToolSet) execCompleteOrder(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var params struct {
		OrderID  string `json:"order_id"`
		TenantID string `json:"tenant_id"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if err := m.lifecycle.CompleteOrder(ctx, params.TenantID, params.OrderID); err != nil {
		return nil, err
	}
	return map[string]string{"status": "completed", "order_id": params.OrderID}, nil
}

func (m *MarketplaceToolSet) execGetWalletBalance(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var params struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	wallet, err := m.wallet.EnsureWallet(ctx, params.TenantID, params.UserID)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (m *MarketplaceToolSet) execMyOrders(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var params struct {
		CreatorID string `json:"creator_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	orders, err := m.store.ListOrdersByCreator(ctx, params.CreatorID, params.Status, 50, 0)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (m *MarketplaceToolSet) RegisterInto(server *Server) {
	for _, td := range m.ListTools() {
		t := td
		server.RegisterTool(t.Name, t.Description, t.InputSchema, func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return m.ExecuteTool(ctx, t.Name, args)
		})
	}
}
