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

	if _, err := r.db.Exec(query); err != nil {
		return err
	}

	query = `
	CREATE TABLE IF NOT EXISTS seed_ranks(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		seed_id INTEGER,
		difficulty INTEGER,
		fun	INTEGER,
		FOREIGN KEY(seed_id) REFERENCES seeds(id)
	);
	`

	if _, err := r.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// TODO: should probably pass pointers into these CRUD methods instead of passing by value
func (r *SQLiteRepository) CreateSeed(dbseed DBSeed, tx *sql.Tx) (*DBSeed, error) {
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

func (r *SQLiteRepository) CreateRank(rank SeedRank, tx *sql.Tx) (*SeedRank, error) {
	var execFunc func(command string, args ...any) (sql.Result, error)

	if tx != nil {
		execFunc = tx.Exec
	} else {
		execFunc = r.db.Exec
	}

	res, err := execFunc(`INSERT INTO seed_ranks(
		seed_id,
		difficulty,
		fun)
		values(?,?,?)`,
		rank.DBSeedId,
		rank.Difficulty,
		rank.Fun)
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
	rank.Id = id

	return &rank, nil
}

func (r *SQLiteRepository) GetUserRank(fileHash string, userID string) (*SeedRank, error) {
	row := r.db.QueryRow(`SELECT
		seed_ranks.id,
		seed_ranks.user_id,
		seed_ranks.seed_id,
		seed_ranks.difficulty,
		seed_ranks.fun
	FROM seed_ranks
	INNER JOIN seeds ON seed_ranks.seed_id = seeds.id
	WHERE seeds.file_hash = ?
	AND seed_ranks.user_id = ?`, fileHash, userID)
	var rank SeedRank
	if err := row.Scan(&rank.Id,
		&rank.UserID,
		&rank.DBSeedId,
		&rank.Difficulty,
		&rank.Fun); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rank, nil
}

func (r *SQLiteRepository) GetAverageRank(fileHash string) (*AvgSeedRank, error) {
	row := r.db.QueryRow(`SELECT
	    seeds.id,
		COUNT(seed_ranks.id),
		AVG(seed_ranks.difficulty),
		AVG(seed_ranks.fun)
	FROM seed_ranks
	INNER JOIN seeds ON seed_ranks.seed_id = seeds.id
	WHERE seeds.file_hash = ?
	GROUP BY seeds.id`, fileHash)
	var avgRank AvgSeedRank
	if err := row.Scan(&avgRank.DBSeedId,
		&avgRank.TotalVotes,
		&avgRank.Difficulty,
		&avgRank.Fun); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &avgRank, nil
}
