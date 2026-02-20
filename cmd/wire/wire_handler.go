package wiring

import (
	ha "stmnplibrary/controller/handler/admin"
	hb "stmnplibrary/controller/handler/auth"
	h "stmnplibrary/controller/handler/user"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/middleware"

	_ "stmnplibrary/docs"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

)

func WireHandler(handlerA *ha.AdminHandler, handlerB *hb.AuthHandler, handler *h.UserHandler, s service.UserService) *gin.Engine {
	router := gin.Default()

	middle := middleware.FnNewMiddle(s)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	router.Use(middleware.Recovery())
	router.Use(middleware.GenerateUUID())
	router.Use(middle.RateLimiter())

	router.POST("/register", handler.Register)
	router.POST("/login", handlerB.Login)
	router.GET("/refresh", handlerB.Refresh)

	router.Use(middle.Auth())
	admin := router.Group("admin")
	admin.Use(middleware.AdminAuth())
	students := router.Group("student")
	students.Use(middleware.StudentAuth())

	admin.GET("/loan", handlerA.GetLoanData)
	admin.GET("/loan/done", handlerA.GetLDDone)
	admin.GET("/loan/dont", handlerA.GetLDDont)
	admin.POST("/loan/confirm", handlerA.Confirm)
	admin.POST("/add/category", middleware.GetIdempotencyKey(), handlerA.AddCategory)
	admin.POST("/add/book", middleware.GetIdempotencyKey(), handlerA.AddBook)

	students.GET("/logout", handler.Logout)
	students.GET("/books", handler.GetBooks)
	students.GET("/books/author", handler.GetBooksByAuthor)
	students.GET("/books/category", handler.GetBooksByCategory)
	students.POST("/book/loan", handler.Loan)

	return router
}