package account

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Repository interface {
	Close()
	PutAccount(ctx context.Context, a Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccountByName(ctx context.Context, name string) (*Account, error)
	UpdateRole(ctx context.Context, id, role string) error
	ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}
func (r *postgresRepository) Ping() {
	r.db.Ping()
}
func (r *postgresRepository) PutAccount(ctx context.Context, a Account) error {

	_, err := r.db.ExecContext(ctx, "INSERT INTO accounts(id,name,role,password_hash) VALUES($1,$2,$3,$4)", a.ID, a.Name, a.Role, a.PasswordHash)
	return err

}

func (r *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name, role, password_hash FROM accounts WHERE id = $1", id)
	a := &Account{} // creating a variable a and intializing
	if err := row.Scan(&a.ID, &a.Name, &a.Role, &a.PasswordHash); err != nil {
		return nil, err
	}
	return a, nil

}

func (r *postgresRepository) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,name,role,password_hash FROM accounts ORDER BY id DESC OFFSET $1 LIMIT $2", skip, take)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}

	for rows.Next() {
		a := &Account{}
		if err = rows.Scan(&a.ID, &a.Name, &a.Role, &a.PasswordHash); err == nil {
			accounts = append(accounts, *a)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *postgresRepository) GetAccountByName(ctx context.Context, name string) (*Account, error) {
	a := &Account{}
	err := r.db.QueryRowContext(ctx, "SELECT id,name,role,password_hash FROM accounts WHERE name = $1", name).
		Scan(&a.ID, &a.Name, &a.Role, &a.PasswordHash)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *postgresRepository) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET role = $1 WHERE id = $2", role, id)
	return err
}
