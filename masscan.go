package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

func runCommand(ip string, ports string, rate int, savePath string) []byte {

	rateStr := strconv.Itoa(rate)

	args := []string{
		"-p", ports,
		"--rate", rateStr,
		ip,
		"-oJ", savePath,
	}
	log.Printf("执行命令: masscan %v", args)
	cmd := exec.Command("masscan", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}
	log.Println("扫描完成，结果已保存到 ", savePath)

	jsonFile, err := os.ReadFile(savePath)
	if err != nil {
		log.Fatalf("无法打开 JSON 文件: %v", err)
	}

	return jsonFile
}
