package service

import (
	"fmt"
	"time"

	"github.com/aelhady03/twerlo-chat-app/internal/auth"
	"github.com/aelhady03/twerlo-chat-app/internal/models"
	"github.com/aelhady03/twerlo-chat-app/internal/repository"
	"github.com/aelhady03/twerlo-chat-app/pkg/utils"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewUserService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *UserService) Register(req *models.UserRegistration) (*models.AuthResponse, error) {
	// Check if email already exists
	emailExists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, fmt.Errorf("email already exists")
	}

	// Check if username already exists
	usernameExists, err := s.userRepo.UsernameExists(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsOnline:  false,
		LastSeen:  time.Now(),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToResponse(),
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(req *models.UserLogin) (*models.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update user online status
	err = s.userRepo.UpdateOnlineStatus(user.ID, true)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update user online status: %v\n", err)
	}

	// Generate JWT token
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToResponse(),
	}, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// GetAllUsers retrieves all users (excluding sensitive information)
func (s *UserService) GetAllUsers() ([]models.UserResponse, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, nil
}

// UpdateOnlineStatus updates a user's online status
func (s *UserService) UpdateOnlineStatus(userID uuid.UUID, isOnline bool) error {
	err := s.userRepo.UpdateOnlineStatus(userID, isOnline)
	if err != nil {
		return fmt.Errorf("failed to update online status: %w", err)
	}

	return nil
}

// Logout updates user's online status to offline
func (s *UserService) Logout(userID uuid.UUID) error {
	err := s.userRepo.UpdateOnlineStatus(userID, false)
	if err != nil {
		return fmt.Errorf("failed to update online status: %w", err)
	}

	return nil
}
