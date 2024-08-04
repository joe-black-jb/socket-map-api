package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joe-black-jb/socket-map-api/internal/database"
)

func GetPlaces(c *gin.Context){
	Places := &[]internal.Place{}
	if err := database.Db.Find(Places).Error; err != nil {
		FormatResponse(c, http.StatusNotFound, err)
	}
	FormatResponse(c, http.StatusOK, Places)
}