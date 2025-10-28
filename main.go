package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/pflag"
)

func main() {
	var ports string
	var ip string
	var rate int

	pflag.IntVarP(&rate, "rate", "r", 1000, "扫描速率")
	pflag.StringVarP(&ports, "ports", "p", "80,443", "要扫描的端口")
	pflag.StringVarP(&ip, "ip", "i", "", "ip范围")
	pflag.Parse()

	fmt.Println("--- 解析 ---")
	fmt.Printf("端口: %s\n", ports)
	fmt.Printf("速率: %d\n", rate)
	fmt.Printf("IP 地址/段: %s\n", ip)
	fmt.Println("--- 结束 ---")
	runCommand(ip, ports, rate)

}

func runCommand(ip string, ports string, rate int) {

	rateStr := strconv.Itoa(rate)

	args := []string{
		"-p", ports,
		"--rate", rateStr,
		ip,
		"-oJ", "scan_results.json",
	}

	cmd := exec.Command("masscan", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}
	fmt.Println("扫描完成，结果已保存到 scan_results.json")

	jsonFile, err := os.ReadFile("scan_results.json")
	if err != nil {
		log.Fatalf("无法打开 JSON 文件: %v", err)
	}
	
	fmt.Printf("%s", string(jsonFile))
}