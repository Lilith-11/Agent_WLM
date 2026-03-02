package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"syscall"
)

type DiskInfo struct {
	Device  string  `json:"device"`
	FSType  string  `json:"fs_type"`
	TotalGB float64 `json:"total_gb"`
	UsedGB  float64 `json:"used_gb"`
	FreeGB  float64 `json:"free_gb"`
	Percent float64 `json:"percent_used"`
}

type CPUInfo struct {
	Manufacturer     string    `json:"manufacturer"`
	Model            string    `json:"model"`
	SpeedMHz         string    `json:"speed_mhz"`
	PhysicalCores    int       `json:"physical_cores"`
	LogicalCores     int       `json:"logical_cores"`
	HyperThreading   bool      `json:"hyper_threading"`
	UsagePerCore     []float64 `json:"usage_per_core"`
	AverageCPUUsage  float64   `json:"average_cpu_usage"`
}

type AgentInfo struct {
	AgentID      int        `json:"agent_id"`
	Hostname     string     `json:"hostname"`
	OS           string     `json:"os"`
	Architecture string     `json:"architecture"`
	BootTime     time.Time  `json:"boot_time"`
	TotalRAMGB   float64    `json:"total_ram_gb"`
	CPU          CPUInfo    `json:"cpu"`
	Disks        []DiskInfo `json:"disks"`
}

func main() {
	if runtime.GOOS != "linux" {
		fmt.Println("This agent supports Linux only.")
		return
	}

	rand.Seed(time.Now().UnixNano())

	agent := AgentInfo{}
	agent.AgentID = rand.Intn(100000)
	agent.Architecture = runtime.GOARCH
	agent.OS = getOSPrettyName()

	hostname, _ := os.Hostname()
	agent.Hostname = hostname

	agent.BootTime = getBootTime()
	agent.TotalRAMGB = getTotalRAM()
	agent.CPU = getCPUInfo()
	agent.Disks = getDisks()
	saveJSON(agent)

	fmt.Println("Agent data collected successfully.")
}

func getOSPrettyName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(line[13:], `"`)
		}
	}
	return "Linux"
}

func getBootTime() time.Time {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "btime") {
			fields := strings.Fields(line)
			sec, _ := strconv.ParseInt(fields[1], 10, 64)
			return time.Unix(sec, 0)
		}
	}
	return time.Time{}
}

func getTotalRAM() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			kb, _ := strconv.ParseFloat(fields[1], 64)
			return kb / (1024 * 1024)
		}
	}
	return 0
}

func getCPUInfo() CPUInfo {
	data, _ := os.ReadFile("/proc/cpuinfo")
	lines := strings.Split(string(data), "\n")

	var manufacturer, model, speed string
	physicalCores := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "vendor_id") && manufacturer == "" {
			manufacturer = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "model name") && model == "" {
			model = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "cpu MHz") && speed == "" {
			speed = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "cpu cores") {
			val, _ := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1]))
			physicalCores = val
		}
	}

	logical := runtime.NumCPU()
	hyper := logical > physicalCores

	usages, avg := getCPUUsage()

	return CPUInfo{
		Manufacturer:    manufacturer,
		Model:           model,
		SpeedMHz:        speed,
		PhysicalCores:   physicalCores,
		LogicalCores:    logical,
		HyperThreading:  hyper,
		UsagePerCore:    usages,
		AverageCPUUsage: avg,
	}
}

func getCPUUsage() ([]float64, float64) {
	readStat := func() [][]uint64 {
		data, _ := os.ReadFile("/proc/stat")
		var stats [][]uint64
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "cpu") && len(line) > 3 && line[3] != ' ' {
				fields := strings.Fields(line)[1:]
				var vals []uint64
				for _, f := range fields {
					v, _ := strconv.ParseUint(f, 10, 64)
					vals = append(vals, v)
				}
				stats = append(stats, vals)
			}
		}
		return stats
	}

	stat1 := readStat()
	time.Sleep(200 * time.Millisecond)
	stat2 := readStat()

	var usages []float64
	var total float64

	for i := range stat1{
		idle1 := stat1[i][3]
		total1 := sum(stat1[i])
		idle2 := stat2[i][3]
		total2 := sum(stat2[i])

		deltaTotal := float64(total2 - total1)
		deltaIdle := float64(idle2 - idle1)

		if deltaTotal == 0 {
			usages = append(usages, 0)
			continue
		}

		usage := (deltaTotal - deltaIdle) / deltaTotal * 100
		usages = append(usages, usage)
		total += usage
	}
   
	avg := total / float64(len(usages))
	return usages, avg
}

func sum(vals []uint64) uint64 {
	var s uint64
	for _, v := range vals {
		s += v
	}
	return s
}

func getDisks() []DiskInfo {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}

	var disks []DiskInfo

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		if !strings.HasPrefix(fields[0], "/dev/") {
			continue
		}

		var stat syscallStatfs
		if err := statfs(fields[1], &stat); err != nil {
			continue
		}

		total := float64(stat.Blocks*uint64(stat.Bsize)) / (1024 * 1024 * 1024)
		free := float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024)
		used := total - free
		percent := (used / total) * 100

		disks = append(disks, DiskInfo{
			Device:  fields[0],
			FSType:  fields[2],
			TotalGB: total,
			UsedGB:  used,
			FreeGB:  free,
			Percent: percent,
		})
	}

	return disks
}

type syscallStatfs struct {
	Type    int64
	Bsize   int64
	Blocks  uint64
	Bfree   uint64
	Bavail  uint64
}

func statfs(path string, stat *syscallStatfs) error {
	var s syscall.Statfs_t
	err := syscall.Statfs(path, &s)
	if err != nil {
		return err
	}
	stat.Type = int64(s.Type)
	stat.Bsize = int64(s.Bsize)
	stat.Blocks = s.Blocks
	stat.Bfree = s.Bfree
	stat.Bavail = s.Bavail
	return nil
}

func saveJSON(agent AgentInfo) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "agent.json")

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(agent)
}