package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mylxsw/coll"
	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/storage"
)

type HistoryController struct {
	cc *container.Container
}

func NewHistoryController(cc *container.Container) Controller {
	return &HistoryController{cc: cc}
}

func (h *HistoryController) Register(router *hades.Router) {
	router.Group("/histories", func(router *hades.Router) {
		router.Get("/", h.Recently)
	})
}

// Recently 最近的同步任务历史记录
func (h *HistoryController) Recently(ctx *hades.WebContext, req *hades.Request, historyStore storage.JobHistoryStore) hades.HTTPResponse {
	limit := req.IntInput("limit", 10)
	if limit <= 0 || limit > 100 {
		return ctx.JSONError("invalid limit argument", http.StatusUnprocessableEntity)
	}

	items, err := historyStore.Recently(limit)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(coll.MustNew(items).Map(func(item storage.JobHistoryItem) map[string]interface{} {
		job := queue.FileSyncJob{}
		job.Decode(item.Payload)

		var col collector.Collector
		_ = json.Unmarshal(item.Output, &col)

		return map[string]interface{}{
			"id":         item.ID,
			"name":       item.Name,
			"status":     item.Status,
			"created_at": item.CreatedAt.Format(time.RFC3339),
			"job":        job,
			"output":     col,
		}
	}).Items())
}
