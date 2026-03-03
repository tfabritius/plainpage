package server

import (
	"net/http"
	"runtime"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
)

func (app App) getStats(w http.ResponseWriter, r *http.Request) {
	// Memory stats from Go runtime
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memory := model.MemoryStats{
		Alloc:      memStats.Alloc,
		TotalAlloc: memStats.TotalAlloc,
		Sys:        memStats.Sys,
	}

	// Disk usage from ContentService
	diskUsage := app.Content.GetDiskUsage()

	render.JSON(w, r, model.GetStatsResponse{
		Memory:    memory,
		DiskUsage: diskUsage,
	})
}
