package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/constanta"
	"stmnplibrary/controller/handler/utils"
	"stmnplibrary/log"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func FnUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{userService: service}
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

func getPage(c *gin.Context) (int, error) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		return 1, nil
	}
	if page == 0 {
		return 1, nil
	}
	return page, err
}

// Logout godoc
// @Summary Logout 
// @Description Log out of account
// @Produce json
// @Tags student
// @Success 200 {object} dto.Response "Successfully confirm logout"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /student/logout [get]
func (uh *UserHandler) Logout(c *gin.Context) {
	var ctx = c.Request.Context()
	if err := uh.userService.Logout(ctx); err != nil {
		var resMsg = "failed logout"
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "logout", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	delCookieToken(c, string(constanta.TokenA))
	delCookieToken(c, string(constanta.TokenR))
	c.JSON(http.StatusCreated, &dto.Response{
		Status:  "true / success",
		Message: "success logout",
	})
}

// Register godoc
// @Summary Register 
// @Description Create account
// @Accept json
// @Produce json
// @Param register body dto.Students true "Student data"
// @Tags student
// @Success 200 {object} dto.Response "Successfully register"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /register [post]
func (uh *UserHandler) Register(c *gin.Context) {
	const status = "failed / error"
	const resMsg = "failed register"
	var data dto.Students
	ctx := c.Request.Context()
	errMsg := utils.GetData(func() error { return c.ShouldBindJSON(&data) }, "failed register")
	if errMsg != nil {
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	msg, err := uh.userService.Register(ctx, &data)
	if len(msg) > 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorMsg("failed register", msg))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, &dto.Response{
			Status:  status,
			Message: resMsg,
			Error:   dto.Errors{Error: "an error occurred"},
		})
		log.LogHSR(ctx, resMsg, "register", c.Request.URL.Path, c.Request.Method, err.Error())
		return
	}
	c.JSON(http.StatusCreated, &dto.Response{
		Status:  "true / success",
		Message: "success register",
	})
}

// GetBooks godoc
// @Summary Get books 
// @Description Get all books from db
// @Produce json
// @Param page query int true "Page"
// @Tags student
// @Success 200 {object} dto.Response "Successfully get books"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /student/books [get]
func (uh *UserHandler) GetBooks(c *gin.Context) {
	const resMsg = "failed get books"
	var ctx = c.Request.Context()
	page, err := getPage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  resMsg,
			Message: err.Error(),
		})
		return
	}
	books, err := uh.userService.GetBooks(ctx, page)
	if (len(books) == 0 || books == nil) && err == nil {
		c.JSON(http.StatusOK, dto.Response{
			Status: "success",
			Message: "no books found",
			Data: []interface{}{},
		})
		return
	}
	if err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "books not found")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "get_books", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		fmt.Printf("\nini status: %d\n", status)
		return
	}
	c.JSON(http.StatusOK, books)
}

// GetBooksByAuthor godoc
// @Summary Get books 
// @Description Get books with author as filter
// @Produce json
// @Param page query int true "Page"
// @Param author query string true "Author"
// @Tags student
// @Success 200 {object} dto.Response "Successfully get books by author"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /student/books/author [get]
func (uh *UserHandler) GetBooksByAuthor(c *gin.Context) {
	const resMsg = "failed get books"
	var ctx = c.Request.Context()
	author := c.Query("author")
	if author == "" {
		c.JSON(400, "author must be filled")
		return
	}
	page, err := getPage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  resMsg,
			Message: err.Error(),
		})
		return
	}
	books, err := uh.userService.GetBooksByAuthor(ctx, author, page)
	if (len(books) == 0 || books == nil) && err == nil {
		c.JSON(http.StatusOK, dto.Response{
			Status: "success",
			Message: "no books found",
			Data: []interface{}{},
		})
		return
	}
	if err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "author not found")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "get_books_by_author", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(200, books)
}

// GetBooksByCategory godoc
// @Summary Get books 
// @Description Get books with category as filter
// @Produce json
// @Param page query int true "Page"
// @Param category query []string true "List category" collectionFormat(multi) minItems(1)
// @Tags student
// @Success 200 {object} dto.Response "Successfully get books by category"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /student/books/category [get]
func (uh *UserHandler) GetBooksByCategory(c *gin.Context) {
	const resMsg = "failed get books"
	var ctx = c.Request.Context()
	category := c.QueryArray("category")
	if len(category) == 0 {
		c.JSON(400, "category must be filled")
		return
	}
	page, err := getPage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  resMsg,
			Message: err.Error(),
		})
		return
	}
	books, err := uh.userService.GetBooksByCategory(ctx, category, page)
	if (len(books) == 0 || books == nil) && err == nil {
		c.JSON(http.StatusOK, dto.Response{
			Status: "success",
			Message: "no books found",
			Data: []interface{}{},
		})
		return
	}
	if err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "category not found")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "get_books_by_category", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(200, books)
}

// Loan godoc
// @Summary Loan book  
// @Description Borrow books from the database
// @Accept json
// @Produce json
// @Param loan body dto.Loan true "Loan data, with book id and returned time as a value"
// @Tags student
// @Success 200 {object} dto.Response "Successfully loan a book"
// @Failure 400 {object} dto.Response "Incorrect client input"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /student/book/loan [post]
func (uh *UserHandler) Loan(c *gin.Context) {
	const status = "failed / error"
	const resMsg = "failed borrow"
	var data dto.Loan
	ctx := c.Request.Context()
	errMsg := utils.GetData(func() error { return c.ShouldBindJSON(&data) }, "failed loan")
	if errMsg != nil {
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	err := uh.userService.Loan(ctx, data)
	if err != nil {
		status, errMsg := utils.ValidateErr(err, resMsg, "")
		if status == 500 {
			log.LogHSR(ctx, resMsg, "loan", c.Request.URL.Path, c.Request.Method, err.Error())
		}
		c.JSON(status, errMsg)
		return
	}
	c.JSON(http.StatusCreated, &dto.Response{
		Status:  "true / success",
		Message: "success borrow",
	})
}
