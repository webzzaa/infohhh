package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v2"
)

// Rule 定义配置文件中的规则结构
type Rule struct {
	ID      int      `yaml:"id"`
	Name    string   `yaml:"name"`
	Enable  bool     `yaml:"enable"`
	Regexes []string `yaml:"regexes"`
}

// Config 定义配置文件的结构
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// MatchResult 定义匹配结果的结构
type MatchResult struct {
	FileName  string
	LineNum   int
	LineContent string
	RuleID    int
	RuleName  string
	MatchText string
}

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	scanDir := flag.String("dir", ".", "要扫描的目录路径")
	outputFile := flag.String("output", "sensitive_info.xlsx", "输出的Excel文件路径")
	flag.Parse()

	// 加载配置文件
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 扫描目录
	results := scanDirectory(*scanDir, config)

	// 输出结果到Excel
	err = outputToExcel(results, *outputFile)
	if err != nil {
		log.Fatalf("输出结果到Excel失败: %v", err)
	}

	log.Printf("扫描完成，共发现 %d 条敏感信息，结果已输出到 %s", len(results), *outputFile)
}

// loadConfig 加载配置文件
func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// scanDirectory 扫描目录并匹配敏感信息
func scanDirectory(dirPath string, config *Config) []MatchResult {
	var results []MatchResult

	// 遍历目录
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("访问文件失败: %v", err)
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理文本文件（可以根据需要调整）
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".js" && ext != ".json" && ext != ".wxml" && ext != ".wxss" && ext != ".ts" && ext != ".jsx" && ext != ".tsx" && ext != ".html" && ext != ".css" {
			return nil
		}

		// 读取文件内容
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("读取文件失败: %v", err)
			return nil
		}

		// 按行处理
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			// 应用每条规则
			for _, rule := range config.Rules {
				if !rule.Enable {
					continue
				}

				// 应用规则中的每个正则表达式
				for _, regexPattern := range rule.Regexes {
					regex, err := regexp.Compile(regexPattern)
					if err != nil {
						log.Printf("编译正则表达式失败 (规则 %d): %v", rule.ID, err)
						continue
					}

					// 查找匹配
					matches := regex.FindAllStringSubmatch(line, -1)
					for _, match := range matches {
						if len(match) > 0 {
							// 使用第一个捕获组作为匹配文本，如果有的话
							matchText := match[0]
							if len(match) > 1 && match[1] != "" {
								matchText = match[1]
							}

							// 添加匹配结果
							results = append(results, MatchResult{
								FileName:    path,
								LineNum:     lineNum + 1,
								LineContent: line,
								RuleID:      rule.ID,
								RuleName:    rule.Name,
								MatchText:   matchText,
							})
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("遍历目录失败: %v", err)
	}

	return results
}

// outputToExcel 将匹配结果输出到Excel文件
func outputToExcel(results []MatchResult, outputPath string) error {
	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 设置工作表名称
	sheetName := "敏感信息扫描结果"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// 删除默认的Sheet1
	f.DeleteSheet("Sheet1")

	// 设置列宽
	f.SetColWidth(sheetName, "A", "F", 20)
	f.SetColWidth(sheetName, "C", "C", 50)
	f.SetColWidth(sheetName, "F", "F", 30)

	// 写入表头
	header := []string{"文件名", "行号", "行内容", "规则ID", "规则名称", "匹配文本"}
	for i, headerText := range header {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, headerText)
		// 设置表头样式
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#CCCCCC"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
		})
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 写入匹配结果
	for rowNum, result := range results {
		row := []interface{}{
			result.FileName,
			result.LineNum,
			result.LineContent,
			result.RuleID,
			result.RuleName,
			result.MatchText,
		}
		for colNum, value := range row {
			cell := fmt.Sprintf("%c%d", 'A'+colNum, rowNum+2)
			f.SetCellValue(sheetName, cell, value)
			// 设置单元格样式
			cellStyle, _ := f.NewStyle(&excelize.Style{
				Alignment: &excelize.Alignment{Vertical: "top"},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
				},
			})
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
		}
	}

	// 保存Excel文件
	return f.SaveAs(outputPath)
}
