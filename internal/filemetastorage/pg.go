package filemetastorage

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"karma8"
	"strconv"
	"strings"
)

type dbFilePart struct {
	StorageURL    string `db:"storage_url"`
	Path          string `db:"file_path"`
	ContentLength int64  `db:"content_length"`
}

func (m dbFilePart) Value() (driver.Value, error) {
	return fmt.Sprintf("(%s,%s,%d)", m.StorageURL, m.Path, m.ContentLength), nil
}

func (m *dbFilePart) Scan(src interface{}) error {
	rawValue := string(src.([]byte))
	rawValue = rawValue[1 : len(rawValue)-1]
	splitValue := strings.Split(rawValue, ",")
	if len(splitValue) != 3 {
		return fmt.Errorf("unexpected split len for '%s': %d", rawValue, len(splitValue))
	}
	m.StorageURL = splitValue[0]
	m.Path = splitValue[1]

	contentLength, err := strconv.ParseInt(splitValue[2], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse ContentLength: %w", err)
	}

	m.ContentLength = contentLength

	return nil
}

func convertFilePart(part *karma8.FilePart) *dbFilePart {
	return (*dbFilePart)(part)
}

func convertFileParts(parts []*karma8.FilePart) []*dbFilePart {
	result := make([]*dbFilePart, 0, len(parts))
	for _, part := range parts {
		result = append(result, convertFilePart(part))
	}
	return result
}

func convertDBFilePart(part *dbFilePart) *karma8.FilePart {
	return (*karma8.FilePart)(part)
}

func convertDBFileParts(parts []dbFilePart) []*karma8.FilePart {
	result := make([]*karma8.FilePart, 0, len(parts))
	for i := range parts {
		result = append(result, convertDBFilePart(&parts[i]))
	}
	return result
}

type pgStorage struct {
	db *sqlx.DB

	logger *zap.Logger
}

func (m *pgStorage) Transact(ctx context.Context, atomic func(tx *sqlx.Tx) error) (err error) {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = atomic(tx)
	return err
}

func (m *pgStorage) PutProcessingFileMeta(ctx context.Context, meta *karma8.FileMeta) error {
	_, err := m.db.ExecContext(
		ctx,
		`INSERT INTO processing_file (name, parts, content_length) VALUES ($1, $2, $3)`,
		meta.Name,
		pq.Array(convertFileParts(meta.Parts)),
		meta.ContentLength,
	)
	if err != nil {
		m.logger.Error("can't put processing file meta", zap.Error(err))
		return err
	}
	return nil
}

func (m *pgStorage) CompleteFileMeta(ctx context.Context, filename string) error {
	err := m.Transact(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`
INSERT INTO file
SELECT * FROM processing_file
WHERE name = $1;
`,
			filename,
		)
		if err != nil {
			m.logger.Error("can't move file meta to completed", zap.Error(err))
			return err
		}

		_, err = tx.ExecContext(
			ctx,
			`
DELETE FROM processing_file
WHERE name = $1;
`,
			filename,
		)
		if err != nil {
			m.logger.Error("can't delete processing meta", zap.Error(err))
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *pgStorage) GetFileMeta(ctx context.Context, filename string) (*karma8.FileMeta, error) {
	var name string
	var parts []dbFilePart
	var contentLength int64
	err := m.db.QueryRowContext(
		ctx,
		`
SELECT * FROM file
WHERE name = $1`,
		filename,
	).Scan(&name, pq.Array(&parts), &contentLength)
	if err != nil {
		return nil, err
	}

	return &karma8.FileMeta{
		Name:          name,
		Parts:         convertDBFileParts(parts),
		ContentLength: contentLength,
	}, nil
}

func NewPGStorage(db *sqlx.DB, logger *zap.Logger) karma8.FileMetaStorage {
	return &pgStorage{
		db:     db,
		logger: logger,
	}
}
