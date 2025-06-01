package routes

import (
	"fmt"
	"net/http"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
)

func signUp(context *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	userService := services.GetUserService()

	// Check if the user already exists
	_, err := userService.FindByEmail(input.Email)

	if err == nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not create user. It already exists."})
		return
	}

	err = userService.Create(models.NewUser(input.Email, input.Password))

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create user. Try again later."})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created.", "user": input.Email})
}

func login(context *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	// Create a new User instance using the constructor
	userService := services.GetUserService()

	user := models.NewUser(input.Email, input.Password)

	err := userService.ValidateCredentials(user)
	if err != nil {
		fmt.Print(err)
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Wrong user credentials."})
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		fmt.Print(err)
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Wrong user credentials."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
}
