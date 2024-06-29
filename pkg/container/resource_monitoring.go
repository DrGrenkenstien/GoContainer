package container

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// ReadFile reads the content of a file and converts it to a string value.
func ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetCPUUsage returns the CPU usage in user mode and kernel mode from /proc/[pid]/stat
func GetCPUUsage(pid int) (int64, int64, error) {

	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := ReadFile(statFile)
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(data)
	utime, err := strconv.ParseInt(fields[13], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	stime, err := strconv.ParseInt(fields[14], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return utime, stime, nil
}

// GetMemoryUsage returns the memory usage from /proc/[pid]/status
func GetMemoryUsage(pid int) (int64, error) {
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	data, err := ReadFile(statusFile)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			value, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return value * 1024, nil // Convert from kB to bytes
		}
	}
	return 0, fmt.Errorf("VmRSS not found in %s", statusFile)
}

func run(pid int) {
	utime, stime, err := GetCPUUsage(pid)
	if err != nil {
		log.Fatalf("Failed to get CPU usage: %v", err)
	}
	fmt.Printf("CPU usage in clock ticks (user mode): %d\n", utime)
	fmt.Printf("CPU usage in clock ticks (kernel mode): %d\n", stime)

	memoryUsage, err := GetMemoryUsage(pid)
	if err != nil {
		log.Fatalf("Failed to get memory usage: %v", err)
	}
	fmt.Printf("Memory usage: %d bytes\n", memoryUsage)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: <program> <process id>")
		os.Exit(1)
	}

	arg1, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error passing argument: %v\n", err)
		return
	}

	run(arg1)
}
