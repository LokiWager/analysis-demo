package service

import (
	"debug/buildinfo"
	"github.com/LokiWager/analysis-demo/pkg/logger"
)

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

func (s *Service) getGoVersion(path string) (string, error) {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return "", err
	}
	return info.GoVersion, nil
}
