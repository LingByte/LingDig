package main

import (
	"fmt"
	"log"
)

func testDig() {
	digService := NewDigService()

	// 测试查询 Google 的 A 记录
	fmt.Println("测试查询 google.com 的 A 记录...")
	results, err := digService.Query("google.com", "A", "8.8.8.8:53")
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 条记录:\n", len(results))
	for _, record := range results {
		fmt.Printf("- %s %s %d %s\n", record.Name, record.Type, record.TTL, record.Value)
	}

	// 测试查询 MX 记录
	fmt.Println("\n测试查询 google.com 的 MX 记录...")
	results, err = digService.Query("google.com", "MX", "8.8.8.8:53")
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 条记录:\n", len(results))
	for _, record := range results {
		fmt.Printf("- %s %s %d %s\n", record.Name, record.Type, record.TTL, record.Value)
	}
}

// 如果直接运行这个文件进行测试
// go run test_dig.go dig_service.go
