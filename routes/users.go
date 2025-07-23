package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,max=128" example:"test@example.com"`
	Password string `json:"password" binding:"required,max=256" example:"Password123-"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,max=100" example:"test@example.com"`
	Password string `json:"password" binding:"required,max=256" example:"Password123-"`
}

type SignUpResponse struct {
	Message string `json:"message" example:"User created."`
	User    string `json:"user" example:"test@example.com"`
}

type LoginResponse struct {
	Message string `json:"message" example:"Login successful!"`
	Token   string `json:"token" example:"token123"`
}

// signUp godoc
// @Summary      Register a new user
// @Description  Creates a new user account. The password must be at least 8 characters long and contain at least 1 number, 1 upper case letter, and 1 special character.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body  SignUpRequest  true  "User registration data"
// @Success      201  {object}  SignUpResponse  "User created."  example({"message": "User created.", "user": "user@example.com"})
// @Failure      400  {object}  ErrorResponse       "Could not parse request data or user already exists."
// @Failure      500  {object}  ErrorResponse       "Could not create user. Try again later."
// @Router       /signup [post]
func signUp(context *gin.Context) {
	log.Debug("Sign Up endpoint called")

	var input SignUpRequest

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data. One or more mandatory attributes are null/empty or at least one of the expected attributes is too large."})
		return
	}

	err := utils.ValidateEmail(input.Email)
	if err != nil {
		log.Errorf("User email is empty or invalid")
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not create user. The provided email is empty or invalid."})
		return
	}

	minUserPasswordLenghtStr := os.Getenv("MIN_USER_PASSWORD_LENGTH")
	minPasswordLength := 8 // Default min password length
	if minUserPasswordLenghtStr != "" {
		minPasswordLength, err = strconv.Atoi(minUserPasswordLenghtStr)
		if err != nil {
			log.Errorf("Unexpected error reading min user password length in environment properties")
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user. Try again later."})
			return
		}
	}

	isPasswordValid := utils.ValidatePassword(input.Password, minPasswordLength)
	if !isPasswordValid {
		log.Errorf("The provided user password is invalid. It must contain at least 1 number, 1 upper case letter and 1 special character")
		errorMsg := fmt.Sprintf("The provided user password is invalid. It must be at least %d characters long, contain at least 1 number, 1 upper case letter and 1 special character (a punctuation sign or a symbol like @,#,*,etc)", minPasswordLength)
		context.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	userService := services.GetUserService()

	// Check if the user already exists
	_, err = userService.FindByEmail(input.Email)

	if err == nil {
		log.Errorf("User with email %s already exists", input.Email)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not create user. It already exists."})
		return
	}

	err = userService.Create(models.NewUser(input.Email, input.Password))

	if err != nil {
		log.Errorf("Error creating user: %v", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not create user. Try again later."})
		return
	}

	log.Debugf("User %s created successfully", input.Email)

	context.JSON(http.StatusCreated, &SignUpResponse{Message: "User created.", User: input.Email})
}

// login godoc
// @Summary      User login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body  LoginRequest  true  "User login credentials"
// @Success      200  {object}  LoginResponse  "Login successful."  example({"message": "Login successful!", "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."})
// @Failure      400  {object}  ErrorResponse  "Could not parse request data."
// @Failure      401  {object}  ErrorResponse  "Wrong user credentials."
// @Router       /login [post]
func login(context *gin.Context) {
	log.Debug("Login endpoint called")

	var input LoginRequest

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data. One or more mandatory attributes are null/empty or at least one of the expected attributes is too large."})
		return
	}

	// Create a new User instance using the constructor
	userService := services.GetUserService()

	user := models.InitUser()
	user.Email = input.Email

	err := userService.ValidateCredentials(user, input.Password)
	if err != nil {
		log.Errorf("Error validating user credentials: %v", err)
		context.JSON(http.StatusUnauthorized, &ErrorResponse{Message: "Wrong user credentials."})
		return
	}

	token, err := userService.GenerateLoginToken(user)
	if err != nil {
		log.Errorf("Error generating token: %v", err)
		context.JSON(http.StatusUnauthorized, &ErrorResponse{Message: "Unexpected error."})
		return
	}

	log.Debugf("User %s logged in successfully", user.Email)
	context.JSON(http.StatusOK, &LoginResponse{Message: "Login successful!", Token: token})
}
