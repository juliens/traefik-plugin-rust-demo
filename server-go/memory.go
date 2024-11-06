package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func formatBytes(bytes uint64) string {
	const (
		B  = 1
		KB = B * 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func monitorMemory(interval time.Duration, duration *time.Duration) {
	fmt.Println("\nDébut du monitoring mémoire:")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("%-25s %-20s %-20s %-20s %s\n",
		"Timestamp",
		"Heap Alloc",
		"Total Alloc",
		"Sys Mem",
		"Num GC")
	fmt.Println(strings.Repeat("-", 100))

	startTime := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)

			fmt.Printf("%-25s %-20s %-20s %-20s %d\n",
				time.Now().Format("2006-01-02 15:04:05"),
				formatBytes(stats.HeapAlloc),
				formatBytes(stats.TotalAlloc),
				formatBytes(stats.Sys),
				stats.NumGC,
			)

			if duration != nil {
				if time.Since(startTime) >= *duration {
					fmt.Println(strings.Repeat("-", 100))
					fmt.Println("Fin du monitoring mémoire (durée atteinte)")
					return
				}
			}
		}
	}
}
