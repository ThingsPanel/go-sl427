package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// 项目名称，用于验证
	PROJECT_NAME = "go-sl427"
	// 项目源码所在的相对路径
	PROJECT_PATH = "../../../" + PROJECT_NAME
)

// 允许的文件类型
var allowedExtensions = map[string]bool{
	".go":   true,
	".mod":  true,
	".yaml": true,
	".yml":  true,
	".md":   true,
}

func main() {
	sourceDir, err := filepath.Abs(PROJECT_PATH)
	if err != nil {
		fmt.Printf("获取项目路径失败: %v\n", err)
		return
	}

	// 验证目录名称
	if !strings.HasSuffix(sourceDir, PROJECT_NAME) {
		fmt.Printf("错误: 当前目录 %s 不是目标项目目录\n", sourceDir)
		fmt.Printf("请确保在正确的位置执行此工具\n")
		return
	}

	// 检查目录是否存在
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Printf("错误: 项目目录不存在: %s\n", sourceDir)
		fmt.Printf("请检查项目路径是否正确\n")
		return
	}

	outputFile := fmt.Sprintf("%s_code_%s.txt", PROJECT_NAME, time.Now().Format("20060102_150405"))
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer out.Close()

	fmt.Printf("开始处理项目: %s\n", PROJECT_NAME)
	fmt.Printf("项目路径: %s\n", sourceDir)

	// 写入文档开始标签
	fmt.Fprintf(out, "<documents>")

	fileCount := 0
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过隐藏文件和目录
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 跳过日志文件和目录
		if strings.Contains(path, "logs") || strings.HasSuffix(path, ".log") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 如果是目录，继续遍历
		if info.IsDir() {
			return nil
		}

		// 检查文件类型
		ext := strings.ToLower(filepath.Ext(path))
		if !allowedExtensions[ext] {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("获取相对路径失败: %v", err)
		}

		// 写入文件信息
		fmt.Fprintf(out, "\n<document index=\"%s\">\n", relPath)
		fmt.Fprintf(out, "<source>%s</source>\n", relPath)
		fmt.Fprintf(out, "<document_content>")

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取文件 %s 失败: %v", path, err)
		}

		// 写入内容
		fmt.Fprintf(out, "%s", content)
		fmt.Fprintf(out, "</document_content>\n")
		fmt.Fprintf(out, "</document>")

		fileCount++
		fmt.Printf("处理文件(%d): %s\n", fileCount, relPath)
		return nil
	})

	fmt.Fprintf(out, "</documents>")

	if err != nil {
		fmt.Printf("\n处理过程中发生错误: %v\n", err)
		return
	}

	fmt.Printf("\n处理完成!\n")
	fmt.Printf("共处理 %d 个文件\n", fileCount)
	fmt.Printf("输出文件: %s\n", outputFile)
}
