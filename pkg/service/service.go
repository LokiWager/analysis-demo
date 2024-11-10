package service

import (
	"debug/buildinfo"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/LokiWager/analysis-demo/pkg/logger"
)

const (
	DefaultTimeout       = 30
	DefaultTraceFilePath = "./trace"
)

type (
	// ServiceConfig is the configuration for service.
	ServiceConfig struct {
		ProcessID int  `yaml:"processID" json:"processID"`
		Persist   bool `yaml:"persist" json:"persist"`
	}

	// Service is the service.
	Service struct {
		config        *ServiceConfig
		process       *process.Process
		restyClient   *resty.Client
		detectEMA     *EMA
		stopCh        chan struct{}
		traceFilePath string
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

	client := resty.New().SetTimeout(DefaultTimeout * time.Second)
	// check or mkdir trace file path
	if _, err := os.Stat(DefaultTraceFilePath); os.IsNotExist(err) {
		err = os.Mkdir(DefaultTraceFilePath, 0755)
		if err != nil {
			logger.Fatalf("mkdir trace file path failed: %v", err)
		}
	}

	return &Service{
		config:        config,
		process:       p,
		restyClient:   client,
		detectEMA:     NewEMA(DefaultAlpha, DefaultThreshold),
		stopCh:        make(chan struct{}),
		traceFilePath: DefaultTraceFilePath,
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
	result := s.collectMetrics()
	return ctx.JSON(200, result)
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

func (s *Service) GetProfile(ctx echo.Context) error {
	// curl -o trace.out http://localhost:6060/debug/pprof/trace\?seconds\=5
	// go tool trace trace.out
	fileName, err := s.dumpTraceFile()
	if err != nil {
		logger.Warnf("dump trace file failed: %v", err)
		return ctx.NoContent(500)
	}

	if _, err := exec.LookPath("go"); err != nil {
		logger.Warnf("go command not found: %v", err)
		return ctx.NoContent(500)
	}

	go func() {
		err = exec.Command("go", "tool", "trace", "-http=:8891", fileName).Run()
		if err != nil {
			logger.Warnf("run go tool trace failed: %v", err)
		}
		defer func() {
			_ = os.Remove(fileName)
		}()
	}()

	return ctx.String(200, "view profile at http://[::]:8891/")
}

func (s *Service) Close() {
	close(s.stopCh)
}

func (s *Service) dumpTraceFile() (string, error) {
	// create file to trace dir with name trace-<date>.out
	fileName := fmt.Sprintf("trace-%d.out", time.Now().UnixMilli())
	tmpFile, err := os.Create(fmt.Sprintf("%s/%s", s.traceFilePath, fileName))
	if err != nil {
		logger.Warnf("create tmp file failed: %v", err)
		return "", err
	}
	fileName = fmt.Sprintf("%s/%s", s.traceFilePath, fileName)

	// write to file
	resp, err := s.restyClient.R().
		SetQueryParam("seconds", "5").
		Execute("GET", "http://localhost:6060/debug/pprof/trace")
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		logger.Warnf("get profile failed: %v", err)
		return "", err
	}

	traceData := resp.Body()
	err = os.WriteFile(fileName, traceData, 0644)
	if err != nil {
		logger.Warnf("write trace data to file failed: %v", err)
		return "", err
	}

	return fileName, nil
}

func (s *Service) dumpTraceFileToMDB() error {
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//
	//traceCollection := mongodbtool.GetCollection("trace")
	//
	//fileName, err := s.dumpTraceFile()
	//if err != nil {
	//	logger.Warnf("dump trace file failed: %v", err)
	//	return err
	//}
	//
	//file, err := os.Open(fileName)
	//if err != nil {
	//	logger.Warnf("open trace file failed: %v", err)
	//	return err
	//}
	//
	//defer func() {
	//	_ = file.Close()
	//	_ = os.Remove(fileName)
	//}()
	//
	//// read file content to string
	//buf := make([]byte, 1024)
	//n, err := file.Read(buf)
	//if err != nil {
	//	logger.Warnf("read trace file failed: %v", err)
	//	return err
	//}
	return nil
}

func (s *Service) collectMetrics() map[string]interface{} {
	result := map[string]interface{}{}

	cpuPercent, err := s.process.CPUPercent()
	if err != nil {
		logger.Warnf("get process cpu cpuPercent failed: %v", err)
	} else {
		result["cpuPercent"] = cpuPercent
	}

	memoryPercent, err := s.process.MemoryPercent()
	if err != nil {
		logger.Warnf("get process memory memoryPercent failed: %v", err)
	} else {
		result["memoryPercent"] = memoryPercent
	}

	pageFaults, err := s.process.PageFaults()
	if err != nil {
		logger.Warnf("get process pageFaults failed: %v", err)
	} else {
		result["pageFaults"] = pageFaults.MajorFaults + pageFaults.MinorFaults + pageFaults.ChildMajorFaults + pageFaults.ChildMinorFaults
	}

	ioCounters, err := s.process.IOCounters()
	if err != nil {
		logger.Warnf("get process ioCounters failed: %v", err)
	} else {
		result["readCount"] = ioCounters.ReadCount
		result["writeCount"] = ioCounters.WriteCount
		result["readBytes"] = ioCounters.ReadBytes
		result["writeBytes"] = ioCounters.WriteBytes
	}

	memoryInfo, err := s.process.MemoryInfo()
	if err != nil {
		logger.Warnf("get process memoryInfo failed: %v", err)
	} else {
		result["rss"] = memoryInfo.RSS
		result["vms"] = memoryInfo.VMS
		result["swap"] = memoryInfo.Swap
	}

	cpuTimes, err := s.process.Times()
	if err != nil {
		logger.Warnf("get process cpuTimes failed: %v", err)
	} else {
		result["userTime"] = cpuTimes.User
		result["systemTime"] = cpuTimes.System
		result["iowait"] = cpuTimes.Iowait
	}

	contextSwitches, err := s.process.NumCtxSwitches()
	if err != nil {
		logger.Warnf("get process contextSwitches failed: %v", err)
	} else {
		result["contextSwitch"] = contextSwitches.Involuntary + contextSwitches.Voluntary
	}

	return result
}
