package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"show_logs/pkg/tui"
	"strings"
)

var (
	printVersion bool
	dirPath      string
	schemaName   string
)

type TableInfo struct {
	SchemaName string
	TableName  string
	TaskData   TaskData
}

type TaskData struct {
	TaskStartTime  string
	TaskEndTime    string
	TotalDuration  string
	AverageTraffic string
	WriteSpeed     string
	TotalReadCount string
	TotalFailCount string
}

func readLastLines(filePath string, numLines int) ([]string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// 获取文件的文件信息
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file stats: %v", err)
	}

	// 获取文件的大小
	fileSize := info.Size()

	// 设置读取的位置，从文件的末尾开始
	const bufferSize = 1024
	var lines []string
	var buffer []byte

	// 从文件的末尾反向读取
	for i := fileSize - 1; i >= 0; i-- {
		// 设置读取位置
		_, err := file.Seek(i, 0)
		if err != nil {
			return nil, fmt.Errorf("error seeking in file: %v", err)
		}

		// 读取数据到缓冲区
		buf := make([]byte, 1)
		_, err = file.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		// 将字符添加到缓冲区
		buffer = append([]byte{buf[0]}, buffer...)

		// 如果我们遇到行尾，则将行添加到结果
		if buf[0] == '\n' {
			line := string(buffer)
			// lines = append([]string{line}, lines...)
			lines = append([]string{strings.TrimSpace(line)}, lines...)
			buffer = nil // 清空缓冲区

			// 如果已经读取足够的行，退出
			if len(lines) == numLines {
				break
			}
		}

		// 如果文件读取位置已经超出缓冲区大小，我们就跳到前面进行更大的读取操作
		if len(buffer) >= bufferSize {
			buffer = nil
		}
	}

	// 如果读取的行数少于要求的行数，返回所有可读取的行
	if len(lines) < numLines {
		// 如果文件总行数少于10行, 返回所有行
		return lines, nil
	}

	return lines, nil
}

// 使用正则表达式提取每行分隔符后的数据
func extractData(line string) (string, error) {
	// 正则表达式匹配每行第二段内容（冒号后面的数据）
	re := regexp.MustCompile(`\s*:\s*(.*)`)
	match := re.FindStringSubmatch(line)

	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("no match found in line: %s", line)
}

// 解析任务信息
func parseTaskInfo(input []string) (*TaskData, error) {
	var taskData TaskData

	// 处理每一行并提取冒号后的内容
	for _, line := range input {
		// 根据行的内容决定将什么值赋给结构体
		switch {
		case strings.HasPrefix(line, "任务启动时刻"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.TaskStartTime = data
		case strings.HasPrefix(line, "任务结束时刻"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.TaskEndTime = data
		case strings.HasPrefix(line, "任务总计耗时"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.TotalDuration = data
		case strings.HasPrefix(line, "任务平均流量"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.AverageTraffic = data
		case strings.HasPrefix(line, "记录写入速度"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.WriteSpeed = data
		case strings.HasPrefix(line, "读出记录总数"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.TotalReadCount = data
		case strings.HasPrefix(line, "读写失败总数"):
			data, err := extractData(line)
			if err != nil {
				return nil, err
			}
			taskData.TotalFailCount = data
		}
	}

	return &taskData, nil
}

// 遍历目录并读取所有日志文件
func parseLogsInDirectory(dirPath string, numLines int) ([]TableInfo, error) {
	//	var allTaskData []TaskData
	var allTableInfo []TableInfo

	// 读取目录中的所有文件
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	// 遍历目录中的所有文件
	for _, file := range files {
		// 确保文件是日志文件（例如以 .log 结尾）
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
			filePath := dirPath + "/" + file.Name()

			// 读取文件的最后几行
			lines, err := readLastLines(filePath, numLines)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", file.Name(), err)
				continue
			}

			// 解析任务信息
			taskData, err := parseTaskInfo(lines)
			if err != nil {
				fmt.Printf("Error parsing task info from file %s: %v\n", file.Name(), err)
				continue
			}

			fileName := filepath.Base(filePath)
			tableName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			// 将解析的数据添加到切片中
			allTableInfo = append(allTableInfo, TableInfo{
				SchemaName: schemaName,
				TableName:  tableName,
				TaskData:   *taskData})
		}
	}

	return allTableInfo, nil
}

func init() {
	flag.StringVar(&dirPath, "log_path", "", "Specify the log path. The default value is empty.")
	flag.StringVar(&schemaName, "schema_name", "", "Specify the schema name.")

	flag.BoolVar(&printVersion, "version", false, "print program build version")
	flag.Parse()
}

func main() {
	if printVersion {
		fmt.Println("version: 0.0.1")
		os.Exit(0)
	}

	// 文件路径和需要读取的行数
	numLines := 10

	// 读取文件的最后10行
	//lines, err := readLastLines(filePath, numLines)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}

	tableInfoList, err := parseLogsInDirectory(dirPath, numLines)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// print table
	table := [][]string{
		// header
		{"Schame", "Table Name", "Start Time", "End Time", "Total Duration", "Average Traffic", "Write Speed", "Total Read Count", "Total Fail Count"},
	}

	for _, tableInfo := range tableInfoList {
		table = append(table, []string{
			tableInfo.SchemaName,
			tableInfo.TableName,
			tableInfo.TaskData.TaskStartTime,
			tableInfo.TaskData.TaskEndTime,
			tableInfo.TaskData.TotalDuration,
			tableInfo.TaskData.AverageTraffic,
			tableInfo.TaskData.WriteSpeed,
			tableInfo.TaskData.TotalReadCount,
			tableInfo.TaskData.TotalFailCount,
		})
	}

	tui.PrintTable(table, true)
}
