package server

import (
	goerrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"lumina/internal/dao"
	"lumina/internal/model"
	"lumina/internal/version"
)

const userKey = "user"

type TokenClaims struct {
	jwt.RegisteredClaims
	UserId int `json:"user_id"`
}

func TrySetUserToContext(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			tokenStr, _ = c.Cookie("token")
		}
		if tokenStr == "" {
			auth := c.GetHeader("Authorization")
			if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
				tokenStr = auth[7:]
			}
		}
		if tokenStr != "" {
			if strings.HasPrefix(tokenStr, "sk-") {
				user, userErr := model.GetUserByToken(tokenStr)
				if userErr != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "invalid token",
					})
					return
				}
				c.Set(userKey, user)
			} else {
				token, tokenErr := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				if tokenErr != nil || !token.Valid {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "invalid token",
					})
					return
				}

				if claims, ok := token.Claims.(*TokenClaims); ok {
					user, userErr := model.GetUserById(claims.UserId)
					if userErr != nil {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error": "invalid user",
						})
						return
					}
					c.Set(userKey, user)
				} else {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "invalid token claims",
					})
					return
				}
			}
			c.Next()
			return
		}

		c.Next()
	}
}

func NeedAuth(needAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get(userKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}
		user := u.(*model.User)
		if needAdmin && !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "admin permission required",
			})
			return
		}
		c.Next()
	}
}

// @Summary 用户登录
// @Description 用户登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body dao.LoginRequest true "请求参数"
// @Success 200 {object} dao.LoginResponse
// @Router /api/v1/login [post]
func (s *Server) handleLogin(c *gin.Context) {
	var req dao.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	user, err := model.GetUserByUsername(req.Username)
	if err != nil {
		if goerrors.Is(err, gorm.ErrRecordNotFound) {
			s.writeError(c, http.StatusUnauthorized, fmt.Errorf("invalid username or password"))
			return
		}
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	if user.Password != req.Password {
		s.writeError(c, http.StatusUnauthorized, fmt.Errorf("invalid username or password"))
		return
	}

	token, err := genJwtToken(user, s.conf.JwtSecret)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	userSpec, err := dao.ToUserSpec(user)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	resp := dao.LoginResponse{
		Token: token,
		User:  *userSpec,
	}
	c.SetCookie("token", token, 7*24*60*60, "/", "", false, true)
	c.JSON(http.StatusOK, resp)
}

func genJwtToken(user *model.User, jwtSecret string) (string, error) {
	claims := TokenClaims{
		UserId: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    version.APP,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// @Summary 用户登出
// @Description 用户登出
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200
// @Router /api/v1/logout [post]
func (s *Server) handleLogout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
}

// @Summary 获取用户信息
// @Description 获取用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} dao.UserSpec
// @Router /api/v1/settings/profile [get]
func (s *Server) handleGetUserProfile(c *gin.Context) {
	user := c.MustGet(userKey).(*model.User)
	resp, err := dao.ToUserSpec(user)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary 获取用户列表
// @Description 获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param start query int true "分页开始位置"
// @Param limit query int true "分页大小"
// @Success 200 {object} dao.ListUsersResponse
// @Router /api/v1/admin/users [get]
func (s *Server) handleAdminListUsers(c *gin.Context) {
	var req dao.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Start < 0 {
		req.Start = 0
	}

	total, err := model.CountUsers()
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	users, err := model.GetUsers(req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.ListUsersResponse{
		Total: total,
		Items: make([]dao.UserSpec, len(users)),
	}
	for i, u := range users {
		userSpec, err := dao.ToUserSpec(u)
		if err != nil {
			s.logger.Errorf("ToUserSpec failed, err: %v", err)
			continue
		}
		resp.Items[i] = *userSpec
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 创建用户
// @Description 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dao.CreateUserRequest true "请求参数"
// @Success 200 {object} dao.CreateUserResponse
// @Router /api/v1/admin/users [post]
func (s *Server) handleAdminCreateUsers(c *gin.Context) {
	var req dao.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	if req.DepartmentId <= 0 {
		s.writeError(c, http.StatusBadRequest, fmt.Errorf("departmentId is required"))
		return
	}
	token := "sk-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	user := &model.User{
		Username:    req.Username,
		Password:    req.Password,
		Nickname:    req.Nickname,
		IsAdmin:     req.IsAdmin,
		AccessToken: token,
	}
	if err := model.CreateUser(user); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, dao.CreateUserResponse{
		Id: user.Id,
	})
}

// @Summary 删除用户
// @Description 删除指定的用户
// @Tags 用户管理
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/user/{user_id} [delete]
func (s *Server) handleAdminDeleteUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	user, err := model.GetUserById(userId)
	if err != nil {
		if goerrors.Is(err, gorm.ErrRecordNotFound) {
			s.writeError(c, http.StatusNotFound, err)
			return
		}
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	if err := model.DeleteUser(user.Id); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
