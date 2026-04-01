package handler

import (
	"net/http"
	"wechat-enterprise-backend/internal/middleware"
	"wechat-enterprise-backend/internal/response"
	"wechat-enterprise-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "用户名和密码不能为空")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, 40100, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := middleware.CurrentClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40103, "登录状态无效")
		return
	}
	user, err := h.authService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, 40104, "用户不存在")
		return
	}
	response.Success(c, user)
}
