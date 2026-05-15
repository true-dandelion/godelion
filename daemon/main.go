package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	// 主程序路径
	MainProgramPath = "/root/godelion/api/main.go"
	MainProgramName = "godelion-api"
	
	// 日志目录
	LogBaseDir = "/root/godelion"
	
	// 监控配置
	CheckInterval = 5 * time.Second
	MaxRestartAttempts = 5
	RestartDelay = 10 * time.Second
)

var (
	mainProcess *exec.Cmd
	mainPID     int
	logDir      string
	restartCount = 0
	isRunning   = true
	ctx, cancel = context.WithCancel(context.Background())
)

// 初始化日志目录
func initLogDir() error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logDir = filepath.Join(LogBaseDir, timestamp)
	
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	// 创建日志子目录
	logsDir := filepath.Join(logDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		return fmt.Errorf("创建日志子目录失败: %v", err)
	}
	
	return nil
}

// 获取当前日志文件路径
func getCurrentLogFile() string {
	return filepath.Join(logDir, "logs", fmt.Sprintf("daemon_%s.log", time.Now().Format("2006-01-02")))
}

// 记录日志到文件和控制台
func writeLog(level string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullMsg := fmt.Sprintf("[%s] [%s] %s", timestamp, level, msg)
	
	// 输出到控制台
	fmt.Println(fullMsg)
	
	// 写入日志文件
	logFile, err := os.OpenFile(getCurrentLogFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("写入日志文件失败: %v", err)
		return
	}
	defer logFile.Close()
	
	logFile.WriteString(fullMsg + "\n")
}

// 信息日志
func logInfo(format string, args ...interface{}) {
	writeLog("INFO", format, args...)
}

// 错误日志
func logError(format string, args ...interface{}) {
	writeLog("ERROR", format, args...)
}

// 警告日志
func logWarn(format string, args ...interface{}) {
	writeLog("WARN", format, args...)
}

// 检查主程序进程是否在运行
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// 发送信号0来检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// 启动主程序
func startMainProgram() error {
	logInfo("正在启动主程序: %s", MainProgramPath)
	
	// 构建主程序（如果需要）
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(LogBaseDir, "bin", MainProgramName), MainProgramPath)
	buildCmd.Dir = filepath.Dir(MainProgramPath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		logError("编译主程序失败: %v, 输出: %s", err, string(buildOutput))
		return fmt.Errorf("编译主程序失败: %v", err)
	}
	logInfo("主程序编译成功")
	
	// 启动主程序
	binaryPath := filepath.Join(LogBaseDir, "bin", MainProgramName)
	cmd := exec.Command(binaryPath)
	cmd.Dir = filepath.Dir(MainProgramPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 创建新的进程组
	}
	
	err = cmd.Start()
	if err != nil {
		logError("启动主程序失败: %v", err)
		return fmt.Errorf("启动主程序失败: %v", err)
	}
	
	mainProcess = cmd
	mainPID = cmd.Process.Pid
	logInfo("主程序已启动, PID: %d", mainPID)
	
	return nil
}

// 停止主程序
func stopMainProgram() error {
	if mainProcess == nil || mainProcess.Process == nil {
		return nil
	}
	
	logInfo("正在停止主程序, PID: %d", mainPID)
	
	// 尝试正常终止
	err := mainProcess.Process.Signal(syscall.SIGTERM)
	if err != nil {
		logWarn("发送 SIGTERM 失败: %v, 尝试强制终止", err)
		err = mainProcess.Process.Kill()
		if err != nil {
			logError("强制终止主程序失败: %v", err)
			return err
		}
	}
	
	// 等待进程结束
	mainProcess.Wait()
	logInfo("主程序已停止")
	
	return nil
}

// 监控主程序
func monitorMainProgram() {
	ticker := time.NewTicker(CheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !isProcessRunning(mainPID) {
				logWarn("检测到主程序 (PID: %d) 未运行", mainPID)
				
				if restartCount >= MaxRestartAttempts {
					logError("已达到最大重启次数 (%d), 守护进程将退出", MaxRestartAttempts)
					isRunning = false
					return
				}
				
				restartCount++
				logWarn("准备重启主程序 (第 %d/%d 次)", restartCount, MaxRestartAttempts)
				
				// 等待一段时间后重启
				time.Sleep(RestartDelay)
				
				err := startMainProgram()
				if err != nil {
					logError("重启主程序失败: %v", err)
				}
			} else {
				// 进程正常运行，重置重启计数
				if restartCount > 0 {
					logInfo("主程序运行正常, 重启计数已重置")
					restartCount = 0
				}
			}
		}
	}
}

// 创建 PID 文件
func createPIDFile() error {
	pidFile := filepath.Join(LogBaseDir, "daemon.pid")
	err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		return fmt.Errorf("创建 PID 文件失败: %v", err)
	}
	logInfo("PID 文件已创建: %s", pidFile)
	return nil
}

// 删除 PID 文件
func removePIDFile() {
	pidFile := filepath.Join(LogBaseDir, "daemon.pid")
	os.Remove(pidFile)
}

// 处理系统信号
func handleSignals() {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	
	// 注册需要捕获的信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	
	for sig := range sigChan {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			logInfo("收到终止信号, 正在关闭守护进程...")
			cancel()
			stopMainProgram()
			removePIDFile()
			logInfo("守护进程已关闭")
			os.Exit(0)
			
		case syscall.SIGUSR1:
			logInfo("收到用户信号 SIGUSR1, 正在重启主程序...")
			stopMainProgram()
			restartCount = 0 // 重置重启计数
			err := startMainProgram()
			if err != nil {
				logError("重启主程序失败: %v", err)
			}
			
		case syscall.SIGUSR2:
			logInfo("收到用户信号 SIGUSR2, 正在重新加载配置...")
			// 重新加载配置（如果需要）
			logInfo("配置已重新加载")
		}
	}
}

// 主函数
func main() {
	fmt.Println("========================================")
	fmt.Println("       Godelion 守护进程启动中...       ")
	fmt.Println("========================================")
	
	// 初始化日志目录
	err := initLogDir()
	if err != nil {
		fmt.Printf("初始化日志目录失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("日志目录: %s\n", logDir)
	
	// 创建 PID 文件
	err = createPIDFile()
	if err != nil {
		logError("创建 PID 文件失败: %v", err)
	}
	
	// 启动主程序
	err = startMainProgram()
	if err != nil {
		logError("启动主程序失败: %v", err)
		os.Exit(1)
	}
	
	// 启动监控协程
	go monitorMainProgram()
	
	// 启动信号处理协程
	go handleSignals()
	
	logInfo("========================================")
	logInfo("  Godelion 守护进程已启动")
	logInfo("  监控主程序: %s", MainProgramPath)
	logInfo("  主程序 PID: %d", mainPID)
	logInfo("  检查间隔: %v", CheckInterval)
	logInfo("  最大重启次数: %d", MaxRestartAttempts)
	logInfo("========================================")
	
	// 保持主程序运行
	<-ctx.Done()
}
