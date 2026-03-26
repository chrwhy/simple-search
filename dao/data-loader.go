package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/chrwhy/simple-search/config"
	"github.com/chrwhy/simple-search/model"
)

// InitData loads sample data from JSON into the FTS5 tables if they are empty.
// Inserts are wrapped in a transaction for performance and atomicity.
func InitData(cfg *config.Config, db *sql.DB) {
	var count int
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", cfg.NameTable)).Scan(&count)
	if count > 0 {
		log.Printf("表已有 %d 条数据，跳过初始化", count)
		return
	}

	docs, err := loadDocs(cfg.DataPath)
	if err != nil {
		log.Printf("加载数据失败: %v", err)
		return
	}

	if err := batchInsert(cfg, db, docs); err != nil {
		log.Printf("批量插入失败: %v", err)
		return
	}
	log.Printf("已插入 %d 条测试数据", len(docs))
}

// ImportData loads docs from a JSON file and upserts them (INSERT OR REPLACE).
// Unlike InitData, this does NOT skip if the table already has data.
func ImportData(cfg *config.Config, db *sql.DB, dataPath string) {
	docs, err := loadDocs(dataPath)
	if err != nil {
		log.Printf("加载数据失败: %v", err)
		return
	}

	if err := upsertBatch(cfg, db, docs); err != nil {
		log.Printf("增量导入失败: %v", err)
		return
	}
	log.Printf("已导入 %d 条数据", len(docs))
}

// loadDocs reads and parses a JSON file into a slice of Doc.
func loadDocs(path string) ([]model.Doc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件 %s: %w", path, err)
	}
	var docs []model.Doc
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("解析 JSON: %w", err)
	}
	return docs, nil
}

// batchInsert inserts all docs in a single transaction.
func batchInsert(cfg *config.Config, db *sql.DB, docs []model.Doc) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务: %w", err)
	}
	for _, doc := range docs {
		InsertDoc(cfg, tx, doc)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务: %w", err)
	}
	return nil
}

// upsertBatch inserts or replaces docs using INSERT OR REPLACE in a transaction.
// Note: FTS5 does not support ON CONFLICT; this deletes+reinserts per fid.
func upsertBatch(cfg *config.Config, db *sql.DB, docs []model.Doc) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务: %w", err)
	}
	for _, doc := range docs {
		// Delete existing rows with the same fid, then insert.
		tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE fid = ?", cfg.NameTable), doc.FID)
		tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE fid = ?", cfg.ContentTable), doc.FID)
		InsertDoc(cfg, tx, doc)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务: %w", err)
	}
	return nil
}
