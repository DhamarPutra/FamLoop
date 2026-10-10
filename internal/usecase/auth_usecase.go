package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"famloop/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrUserAlreadyExists  = errors.New("username, email, atau nomor telepon sudah terdaftar")
	ErrInvalidToken       = errors.New("token tidak valid atau sudah kedaluwarsa")
)

type RegisterInput struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	Avatar      string `json:"avatar,omitempty"`
}

type LoginInput struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
}

type AuthResult struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type AuthUseCase struct {
	userRepo  domain.UserRepository
	jwtSecret []byte
}

func NewAuthUseCase(userRepo domain.UserRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)

	if len(input.Username) < 3 {
		return nil, errors.New("username minimal 3 karakter")
	}
	if !strings.Contains(input.Email, "@") {
		return nil, errors.New("format email tidak valid")
	}
	if len(input.PhoneNumber) < 8 {
		return nil, errors.New("nomor telepon minimal 8 digit")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("password minimal 6 karakter")
	}

	// Cek duplikasi
	if existing, _ := uc.userRepo.GetByUsername(ctx, input.Username); existing != nil {
		return nil, ErrUserAlreadyExists
	}
	if existing, _ := uc.userRepo.GetByEmail(ctx, input.Email); existing != nil {
		return nil, ErrUserAlreadyExists
	}
	if existing, _ := uc.userRepo.GetByPhone(ctx, input.PhoneNumber); existing != nil {
		return nil, ErrUserAlreadyExists
	}

	avatar := input.Avatar
	if avatar == "" {
		avatar = domain.GetRandomDefaultAvatar()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		ID:           uuid.New().String(),
		Username:     input.Username,
		Email:        input.Email,
		PhoneNumber:  input.PhoneNumber,
		Avatar:       avatar,
		PasswordHash: string(hashedPassword),
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	token, err := uc.generateToken(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResult{
		Token: token,
		User:  newUser,
	}, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	loginKey := strings.TrimSpace(input.UsernameOrEmail)
	if loginKey == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	var user *domain.User
	var err error

	if strings.Contains(loginKey, "@") {
		user, err = uc.userRepo.GetByEmail(ctx, strings.ToLower(loginKey))
	} else {
		user, err = uc.userRepo.GetByUsername(ctx, loginKey)
	}

	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := uc.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResult{
		Token: token,
		User:  user,
	}, nil
}

func (uc *AuthUseCase) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (uc *AuthUseCase) GetUserProfile(ctx context.Context, userID string) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

func (uc *AuthUseCase) UpdateProfile(ctx context.Context, userID, username, email, phone string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	phone = strings.TrimSpace(phone)

	if username != "" && len(username) >= 3 {
		user.Username = username
	}
	if email != "" && strings.Contains(email, "@") {
		user.Email = email
	}
	if phone != "" && len(phone) >= 8 {
		user.PhoneNumber = phone
	}

	if err := uc.userRepo.UpdateProfile(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (uc *AuthUseCase) generateToken(user *domain.User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 hari
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtSecret)
}
