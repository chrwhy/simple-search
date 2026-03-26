package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chrwhy/simple-parser"
	"github.com/chrwhy/simple-search/config"
	"github.com/chrwhy/simple-search/dao"
)

func main() {
	// CLI flags for one-shot queries
	queryName := flag.String("query", "", "Search by name (title) and exit")
	queryContent := flag.String("content", "", "Search by content and exit")
	rawSQL := flag.String("sql", "", "Execute raw SQL and exit")
	importPath := flag.String("import", "", "Import/upsert data from JSON file and exit")
	flag.Parse()

	cfg := config.Load()
	db := dao.InitDB(cfg)
	defer db.Close()
	dao.CreateTable(cfg, db)
	dao.InitData(cfg, db)

	parser.InitJieba()
	defer parser.FreeJieba()

	// One-shot mode: execute and exit
	if *queryName != "" {
		clause := parser.ParseJiebaClause(*queryName)
		fmt.Printf("MATCH clause: %s\n", clause)
		dao.QueryByName(cfg, db, clause)
		return
	}
	if *queryContent != "" {
		clause := parser.ParseJiebaClause(*queryContent)
		fmt.Printf("MATCH clause: %s\n", clause)
		dao.QueryByContent(cfg, db, clause)
		return
	}
	if *rawSQL != "" {
		dao.ExecQuery(db, *rawSQL)
		return
	}
	if *importPath != "" {
		dao.ImportData(cfg, db, *importPath)
		return
	}

	// Interactive REPL mode
	runREPL(cfg, db)
}

func runREPL(cfg *config.Config, db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Simple FTS5 REPL (Jieba) ===")
	fmt.Printf("%s(fid INTEGER, name, cate, ctime)       tokenize='%s'\n", cfg.NameTable, cfg.NameTokenizer)
	fmt.Printf("%s(fid INTEGER, content, cate, ctime) tokenize='%s'\n", cfg.ContentTable, cfg.ContentTokenizer)
	fmt.Println()

	for {
		fmt.Println("选择模式:")
		fmt.Println("  1. 搜索 name (标题)")
		fmt.Println("  2. 搜索 content (内容)")
		fmt.Println("  3. 原始 SQL")
		fmt.Println("  4. 退出")
		fmt.Print("> ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nBye!")
				return
			}
			log.Fatalf("读取输入失败: %v", err)
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			queryLoop(reader, "name", func(clause string) {
				dao.QueryByName(cfg, db, clause)
			})
		case "2":
			queryLoop(reader, "content", func(clause string) {
				dao.QueryByContent(cfg, db, clause)
			})
		case "3":
			sqlLoop(reader, db)
		case "4":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("无效选择")
		}
		fmt.Println()
	}
}

func queryLoop(reader *bufio.Reader, field string, queryFn func(string)) {
	fmt.Printf("\n--- 搜索 %s (输入 exit 返回) ---\n", field)
	for {
		fmt.Printf("[%s] > ", field)
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatalf("读取输入失败: %v", err)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" {
			return
		}

		clause := parser.ParseJiebaClause(input)
		fmt.Printf("  MATCH clause: %s\n", clause)
		queryFn(clause)
		fmt.Println()
	}
}

func sqlLoop(reader *bufio.Reader, db *sql.DB) {
	fmt.Println("\n--- 原始 SQL (输入 exit 返回) ---")
	for {
		fmt.Print("[sql] > ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatalf("读取输入失败: %v", err)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" {
			return
		}
		dao.ExecQuery(db, input)
		fmt.Println()
	}
}
