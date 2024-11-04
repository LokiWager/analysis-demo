package service

import (
	"debug/buildinfo"
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/LokiWager/analysis-demo/pkg/logger"
)

type (
	// ServiceConfig is the configuration for service.
	ServiceConfig struct {
		ProcessID int `yaml:"processID" json:"processID"`
	}

	// Service is the service.
	Service struct {
		config  *ServiceConfig
		process *process.Process
	}

	// ProcessInfo is the information of a process.
	ProcessInfo struct {
		Name       string   `json:"name"`
		PID        int      `json:"pid"`
		Ppid       int32    `json:"ppid"`
		Cmd        string   `json:"cmd"`
		Pwd        string   `json:"pwd"`
		Envs       []string `json:"envs"`
		NumFDs     int32    `json:"numFDs"`
		NumThreads int32    `json:"numThreads"`
		UserName   string   `json:"userName"`
		Path       string   `json:"path"`
		GoVersion  string   `json:"goVersion"`
	}

	// Connection is the information of a connection.
	Connection struct {
		SourceIP   string `json:"sourceIP"`
		SourcePort int    `json:"sourcePort"`
		DestIP     string `json:"destIP"`
		DestPort   int    `json:"destPort"`
		State      string `json:"state"`
	}
)

// NewService creates a service.
func NewService(config *ServiceConfig) *Service {
	p, err := process.NewProcess(int32(config.ProcessID))
	if err != nil {
		panic(err)
	}

	return &Service{
		config:  config,
		process: p,
	}
}

func (s *Service) GetProcessInfo(ctx echo.Context) error {
	name, err := s.process.Name()
	if err != nil {
		logger.Warnf("get process name failed: %v", err)
	}

	ppid, err := s.process.Ppid()
	if err != nil {
		logger.Warnf("get process ppid failed: %v", err)
	}

	cmd, err := s.process.Cmdline()
	if err != nil {
		logger.Warnf("get process cmdline failed: %v", err)
	}

	pwd, err := s.process.Cwd()
	if err != nil {
		logger.Warnf("get process cwd failed: %v", err)
	}

	envs, err := s.process.Environ()
	if err != nil {
		logger.Warnf("get process envs failed: %v", err)
	}

	numFDs, err := s.process.NumFDs()
	if err != nil {
		logger.Warnf("get process numFDs failed: %v", err)
	}

	numThreads, err := s.process.NumThreads()
	if err != nil {
		logger.Warnf("get process numThreads failed: %v", err)
	}

	userName, err := s.process.Username()
	if err != nil {
		logger.Warnf("get process username failed: %v", err)
	}

	path, err := s.process.Exe()
	if err != nil {
		logger.Warnf("get process path failed: %v", err)
	}

	goVersion, err := s.getGoVersion(path)
	if err != nil {
		logger.Warnf("get go version failed: %v", err)
	}

	info := &ProcessInfo{
		Name:       name,
		PID:        s.config.ProcessID,
		Ppid:       ppid,
		Cmd:        cmd,
		Pwd:        pwd,
		Envs:       envs,
		NumFDs:     numFDs,
		NumThreads: numThreads,
		UserName:   userName,
		Path:       path,
		GoVersion:  goVersion,
	}

	return ctx.JSON(200, info)
}

func (s *Service) getGoVersion(path string) (string, error) {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return "", err
	}
	return info.GoVersion, nil
}

func (s *Service) GetUsage(ctx echo.Context) error {
	cpuPercent, err := s.process.CPUPercent()
	if err != nil {
		logger.Warnf("get process cpu cpuPercent failed: %v", err)
	}

	memoryPercent, err := s.process.MemoryPercent()
	if err != nil {
		logger.Warnf("get process memory memoryPercent failed: %v", err)
	}

	pageFaults, err := s.process.PageFaults()
	if err != nil {
		logger.Warnf("get process pageFaults failed: %v", err)
	}

	ioCounters, err := s.process.IOCounters()
	if err != nil {
		logger.Warnf("get process ioCounters failed: %v", err)
	}

	memoryInfo, err := s.process.MemoryInfo()
	if err != nil {
		logger.Warnf("get process memoryInfo failed: %v", err)
	}

	cpuTimes, err := s.process.Times()
	if err != nil {
		logger.Warnf("get process cpuTimes failed: %v", err)
	}

	contextSwitches, err := s.process.NumCtxSwitches()
	if err != nil {
		logger.Warnf("get process contextSwitches failed: %v", err)
	}

	return ctx.JSON(200, echo.Map{
		"cpuPercent":    cpuPercent,
		"memoryPercent": memoryPercent,
		"pageFaults":    pageFaults.MajorFaults + pageFaults.MinorFaults + pageFaults.ChildMajorFaults + pageFaults.ChildMinorFaults,
		"readCount":     ioCounters.ReadCount,
		"writeCount":    ioCounters.WriteCount,
		"readBytes":     ioCounters.ReadBytes,
		"writeBytes":    ioCounters.WriteBytes,
		"rss":           memoryInfo.RSS,
		"vms":           memoryInfo.VMS,
		"swap":          memoryInfo.Swap,
		"userTime":      cpuTimes.User,
		"systemTime":    cpuTimes.System,
		"iowait":        cpuTimes.Iowait,
		"contextSwitch": contextSwitches.Involuntary + contextSwitches.Voluntary,
	})
}

func (s *Service) GetOpenFiles(ctx echo.Context) error {
	files, err := s.process.OpenFiles()
	if err != nil {
		logger.Warnf("get process open files failed: %v", err)
	}

	return ctx.JSON(200, files)
}

func (s *Service) GetConnections(ctx echo.Context) error {
	conns, err := s.process.Connections()
	if err != nil {
		logger.Warnf("get process connections failed: %v", err)
	}

	connections := make([]Connection, 0, len(conns))
	for _, conn := range conns {
		connections = append(connections, Connection{
			SourceIP:   conn.Laddr.IP,
			SourcePort: int(conn.Laddr.Port),
			DestIP:     conn.Raddr.IP,
			DestPort:   int(conn.Raddr.Port),
			State:      conn.Status,
		})
	}

	return ctx.JSON(200, connections)
}
