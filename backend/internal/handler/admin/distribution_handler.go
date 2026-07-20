package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type DistributionHandler struct{ service *service.DistributionService }

func NewDistributionHandler(distributionService *service.DistributionService) *DistributionHandler {
	return &DistributionHandler{service: distributionService}
}

func adminSubjectID(c *gin.Context) int64 {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		return 0
	}
	return subject.UserID
}

func (h *DistributionHandler) GetSettings(c *gin.Context) {
	v, err := h.service.AdminGetSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, v)
}

func (h *DistributionHandler) GetOverview(c *gin.Context) {
	v, err := h.service.AdminGetOverview(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, v)
}

func (h *DistributionHandler) GetMaturityStatus(c *gin.Context) {
	response.Success(c, h.service.MaturityStatus())
}

func (h *DistributionHandler) ListAnomalies(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListAnomalies(c.Request.Context(), service.DistributionAdminAnomalyListFilter{
		Page: page, PageSize: pageSize, Type: c.Query("type"), Severity: c.Query("severity"),
		Search: c.Query("search"), SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) UpdateSettings(c *gin.Context) {
	var req service.DistributionSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminUpdateSettings(c.Request.Context(), req, adminSubjectID(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

func (h *DistributionHandler) SetFXRate(c *gin.Context) {
	var req struct {
		Currency string          `json:"currency" binding:"required"`
		Rate     decimal.Decimal `json:"rate_to_cny" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminSetFXRate(c.Request.Context(), req.Currency, req.Rate, adminSubjectID(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"currency": strings.ToUpper(strings.TrimSpace(req.Currency)), "rate_to_cny": req.Rate})
}

func (h *DistributionHandler) GrantAgent(c *gin.Context) {
	var req service.DistributionGrantAgentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	req.GrantedBy = adminSubjectID(c)
	req.Depth = 1
	req.ParentAgentID = nil
	agent, err := h.service.AdminGrantAgent(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, agent)
}

func (h *DistributionHandler) UpdateAgentStatus(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminUpdateAgentStatus(c.Request.Context(), agentID, adminSubjectID(c), req.Status, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

func (h *DistributionHandler) UpdateAgentRate(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	var req struct {
		UseDefault      bool   `json:"use_default"`
		RateOverrideBPS *int   `json:"rate_override_bps"`
		Reason          string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	rate := req.RateOverrideBPS
	if req.UseDefault {
		rate = nil
	}
	if err := h.service.AdminUpdateAgentRate(c.Request.Context(), agentID, adminSubjectID(c), rate, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

func (h *DistributionHandler) UpdateAgentRecruitmentPermission(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	var req struct {
		Enabled bool   `json:"enabled"`
		Reason  string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminUpdateAgentRecruitmentPermission(c.Request.Context(), agentID, adminSubjectID(c), req.Enabled, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}
func (h *DistributionHandler) ReviewWithdrawal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("withdrawal_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid withdrawal id")
		return
	}
	var req struct {
		Status           string `json:"status" binding:"required"`
		Note             string `json:"note"`
		PaymentReference string `json:"payment_reference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminReviewWithdrawal(c.Request.Context(), id, adminSubjectID(c), req.Status, req.Note, req.PaymentReference); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

func (h *DistributionHandler) BatchReviewWithdrawals(c *gin.Context) {
	var req service.DistributionBatchWithdrawalReviewInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.AdminBatchReviewWithdrawals(c.Request.Context(), req, adminSubjectID(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
func (h *DistributionHandler) ListWithdrawals(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListWithdrawals(c.Request.Context(), service.DistributionAdminWithdrawalListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Status: c.Query("status"),
		AgentID: parseDistributionInt64(c.Query("agent_id")), DateFrom: parseDistributionTime(c.Query("date_from")),
		DateTo: parseDistributionTime(c.Query("date_to")), SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) GetWithdrawal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("withdrawal_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid withdrawal id")
		return
	}
	item, err := h.service.AdminGetWithdrawal(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributionHandler) GetWithdrawalEvidenceCapabilities(c *gin.Context) {
	response.Success(c, h.service.EvidenceCapabilities())
}

func (h *DistributionHandler) UploadWithdrawalEvidence(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("withdrawal_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid withdrawal id")
		return
	}
	capabilities := h.service.EvidenceCapabilities()
	if !capabilities.Available {
		response.BadRequest(c, "payment evidence storage is unavailable")
		return
	}
	uploads, err := parseDistributionEvidenceMultipart(c, capabilities)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, err := h.service.AdminUploadWithdrawalEvidence(
		c.Request.Context(), id, adminSubjectID(c), c.PostForm("evidence_type"), c.PostForm("note"), uploads,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *DistributionHandler) GetWithdrawalEvidenceURL(c *gin.Context) {
	withdrawalID, err := strconv.ParseInt(c.Param("withdrawal_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid withdrawal id")
		return
	}
	attachmentID, err := strconv.ParseInt(c.Param("attachment_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid attachment id")
		return
	}
	url, err := h.service.AdminWithdrawalAttachmentURL(c.Request.Context(), withdrawalID, attachmentID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"url": url})
}

func parseDistributionEvidenceMultipart(c *gin.Context, capabilities service.DistributionEvidenceCapabilities) ([]service.TicketUpload, error) {
	limit := capabilities.MaxTotalBytes + 2*1024*1024
	if limit < 2*1024*1024 {
		limit = 2 * 1024 * 1024
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	if err := c.Request.ParseMultipartForm(limit); err != nil {
		return nil, service.ErrTicketAttachmentLimit
	}
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 || len(files) > capabilities.MaxFilesPerUpload {
		return nil, service.ErrTicketAttachmentLimit
	}
	uploads := make([]service.TicketUpload, 0, len(files))
	var total int64
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(io.LimitReader(file, capabilities.MaxFileBytes+1))
		_ = file.Close()
		if readErr != nil {
			return nil, readErr
		}
		total += int64(len(data))
		if len(data) == 0 || int64(len(data)) > capabilities.MaxFileBytes || total > capabilities.MaxTotalBytes {
			return nil, service.ErrTicketAttachmentLimit
		}
		uploads = append(uploads, service.TicketUpload{Name: header.Filename, ContentType: header.Header.Get("Content-Type"), Data: data})
	}
	return uploads, nil
}

func (h *DistributionHandler) UpdateLevel(c *gin.Context) {
	depth64, err := strconv.ParseInt(c.Param("depth"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid depth")
		return
	}
	var req struct {
		DefaultRateBPS  int  `json:"default_rate_bps"`
		MaxChildRateBPS int  `json:"max_child_rate_bps"`
		Active          bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminUpdateLevel(c.Request.Context(), int(depth64), req.DefaultRateBPS, req.MaxChildRateBPS, req.Active); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}
func (h *DistributionHandler) CorrectBinding(c *gin.Context) {
	var req struct {
		UserID  int64  `json:"user_id" binding:"required"`
		AgentID int64  `json:"agent_id" binding:"required"`
		Reason  string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AdminCorrectCustomerBinding(c.Request.Context(), req.UserID, req.AgentID, adminSubjectID(c), req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}
func (h *DistributionHandler) ListAgents(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	depth, _ := strconv.Atoi(c.Query("depth"))
	items, total, err := h.service.AdminListAgents(c.Request.Context(), service.DistributionAdminListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Status: c.Query("status"), Depth: depth,
		SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) LookupAgentCandidates(c *gin.Context) {
	items, err := h.service.AdminLookupAgentCandidates(c.Request.Context(), c.Query("q"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *DistributionHandler) LookupAgents(c *gin.Context) {
	items, err := h.service.AdminLookupAgents(c.Request.Context(), c.Query("q"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *DistributionHandler) ListAgentEvents(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	items, err := h.service.AdminListAgentEvents(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *DistributionHandler) ListBindingEvents(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	items, err := h.service.AdminListBindingEvents(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}
func (h *DistributionHandler) ListCommissions(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListCommissions(c.Request.Context(), service.DistributionAdminCommissionListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Status: c.Query("status"),
		EntryType: c.Query("entry_type"), PaymentType: c.Query("payment_type"), AgentID: parseDistributionInt64(c.Query("agent_id")),
		DateFrom: parseDistributionTime(c.Query("date_from")), DateTo: parseDistributionTime(c.Query("date_to")),
		SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListCustomers(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	depth, _ := strconv.Atoi(c.Query("depth"))
	agentID, _ := strconv.ParseInt(c.Query("agent_id"), 10, 64)
	items, total, err := h.service.AdminListCustomers(c.Request.Context(), service.DistributionAdminListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Depth: depth, AgentID: agentID,
		SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func parseDistributionInt64(value string) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return parsed
}

func parseDistributionTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}

const distributionExportLimit = 50000

func (h *DistributionHandler) ExportAgents(c *gin.Context) {
	depth, _ := strconv.Atoi(c.Query("depth"))
	filter := service.DistributionAdminListFilter{Search: c.Query("search"), Status: c.Query("status"), Depth: depth, SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order")}
	rows := make([][]string, 0)
	for page := 1; len(rows) < distributionExportLimit; page++ {
		filter.Page, filter.PageSize = page, 100
		items, total, err := h.service.AdminListAgents(c.Request.Context(), filter)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for _, item := range items {
			rows = append(rows, []string{strconv.FormatInt(item.ID, 10), item.Email, item.Username, strconv.Itoa(item.Depth), item.ParentEmail, item.PromotionCode,
				fmt.Sprintf("%.2f%%", float64(item.EffectiveRateBPS)/100), item.Status, item.CustomerPaidCNY.StringFixed(2), item.TotalEarnedCNY.StringFixed(2),
				item.ThisMonthCommissionCNY.StringFixed(2), item.AvailableCNY.StringFixed(2), item.FrozenCNY.StringFixed(2), item.ReservedCNY.StringFixed(2), item.DebtCNY.StringFixed(2), item.CreatedAt.Format(time.RFC3339)})
			if len(rows) >= distributionExportLimit {
				break
			}
		}
		if len(items) == 0 || int64(page*filter.PageSize) >= total {
			break
		}
	}
	writeDistributionCSV(c, "distribution-agents", []string{"代理ID", "邮箱", "用户名", "层级", "上级代理", "推广码", "返佣比例", "状态", "客户实付(CNY)", "累计佣金(CNY)", "本月佣金(CNY)", "可用佣金(CNY)", "冻结佣金(CNY)", "提现占用(CNY)", "负债(CNY)", "创建时间"}, rows)
}

func (h *DistributionHandler) ExportCustomers(c *gin.Context) {
	depth, _ := strconv.Atoi(c.Query("depth"))
	filter := service.DistributionAdminListFilter{Search: c.Query("search"), Depth: depth, AgentID: parseDistributionInt64(c.Query("agent_id")), SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order")}
	rows := make([][]string, 0)
	for page := 1; len(rows) < distributionExportLimit; page++ {
		filter.Page, filter.PageSize = page, 100
		items, total, err := h.service.AdminListCustomers(c.Request.Context(), filter)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for _, item := range items {
			lastPaid := ""
			if item.LastPaidAt != nil {
				lastPaid = item.LastPaidAt.Format(time.RFC3339)
			}
			rows = append(rows, []string{strconv.FormatInt(item.UserID, 10), item.Email, item.Username, item.AgentEmail, item.AgentPromotionCode, strconv.Itoa(item.AgentDepth),
				strconv.FormatInt(item.OrderCount, 10), item.TotalPaidCNY.StringFixed(2), item.CommissionCNY.StringFixed(2), item.RefundedCNY.StringFixed(2), lastPaid, item.BoundAt.Format(time.RFC3339)})
			if len(rows) >= distributionExportLimit {
				break
			}
		}
		if len(items) == 0 || int64(page*filter.PageSize) >= total {
			break
		}
	}
	writeDistributionCSV(c, "distribution-customers", []string{"用户ID", "邮箱", "用户名", "归属代理", "代理推广码", "代理层级", "订单数", "累计实付(CNY)", "贡献佣金(CNY)", "退款金额(CNY)", "最近付费", "绑定时间"}, rows)
}

func (h *DistributionHandler) ExportCommissions(c *gin.Context) {
	filter := service.DistributionAdminCommissionListFilter{Search: c.Query("search"), Status: c.Query("status"), EntryType: c.Query("entry_type"), PaymentType: c.Query("payment_type"),
		AgentID: parseDistributionInt64(c.Query("agent_id")), DateFrom: parseDistributionTime(c.Query("date_from")), DateTo: parseDistributionTime(c.Query("date_to")), SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order")}
	rows := make([][]string, 0)
	for page := 1; len(rows) < distributionExportLimit; page++ {
		filter.Page, filter.PageSize = page, 100
		items, total, err := h.service.AdminListCommissions(c.Request.Context(), filter)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for _, item := range items {
			rows = append(rows, []string{strconv.FormatInt(item.ID, 10), item.OrderNo, item.CustomerEmail, item.AgentEmail, item.EntryType, item.PaymentType, item.PaymentCurrency,
				item.ActualPaidAmount.StringFixed(8), item.FXRateToCNY.StringFixed(8), item.CommissionBaseCNY.StringFixed(8), fmt.Sprintf("%.2f%%", float64(item.RateBPS)/100),
				item.OriginalAmountCNY.StringFixed(8), item.ReversedAmountCNY.StringFixed(8), item.OriginalAmountCNY.Sub(item.ReversedAmountCNY).StringFixed(8), item.Status, item.AvailableAt.Format(time.RFC3339), item.CreatedAt.Format(time.RFC3339)})
			if len(rows) >= distributionExportLimit {
				break
			}
		}
		if len(items) == 0 || int64(page*filter.PageSize) >= total {
			break
		}
	}
	writeDistributionCSV(c, "distribution-commissions", []string{"佣金ID", "订单号", "客户", "受益代理", "佣金类型", "支付类型", "实付币种", "实付金额", "计佣汇率", "计佣基数(CNY)", "返佣比例", "原始佣金(CNY)", "冲正金额(CNY)", "净佣金(CNY)", "状态", "可用时间", "创建时间"}, rows)
}

func (h *DistributionHandler) ExportWithdrawals(c *gin.Context) {
	filter := service.DistributionAdminWithdrawalListFilter{Search: c.Query("search"), Status: c.Query("status"), AgentID: parseDistributionInt64(c.Query("agent_id")),
		DateFrom: parseDistributionTime(c.Query("date_from")), DateTo: parseDistributionTime(c.Query("date_to")), SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order")}
	rows := make([][]string, 0)
	for page := 1; len(rows) < distributionExportLimit; page++ {
		filter.Page, filter.PageSize = page, 100
		items, total, err := h.service.AdminListWithdrawals(c.Request.Context(), filter)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for _, item := range items {
			reviewedAt, paidAt := "", ""
			if item.ReviewedAt != nil {
				reviewedAt = item.ReviewedAt.Format(time.RFC3339)
			}
			if item.PaidAt != nil {
				paidAt = item.PaidAt.Format(time.RFC3339)
			}
			approvedAt := ""
			if item.ApprovedAt != nil {
				approvedAt = item.ApprovedAt.Format(time.RFC3339)
			}
			rows = append(rows, []string{item.RequestNo, item.AgentEmail, item.AgentUsername, item.AlipayName, item.AlipayAccount, item.AmountCNY.StringFixed(2), item.FeeCNY.StringFixed(2), item.PayoutCNY.StringFixed(2),
				item.Status, item.ApproverEmail, approvedAt, item.ReviewNote, item.ReviewerEmail, item.PaymentReference, strconv.FormatInt(item.AttachmentCount, 10), reviewedAt, paidAt, item.CreatedAt.Format(time.RFC3339)})
			if len(rows) >= distributionExportLimit {
				break
			}
		}
		if len(items) == 0 || int64(page*filter.PageSize) >= total {
			break
		}
	}
	writeDistributionCSV(c, "distribution-withdrawals", []string{"申请编号", "代理邮箱", "代理用户名", "收款人", "支付宝账号", "申请金额(CNY)", "手续费(CNY)", "到账金额(CNY)", "状态", "审批人", "审批时间", "处理备注", "最后操作人", "付款流水号", "凭证数", "首次审核时间", "付款时间", "申请时间"}, rows)
}

func writeDistributionCSV(c *gin.Context, prefix string, header []string, rows [][]string) {
	var buffer bytes.Buffer
	_, _ = buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buffer)
	_ = writer.Write(header)
	_ = writer.WriteAll(rows)
	if err := writer.Error(); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	filename := fmt.Sprintf("%s-%s.csv", prefix, time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-Export-Row-Count", strconv.Itoa(len(rows)))
	if len(rows) >= distributionExportLimit {
		c.Header("X-Export-Truncated", "true")
	}
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buffer.Bytes())
}
