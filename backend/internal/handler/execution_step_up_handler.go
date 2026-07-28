package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ExecutionStepUpHandler struct {
	service *service.ExecutionStepUpService
}

func NewExecutionStepUpHandler(stepUpService *service.ExecutionStepUpService) *ExecutionStepUpHandler {
	return &ExecutionStepUpHandler{service: stepUpService}
}

type executionStepUpRequest struct {
	Password          string `json:"password" binding:"required"`
	Purpose           string `json:"purpose" binding:"required"`
	TargetFingerprint string `json:"target_fingerprint" binding:"required"`
}

type executionStepUpConsumeRequest struct {
	Proof             string `json:"proof" binding:"required"`
	Purpose           string `json:"purpose" binding:"required"`
	TargetFingerprint string `json:"target_fingerprint" binding:"required"`
}

func (h *ExecutionStepUpHandler) Issue(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req executionStepUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "password, purpose and target_fingerprint are required")
		return
	}
	result, err := h.service.Issue(c.Request.Context(), subject.UserID, req.Password, req.Purpose, req.TargetFingerprint)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"proof":      result.Proof,
		"expires_at": result.ExpiresAt.Format(time.RFC3339),
		"expires_in": int64(service.ExecutionStepUpProofTTL.Seconds()),
	})
}

func (h *ExecutionStepUpHandler) Consume(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req executionStepUpConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "proof, purpose and target_fingerprint are required")
		return
	}
	grant, err := h.service.Consume(c.Request.Context(), subject.UserID, req.Proof, req.Purpose, req.TargetFingerprint)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"valid":              true,
		"user_id":            grant.UserID,
		"purpose":            grant.Purpose,
		"target_fingerprint": grant.TargetFingerprint,
	})
}
