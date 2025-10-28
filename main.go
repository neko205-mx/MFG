package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/pflag"
)

func main() {
	var ports string
	var ip string
	var rate int
	var savePath string
	var mode string
	var help bool

	pflag.BoolVarP(&help, "help", "h", false, "显示帮助信息")
	pflag.IntVarP(&rate, "rate", "r", 1000, "扫描速率")
	pflag.StringVarP(&ports, "ports", "p", "80,443", "要扫描的端口")
	pflag.StringVarP(&ip, "ip", "i", "", "ip范围")
	pflag.StringVarP(&savePath, "save", "s", "./output.json", "结果保存路径")
	pflag.StringVarP(&mode, "mode", "m", "scan", "功能选择")
	pflag.Parse()

	if help {
		pflag.Usage()
		return
	}

	switch mode {
	case "scan":
		fmt.Println("--- 解析 ---")
		fmt.Printf("端口: %s\n", ports)
		fmt.Printf("速率: %d\n", rate)
		fmt.Printf("IP 地址/段: %s\n", ip)
		fmt.Println("--- 结束 ---")

		db, err := sqlx.Connect("sqlite3", "./ips.db")
		if err != nil {
			db.Close()
			log.Fatalf("数据库连接失败: %v", err)
		}
		log.Println("数据库连接成功")
		defer db.Close()
		jsonData := runCommand(ip, ports, rate, savePath)

		err = saveToDB(db, jsonData)
		if err != nil {
			log.Fatalf("保存到数据库失败: %v", err)
		}

	case "init":
		fmt.Println("--- sqlite 初始化 ---")
		db, err := initDB("./ips.db")
		if err != nil {
			log.Fatalf("数据库初始化失败: %v", err)
		}
		defer db.Close()
		fmt.Println("--- 数据库初始化完成 ---")

	case "query":
		fmt.Println("--- 查询 ---")
		db, err := sqlx.Connect("sqlite3", "./ips.db")
		if err != nil {
			log.Fatalf("数据库连接失败: %v", err)
		}
		defer db.Close()

		if !pflag.CommandLine.Changed("ip") && !pflag.CommandLine.Changed("ports") {
			log.Fatalf("必须提供 -ip 或 -ports 参数")
		} else if pflag.CommandLine.Changed("ip") && pflag.CommandLine.Changed("ports") {
			log.Fatalf("不能同时提供 -ip 和 -ports 参数")
		}

		var results []string
		if pflag.CommandLine.Changed("ip") {
			results, err = queryByIP(db, ip)
		} else {
			results, err = queryByPort(db, ports)
		}

		if err != nil {
			log.Fatalf("查询失败: %v", err)
		}

		if len(results) > 0 {
			fmt.Println("\n查询结果:")
			for _, line := range results {
				fmt.Println(line)
			}
			fmt.Printf("\n总计: %d 条记录\n", len(results))
		}

	default:
		log.Fatalf("未知模式: %s", mode)
	}

}
