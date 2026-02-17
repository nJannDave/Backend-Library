package handler

import (
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/controller/handler/utils"
	"stmnplibrary/dto"
	"stmnplibrary/constanta"
	"stmnplibrary/log"

	"github.com/gin-gonic/gin"

	"net/http"
)

type AuthHandler struct {
	service service.AuthService
}

func FnAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func delCookieToken(c *gin.Context, key string) {
	const maxAge = -1
	c.SetSameSite(http.SameSiteStrictMode)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     key,
		Value:    "token",
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   maxAge,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func setCookieToken(c *gin.Context, key string, token string) {
	const maxAge = 900
	c.SetSameSite(http.SameSiteStrictMode)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     key,
		Value:    token,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   maxAge,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (ah *AuthHandler) Refresh(c *gin.Context) {
	const resMsg = "failed refresh token"
	var ctx = c.Request.Context()
	refreshTkn, err := c.Cookie(string(constanta.TokenR))
	if err != nil {
		c.JSON(401, dto.Response{
			Status:  "false / failed Authentication",
			Message: "please login or maybe cookie are missing",
		})
		return
	}
	token, err := ah.service.Refresh(ctx, refreshTkn)
	if err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "refresh", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	setCookieToken(c, string(constanta.TokenA), token.AccessToken)
	setCookieToken(c, string(constanta.TokenR), token.RefreshToken)
	c.JSON(http.StatusCreated, &dto.Response{
		Status:  "true / success",
		Message: "success refresh",
	})
}

func (ah *AuthHandler) Login(c *gin.Context) {
	const resMsg = "failed login"
	var data dto.Login
	ctx := c.Request.Context()
	errMsg := utils.GetData(func() error { return c.ShouldBindJSON(&data) }, "failed login")
	if errMsg != nil {
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	token, err := ah.service.Login(ctx, data)
	if err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "login", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	setCookieToken(c, string(constanta.TokenA), token.AccessToken)
	setCookieToken(c, string(constanta.TokenR), token.RefreshToken)
	c.JSON(http.StatusOK, &dto.Response{
		Status:  "true / success",
		Message: "success login",
	})
}
