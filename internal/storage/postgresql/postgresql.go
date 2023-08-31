package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"log/slog"

	seg "github.com/i-pankrat/avito-internship-assignment-2023/internal/data/segment"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/logger/sl"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func New(connectiongString string) (*Storage, error) {
	db, err := sql.Open("pgx", connectiongString)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS segments(
		id bigserial PRIMARY KEY,
		slug varchar(64) UNIQUE NOT NULL);

	CREATE TABLE IF NOT EXISTS user_segments(
		user_id bigint NOT NULL,
		slug varchar(64) REFERENCES segments(slug) ON DELETE CASCADE NOT NULL,
		expiration_date timestamp);
    `)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db}, nil
}

func (s *Storage) AddSegment(slug string) (int64, error) {
	var id int64
	err := s.db.QueryRow("INSERT INTO segments(slug) VALUES($1) RETURNING id", slug).Scan(&id)

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return 0, fmt.Errorf("%w: %s", storage.ErrSegmentExists, slug)
		}

		return 0, err
	}

	return id, nil
}

func (s *Storage) RemoveSegment(slug string) error {
	var id int64
	err := s.db.QueryRow("DELETE FROM segments WHERE slug=$1 RETURNING id", slug).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s", storage.ErrSegmentDoesNotExist, slug)
		}

		return err
	}

	return nil
}

func (s *Storage) ChangeUserSegments(id int64, segmentsToAdd []seg.Segment, segmentsToDelete []seg.Slug) error {
	if len(segmentsToAdd) == 0 && len(segmentsToDelete) == 0 {
		return nil
	}

	tx, err := s.db.Begin()

	if err != nil {
		return err
	}

	// If an error occurred, checks whether the error occurred
	// because the segment does not exist.
	throwSlugError := func(err error, slug seg.Slug) error {
		if err != nil {
			tx.Rollback()
			var e *pgconn.PgError
			if errors.As(err, &e) && e.Code == pgerrcode.ForeignKeyViolation {
				return fmt.Errorf("%w: %s", storage.ErrSegmentDoesNotExist, slug)
			}

			return err
		}

		return nil
	}

	// Remove segments
	if len(segmentsToDelete) > 0 {
		stmtDelete, err := tx.Prepare("DELETE FROM user_segments WHERE user_id=$1 AND slug=$2")

		if err != nil {
			tx.Rollback()
			return err
		}

		defer stmtDelete.Close()

		for _, slug := range segmentsToDelete {
			_, err := stmtDelete.Exec(id, slug)
			if err != nil {
				return throwSlugError(err, slug)
			}
		}
	}

	// Add segments
	if len(segmentsToAdd) > 0 {
		stmtAddWithoutDate, err := tx.Prepare("INSERT INTO user_segments(user_id, slug) VALUES($1, $2);")

		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmtAddWithoutDate.Close()

		stmtAddWithDate, err := tx.Prepare("INSERT INTO user_segments(user_id, slug, expiration_date) VALUES($1, $2, $3);")

		if err != nil {
			tx.Rollback()
			return err
		}

		defer stmtAddWithDate.Close()

		for _, segment := range segmentsToAdd {

			var count int
			err := tx.QueryRow("SELECT COUNT(*) FROM user_segments WHERE user_id=$1 AND slug=$2;",
				id, segment.Slug).Scan(&count)

			if err != nil {
				return throwSlugError(err, segment.Slug)
			}

			if count > 0 {
				tx.Rollback()
				return fmt.Errorf("%w: %d-%s", storage.ErrUserHasAlreadyAddedToSegment, id, segment.Slug)
			}

			if segment.ExpirationDate.IsZero() {
				_, err = stmtAddWithoutDate.Exec(id, segment.Slug)

			} else {
				_, err = stmtAddWithDate.Exec(id, segment.Slug, segment.ExpirationDate)

			}

			if err != nil {
				return throwSlugError(err, segment.Slug)
			}
		}
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetUserSegments(user_id int64) ([]seg.Slug, error) {
	rows, err := s.db.Query("SELECT slug FROM user_segments WHERE user_id=$1", user_id)
	if err != nil {
		return nil, err
	}

	var segment seg.Slug
	segments := make([]seg.Slug, 0, 2)
	for rows.Next() {
		err = rows.Scan(&segment)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

func (s *Storage) StartTTLChecker(log *slog.Logger, secondsTimer int) {
	ticker := time.NewTicker(time.Duration(secondsTimer) * time.Second)

	log.Info(fmt.Sprintf("start ttl checker with %d-second timer", secondsTimer))

	for range ticker.C {
		_, err := s.db.Exec("DELETE FROM user_segments WHERE " +
			"expiration_date IS NOT NULL AND expiration_date <= NOW();")

		if err != nil {
			log.Error("failed to delete users whose time in the segment has expired", sl.Err(err))
		} else {
			log.Info("TTL Checker has deleted users whose time in the segment has expired")
		}
	}
}
