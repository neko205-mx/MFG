package main
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

// ipPortResult 用于查询结果返回
type ipPortResult struct {
	Ip    string	`db:"ip"`
	Port  int		`db:"port"`
}