package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/models"
	"golang.org/x/crypto/bcrypt"
)

// TestPasswordHashing tests that passwords are properly hashed with the correct cost factor
func TestPasswordHashing(t *testing.T) {
	// Create a user service without database for testing
	service := NewUserServiceWithoutDB()

	// Test password
	password := "testpassword123"

	// Create a register request
	req := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: password,
	}

	// Call register method (which includes password hashing)
	user, err := service.Register(req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Verify password is hashed
	if user.Password == password {
		t.Error("Password was not hashed")
	}

	// Verify password can be verified with bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		t.Errorf("Failed to verify password hash: %v", err)
	}

	// Verify the cost factor is 12
	cost, err := bcrypt.Cost([]byte(user.Password))
	if err != nil {
		t.Errorf("Failed to get cost factor: %v", err)
	}

	if cost != 12 {
		t.Errorf("Expected cost factor 12, got %d", cost)
	}
}
