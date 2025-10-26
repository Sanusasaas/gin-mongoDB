package routes

import (
	"github.com/gin-gonic/gin"
	controller "ginProj/controllers"
)

func CarRoutes(incomingRoutes *gin.Engine) {
	carRoutes := incomingRoutes.Group("/cars")
	{
		carRoutes.POST("/create", controller.CreateCar())
		carRoutes.GET("/get", controller.GetCars())
		carRoutes.PUT("/book/:car_id/:user_id", controller.BookCar())
		carRoutes.PUT("/update/:car_id", controller.UpdateCar())
		carRoutes.DELETE("/delete/:car_id", controller.DeleteCar())
	}
}