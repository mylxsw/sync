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
		router.Get("/{id}/", h.Item)
	})
}

type History struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Status    string              `json:"status"`
	CreatedAt time.Time           `json:"created_at"`
	Job       queue.FileSyncJob   `json:"job"`
	Output    collector.Collector `json:"output"`
}

// Recently 最近的同步任务历史记录
// @Summary 查询最近的文件同步记录
// @Tags Histories
// @Param limit query int false "返回记录数目"
// @Success 200 {array} controller.History
// @Router /histories/ [get]
func (h *HistoryController) Recently(ctx *hades.WebContext, req *hades.HttpRequest, historyStore storage.JobHistoryStore) hades.HTTPResponse {
	limit := req.IntInput("limit", 10)
	if limit <= 0 || limit > 100 {
		return ctx.JSONError("invalid limit argument", http.StatusUnprocessableEntity)
	}

	items, err := historyStore.Recently(limit)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(coll.MustNew(items).Map(func(item storage.JobHistoryItem) History {
		job := queue.FileSyncJob{}
		job.Decode(item.Payload)

		return History{
			ID:        item.ID,
			Name:      item.Name,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			Job:       job,
		}
	}).Items())
}

// Item 返回指定ID的历史纪录详情
// @Summary 返回指定ID的历史纪录详情
// @Tags Histories
// @Param id path string true "记录ID"
// @Success 200 {object} controller.History
// @Router /histories/{id}/ [get]
func (h *HistoryController) Item(ctx *hades.WebContext, req *hades.HttpRequest, historyStore storage.JobHistoryStore) hades.HTTPResponse {
	id := req.PathVar("id")
	if id == "" {
		return ctx.JSONError("invalid argument id", http.StatusUnprocessableEntity)
	}

	items, err := historyStore.Recently(100)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	for _, item := range items {
		if item.ID == id {
			job := queue.FileSyncJob{}
			job.Decode(item.Payload)

			var col collector.Collector
			_ = json.Unmarshal(item.Output, &col)

			return ctx.JSON(History{
				ID:        item.ID,
				Name:      item.Name,
				Status:    item.Status,
				CreatedAt: item.CreatedAt,
				Job:       job,
				Output:    col,
			})
		}
	}

	return ctx.JSONError("no such item", http.StatusNotFound)
}
