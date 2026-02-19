package middleware

import (
	"fmt"
	"net/http"
	"stmnplibrary/constanta"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/log"
	token "stmnplibrary/security/jwt"

	"context"
	"strings"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type middle struct {
	service service.UserService
}

func FnNewMiddle(service service.UserService) *middle {
	return &middle{
		service: service,
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var stack = debug.Stack()
				var errStr = fmt.Sprintf("panic: err: %v stack: %s", err, stack)
				var msg = "an error occured"
				log.LogHSR(c.Request.Context(), msg, "backend - library", c.Request.URL.Path, c.Request.Method, errStr)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response {
					Status: "failed / false",
					Message: msg,
				})
				return
			}
		}()
		c.Next()
	}
}

func GenerateUUID() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := uuid.New().String()
		ctx := context.WithValue(c.Request.Context(), constanta.TI, uuid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetIdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx = c.Request.Context()
		key := c.GetHeader(string(constanta.IK))
		if key == "" {
			c.JSON(http.StatusBadRequest, dto.Response{
				Status: "false / failed",
				Message: "missing idempotency key",
			})
			c.Abort()
			return 
		}
		ctx = context.WithValue(ctx, string(constanta.IK), key)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx = c.Request.Context()
		role, ok := ctx.Value(string(constanta.RL)).(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Status: "false/ failed authorization",
				Message: "please login or maybe cookie are missing",
			})
			c.Abort()
			return
		}
		if role != "admin" {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Status: "false/ failed authorization",
				Message: "who are you?? must be admin",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func StudentAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx = c.Request.Context()
		role, ok := ctx.Value(string(constanta.RL)).(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Status: "false/ failed authorization",
				Message: "please login or maybe cookie are missing",
			})
			c.Abort()
			return
		}
		if role != "students" {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Status: "false/ failed authorization",
				Message: "who are you?? must be student",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *middle) RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx = c.Request.Context()
		if err := m.service.RateLimiter(ctx, c.ClientIP()); err != nil {
			var errMsg = "something happened"
			if strings.Contains(err.Error(), "too many request") {
				errMsg = err.Error()
				c.JSON(http.StatusTooManyRequests, dto.Response{
					Status: "false / failed",
					Message: errMsg,
				})
				c.Abort()
			} else {
				c.JSON(http.StatusTooManyRequests, dto.Response{
					Status: "false / failed",
					Message: errMsg,
				})
				c.Abort()
			}
		}
	}
}

func (m *middle) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tkn, err := c.Cookie(string(constanta.TokenA))
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Status:  "false / failed Authentication",
				Message: "please login or maybe cookie are missing",
			})
			c.Abort()
			return
		}
		if tkn == "" {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Status:  "false / failed Authentication",
				Message: "token not found",
			})
			c.Abort()
			return
		}
		ctx := context.WithValue(c.Request.Context(), constanta.TokenA, tkn)
		data, err := token.ValidateToken(tkn)
		if err != nil {
			if c.Request.URL.Path == "/refresh" {
				c.Next()
				return
			}
			status := http.StatusUnauthorized
			if !strings.Contains(err.Error(), "invalid") {
				status = http.StatusInternalServerError
			}
			if strings.Contains(err.Error(), "internal server error") {
				c.JSON(status, dto.Response{
					Status:  "false / failed Authentication",
					Message: "an error occured",
				})
			} else {
				c.JSON(status, dto.Response{
					Status:  "false / failed Authentication",
					Message: err.Error(),
				})
			}
			c.Abort()
			return
		}
		if err := m.service.CheckAccTkn(ctx, tkn); err != nil {
			if strings.Contains(err.Error(), "blacklist") {
				c.JSON(http.StatusUnauthorized, dto.Response{
					Status:  "false / failed Authentication",
					Message: err.Error(),
				})
			} else {
				c.JSON(http.StatusInternalServerError, dto.Response{
					Status:  "false / failed Authentication",
					Message: err.Error(),
				})
			}
			c.Abort()
		}
		ctx = context.WithValue(ctx, string(constanta.RL), data.Role)
		ctx = context.WithValue(ctx, constanta.UI, data.UserId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
