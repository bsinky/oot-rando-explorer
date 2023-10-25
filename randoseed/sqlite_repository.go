package randoseed

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS seeds(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		upload_time DATETIME NOT NULL,
		seed TEXT NOT NULL,
		version TEXT NOT NULL,
		file_hash TEXT NOT NULL UNIQUE,
		logic TEXT NOT NULL,
		shopsanity TEXT NOT NULL,
		tokensanity TEXT NOT NULL,
		scrubsanity TEXT NOT NULL,
		raw_settings TEXT NOT NULL
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *SQLiteRepository) Create(dbseed DBSeed, tx *sql.Tx) (*DBSeed, error) {
	var execFunc func(command string, args ...any) (sql.Result, error)

	if tx != nil {
		execFunc = tx.Exec
	} else {
		execFunc = r.db.Exec
	}

	res, err := execFunc(`INSERT INTO seeds(
		upload_time,
		seed,
		version,
		file_hash,
		logic,
		shopsanity,
		tokensanity,
		scrubsanity,
		raw_settings)
		values(?,?,?,?,?,?,?,?,?)`,
		dbseed.UploadTime,
		dbseed.Seed,
		dbseed.Version,
		dbseed.FileHash,
		dbseed.Logic,
		dbseed.Shopsanity,
		dbseed.Tokensanity,
		dbseed.Scrubsanity,
		dbseed.RawSettings)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	dbseed.Id = id

	return &dbseed, nil
}

func (r *SQLiteRepository) MostRecent(n int) ([]DBSeed, error) {
	rows, err := r.db.Query(`SELECT
		id,
		upload_time,
		seed,
		version,
		file_hash,
		logic,
		shopsanity,
		tokensanity,
		scrubsanity,
		raw_settings
	FROM seeds
	ORDER BY id DESC
	LIMIT ?`, n)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recent []DBSeed
	for rows.Next() {
		var seed DBSeed
		if err := rows.Scan(&seed.Id,
			&seed.UploadTime,
			&seed.Seed,
			&seed.Version,
			&seed.FileHash,
			&seed.Logic,
			&seed.Shopsanity,
			&seed.Tokensanity,
			&seed.Scrubsanity,
			&seed.RawSettings); err != nil {
			return nil, err
		}
		recent = append(recent, seed)
	}
	return recent, nil
}

func (r *SQLiteRepository) GetByFileHash(fileHash string) (*DBSeed, error) {
	row := r.db.QueryRow(`SELECT
		id,
		upload_time,
		seed,
		version,
		file_hash,
		logic,
		shopsanity,
		tokensanity,
		scrubsanity,
		raw_settings
	FROM seeds WHERE file_hash = ?`, fileHash)
	var seed DBSeed
	if err := row.Scan(&seed.Id,
		&seed.UploadTime,
		&seed.Seed,
		&seed.Version,
		&seed.FileHash,
		&seed.Logic,
		&seed.Shopsanity,
		&seed.Tokensanity,
		&seed.Scrubsanity,
		&seed.RawSettings); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &seed, nil
}
