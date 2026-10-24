package controllers

import (
	"errors"
	"homework-blog/config"
	"homework-blog/models"
	"homework-blog/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct{}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Register 用户注册
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	db := config.DB

	// 检查用户名是否已存在 ，验证用户名和邮箱
	var user models.User
	err := db.Where("username = ?", req.Username).First(&user).Error
	if err == nil {
		utils.Conflict(c, "用户已注册")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.InternalServerError(c, "Failed to check username")
		return
	}

	err = db.Where("email = ?", req.Email).First(&user).Error
	if err == nil {
		utils.Conflict(c, "邮箱已注册")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.InternalServerError(c, "Failed to check email")
		return
	}

	// 创建新用户

	newUser := models.User{
		UserName: req.Username,
		PassWord: req.Password,
		Email:    req.Email,
	}

	if db.Create(&newUser).Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}
	// 生成JWT token

	token, tokenErr := utils.GenerateToken(newUser.ID, newUser.UserName)

	if tokenErr != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.Success(c, AuthResponse{
		Token: token,
		User:  newUser,
	})

}

// Login 用户登录
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 查找用户
	var user models.User

	db := config.DB
	if db.Where("username = ?", req.Username).First(&user).Error != nil {
		utils.Error(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// 验证密码

	if err := bcrypt.CompareHashAndPassword([]byte(user.PassWord), []byte(req.Password)); err != nil {
		utils.Error(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// 生成JWT token

	token, err := utils.GenerateToken(user.ID, user.UserName)
	if err != nil {
		utils.InternalServerError(c, "Failed to generate token")
		return
	}

	utils.Success(c, AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile 获取用户信息
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.NotFound(c, "User not found")
		return
	}

	utils.Success(c, user)
}
