package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/stupidrun/mon/api/proto"
	"time"
)

// GetCurrentMetrics 获取当前系统的监控指标
func GetCurrentMetrics(threshold int) (*proto.Metric, error) {
	cpuUsage, err := GetCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("获取CPU使用率失败: %w", err)
	}

	memUsage, err := GetMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("获取内存使用率失败: %w", err)
	}

	netIn, netOut, err := GetNetworkIO(threshold)
	if err != nil {
		return nil, fmt.Errorf("获取网络流量失败: %w", err)
	}

	return &proto.Metric{
		CpuUsage:    cpuUsage,
		MemoryUsage: memUsage,
		NetworkIn:   netIn,
		NetworkOut:  netOut,
		Timestamp:   time.Now().UTC().Unix(),
	}, nil
}

// GetCPUUsage 获取CPU使用率(百分比)
func GetCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	if len(percentages) == 0 {
		return 0, fmt.Errorf("无法获取CPU使用率")
	}
	return float64(percentages[0]), nil
}

// GetMemoryUsage 获取内存使用量(MB)
func GetMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	// 转换为MB
	memUsedMB := float64(vmStat.Used) / 1024 / 1024
	return memUsedMB, nil
}

// GetNetworkIO 获取网络IO(KB/s)
func GetNetworkIO(threshold int) (float64, float64, error) {
	// 获取第一次快照
	ioCounters1, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}
	if len(ioCounters1) == 0 {
		return 0, 0, fmt.Errorf("无法获取网络接口信息")
	}

	// 等待一秒以计算速率
	time.Sleep(time.Second * time.Duration(threshold))

	// 获取第二次快照
	ioCounters2, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}
	if len(ioCounters2) == 0 {
		return 0, 0, fmt.Errorf("无法获取网络接口信息")
	}

	// 计算差值并转换为KB/s
	netIn := float64(ioCounters2[0].BytesRecv-ioCounters1[0].BytesRecv) / float64(1024/threshold)
	netOut := float64(ioCounters2[0].BytesSent-ioCounters1[0].BytesSent) / float64(1024/threshold)

	return netIn, netOut, nil
}
