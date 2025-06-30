package routes

import (
	"net/http"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func signUp(context *gin.Context) {
	log.Debug("Sign Up endpoint called")

	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing request data: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	userService := services.GetUserService()

	// Check if the user already exists
	_, err := userService.FindByEmail(input.Email)

	if err == nil {
		log.Errorf("User with email %s already exists", input.Email)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not create user. It already exists."})
		return
	}

	err = userService.Create(models.NewUser(input.Email, input.Password))

	if err != nil {
		log.Errorf("Error creating user: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create user. Try again later."})
		return
	}

	log.Debugf("User %s created successfully", input.Email)

	context.JSON(http.StatusCreated, gin.H{"message": "User created.", "user": input.Email})
}

func login(context *gin.Context) {
	log.Debug("Login endpoint called")

	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing request data: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	// Create a new User instance using the constructor
	userService := services.GetUserService()

	user := models.InitUser()
	user.Email = input.Email

	err := userService.ValidateCredentials(user, input.Password)
	if err != nil {
		log.Errorf("Error validating user credentials: %v", err)
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Wrong user credentials."})
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		log.Errorf("Error generating token: %v", err)
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Wrong user credentials."})
		return
	}

	log.Debugf("User %s logged in successfully", user.Email)
	context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
}
