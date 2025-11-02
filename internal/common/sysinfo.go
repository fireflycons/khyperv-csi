package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// LatestRelease holds minimal release info from GitHub API.
type LatestRelease struct {
	Version   string    `json:"tag_name"`
	Published time.Time `json:"published_at"`
}

type ApplicationInfo struct {
	Name          string         `json:"name"`
	Version       string         `json:"version"`
	CommitHash    string         `json:"commit_hash"`
	BuildDate     string         `json:"build_date"`
	LatestRelease *LatestRelease `json:"latest_release,omitempty"`
}

type RuntimeInfo struct {
	GoVersion string `json:"go_version"`
	GOOS      string `json:"goos"`
	GOARCH    string `json:"goarch"`
	NumCPU    int    `json:"num_cpu"`
	Compiler  string `json:"compiler"`
}

type OSInfo struct {
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	Hypervisor      string `json:"hypervisor,omitempty"`
}

type CPUInfo struct {
	ModelName  string  `json:"model_name"`
	Sockets    int     `json:"sockets"`
	Threads    int     `json:"threads"`
	Mhz        float64 `json:"mhz_per_core"`
	Hypervisor string  `json:"hypervisor,omitempty"`
}

type MemoryInfo struct {
	TotalGB     float64 `json:"total_gb"`
	AvailableGB float64 `json:"available_gb"`
	UsedGB      float64 `json:"used_gb"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskPartition struct {
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	TotalGB     float64 `json:"total_gb"`
	FreeGB      float64 `json:"free_gb"`
	UsedGB      float64 `json:"used_gb"`
	UsedPercent float64 `json:"used_percent"`
}

type SystemReport struct {
	Application ApplicationInfo `json:"application"`
	Runtime     RuntimeInfo     `json:"runtime"`
	OS          OSInfo          `json:"os"`
	CPU         CPUInfo         `json:"cpu"`
	Memory      MemoryInfo      `json:"memory"`
	Disks       []DiskPartition `json:"disks"`
}

func GetSystemInfo() SystemReport {
	return SystemReport{
		Application: collectApplicationInfo(),
		Runtime:     collectRuntimeInfo(),
		OS:          collectOSInfo(),
		CPU:         collectCPUInfo(),
		Memory:      collectMemoryInfo(),
		Disks:       collectDiskInfo(),
	}
}

// -------- Collectors --------

func collectReleaseInfo() *LatestRelease {

	getLatest := func(owner, repo string) (*LatestRelease, error) {
		const reqTimeout = 5 * time.Second
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		client := &http.Client{
			// Default http.Client follows redirects automatically.
			Timeout: 10 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("perform request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status: %s", resp.Status)
		}

		var release LatestRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		return &release, nil
	}

	r, err := getLatest("fireflycons", constants.RepoName)
	if err != nil {
		fmt.Printf("Unable to get latest release info: %v\n\n", err)
		return nil
	}

	return r
}

func collectApplicationInfo() ApplicationInfo {
	return ApplicationInfo{
		Name:          filepath.Base(os.Args[0]),
		Version:       Version,
		BuildDate:     BuildDate,
		CommitHash:    CommitHash,
		LatestRelease: collectReleaseInfo(),
	}
}

func collectRuntimeInfo() RuntimeInfo {
	return RuntimeInfo{
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		Compiler:  runtime.Compiler,
	}
}

func collectOSInfo() OSInfo {
	info, _ := host.Info()

	osinfo := OSInfo{
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
	}
	if info.VirtualizationSystem != "" {
		osinfo.Hypervisor = fmt.Sprintf("%s (%s)", info.VirtualizationSystem, info.VirtualizationRole)
	}
	return osinfo
}

func collectCPUInfo() CPUInfo {
	cpus, err := cpu.Info()
	if err != nil || len(cpus) == 0 {
		return CPUInfo{}
	}

	model := strings.TrimSpace(cpus[0].ModelName)
	if model == "" {
		model = "Unknown"
	}

	socketSet := make(map[string]struct{})

	for i := range cpus {
		socketSet[cpus[i].PhysicalID] = struct{}{}
	}

	sockets := len(socketSet)
	if sockets == 0 {
		sockets = 1
	}

	return CPUInfo{
		ModelName: model,
		Sockets:   sockets,
		Threads:   int(cpus[0].Cores),
		Mhz:       cpus[0].Mhz,
	}
}

func collectMemoryInfo() MemoryInfo {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}
	}
	return MemoryInfo{
		TotalGB:     float64(vm.Total) / (1024 * 1024 * 1024),
		AvailableGB: float64(vm.Available) / (1024 * 1024 * 1024),
		UsedGB:      float64(vm.Used) / (1024 * 1024 * 1024),
		UsedPercent: vm.UsedPercent,
	}
}

func collectDiskInfo() []DiskPartition {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil
	}

	ignoreFSTypes := map[string]struct{}{
		"proc": {}, "sysfs": {}, "devtmpfs": {}, "tmpfs": {},
		"overlay": {}, "squashfs": {}, "autofs": {}, "cgroup2": {},
		"nsfs": {}, "debugfs": {}, "tracefs": {}, "securityfs": {},
		"pstore": {}, "efivarfs": {}, "mqueue": {}, "rpc_pipefs": {},
		"fusectl": {}, "binfmt_misc": {}, "ramfs": {}, "selinuxfs": {},
		"hugetlbfs": {}, "bpf": {}, "configfs": {}, "devpts": {},
	}

	var out []DiskPartition
	for _, p := range partitions {
		fs := strings.ToLower(p.Fstype)
		if _, skip := ignoreFSTypes[fs]; skip {
			continue
		}

		// Skip floppy/CD drives on Windows
		if runtime.GOOS == "windows" {
			if strings.HasPrefix(strings.ToUpper(p.Device), "A:") ||
				strings.HasPrefix(strings.ToUpper(p.Device), "B:") {
				continue
			}
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		out = append(out, DiskPartition{
			Mountpoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			TotalGB:     float64(usage.Total) / (1024 * 1024 * 1024),
			FreeGB:      float64(usage.Free) / (1024 * 1024 * 1024),
			UsedGB:      float64(usage.Used) / (1024 * 1024 * 1024),
			UsedPercent: usage.UsedPercent,
		})
	}

	return out
}

// -------- Rendering --------

func (r SystemReport) String() string {
	var sb strings.Builder

	sb.WriteString("===== System Information Report =====\n")
	latestVersion := r.Application.Version
	if r.Application.LatestRelease != nil {
		latestVersion = strings.TrimLeft(r.Application.LatestRelease.Version, "v")
	}

	outdated := func() string {
		if latestVersion != r.Application.Version {
			return " (newer version available)"
		}
		return ""
	}()

	sb.WriteString("\n--- Application Info ---\n")
	sb.WriteString(fmt.Sprintf("Application:    %s\n", r.Application.Name))
	sb.WriteString(fmt.Sprintf("Version:        %s%s\n", r.Application.Version, outdated))
	sb.WriteString(fmt.Sprintf("Commit Hash:    %s\n", r.Application.CommitHash))
	sb.WriteString(fmt.Sprintf("Build Date:     %s\n", r.Application.BuildDate))
	if r.Application.LatestRelease != nil {
		sb.WriteString("Latest Release:\n")
		sb.WriteString(fmt.Sprintf("  Version:        %s\n", strings.TrimLeft(r.Application.LatestRelease.Version, "v")))
		sb.WriteString(fmt.Sprintf("  Publish Date:   %s\n", r.Application.LatestRelease.Published.UTC().Format("Mon Jan 02 15:04:05 MST 2006")))
	}

	sb.WriteString("\n--- Runtime Info ---\n")
	sb.WriteString(fmt.Sprintf("Go Version:     %s\n", r.Runtime.GoVersion))
	sb.WriteString(fmt.Sprintf("GOOS/GOARCH:    %s/%s\n", r.Runtime.GOOS, r.Runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("NumCPU:         %d\n", r.Runtime.NumCPU))
	sb.WriteString(fmt.Sprintf("Compiler:       %s\n", r.Runtime.Compiler))

	sb.WriteString("\n--- Operating System ---\n")
	sb.WriteString(fmt.Sprintf("OS:             %s\n", r.OS.OS))
	sb.WriteString(fmt.Sprintf("Platform:       %s %s (%s)\n", r.OS.Platform, r.OS.PlatformVersion, r.OS.PlatformFamily))
	sb.WriteString(fmt.Sprintf("Kernel Version: %s\n", r.OS.KernelVersion))
	if r.OS.Hypervisor != "" {
		sb.WriteString(fmt.Sprintf("Hypervisor:     %s\n", r.OS.Hypervisor))
	}

	sb.WriteString("\n--- CPU Info ---\n")
	sb.WriteString(fmt.Sprintf("CPU Model:      %s\n", r.CPU.ModelName))
	sb.WriteString(fmt.Sprintf("Sockets:        %d\n", r.CPU.Sockets))
	sb.WriteString(fmt.Sprintf("Threads:        %d\n", r.CPU.Threads))
	sb.WriteString(fmt.Sprintf("Mhz per core:   %.2f\n", r.CPU.Mhz))

	sb.WriteString("\n--- Memory Info ---\n")
	sb.WriteString(fmt.Sprintf("Total:          %.2f GB\n", r.Memory.TotalGB))
	sb.WriteString(fmt.Sprintf("Available:      %.2f GB\n", r.Memory.AvailableGB))
	sb.WriteString(fmt.Sprintf("Used:           %.2f GB (%.2f%%)\n",
		r.Memory.UsedGB, r.Memory.UsedPercent))

	sb.WriteString("\n--- Disk Info ---\n")
	for _, d := range r.Disks {
		sb.WriteString(fmt.Sprintf("Mountpoint:     %s\n", d.Mountpoint))
		sb.WriteString(fmt.Sprintf("  Filesystem:   %s\n", d.Fstype))
		sb.WriteString(fmt.Sprintf("  Total:        %.2f GB\n", d.TotalGB))
		sb.WriteString(fmt.Sprintf("  Free:         %.2f GB\n", d.FreeGB))
		sb.WriteString(fmt.Sprintf("  Used:         %.2f GB (%.2f%%)\n\n",
			d.UsedGB, d.UsedPercent))
	}

	return sb.String()
}
