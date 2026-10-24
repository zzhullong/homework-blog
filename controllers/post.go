package controllers

import (
	"homework-blog/config"
	"homework-blog/models"
	"homework-blog/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct{}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
}

// CreatePost 创建文章
func (pc *PostController) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, http.StatusBadRequest, "User not authenticated")
		return
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID.(uint),
	}

	if err := config.DB.Create(&post).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create post")
		return
	}

	// 预加载用户信息
	config.DB.Preload("User").First(&post, post.ID)

	utils.Success(c, post)
}

// GetPosts 获取文章列表
func (pc *PostController) GetPosts(c *gin.Context) {
	var posts []models.Post

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	//获取查询条件
	query := config.DB.Model(&models.Post{}).Order("created_at DESC")

	//获取查询条件下的count总数
	var total int64
	query.Count(&total)

	// 查询文章列表，预加载用户信息
	if err := query.Preload("User").Scopes(utils.Paginate(page, pageSize)).
		Find(&posts).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	utils.Success(c, gin.H{
		"posts":     posts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPost 获取单个文章详情
func (pc *PostController) GetPost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	var post models.Post
	if err := config.DB.Preload("User").Preload("Comments.User").First(&post, postID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Post not found")
		return
	}

	utils.Success(c, post)
}

// UpdatePost 更新文章
func (pc *PostController) UpdatePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	udpateData := map[string]any{
		"title":   req.Title,
		"content": req.Content,
	}

	//用了map结构体，指定更新不需要用select来控制了，同时保证零值更新
	result := config.DB.Model(&models.Post{}).Where("id = ? and user_id=?", postID, userID).Updates(udpateData)

	if err := result.Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "fail to update")
		return
	}

	if result.RowsAffected == 0 {
		utils.Error(c, http.StatusBadRequest, "文章不存在或无权限修改")
		return
	}

	var post models.Post

	// 预加载用户信息
	if err := config.DB.Preload("User").First(&post, postID).Error; err != nil {
		utils.NotFound(c, "Post not found")
		return
	}

	utils.Success(c, post)
}

// DeletePost 删除文章
func (pc *PostController) DeletePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		utils.NotFound(c, "Post not found")
		return
	}

	// 检查是否是文章作者
	if post.UserID != userID.(uint) {
		utils.Forbidden(c, "You can only delete your own posts")
		return
	}

	// 删除文章（软删除）
	if err := config.DB.Delete(&post).Error; err != nil {
		utils.InternalServerError(c, "Failed to delete post")
		return
	}

	utils.Success(c, gin.H{"message": "Post deleted successfully"})
}
