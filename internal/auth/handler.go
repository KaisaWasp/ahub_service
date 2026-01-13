package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *AuthService
}

func NewHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Login     string `json:"login"`
	Password  string `json:"password"`
}

type RegisterResponse struct {
	RegistrationToken string `json:"registration_token"`
}

type ConfirmRegisterRequest struct {
	Token string `json:"token"`
	Code  string `json:"code"`
}

type ConfirmRegisterResponse struct {
	Id int64 `json:"id"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) StartRegistration(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	token, err := h.service.StartRegistration(ctx, req.FirstName, req.LastName, req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{RegistrationToken: token})
}

func (h *AuthHandler) CreateNewUser(c *gin.Context) {
	var req ConfirmRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	accessToken, refreshToken, err := h.service.ConfirmRegistration(ctx, req.Token, req.Code)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		60*60*24*30, // 30 дней
		"/",
		"",
		true,
		true,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	accessToken, refreshToken, err := h.service.Login(ctx, req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		refreshToken,
		60*60*24*30, // 30 дней
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"error": "refresh token missing"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	accessToken, newRefreshToken, err := h.service.Refresh(ctx, refreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		60*60*24*30,
		"/",
		"",
		true,
		true,
	)

	c.JSON(200, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if refreshToken != "" {
		_ = h.service.storage.DeleteRefreshToken(ctx, refreshToken)
	}

	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	c.JSON(200, gin.H{"message": "logged out"})
}
