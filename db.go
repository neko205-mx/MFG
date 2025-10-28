package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

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

func saveToDB(db *sqlx.DB, jsonData []byte) error {

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	insertSql := `INSERT INTO scan_results (ip, port, proto, ttl, reason, scan_time_unix)
				VALUES (?, ?, ?, ?, ?, ?)`
	deleteSql := `DELETE FROM scan_results WHERE ip = ? `

	// 解析 处理 JSON 数据
	var results []ScanResult

	err = json.Unmarshal(jsonData, &results)
	if err != nil {
		return fmt.Errorf("JSON 解析失败: %w", err)
	}

	// 用于记录已删除的 IP，避免重复删除
	deleteIpMap := make(map[string]bool)

	for _, result := range results {

		if !deleteIpMap[result.Ip] {
			log.Printf("删除旧记录: IP %s", result.Ip)
			_, err := tx.Exec(deleteSql, result.Ip)
			if err != nil {
				log.Printf("删除旧记录失败: %v", err)
				continue
			}
			deleteIpMap[result.Ip] = true
		}

		// 插入新的扫描结果
		log.Printf("插入新记录: IP %s", result.Ip)
		timestamp, err := strconv.ParseInt(result.Timestamp, 10, 64)
		if err != nil {
			log.Printf("时间戳解析失败: %v", err)
			continue
		}

		for _, portInfo := range result.Ports {
			_, err := tx.Exec(insertSql,
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
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}
	log.Println("save to db successfully")
	return nil

}

func formatResults(results []ipPortResult) []string {
	formatted := make([]string, len(results))
	for i, res := range results {
		formatted[i] = fmt.Sprintf("%s:%d", res.Ip, res.Port)
	}
	return formatted
}

// 根据 IP 查询信息
func queryByIP(db *sqlx.DB, ip string) ([]string, error) {
	var results []ipPortResult
	sql := `SELECT DISTINCT ip, port FROM scan_results WHERE ip = ? ORDER BY port`
	err := db.Select(&results, sql, ip)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	if len(results) == 0 {
		log.Printf("未找到IP %s 的扫描记录", ip)
		return nil, nil
	}

	return formatResults(results), nil

}

// 根据端口查询信息
func queryByPort(db *sqlx.DB, portStr string) ([]string, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("端口转换失败: %w", err)
	}

	var results []ipPortResult
	sql := `SELECT DISTINCT ip, port FROM scan_results WHERE port = ? ORDER BY ip`
	err = db.Select(&results, sql, port)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	if len(results) == 0 {
		log.Printf("未找到端口 %d 的扫描记录", port)
		return nil, nil
	}

	return formatResults(results), nil

}
