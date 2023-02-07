package vault

import (
	"database/sql"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	repository := Repository{db: db}

	return &repository
}

func (r *Repository) Create(ctx context.Context, vault *VaultModel) error {
	_, err := r.db.ExecContext(
		ctx,
		`insert into vaults (id, user_id, vault, version, s3) values ($1, $2::uuid, $3::bytea, $4, $5);`,
		vault.ID,
		vault.UserID,
		vault.Vault,
		vault.Version,
		vault.S3,
	)

	if err != nil {
		return err
	}

	return err
}

func (r *Repository) checkIsExists(ctx context.Context, userID, id uuid.UUID) error {
	var check bool
	err := r.db.QueryRowContext(
		ctx,
		`select true from vaults where id = $1 and user_id = $2 and is_deleted = false;`,
		id,
		userID,
	).Scan(&check)

	if err == sql.ErrNoRows {
		return ErrVaultNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateVault(ctx context.Context, userID, id uuid.UUID, vault []byte, version int) (int, error) {
	err := r.checkIsExists(ctx, userID, id)

	if err != nil {
		return 0, err
	}

	var updatedVersion int

	err = r.db.QueryRowContext(
		ctx,
		`update vaults set vault = $2, version = version + 1 where id = $1 and version = $3 returning version;`,
		id,
		vault,
		version,
	).Scan(&updatedVersion)

	if err == sql.ErrNoRows {
		return 0, ErrVaultConflict
	}

	if err != nil {
		return 0, err
	}

	return updatedVersion, nil
}

func (r *Repository) Delete(ctx context.Context, userID, id uuid.UUID, version int) error {
	result, err := r.db.ExecContext(
		ctx,
		`update vaults set vault = null, is_deleted = true where id = $1 and version = $2 and user_id = $3;`,
		id,
		version,
		userID,
	)

	if err != nil {
		return err
	}

	countAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if countAffected == 0 {
		return ErrVaultConflict
	}

	return nil
}

func (r *Repository) LoadUpdatedVaults(ctx context.Context, userID uuid.UUID, dto []VaultVersionDTO) ([]VaultModel, error) {
	mapVaultsVersions := make(map[string]int)
	vaultIDs := make([]uuid.UUID, len(dto))
	for i := 0; i < len(dto); i++ {
		vaultIDs[i] = dto[i].ID
		mapVaultsVersions[dto[i].ID.String()] = dto[i].Version
	}

	rows, err := r.db.QueryContext(
		ctx,
		`select id, version from vaults where user_id = $1 and (id = any($2) or is_deleted = false);`,
		userID,
		vaultIDs,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	mapActionVersion := make(map[string]int)

	for rows.Next() {
		var id string
		var version int

		if err := rows.Scan(&id, &version); err != nil {
			return nil, err
		}

		mapActionVersion[id] = version
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	needUpdatedIds := make([]string, 0)

	for id, version := range mapActionVersion {
		oldVersion, ok := mapVaultsVersions[id]
		if !ok || version > oldVersion {
			needUpdatedIds = append(needUpdatedIds, id)
		}

	}

	vaultRows, err := r.db.QueryContext(
		ctx,
		`select id, user_id, vault, version, is_deleted, s3 from vaults where user_id = $1 and id = any($2);`,
		userID,
		needUpdatedIds,
	)

	if err != nil {
		return nil, err
	}
	defer vaultRows.Close()

	vaults := make([]VaultModel, 0)

	for vaultRows.Next() {
		vault := VaultModel{}

		if err := vaultRows.Scan(
			&vault.ID,
			&vault.UserID,
			&vault.Vault,
			&vault.Version,
			&vault.IsDeleted,
			&vault.S3,
		); err != nil {
			return nil, err
		}

		vaults = append(vaults, vault)
	}

	if vaultRows.Err() != nil {
		return nil, vaultRows.Err()
	}

	return vaults, nil
}
