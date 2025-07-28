package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/requests"
	"example.com/travel-advisor/responses"
	"example.com/travel-advisor/services"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// signUp godoc
// @Summary      Register a new user
// @Description  Creates a new user account. The password must be at least 8 characters long and contain at least 1 number, 1 upper case letter, and 1 special character.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body  requests.SignUpRequest  true  "User registration data"
// @Success      201  {object}  responses.SignUpResponse  "User created."
// @Failure      400  {object}  responses.ErrorResponse   "Could not parse request data or user already exists."
// @Failure      500  {object}  responses.ErrorResponse   "Could not create user. Try again later."
// @Router       /signup [post]
func signUp(context *gin.Context) {
	log.Debug("Sign Up endpoint called")

	var input requests.SignUpRequest

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON: %v", err)
		context.JSON(http.StatusBadRequest, &responses.ErrorResponse{Message: "Could not parse request data. One or more mandatory attributes are null/empty or at least one of the expected attributes is too large."})
		return
	}

	err := utils.ValidateEmail(input.Email)
	if err != nil {
		log.Errorf("User email is empty or invalid")
		context.JSON(http.StatusBadRequest, &responses.ErrorResponse{Message: "Could not create user. The provided email is empty or invalid."})
		return
	}

	minUserPasswordLenghtStr := os.Getenv("MIN_USER_PASSWORD_LENGTH")
	minPasswordLength := 8 // Default min password length
	if minUserPasswordLenghtStr != "" {
		minPasswordLength, err = strconv.Atoi(minUserPasswordLenghtStr)
		if err != nil {
			log.Errorf("Unexpected error reading min user password length in environment properties")
			context.JSON(http.StatusInternalServerError, &responses.ErrorResponse{Message: "Could not create user. Try again later."})
			return
		}
	}

	isPasswordValid := utils.ValidatePassword(input.Password, minPasswordLength)
	if !isPasswordValid {
		log.Errorf("The provided user password is invalid. It must contain at least 1 number, 1 upper case letter and 1 special character")
		errorMsg := fmt.Sprintf("The provided user password is invalid. It must be at least %d characters long, contain at least 1 number, 1 upper case letter and 1 special character (a punctuation sign or a symbol like @,#,*,etc)", minPasswordLength)
		context.JSON(http.StatusBadRequest, &responses.ErrorResponse{Message: errorMsg})
		return
	}

	userService := services.GetUserService()

	// Check if the user already exists
	_, err = userService.FindByEmail(input.Email)

	if err == nil {
		log.Errorf("User with email %s already exists", input.Email)
		context.JSON(http.StatusBadRequest, &responses.ErrorResponse{Message: "Could not create user. It already exists."})
		return
	}

	err = userService.Create(models.NewUser(input.Email, input.Password))

	if err != nil {
		log.Errorf("Error creating user: %v", err)
		context.JSON(http.StatusInternalServerError, &responses.ErrorResponse{Message: "Could not create user. Try again later."})
		return
	}

	log.Debugf("User %s created successfully", input.Email)

	context.JSON(http.StatusCreated, &responses.SignUpResponse{Message: "User created.", User: input.Email})
}

// login godoc
// @Summary      User login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body  requests.LoginRequest  true  "User login credentials"
// @Success      200  {object}  responses.LoginResponse  "Login successful."
// @Failure      400  {object}  responses.ErrorResponse  "Could not parse request data."
// @Failure      401  {object}  responses.ErrorResponse  "Wrong user credentials."
// @Router       /login [post]
func login(context *gin.Context) {
	log.Debug("Login endpoint called")

	var input requests.LoginRequest

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON: %v", err)
		context.JSON(http.StatusBadRequest, &responses.ErrorResponse{Message: "Could not parse request data. One or more mandatory attributes are null/empty or at least one of the expected attributes is too large."})
		return
	}

	// Create a new User instance using the constructor
	userService := services.GetUserService()

	user := models.InitUser()
	user.Email = input.Email

	err := userService.ValidateCredentials(user, input.Password)
	if err != nil {
		log.Errorf("Error validating user credentials: %v", err)
		context.JSON(http.StatusUnauthorized, &responses.ErrorResponse{Message: "Wrong user credentials."})
		return
	}

	token, err := userService.GenerateLoginToken(user)
	if err != nil {
		log.Errorf("Error generating token: %v", err)
		context.JSON(http.StatusUnauthorized, &responses.ErrorResponse{Message: "Unexpected error."})
		return
	}

	log.Debugf("User %s logged in successfully", user.Email)
	context.JSON(http.StatusOK, &responses.LoginResponse{Message: "Login successful!", Token: token})
}
