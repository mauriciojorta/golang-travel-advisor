package middlewares

import (
	"net/http"

	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func Authenticate(context *gin.Context) {

	token := context.Request.Header.Get("Authorization")

	if token == "" {
		log.Error("Authorization header is missing")
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	userId, err := utils.VerifyToken(token)

	if err != nil {
		log.Errorf("Error verifying token: %v", err)
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	context.Set("userId", userId)

	context.Next()
}
