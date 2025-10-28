package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/pflag"
)

// PortInfo 对应 "ports" 数组中的对象
type PortInfo struct {
	Port   int    `json:"port"`
	Proto  string `json:"proto"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	Ttl    int    `json:"ttl"`
}

// ScanResult 对应顶层数组中的对象
type ScanResult struct {
	Ip        string     `json:"ip"`
	Timestamp string     `json:"timestamp"`
	Ports     []PortInfo `json:"ports"`
}

func main() {
	var ports string
	var ip string
	var rate int
	var savePath string
	var mode string
	

	pflag.IntVarP(&rate, "rate", "r", 1000, "扫描速率")
	pflag.StringVarP(&ports, "ports", "p", "80,443", "要扫描的端口")
	pflag.StringVarP(&ip, "ip", "i", "", "ip范围")
	pflag.StringVarP(&savePath, "save", "s", "./output.json", "结果保存路径")
	pflag.StringVarP(&mode, "mode", "m", "scan", "功能选择")
	pflag.Parse()

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
		jsonData := runCommand(ip, ports, rate)

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

	default:
		log.Fatalf("未知模式: %s", mode)
	}

}

// 初始化数据库连接
func initDB(dbPatch string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbPatch)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	newTable := `
    CREATE TABLE IF NOT EXISTS scan_results (
        id             INTEGER PRIMARY KEY AUTOINCREMENT,
        ip             TEXT NOT NULL,
        port           INTEGER NOT NULL,
        proto          TEXT NOT NULL,
        ttl            INTEGER,
        reason         TEXT,
        scan_time_unix INTEGER NOT NULL
    );`
	_, err = db.Exec(newTable)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table Err: %v", err)
	}
	fmt.Println("-- new db successfully --")
	return db, nil
}

func runCommand(ip string, ports string, rate int) []byte {

	rateStr := strconv.Itoa(rate)

	args := []string{
		"-p", ports,
		"--rate", rateStr,
		ip,
		"-oJ", "scan_results.json",
	}
	log.Printf("执行命令: masscan %v", args)
	cmd := exec.Command("masscan", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}
	log.Println("扫描完成，结果已保存到 scan_results.json")

	jsonFile, err := os.ReadFile("scan_results.json")
	if err != nil {
		log.Fatalf("无法打开 JSON 文件: %v", err)
	}

	return jsonFile
}

func saveToDB(db *sqlx.DB, jsonData []byte) error {
	sql := `INSERT INTO scan_results (ip, port, proto, ttl, reason, scan_time_unix)
			VALUES (?, ?, ?, ?, ?, ?)`

	// 解析 处理 JSON 数据
	var results []ScanResult
	err := json.Unmarshal(jsonData, &results)
	if err != nil {
		return fmt.Errorf("JSON 解析失败: %w", err)
	}

	for _, result := range results {

		timestamp, err := strconv.ParseInt(result.Timestamp, 10, 64)
		if err != nil {
			log.Printf("时间戳解析失败: %v", err)
			continue
		}

		for _, portInfo := range result.Ports {
			_, err := db.Exec(sql,
				result.Ip,
				portInfo.Port,
				portInfo.Proto,
				portInfo.Ttl,
				portInfo.Reason,
				timestamp,
			)
			if err != nil {
				log.Printf("插入数据库失败: %v", err)
				continue
			}
		}
	}
	log.Println("save to db successfully")
	return nil

}
