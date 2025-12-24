package command

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type PayOrderCommand struct {
	MessageID string `json:"message_id"`
	OrderID   string `json:"order_id"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
}

type Handler struct {
	DB          *sqlx.DB
	AccountRepo AccountRepository
	InboxRepo   InboxRepository
	TxRepo      TxRepository
	OutboxRepo  OutboxRepository
}

func (h *Handler) Handle(ctx context.Context, cmd PayOrderCommand) error {
	tx, err := h.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if h.InboxRepo.Exists(ctx, tx.Tx, cmd.MessageID) {
		return nil
	}
	if err := h.InboxRepo.Save(ctx, tx.Tx, cmd.MessageID); err != nil {
		return err
	}

	if h.TxRepo.ExistsByOrderID(ctx, tx.Tx, cmd.OrderID) {
		return nil
	}

	if err := h.AccountRepo.WithdrawInTx(ctx, tx.Tx, cmd.UserID, cmd.Amount); err != nil {
		h.OutboxRepo.SaveFailed(ctx, tx.Tx, cmd.OrderID, cmd.UserID, err.Error())
		return tx.Commit()
	}

	if err := h.TxRepo.Save(ctx, tx.Tx, cmd.OrderID, cmd.UserID, cmd.Amount); err != nil {
		return err
	}
	if err := h.OutboxRepo.SaveSuccess(ctx, tx.Tx, cmd.OrderID, cmd.UserID); err != nil {
		return err
	}

	return tx.Commit()
}
