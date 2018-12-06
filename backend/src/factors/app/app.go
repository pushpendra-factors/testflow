package main

import (
	C "factors/config"
	H "factors/handler"
	Mid "factors/middleware"
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
	r.Use(Mid.CustomCors())

	// Initialize routes.
	H.InitAppRoutes(r)
	H.InitSDKRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
