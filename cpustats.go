package main

import (
	"fmt"
	"github.com/docker/engine-api/types"
	ui "github.com/gizak/termui"
)

const CPU_RANGE_SIZE = 600
const (
	cpu_graph_right_pad = 9
	cpu_graph_top_pad   = 1
	cpu_graph_height    = 10
)

type CpuUsageWidget struct {
	Views   []ui.GridBufferer
	Handler func(ui.Event)
}

type CPUUsagePercent struct {
	Pct  float64
	Data []float64
}

func NewCpuUsageWidget() *CpuUsageWidget {
	cpuGraph := ui.NewLineChart()
	cpuGraph.BorderLabel = "CPU Usage"
	cpuGraph.Height = cpu_graph_height
	cpuGraph.PaddingTop = cpu_graph_top_pad
	cpuGraph.PaddingRight = cpu_graph_right_pad
	var currentCPUUsage = uint64(0)
	var currentSystemUsage = uint64(0)
	var cpuHistory = make([]float64, CPU_RANGE_SIZE)
	var cpuHead = 0
	return &CpuUsageWidget{Views: []ui.GridBufferer{cpuGraph}, Handler: func(e ui.Event) {
		stats := e.Data.(types.StatsJSON)
		var cpuPct = 0.0
		cpuPct, currentCPUUsage, currentSystemUsage = computeCpu(stats, currentCPUUsage, currentSystemUsage)
		if cpuHead < len(cpuHistory)-1 {
			cpuHead = cpuHead + 1
		} else {
			cpuHead = 0
			//reset the data range
			cpuHistory = make([]float64, CPU_RANGE_SIZE)
		}
		numPoints := computeNumPoints(cpuGraph)
		cpuHistory[cpuHead] = cpuPct * 100.0
		cpuGraph.BorderLabel = fmt.Sprintf("CPU Usage: %5.2f%%", cpuPct*100)
		cpuGraph.Data = getDataRange(cpuGraph, cpuHistory, cpuHead)
		cpuGraph.DataLabels = computeLabels(cpuHead, numPoints)
	}}
}

func computeLabels(head int, numPoints int) []string {
	result := make([]string, numPoints)
	var offset = 0
	if head > len(result) {
		offset = head - numPoints
	}
	for i := 0; i < len(result); i++ {
		result[i] = fmt.Sprintf("%d", offset+i)
	}
	return result
}
func computeNumPoints(lc *ui.LineChart) int {
	//2x multiplier if using braille mode
	padding := cpu_graph_right_pad * 2
	return (lc.Width - padding) * 2
}
func getDataRange(lc *ui.LineChart, data []float64, head int) []float64 {
	points := computeNumPoints(lc)
	if head < points {
		return data[:points]
	}
	var end int
	if head+points < len(data) {
		end = head
	} else {
		end = len(data) - 1
	}
	return data[(head - points):end]
}
func computeCpu(stats types.StatsJSON, currentUsage uint64, currentSystemUsage uint64) (cpuPct float64, cpuUsage uint64, systemUsage uint64) {
	//compute the cpu usage percentage
	//via https://github.com/docker/docker/blob/e884a515e96201d4027a6c9c1b4fa884fc2d21a3/api/client/container/stats_helpers.go#L199-L212
	newCpuUsage := stats.CPUStats.CPUUsage.TotalUsage
	newSystemUsage := stats.CPUStats.SystemUsage
	cpuDiff := float64(newCpuUsage) - float64(currentUsage)
	systemDiff := float64(newSystemUsage) - float64(currentSystemUsage)
	return cpuDiff / systemDiff * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)), newCpuUsage, newSystemUsage

}
