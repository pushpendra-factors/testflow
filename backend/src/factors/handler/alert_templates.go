package handler

import (
	"factors/model/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAlertTemplateHandler(c *gin.Context){
	alertTemplates, error := store.GetStore().GetAlertTemplates()
	c.JSON(error, alertTemplates)
}

func DeleteAlertTemplateHandler(c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be int"})
		return
	}
	error := store.GetStore().DeleteAlertTemplate(id)
	if error != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete template"})
		return
	}

	c.JSON(200,gin.H{})
}