package dao

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	sql.Register("sqlite3_simple",
		&sqlite3.SQLiteDriver{
			Extensions: []string{
				"./libsimple-osx-x64/libsimple",
			},
		})

	db, err := sql.Open("sqlite3_simple", "example.db")
	if err != nil {
		log.Fatalf("open error: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("ping error: ", err)
	}
	return db
}

func CreateTable(db *sql.DB) {
	nameSQL := `CREATE VIRTUAL TABLE IF NOT EXISTS docs_name USING fts5(
		fid, name, cate, ctime,
		tokenize = 'simple'
	);`
	contentSQL := `CREATE VIRTUAL TABLE IF NOT EXISTS docs_content USING fts5(
		fid, content, cate, ctime,
		tokenize = 'simple 0'
	);`

	if _, err := db.Exec(nameSQL); err != nil {
		log.Fatal("create docs_name: ", err)
	}
	if _, err := db.Exec(contentSQL); err != nil {
		log.Fatal("create docs_content: ", err)
	}
	log.Println("Tables 'docs_name' (pinyin ON) and 'docs_content' (pinyin OFF) ready")
}

type Doc struct {
	FID     int
	Name    string
	Content string
	Cate    string
	CTime   string
}

func InsertDoc(db *sql.DB, doc Doc) {
	_, err := db.Exec(
		`INSERT INTO docs_name(fid, name, cate, ctime) VALUES (?, ?, ?, ?)`,
		doc.FID, doc.Name, doc.Cate, doc.CTime,
	)
	if err != nil {
		log.Printf("insert docs_name error: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO docs_content(fid, content, cate, ctime) VALUES (?, ?, ?, ?)`,
		doc.FID, doc.Content, doc.Cate, doc.CTime,
	)
	if err != nil {
		log.Printf("insert docs_content error: %v", err)
	}
}

func QueryByName(db *sql.DB, matchClause string) {
	sqlStr := fmt.Sprintf(
		"SELECT fid, simple_highlight(docs_name, 1, '[', ']'), cate, ctime FROM docs_name WHERE name MATCH '%s'",
		matchClause,
	)
	execQueryName(db, sqlStr)
}

func QueryByContent(db *sql.DB, matchClause string) {
	sqlStr := fmt.Sprintf(
		"SELECT fid, simple_highlight(docs_content, 1, '[', ']'), cate, ctime FROM docs_content WHERE content MATCH '%s'",
		matchClause,
	)
	execQueryContent(db, sqlStr)
}

func execQueryName(db *sql.DB, sqlStr string) {
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
		var fid int
		var name, cate, ctime string
		rows.Scan(&fid, &name, &cate, &ctime)
		fmt.Printf("  [%d] fid=%d cate=%s ctime=%s\n", i+1, fid, cate, ctime)
		fmt.Printf("    name: %s\n", name)
		i++
	}
	if i == 0 {
		fmt.Println("  (无匹配结果)")
	} else {
		fmt.Printf("  共 %d 条结果\n", i)
	}
}

func execQueryContent(db *sql.DB, sqlStr string) {
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
		var fid int
		var content, cate, ctime string
		rows.Scan(&fid, &content, &cate, &ctime)
		fmt.Printf("  [%d] fid=%d cate=%s ctime=%s\n", i+1, fid, cate, ctime)
		fmt.Printf("    content: %s\n", content)
		i++
	}
	if i == 0 {
		fmt.Println("  (无匹配结果)")
	} else {
		fmt.Printf("  共 %d 条结果\n", i)
	}
}

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
		vals := make([]string, len(cols))
		ptrs := make([]interface{}, len(cols))
		for j := range vals {
			ptrs[j] = &vals[j]
		}
		rows.Scan(ptrs...)
		fmt.Printf("  [%d]", i+1)
		for j, col := range cols {
			fmt.Printf(" %s=%s", col, vals[j])
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
