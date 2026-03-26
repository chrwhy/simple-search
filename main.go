package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/chrwhy/simple-search/dao"
	"github.com/chrwhy/simple-search/parser"
)

func main() {
	db := dao.InitDB()
	defer db.Close()
	dao.CreateTable(db)
	dao.InitData(db)

	parser.InitJieba()
	defer parser.FreeJieba()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Simple FTS5 REPL (Jieba) ===")
	fmt.Println("docs_name(fid INTEGER, name, cate, ctime)       tokenize='simple'   (pinyin ON)")
	fmt.Println("docs_content(fid INTEGER, content, cate, ctime) tokenize='simple 0'  (pinyin OFF)")
	fmt.Println()

	for {
		fmt.Println("选择模式:")
		fmt.Println("  1. 搜索 name (标题)")
		fmt.Println("  2. 搜索 content (内容)")
		fmt.Println("  3. 原始 SQL")
		fmt.Println("  4. 退出")
		fmt.Print("> ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			queryLoop(reader, "name", func(clause string) {
				dao.QueryByName(db, clause)
			})
		case "2":
			queryLoop(reader, "content", func(clause string) {
				dao.QueryByContent(db, clause)
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
		input, _ := reader.ReadString('\n')
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
		input, _ := reader.ReadString('\n')
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
