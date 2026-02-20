package handler

import (
	"fmt"
	"net/http"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/controller/handler/utils"
	"stmnplibrary/log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService service.AdminService
}

func FnAdminHandler(service service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: service}
}

func getPage(c *gin.Context) (int, error) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		return 0, fmt.Errorf("an error occured")
	}
	if page == 0 {
		return 0, fmt.Errorf("page must be filled")
	}
	return page, err
}

// GetLoanData godoc
// @Summary Get loan data
// @Description Get all loan data, whether it has been returned or not
// @Produce json
// @Param page query int true "Page"
// @Tags Admin
// @Success 200 {object} dto.Response "Successfully get loan data"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /admin/loan [get]
func (ah *AdminHandler) GetLoanData(c *gin.Context) {
	var ctx = c.Request.Context()
	page, err := getPage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  "failed - get loan data",
			Message: err.Error(),
		})
		return
	}
	ld, err := ah.adminService.GetLoanData(ctx, page)
	if err != nil {
		var resMsg = "failed get loan data"
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "get loan data", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Status: "success get loan data",
		Data:   ld,
	})
}

// GetLDDone godoc
// @Summary Get loan data
// @Description Get all loan data that has been returned
// @Produce json
// @Param page query int true "Page"
// @Tags Admin
// @Success 200 {object} dto.Response "Successfully get loan data"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /admin/loan/done [get]
func (ah *AdminHandler) GetLDDone(c *gin.Context) {
	var ctx = c.Request.Context()
	page, err := getPage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  "failed - get loan data done",
			Message: err.Error(),
		})
		return
	}
	ld, err := ah.adminService.GetLDDone(ctx, page)
	if err != nil {
		var resMsg = "failed get loan data done"
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "get loan data done", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Status: "success get loan data done",
		Data:   ld,
	})
}

// GetLDDont godoc
// @Summary Get loan data
// @Description Get all loan data that has not been returned
// @Produce json
// @Param page query int true "Page"
// @Tags Admin
// @Success 200 {object} dto.Response "Successfully get loan data"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /admin/loan/dont [get]
func (ah *AdminHandler) GetLDDont(c *gin.Context) {
	var ctx = c.Request.Context()
	page, err := getPage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  "failed - get loan data dont",
			Message: err.Error(),
		})
		return
	}
	if page == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{
			Status:  "failed - get loan data",
			Message: "page must be filled",
		})
		return
	}
	ld, err := ah.adminService.GetLDDont(ctx, page)
	if err != nil {
		var resMsg = "failed get loan data dont"
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "get loan data dont", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Status: "success get loan data dont",
		Data:   ld,
	})
}

// AddCategory godoc
// @Summary Add category
// @Description Add new category to database
// @Accept json
// @Produce json
// @Param category body dto.Category true "Category name"
// @Tags Admin
// @Success 200 {object} dto.Response "Successfully add new category"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /admin/add/category [post]
func (ah *AdminHandler) AddCategory(c *gin.Context) {
	var (
		data dto.Category
		ctx  = c.Request.Context()
	)
	if errMsg := utils.GetData(func() error { return c.ShouldBindJSON(&data) }, "failed add category"); errMsg != nil {
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	if err := ah.adminService.AddCategory(ctx, data); err != nil {
		var resMsg = "failed add category"
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "add category", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(http.StatusCreated, dto.Response{
		Status:  "success",
		Message: "success add category to database",
	})
}

// AddBook godoc
// @Summary Add book
// @Description Add new book to database
// @Accept json
// @Produce json
// @Param book body dto.BookData true "Book data"
// @Tags Admin
// @Success 200 {object} dto.Response "Successfully add new book"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /admin/add/book [post]
func (ah *AdminHandler) AddBook(c *gin.Context) {
	var (
		data dto.BookData
		ctx = c.Request.Context()
	)
	if err := utils.GetData(func() error { return c.ShouldBindJSON(&data) }, "failed add book"); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if err := ah.adminService.AddBook(ctx, data); err != nil {
		resMsg := "failed add book"
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(c, resMsg, "add book", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(http.StatusCreated, dto.Response{
		Status: "success add book",
	})
}

// Confirm godoc
// @Summary Confirm
// @Description Confirm book loan
// @Accept json
// @Produce json
// @Param confirm body dto.Confirm true "Student and book data"
// @Tags Admin
// @Success 200 {object} dto.Response "Successfully confirm book loan"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /admin/loan/confirm [post]
func (ah *AdminHandler) Confirm(c *gin.Context) {
	var (
		data dto.Confirm
		ctx = c.Request.Context()
		resMsg = "failed confirm loan"
	)
	if err := utils.GetData(func() error { return c.ShouldBindJSON(&data) }, "failed confirm loan"); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return 
	}
	if err := ah.adminService.Confirm(ctx, data); err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "confirm loan", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		fmt.Printf("debug: err: %v \n", err)
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Status: "success confirm loan",
	})
}