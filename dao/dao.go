// Package dao provides data access operations for the FTS5 search database.
//
// Error handling strategy:
//   - Fatal (log.Fatalf): initialization failures that prevent the application from running
//     (DB open, ping, table creation). These are unrecoverable.
//   - Log-and-continue (log.Printf): per-record failures (insert errors, query errors, scan errors).
//     These are recoverable and should not crash the application.
package dao

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chrwhy/simple-search/config"
	"github.com/chrwhy/simple-search/model"
	"github.com/mattn/go-sqlite3"
)

// registerOnce ensures the sqlite3_simple driver is registered exactly once,
// preventing panic from sql.Register's duplicate-name check.
var registerOnce sync.Once

// allowedTables whitelists table names that can appear in dynamically-built SQL.
// This prevents SQL injection when table names originate from config/env.
var allowedTables = map[string]bool{
	"docs_name":    true,
	"docs_content": true,
}

// allowedFields whitelists column names for FTS5 queries.
var allowedFields = map[string]bool{
	"name":    true,
	"content": true,
}

// DBExecutor abstracts *sql.DB and *sql.Tx so InsertDoc works in both contexts.
//
//go:generate go run github.com/golang/mock/mockgen -destination=mock_dao_test.go -package=dao -source=dao.go DBExecutor
type DBExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// InitDB registers the sqlite3_simple driver with the libsimple extension and opens the database.
// Calls log.Fatalf on failure since the application cannot run without a database connection.
// Safe to call multiple times — driver registration is guarded by sync.Once.
func InitDB(cfg *config.Config) *sql.DB {
	registerOnce.Do(func() {
		sql.Register("sqlite3_simple",
			&sqlite3.SQLiteDriver{
				Extensions: []string{
					cfg.LibPath,
				},
			})
	})

	db, err := sql.Open("sqlite3_simple", cfg.DBPath)
	if err != nil {
		log.Fatalf("open error: %v", err)
	}

	// SQLite supports only one concurrent writer; cap connections accordingly.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	err = db.Ping()
	if err != nil {
		log.Fatal("ping error: ", err)
	}
	return db
}

// CreateTable creates the FTS5 virtual tables if they don't already exist.
// Table names are validated against the allowlist before use.
func CreateTable(cfg *config.Config, db *sql.DB) {
	if err := validateTable(cfg.NameTable); err != nil {
		log.Fatalf("invalid name table: %v", err)
	}
	if err := validateTable(cfg.ContentTable); err != nil {
		log.Fatalf("invalid content table: %v", err)
	}

	nameSQL := fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(
		fid, name, cate, ctime,
		tokenize = '%s'
	);`, cfg.NameTable, cfg.NameTokenizer)
	contentSQL := fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(
		fid, content, cate, ctime,
		tokenize = '%s'
	);`, cfg.ContentTable, cfg.ContentTokenizer)

	if _, err := db.Exec(nameSQL); err != nil {
		log.Fatalf("create %s: %v", cfg.NameTable, err)
	}
	if _, err := db.Exec(contentSQL); err != nil {
		log.Fatalf("create %s: %v", cfg.ContentTable, err)
	}
	log.Printf("Tables '%s' (pinyin ON) and '%s' (pinyin OFF) ready", cfg.NameTable, cfg.ContentTable)
}

// validateTable checks that a table name is in the allowlist.
func validateTable(name string) error {
	if !allowedTables[name] {
		return fmt.Errorf("table %q not in allowlist; update allowedTables if this is intentional", name)
	}
	return nil
}

// validateField checks that a field name is in the allowlist.
func validateField(field string) error {
	if !allowedFields[field] {
		return fmt.Errorf("field %q not in allowlist; update allowedFields if this is intentional", field)
	}
	return nil
}

// InsertDoc inserts a document into both the name and content FTS5 tables.
// Accepts DBExecutor (*sql.DB or *sql.Tx) so callers can batch in a transaction.
// Logs errors per-insert but does not abort, so partial failures are tolerated.
func InsertDoc(cfg *config.Config, exec DBExecutor, doc model.Doc) {
	_, err := exec.Exec(
		fmt.Sprintf(`INSERT INTO %s(fid, name, cate, ctime) VALUES (?, ?, ?, ?)`, cfg.NameTable),
		doc.FID, doc.Name, doc.Cate, doc.CTime,
	)
	if err != nil {
		log.Printf("insert %s error: %v", cfg.NameTable, err)
	}
	_, err = exec.Exec(
		fmt.Sprintf(`INSERT INTO %s(fid, content, cate, ctime) VALUES (?, ?, ?, ?)`, cfg.ContentTable),
		doc.FID, doc.Content, doc.Cate, doc.CTime,
	)
	if err != nil {
		log.Printf("insert %s error: %v", cfg.ContentTable, err)
	}
}

// sanitizeMatch escapes single quotes in FTS5 MATCH clauses to prevent SQL injection.
func sanitizeMatch(clause string) string {
	return strings.ReplaceAll(clause, "'", "''")
}

// QueryFTS performs a highlighted FTS5 MATCH query. Table and field are validated
// against the allowlist to prevent SQL injection.
func QueryFTS(db *sql.DB, table, field, matchClause string) {
	if err := validateTable(table); err != nil {
		log.Printf("query error: %v", err)
		return
	}
	if err := validateField(field); err != nil {
		log.Printf("query error: %v", err)
		return
	}
	safe := sanitizeMatch(matchClause)
	sqlStr := fmt.Sprintf(
		"SELECT fid, simple_highlight(%s, 1, '[', ']'), cate, ctime FROM %s WHERE %s MATCH '%s'",
		table, table, field, safe,
	)
	execQuery(db, sqlStr, func(rows *sql.Rows) (string, error) {
		var fid int
		var text, cate, ctime string
		err := rows.Scan(&fid, &text, &cate, &ctime)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(" fid=%d cate=%s ctime=%s\n    %s: %s", fid, cate, ctime, field, text), nil
	})
}

func QueryByName(cfg *config.Config, db *sql.DB, matchClause string) {
	QueryFTS(db, cfg.NameTable, "name", matchClause)
}

func QueryByContent(cfg *config.Config, db *sql.DB, matchClause string) {
	QueryFTS(db, cfg.ContentTable, "content", matchClause)
}

func execQuery(db *sql.DB, sqlStr string, scanFn func(rows *sql.Rows) (string, error)) {
	log.Printf("SQL: %s", sqlStr)
	t0 := time.Now()
	rows, err := db.Query(sqlStr)
	log.Printf("Query cost: %v", time.Since(t0))
	if err != nil {
		log.Printf("query error: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		output, err := scanFn(rows)
		if err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		fmt.Printf("  [%d]%s\n", i+1, output)
		i++
	}
	if i == 0 {
		fmt.Println("  (无匹配结果)")
	} else {
		fmt.Printf("  共 %d 条结果\n", i)
	}
}

// ExecQuery executes arbitrary SQL and prints results. Intended for debugging/experimentation.
// Note: this function does NOT sanitize input — use only in trusted contexts.
// NULL values are rendered as "<nil>" rather than causing scan errors.
func ExecQuery(db *sql.DB, sqlStr string) {
	log.Printf("SQL: %s", sqlStr)
	t0 := time.Now()
	rows, err := db.Query(sqlStr)
	log.Printf("Query cost: %v", time.Since(t0))
	if err != nil {
		log.Printf("query error: %v", err)
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	i := 0
	for rows.Next() {
		ptrs := make([]interface{}, len(cols))
		vals := make([]sql.NullString, len(cols))
		for j := range vals {
			ptrs[j] = &vals[j]
		}
		if err := rows.Scan(ptrs...); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		fmt.Printf("  [%d]", i+1)
		for j, col := range cols {
			if vals[j].Valid {
				fmt.Printf(" %s=%s", col, vals[j].String)
			} else {
				fmt.Printf(" %s=<nil>", col)
			}
		}
		fmt.Println()
		i++
	}
	if i == 0 {
		fmt.Println("  (无匹配结果)")
	} else {
		fmt.Printf("  共 %d 条结果\n", i)
	}
}
