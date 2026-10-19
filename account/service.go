package account

import (
	"context"
	"errors"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	PostAccount(ctx context.Context, name, password string) (*Account, error)
	Login(ctx context.Context, name, password string) (*Account, error)
	BootstrapOperator(ctx context.Context, name, password string) error
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type Account struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	PasswordHash string `json:"-"`
}

type accountService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &accountService{r}
}

func (s *accountService) PostAccount(ctx context.Context, name, password string) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	a := &Account{
		Name:         name,
		ID:           ksuid.New().String(),
		Role:         "CUSTOMER",
		PasswordHash: string(hash),
	}
	if err := s.repository.PutAccount(ctx, *a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *accountService) Login(ctx context.Context, name, password string) (*Account, error) {
	a, err := s.repository.GetAccountByName(ctx, name)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return a, nil
}

func (s *accountService) BootstrapOperator(ctx context.Context, name, password string) error {
	if name == "" || password == "" {
		return nil
	}
	if _, err := s.repository.GetAccountByName(ctx, name); err == nil {
		return nil
	}
	_, err := s.PostAccount(ctx, name, password)
	if err != nil {
		return err
	}
	a, err := s.repository.GetAccountByName(ctx, name)
	if err != nil {
		return err
	}
	a.Role = "OPERATOR"
	return s.repository.UpdateRole(ctx, a.ID, a.Role)
}

func (s *accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	a, err := s.repository.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return a, nil
}
func (s *accountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	listAccts, err := s.repository.ListAccounts(ctx, skip, take)

	if err != nil {
		return nil, err
	}
	return listAccts, nil
}
