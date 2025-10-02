package postgres

import (
	"github.com/deimossy/order-processing-system/internal/user/usecase"
)

type PgRepos struct {
	refreshToken *PgRefreshTokenRepo
	user         *PgUserRepo
}

func NewPgRepos(userRepo *PgUserRepo, refreshTokenRepo *PgRefreshTokenRepo) *PgRepos {
	return &PgRepos{
		user:         userRepo,
		refreshToken: refreshTokenRepo,
	}
}

func (r *PgRepos) UserRepo() usecase.UserRepo {
	return r.user
}

func (r *PgRepos) RefreshTokenRepo() usecase.RefreshTokenRepo {
	return r.refreshToken
}
