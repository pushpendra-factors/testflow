package main

import (
	C "factors/config"
	H "factors/handler"
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
	// Initialize routes.
	H.InitRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
