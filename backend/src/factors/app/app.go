package main

import (
	C "factors/config"
	H "factors/handler"
	mid "factors/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Fatal("Failed to initialize.")
		return
	}

	r := gin.Default()
	// Group based middlewares should be registered on corresponding init methods.
	// Root middleware for cors.
	r.Use(mid.CustomCors())

	// Initialize routes.
	H.InitAppRoutes(r)
	H.InitSDKRoutes(r)
	H.InitIntRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
