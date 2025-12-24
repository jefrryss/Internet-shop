package query

import (
	"context"

	"payment-service/internal/domain/account"
)

type Service struct {
	Repo account.Repository
}

func (s *Service) Balance(ctx context.Context, userID int64) (int64, error) {
	acc, err := s.Repo.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}
