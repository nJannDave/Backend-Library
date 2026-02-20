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

func setCookieToken(c *gin.Context, key string, token string, tokenType string) {
	var maxAge = 180
	if tokenType == "refresh" {
		maxAge = 604800
	}
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

// Refresh godoc
// @Summary Refresh session user
// @Description Get the refresh token in the cookie and if it is valid, generate a new access and refresh token.
// @Tags Authentication
// @Produce json
// @Success 200 {object} dto.Response "Successfully refreshed token"
// @Failure 401 {object} dto.Response "Not logged in yet"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /refresh [get]
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
	setCookieToken(c, string(constanta.TokenA), token.AccessToken, "access")
	setCookieToken(c, string(constanta.TokenR), token.RefreshToken, "refresh")
	c.JSON(http.StatusCreated, &dto.Response{
		Status:  "true / success",
		Message: "success refresh",
	})
}

// Login godoc
// @Summary Login for access library API
// @Description Login with NIK & Password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param loginData body dto.Login true "Data for login"
// @Success 200 {object} dto.Response "Successfully Login"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /login [post]
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
	setCookieToken(c, string(constanta.TokenA), token.AccessToken, "access")
	setCookieToken(c, string(constanta.TokenR), token.RefreshToken, "refresh")
	c.JSON(http.StatusOK, &dto.Response{
		Status:  "true / success",
		Message: "success login",
	})
}
