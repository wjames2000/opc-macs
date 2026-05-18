package saas

import (
	"context"
	"fmt"
	"log"
)

type OrderLifecycle struct {
	marketplace *MarketplaceStore
	wallet      *WalletStore
}

func NewOrderLifecycle(marketplace *MarketplaceStore, wallet *WalletStore) *OrderLifecycle {
	return &OrderLifecycle{
		marketplace: marketplace,
		wallet:      wallet,
	}
}

func (l *OrderLifecycle) CompleteOrder(ctx context.Context, tenantID, orderID string) error {
	order, err := l.marketplace.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("lifecycle: get order: %w", err)
	}

	task, err := l.marketplace.GetTask(ctx, order.TaskID)
	if err != nil {
		return fmt.Errorf("lifecycle: get task: %w", err)
	}

	if err := l.marketplace.UpdateOrderStatus(ctx, orderID, "completed"); err != nil {
		return fmt.Errorf("lifecycle: update status: %w", err)
	}

	platformFee := 0.08
	if task.Budget > 10000 {
		platformFee = 0.06
	}
	if task.Budget > 100000 {
		platformFee = 0.04
	}
	creatorAmount := float64(task.Budget) * (1 - platformFee)

	creatorWallet, err := l.wallet.EnsureWallet(ctx, tenantID, order.CreatorID)
	if err != nil {
		return fmt.Errorf("lifecycle: wallet: %w", err)
	}

	if err := l.wallet.Settle(ctx, creatorWallet.ID, creatorAmount, orderID,
		fmt.Sprintf("任务结算: %s", task.Title)); err != nil {
		return fmt.Errorf("lifecycle: settle: %w", err)
	}

	log.Printf("[LIFECYCLE] 订单 %s 完成，创作者 %s 获得 ¥%.2f", orderID, order.CreatorID, creatorAmount)
	return l.marketplace.UpdateOrderSettlement(ctx, orderID, creatorAmount)
}

func (l *OrderLifecycle) CancelOrder(ctx context.Context, orderID, reason string) error {
	if err := l.marketplace.UpdateOrderStatus(ctx, orderID, "cancelled"); err != nil {
		return fmt.Errorf("lifecycle: cancel: %w", err)
	}
	if reason != "" {
		if err := l.marketplace.UpdateOrderNote(ctx, orderID, reason); err != nil {
			return fmt.Errorf("lifecycle: note: %w", err)
		}
	}
	log.Printf("[LIFECYCLE] 订单 %s 已取消: %s", orderID, reason)
	return nil
}
