package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteHitStore struct {
	db *sql.DB
}

func NewSQLiteHitStore(path string) (*SQLiteHitStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer
	if err := initSQLiteSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteHitStore{db: db}, nil
}

func initSQLiteSchema(db *sql.DB) error {
	stmts := []string{
		"PRAGMA journal_mode = WAL",
		`CREATE TABLE IF NOT EXISTS hits (
			ts            TEXT NOT NULL,
			path          TEXT NOT NULL,
			method        TEXT NOT NULL,
			caller_id     TEXT,
			user_agent    TEXT,
			days_left     INTEGER,
			endpoint_key  TEXT NOT NULL,
			enforced      INTEGER NOT NULL DEFAULT 0
		)`,
		"CREATE INDEX IF NOT EXISTS idx_endpoint ON hits(endpoint_key)",
		"CREATE INDEX IF NOT EXISTS idx_ts ON hits(ts)",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteHitStore) Record(hit HitEvent) error {
	enforced := 0
	if hit.Enforced {
		enforced = 1
	}
	_, err := s.db.ExecContext(context.Background(),
		"INSERT INTO hits VALUES (?,?,?,?,?,?,?,?)",
		hit.Ts.UTC().Format(time.RFC3339Nano),
		hit.Path,
		hit.Method,
		nullStr(hit.CallerID),
		nullStr(hit.UserAgent),
		nullInt(hit.DaysUntilSunset),
		hit.EndpointKey,
		enforced,
	)
	return err
}

func (s *SQLiteHitStore) RecentHits(query HitQuery) ([]HitEvent, error) {
	sinceDays := query.SinceDays
	if sinceDays == 0 {
		sinceDays = 30
	}
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}
	since := time.Now().UTC().AddDate(0, 0, -sinceDays).Format(time.RFC3339Nano)

	q := `SELECT ts, path, method, caller_id, user_agent, days_left, endpoint_key, enforced
	      FROM hits WHERE ts >= ? `
	args := []interface{}{since}

	if query.EndpointKey != nil {
		q += "AND endpoint_key = ? "
		args = append(args, *query.EndpointKey)
	}
	if query.CallerID != nil {
		q += "AND caller_id = ? "
		args = append(args, *query.CallerID)
	}

	q += "ORDER BY ts DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hits []HitEvent
	for rows.Next() {
		var (
			tsStr       string
			path        string
			method      string
			callerID    sql.NullString
			userAgent   sql.NullString
			daysLeft    sql.NullInt64
			endpointKey string
			enforced    int
		)
		if err := rows.Scan(&tsStr, &path, &method, &callerID, &userAgent, &daysLeft, &endpointKey, &enforced); err != nil {
			return nil, err
		}
		ts, _ := time.Parse(time.RFC3339Nano, tsStr)
		hit := HitEvent{
			Ts:          ts,
			Path:        path,
			Method:      method,
			EndpointKey: endpointKey,
			Enforced:    enforced == 1,
		}
		if callerID.Valid {
			hit.CallerID = &callerID.String
		}
		if userAgent.Valid {
			hit.UserAgent = &userAgent.String
		}
		if daysLeft.Valid {
			d := int(daysLeft.Int64)
			hit.DaysUntilSunset = &d
		}
		hits = append(hits, hit)
	}
	return hits, rows.Err()
}

func (s *SQLiteHitStore) EndpointSummaries(sinceDays int) ([]EndpointSummary, error) {
	if sinceDays == 0 {
		sinceDays = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -sinceDays).Format(time.RFC3339Nano)
	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx, `
		SELECT endpoint_key, COUNT(*) AS total_hits,
		       COUNT(DISTINCT caller_id) AS unique_callers, MAX(ts) AS last_seen
		FROM hits WHERE ts >= ?
		GROUP BY endpoint_key ORDER BY total_hits DESC
	`, since)
	if err != nil {
		return nil, err
	}

	// Collect all outer rows before closing — nested queries below need the connection.
	type epRow struct {
		epKey         string
		totalHits     int
		uniqueCallers int
		lastSeenStr   sql.NullString
	}
	var collected []epRow
	for rows.Next() {
		var r epRow
		if err := rows.Scan(&r.epKey, &r.totalHits, &r.uniqueCallers, &r.lastSeenStr); err != nil {
			rows.Close()
			return nil, err
		}
		collected = append(collected, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var summaries []EndpointSummary
	for _, r := range collected {
		callerRows, err := s.db.QueryContext(ctx, `
			SELECT caller_id, COUNT(*) AS cnt FROM hits
			WHERE endpoint_key = ? AND ts >= ? AND caller_id IS NOT NULL
			GROUP BY caller_id ORDER BY cnt DESC LIMIT 10
		`, r.epKey, since)
		if err != nil {
			return nil, err
		}
		var topCallers []TopCaller
		for callerRows.Next() {
			var tc TopCaller
			if err := callerRows.Scan(&tc.CallerID, &tc.Count); err != nil {
				callerRows.Close()
				return nil, err
			}
			topCallers = append(topCallers, tc)
		}
		callerRows.Close()

		sum := EndpointSummary{
			EndpointKey:   r.epKey,
			TotalHits:     r.totalHits,
			UniqueCallers: r.uniqueCallers,
			TopCallers:    topCallers,
		}
		if r.lastSeenStr.Valid {
			t, _ := time.Parse(time.RFC3339Nano, r.lastSeenStr.String)
			sum.LastSeen = &t
		}
		summaries = append(summaries, sum)
	}
	return summaries, nil
}

func (s *SQLiteHitStore) TotalSummary(sinceDays int) (TotalSummary, error) {
	if sinceDays == 0 {
		sinceDays = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -sinceDays).Format(time.RFC3339Nano)
	ctx := context.Background()

	var total TotalSummary
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT caller_id), COUNT(DISTINCT endpoint_key)
		FROM hits WHERE ts >= ?
	`, since).Scan(&total.TotalHits, &total.UniqueCallers, &total.EndpointsWithTraffic); err != nil {
		return TotalSummary{}, err
	}

	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT endpoint_key)
		FROM hits WHERE ts >= ? AND days_left < 0
	`, since).Scan(&total.PastSunsetWithTraffic); err != nil {
		return TotalSummary{}, err
	}

	return total, nil
}

func (s *SQLiteHitStore) Close() error {
	return s.db.Close()
}

func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}
