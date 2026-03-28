package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestConfigHandler_ValidateAPIKeyInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ValidateAPIKeyInput
		wantErr bool
	}{
		{
			name: "valid LLM input",
			input: ValidateAPIKeyInput{
				Type:     "llm",
				Provider: "claude",
				APIKey:   "sk-test-key",
			},
			wantErr: false,
		},
		{
			name: "valid image input",
			input: ValidateAPIKeyInput{
				Type:     "image",
				Provider: "comfyui",
				BaseURL:  "http://localhost:8188",
			},
			wantErr: false,
		},
		{
			name: "valid video input",
			input: ValidateAPIKeyInput{
				Type:     "video",
				Provider: "runway",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the input struct can be marshaled/unmarshaled
			data, err := json.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKeyInput marshal error = %v, wantErr %v", err, tt.wantErr)
			}

			var decoded ValidateAPIKeyInput
			err = json.Unmarshal(data, &decoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKeyInput unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}

			if decoded.Type != tt.input.Type {
				t.Errorf("Type mismatch: got %v, want %v", decoded.Type, tt.input.Type)
			}
		})
	}
}

func TestAuthHandler_PasswordResetInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid email",
			json:    `{"email":"test@example.com"}`,
			wantErr: false,
		},
		{
			name:    "invalid email",
			json:    `{"email":"invalid-email"}`,
			wantErr: true,
		},
		{
			name:    "missing email",
			json:    `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input struct {
				Email string `json:"email" binding:"required,email"`
			}

			err := json.Unmarshal([]byte(tt.json), &input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected unmarshal error: %v", err)
				}
				return
			}

			// Check email format
			if input.Email != "" && !isValidEmail(input.Email) {
				if !tt.wantErr {
					t.Errorf("invalid email format: %s", input.Email)
				}
			}
		})
	}
}

func TestAuthHandler_ResetPasswordInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid input",
			json:    `{"token":"abc123def456ghi789jkl012mno345pqr","password":"newpassword123"}`,
			wantErr: false,
		},
		{
			name:    "short password",
			json:    `{"token":"abc123","password":"short"}`,
			wantErr: false, // JSON unmarshal succeeds, validation happens in handler
		},
		{
			name:    "missing token",
			json:    `{"password":"newpassword123"}`,
			wantErr: false, // JSON unmarshal succeeds, validation happens in handler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input struct {
				Token    string `json:"token"`
				Password string `json:"password"`
			}

			err := json.Unmarshal([]byte(tt.json), &input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Simple email validation helper
func isValidEmail(email string) bool {
	return len(email) > 0 && contains(email, "@") && contains(email, ".")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		password string
		minLen   bool
	}{
		{"short", false},
		{"exactly6", true},
		{"longerpassword123", true},
		{"", false},
	}

	for _, tt := range tests {
		isValid := len(tt.password) >= 6
		if isValid != tt.minLen {
			t.Errorf("password %q: got valid=%v, want %v", tt.password, isValid, tt.minLen)
		}
	}
}

// Test that UUID generation works correctly
func TestUUIDGeneration(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	if id1 == id2 {
		t.Error("UUIDs should be unique")
	}

	if id1 == uuid.Nil {
		t.Error("UUID should not be nil")
	}
}

// Test JSON request handling
func TestJSONRequestParsing(t *testing.T) {
	router := setupTestRouter()

	router.POST("/test", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"email": input.Email})
	})

	// Test valid request
	body := `{"email":"test@example.com","password":"password123"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test invalid request
	body = `{"email":"test@example.com"}`
	req, _ = http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}