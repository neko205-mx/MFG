package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

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
		fmt.Println("mode init, scan, query \n 使用方法 --mode query --ip/ports 1.1.1.1")
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
		log.Println("数据库连接成功")
		defer db.Close()

		if !pflag.CommandLine.Changed("ip") && !pflag.CommandLine.Changed("ports") {
			log.Fatalf("必须提供 -ip 或 -ports 参数")
		} else if pflag.CommandLine.Changed("ip") && pflag.CommandLine.Changed("ports") {
			log.Fatalf("不能同时提供 -ip 和 -ports 参数")
		}

		var results []string
		if pflag.CommandLine.Changed("ip") {
			_, _, err := net.ParseCIDR(ip)
			if err == nil {
				results, err = queryByIP(db, Crid2ips(strings.Split(ip, ",")))
			} else {
				results, err = queryByIP(db, strings.Split(ip, ","))
			}

		} else {
			results, err = queryByPort(db, strings.Split(ports, ","))
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

		fmt.Printf("正在写入%s\n", savePath)
		fileData := []byte(strings.Join(results, "\n"))
		err = os.WriteFile(savePath, fileData, 0644)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("未知模式: %s", mode)
	}

}

// Crid2ips crid to ips
func Crid2ips(crid []string) []string {
	var ips []string

	ipAddr, ipNet, err := net.ParseCIDR(crid[0])

	if err != nil {
		log.Print(err)
	}

	for ip := ipAddr.Mask(ipNet.Mask); ipNet.Contains(ip); increment(ip) {
		ips = append(ips, ip.String())
	}

	// CIDR too small eg. /31
	if len(ips) <= 2 {
		log.Print("err")
	}

	return ips
}

// increment 配合Crid2ips使用
func increment(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}
