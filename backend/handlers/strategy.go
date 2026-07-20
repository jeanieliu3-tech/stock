package handlers

import (
	"github.com/gin-gonic/gin"
	"stock-app/models"
	"stock-app/services"
)

type StrategyHandler struct {
	svc *services.StrategyService
}

func NewStrategyHandler(svc *services.StrategyService) *StrategyHandler {
	return &StrategyHandler{svc: svc}
}

func (h *StrategyHandler) List(c *gin.Context) {
	strategies := h.svc.List()
	ok(c, strategies)
}

func (h *StrategyHandler) Get(c *gin.Context) {
	id := c.Param("id")
	s := h.svc.Get(id)
	if s == nil {
		fail(c, 404, "策略不存在")
		return
	}
	ok(c, s)
}

func (h *StrategyHandler) Create(c *gin.Context) {
	var s models.Strategy
	if err := c.ShouldBindJSON(&s); err != nil {
		fail(c, 400, "请求参数错误")
		return
	}
	if s.Name == "" {
		fail(c, 400, "策略名称不能为空")
		return
	}
	if len(s.Indicators) == 0 {
		fail(c, 400, "至少需要一个指标")
		return
	}
	if len(s.Indicators) > 6 {
		fail(c, 400, "最多6个指标")
		return
	}
	created, err := h.svc.Create(s)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, created)
}

func (h *StrategyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var s models.Strategy
	if err := c.ShouldBindJSON(&s); err != nil {
		fail(c, 400, "请求参数错误")
		return
	}
	updated, err := h.svc.Update(id, s)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, updated)
}

func (h *StrategyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, nil)
}

func (h *StrategyHandler) SetDefault(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.SetDefault(id); err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, nil)
}

func (h *StrategyHandler) Copy(c *gin.Context) {
	id := c.Param("id")
	copied, err := h.svc.Copy(id)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, copied)
}

func (h *StrategyHandler) GetDefault(c *gin.Context) {
	s := h.svc.GetDefault()
	if s == nil {
		fail(c, 404, "没有可用策略")
		return
	}
	ok(c, s)
}

func (h *StrategyHandler) Evaluate(c *gin.Context) {
	var req models.StrategyEvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "请求参数错误")
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	var strategy models.Strategy
	if req.Strategy != nil {
		strategy = *req.Strategy
	} else if req.StrategyID != "" {
		s := h.svc.Get(req.StrategyID)
		if s == nil {
			fail(c, 404, "策略不存在")
			return
		}
		strategy = *s
	} else {
		s := h.svc.GetDefault()
		if s == nil {
			fail(c, 404, "没有可用策略")
			return
		}
		strategy = *s
	}

	result, err := h.svc.Evaluate(strategy, page, pageSize)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, result)
}

func (h *StrategyHandler) GetIndicatorLib(c *gin.Context) {
	ok(c, models.IndicatorLibrary)
}
