package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/net/context"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	userRepository := Repository{db: db}

	return &userRepository
}

func (r *Repository) Create(ctx context.Context, user *UserModel) error {
	_, err := r.db.ExecContext(
		ctx,
		`insert into users (id, login, password) values ($1, $2, $3);`,
		user.ID,
		user.Login,
		user.password,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if !errors.As(err, &pgErr) {
			return err
		}

		if pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%q: %w", user.Login, ErrLoginAlreadyExist)
		}

		return err
	}

	return nil
}

func (r *Repository) FindByLogin(ctx context.Context, login string) (*UserModel, error) {
	row := r.db.QueryRowContext(
		ctx,
		`select id, login, password from users u where u.login = $1 limit 1;`,
		login,
	)

	if row.Err() != nil {
		return nil, row.Err()
	}

	userModel := UserModel{}

	err := row.Scan(&userModel.ID, &userModel.Login, &userModel.password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &userModel, nil
}
