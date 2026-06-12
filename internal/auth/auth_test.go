package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestCheckJWT(t *testing.T) {
	uuidString1 := "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862"
	id1, _ := uuid.Parse(uuidString1)
	secretString1 := "secretToken"
	tokenStr1, _ := MakeJWT(id1, secretString1, 10000000000)
	uuidString2 := "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b863"
	id2, _ := uuid.Parse(uuidString2)
	secretString2 := "secretToken1"
	tokenStr2, _ := MakeJWT(id2, secretString2, 10000000000)
	tokenStr3, _ := MakeJWT(id2, secretString2, -time.Hour)

	tests := []struct {
		name         string
		ID           uuid.UUID
		secretString string
		tokenString  string
		wantErr      bool
		match        bool
	}{
		{
			name:         "ID1, Token1, TRUE",
			ID:           id1,
			secretString: secretString1,
			tokenString:  tokenStr1,
			wantErr:      false,
		},
		{
			name:         "ID1, Token2, FALSE",
			ID:           id1,
			secretString: secretString2,
			tokenString:  tokenStr1,
			wantErr:      true,
		},
		{
			name:         "ID2, Token1, FALSE",
			ID:           id2,
			secretString: secretString1,
			tokenString:  tokenStr2,
			wantErr:      true,
		},
		{
			name:         "ID2, Token2, TRUE",
			ID:           id2,
			secretString: secretString2,
			tokenString:  tokenStr2,
			wantErr:      false,
		},
		{
			name:         "ID2, Token2, FALSE, token expires",
			ID:           id2,
			secretString: secretString2,
			tokenString:  tokenStr3,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := ValidateJWT(tt.tokenString, tt.secretString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.ID {
				t.Errorf("ValidateJWT() expects %v, got %v", tt.ID, match)
			}
		})
	}
}
