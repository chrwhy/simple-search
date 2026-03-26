//go:build fts5

package dao

import (
	"database/sql"
	"os"
	"testing"

	"github.com/chrwhy/simple-search/config"
	"github.com/chrwhy/simple-search/model"

	// Standard sqlite3 driver (no libsimple extension) for testing basic FTS5.
	_ "github.com/mattn/go-sqlite3"
)

// testDB opens an in-memory SQLite database with standard FTS5 support.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// testCfg returns a Config pointing to a temp data file with unicode61 tokenizer
// (available without libsimple extension).
func testCfg(t *testing.T, dataPath string) *config.Config {
	t.Helper()
	return &config.Config{
		DBPath:           ":memory:",
		LibPath:          "unused",
		DataPath:         dataPath,
		NameTable:        "test_name",
		ContentTable:     "test_content",
		NameTokenizer:    "unicode61",
		ContentTokenizer: "unicode61",
	}
}

func TestCreateTable_Integration(t *testing.T) {
	db := testDB(t)
	_ = testCfg(t, "")

	// Manually create tables (CreateTable checks allowlist, which only has docs_*).
	// For integration testing, we use direct SQL.
	for _, tbl := range []string{"test_name", "test_content"} {
		var colDef string
		if tbl == "test_name" {
			colDef = "fid, name, cate, ctime, tokenize = 'unicode61'"
		} else {
			colDef = "fid, content, cate, ctime, tokenize = 'unicode61'"
		}
		_, err := db.Exec("CREATE VIRTUAL TABLE " + tbl + " USING fts5(" + colDef + ")")
		if err != nil {
			t.Fatalf("create %s: %v", tbl, err)
		}
	}

	// Verify tables exist by inserting a row.
	_, err := db.Exec("INSERT INTO test_name(fid, name, cate, ctime) VALUES (1, 'test', 'A', '2024-01-01')")
	if err != nil {
		t.Fatalf("insert test_name: %v", err)
	}
	_, err = db.Exec("INSERT INTO test_content(fid, content, cate, ctime) VALUES (1, 'hello world', 'A', '2024-01-01')")
	if err != nil {
		t.Fatalf("insert test_content: %v", err)
	}
}

func TestInsertDoc_Integration(t *testing.T) {
	db := testDB(t)
	cfg := testCfg(t, "")

	// Create tables with standard tokenizer.
	for _, tbl := range []string{"test_name", "test_content"} {
		var colDef string
		if tbl == "test_name" {
			colDef = "fid, name, cate, ctime, tokenize = 'unicode61'"
		} else {
			colDef = "fid, content, cate, ctime, tokenize = 'unicode61'"
		}
		db.Exec("CREATE VIRTUAL TABLE " + tbl + " USING fts5(" + colDef + ")")
	}

	doc := model.Doc{
		FID:     1,
		Name:    "Go语言编程",
		Content: "Go是一种静态类型的编译型语言",
		Cate:    "tech",
		CTime:   "2024-06-01",
	}
	InsertDoc(cfg, db, doc)

	// Verify data was inserted.
	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_name").Scan(&count)
	if count != 1 {
		t.Errorf("test_name count = %d, want 1", count)
	}
	db.QueryRow("SELECT COUNT(*) FROM test_content").Scan(&count)
	if count != 1 {
		t.Errorf("test_content count = %d, want 1", count)
	}
}

func TestBatchInsert_Integration(t *testing.T) {
	db := testDB(t)
	cfg := testCfg(t, "")

	for _, tbl := range []string{"test_name", "test_content"} {
		var colDef string
		if tbl == "test_name" {
			colDef = "fid, name, cate, ctime, tokenize = 'unicode61'"
		} else {
			colDef = "fid, content, cate, ctime, tokenize = 'unicode61'"
		}
		db.Exec("CREATE VIRTUAL TABLE " + tbl + " USING fts5(" + colDef + ")")
	}

	docs := []model.Doc{
		{FID: 1, Name: "Alpha", Content: "first doc", Cate: "A", CTime: "2024-01-01"},
		{FID: 2, Name: "Beta", Content: "second doc", Cate: "B", CTime: "2024-02-01"},
		{FID: 3, Name: "Gamma", Content: "third doc", Cate: "A", CTime: "2024-03-01"},
	}

	err := batchInsert(cfg, db, docs)
	if err != nil {
		t.Fatalf("batchInsert: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_name").Scan(&count)
	if count != 3 {
		t.Errorf("test_name count = %d, want 3", count)
	}
}

func TestUpsertBatch_Integration(t *testing.T) {
	db := testDB(t)
	cfg := testCfg(t, "")

	for _, tbl := range []string{"test_name", "test_content"} {
		var colDef string
		if tbl == "test_name" {
			colDef = "fid, name, cate, ctime, tokenize = 'unicode61'"
		} else {
			colDef = "fid, content, cate, ctime, tokenize = 'unicode61'"
		}
		db.Exec("CREATE VIRTUAL TABLE " + tbl + " USING fts5(" + colDef + ")")
	}

	// Initial insert.
	docs := []model.Doc{
		{FID: 1, Name: "Alpha", Content: "first", Cate: "A", CTime: "2024-01-01"},
	}
	if err := batchInsert(cfg, db, docs); err != nil {
		t.Fatalf("initial batchInsert: %v", err)
	}

	// Upsert with updated content for fid=1 and a new fid=2.
	updated := []model.Doc{
		{FID: 1, Name: "Alpha Updated", Content: "first updated", Cate: "A", CTime: "2024-01-02"},
		{FID: 2, Name: "Beta", Content: "second", Cate: "B", CTime: "2024-02-01"},
	}
	if err := upsertBatch(cfg, db, updated); err != nil {
		t.Fatalf("upsertBatch: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_name").Scan(&count)
	if count != 2 {
		t.Errorf("test_name count = %d, want 2", count)
	}

	// Verify fid=1 was updated.
	var name string
	db.QueryRow("SELECT name FROM test_name WHERE fid = 1").Scan(&name)
	if name != "Alpha Updated" {
		t.Errorf("fid=1 name = %q, want %q", name, "Alpha Updated")
	}
}

func TestInitData_Integration(t *testing.T) {
	// Write a temp JSON file.
	tmpFile := t.TempDir() + "/test.json"
	data := `[
		{"fid":1,"name":"Test","Content":"hello","cate":"A","ctime":"2024-01-01"},
		{"fid":2,"name":"World","Content":"foo bar","cate":"B","ctime":"2024-02-01"}
	]`
	if err := os.WriteFile(tmpFile, []byte(data), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	db := testDB(t)
	cfg := testCfg(t, tmpFile)

	// Create tables.
	for _, tbl := range []string{"test_name", "test_content"} {
		var colDef string
		if tbl == "test_name" {
			colDef = "fid, name, cate, ctime, tokenize = 'unicode61'"
		} else {
			colDef = "fid, content, cate, ctime, tokenize = 'unicode61'"
		}
		db.Exec("CREATE VIRTUAL TABLE " + tbl + " USING fts5(" + colDef + ")")
	}

	// Manually call the loading logic (InitData reads from cfg.NameTable which is "test_name").
	// We replicate the logic here since InitData checks the allowlist.
	docs, err := loadDocs(tmpFile)
	if err != nil {
		t.Fatalf("loadDocs: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("loaded %d docs, want 2", len(docs))
	}

	if err := batchInsert(cfg, db, docs); err != nil {
		t.Fatalf("batchInsert: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_name").Scan(&count)
	if count != 2 {
		t.Errorf("test_name count = %d, want 2", count)
	}
}

func TestFTS5Match_Integration(t *testing.T) {
	db := testDB(t)

	_, err := db.Exec(`CREATE VIRTUAL TABLE test_fts USING fts5(title, body, tokenize='unicode61')`)
	if err != nil {
		t.Fatalf("create fts: %v", err)
	}

	_, err = db.Exec(`INSERT INTO test_fts(title, body) VALUES ('SQLite FTS5 搜索', '全文检索引擎测试')`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	_, err = db.Exec(`INSERT INTO test_fts(title, body) VALUES ('Go 语言', '并发编程与网络服务')`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// MATCH query.
	rows, err := db.Query("SELECT title FROM test_fts WHERE test_fts MATCH '搜索'")
	if err != nil {
		t.Fatalf("match query: %v", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		rows.Scan(&title)
		titles = append(titles, title)
	}
	if len(titles) != 1 {
		t.Errorf("match '搜索' returned %d results, want 1", len(titles))
	}
	if len(titles) > 0 && titles[0] != "SQLite FTS5 搜索" {
		t.Errorf("matched title = %q, want %q", titles[0], "SQLite FTS5 搜索")
	}
}
