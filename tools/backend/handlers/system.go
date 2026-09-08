package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// ===== CPU 采样（逐核统计，取所有核的平均使用率）=====
type cpuCore struct {
	idle  uint64 // idle + iowait
	total uint64 // 全部时间
}

var (
	cpuMu          sync.Mutex
	prevPerCore    []cpuCore // 上一帧每核样本
	prevSampleTime time.Time
)

// readPerCoreCPUStat 解析 /proc/stat 的 cpu0/cpu1/... 各核累计时间
func readPerCoreCPUStat() ([]cpuCore, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}
	var cores []cpuCore
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || !strings.HasPrefix(fields[0], "cpu") || fields[0] == "cpu" {
			continue // 跳过聚合行 "cpu"
		}
		var vals [8]uint64
		for i := 1; i < len(fields) && i <= 8; i++ {
			vals[i-1], _ = strconv.ParseUint(fields[i], 10, 64)
		}
		idle := vals[3] + vals[4] // idle + iowait
		total := uint64(0)
		for i := 0; i < 8; i++ {
			total += vals[i]
		}
		cores = append(cores, cpuCore{idle: idle, total: total})
	}
	if len(cores) == 0 {
		return nil, fmt.Errorf("no per-core cpu lines in /proc/stat")
	}
	return cores, nil
}

// cpuUsage 返回所有核平均使用率 0~100；同时回传每核明细。
// 首次调用内部短采样两次保证首帧有值。
func cpuUsage() (avg float64, perCore []float64) {
	cpuMu.Lock()
	defer cpuMu.Unlock()

	now, err := readPerCoreCPUStat()
	if err != nil {
		return -1, nil
	}
	if prevPerCore == nil || len(prevPerCore) != len(now) {
		prevPerCore = now
		prevSampleTime = time.Now()
		time.Sleep(150 * time.Millisecond)
		now, err = readPerCoreCPUStat()
		if err != nil {
			return -1, nil
		}
	}
	dt := time.Since(prevSampleTime).Seconds()
	if dt <= 0 {
		dt = 0.001
	}

	perCore = make([]float64, len(now))
	sum := 0.0
	for i := 0; i < len(now); i++ {
		dIdle := float64(now[i].idle) - float64(prevPerCore[i].idle)
		dTotal := float64(now[i].total) - float64(prevPerCore[i].total)
		u := 0.0
		if dTotal > 0 {
			u = (1 - dIdle/dTotal) * 100
		}
		if u < 0 {
			u = 0
		}
		if u > 100 {
			u = 100
		}
		perCore[i] = u
		sum += u
	}
	prevPerCore = now
	prevSampleTime = time.Now()

	avg = sum / float64(len(now))
	if avg < 0 {
		avg = 0
	}
	if avg > 100 {
		avg = 100
	}
	return avg, perCore
}

func cpuCores() int {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}
	return count
}

// ===== vcgencmd（树莓派）=====
func vcgencmdFloat(args ...string) float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/vcgencmd", args...)
	out, err := cmd.Output()
	if err != nil {
		return -1
	}
	s := strings.TrimSpace(string(out))
	if idx := strings.IndexByte(s, '='); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSuffix(s, "'C")
	s = strings.TrimSuffix(s, "V")
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return -1
	}
	return f
}

func vcgencmdRaw(args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/vcgencmd", args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// SystemHandler returns live server stats.
type SystemHandler struct{}

// GetSystem returns CPU/mem/voltage/temp/freq/load/disk/uptime.
// GET /api/system
func (h *SystemHandler) GetSystem(c *gin.Context) {
	// 内存
	memTotal, memAvail := readMemInfo()

	// 磁盘（根分区）
	disk := readDisk("/")

	// 负载
	load := readLoadAvg()

	// 运行时间
	uptime := readUptime()

	// CPU（逐核平均）
	cpuAvg, cpuPerCore := cpuUsage()

	resp := gin.H{
		"ok":   true,
		"time": time.Now().Format("2006-01-02 15:04:05"),
		"cpu": gin.H{
			"usage":    cpuAvg,
			"cores":    cpuCores(),
			"per_core": cpuPerCore,
		},
		"mem": gin.H{
			"total_gb":     round1(float64(memTotal) / 1024 / 1024),
			"available_gb": round1(float64(memAvail) / 1024 / 1024),
			"used_gb":      round1(float64(memTotal-memAvail) / 1024 / 1024),
			"usage":        pct(memTotal, memAvail),
		},
		"temp_c":  vcgencmdFloat("measure_temp"),
		"voltage": vcgencmdFloat("measure_volts"),
		"freq": gin.H{
			"arm_hz":  vcgencmdRaw("measure_clock", "arm"),
			"core_hz": vcgencmdRaw("measure_clock", "core"),
		},
		"throttled": vcgencmdRaw("get_throttled"),
		"load":      load,
		"uptime_s":  uptime,
		"disk":      disk,
	}

	c.JSON(http.StatusOK, resp)
}

// readMemInfo returns (total, available) in KiB from /proc/meminfo.
func readMemInfo() (uint64, uint64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	var total, avail uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			total = val
		case "MemAvailable:":
			avail = val
		}
	}
	if avail == 0 {
		avail = total
	}
	return total, avail
}

// readLoadAvg returns [1m, 5m, 15m] from /proc/loadavg.
func readLoadAvg() []float64 {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return []float64{-1, -1, -1}
	}
	fields := strings.Fields(string(data))
	out := make([]float64, 3)
	for i := 0; i < 3 && i < len(fields); i++ {
		out[i], _ = strconv.ParseFloat(fields[i], 64)
	}
	return out
}

// readUptime returns uptime in seconds from /proc/uptime.
func readUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return -1
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return -1
	}
	up, _ := strconv.ParseFloat(fields[0], 64)
	return up
}

// readDisk returns usage of the given path via statfs.
func readDisk(path string) gin.H {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return gin.H{"total_gb": -1, "free_gb": -1, "used_gb": -1, "usage": -1}
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	used := total - free
	return gin.H{
		"total_gb": round1(float64(total) / 1024 / 1024 / 1024),
		"free_gb":  round1(float64(free) / 1024 / 1024 / 1024),
		"used_gb":  round1(float64(used) / 1024 / 1024 / 1024),
		"usage":    pct(total, free),
	}
}

func pct(total, free uint64) float64 {
	if total == 0 {
		return 0
	}
	used := float64(total) - float64(free)
	if used < 0 {
		used = 0
	}
	p := used / float64(total) * 100
	if p > 100 {
		p = 100
	}
	return p
}

func round1(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}
