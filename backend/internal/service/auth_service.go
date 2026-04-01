package service

import (
	"context"
	"errors"
	"wechat-enterprise-backend/internal/domain"
	"wechat-enterprise-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *gorm.DB
	jwtManager *jwt.Manager
}

type LoginResult struct {
	Token     string           `json:"token"`
	ExpiresAt string           `json:"expiresAt"`
	User      domain.AdminUser `json:"user"`
}

func NewAuthService(db *gorm.DB, jwtManager *jwt.Manager) *AuthService {
	return &AuthService{db: db, jwtManager: jwtManager}
}

func (s *AuthService) EnsureSeedAdmin(ctx context.Context, username, password, name string) error {
	var user domain.AdminUser
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user = domain.AdminUser{
		Username:     username,
		PasswordHash: string(hash),
		Name:         name,
		Status:       "active",
	}
	return s.db.WithContext(ctx).Create(&user).Error
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var user domain.AdminUser
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("账号或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("账号或密码错误")
	}
	token, expiresAt, err := s.jwtManager.Sign(user.ID, user.Username, user.Name)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token:     token,
		ExpiresAt: expiresAt.Format(timeLayout),
		User:      user,
	}, nil
}

func (s *AuthService) GetByID(ctx context.Context, userID uint) (*domain.AdminUser, error) {
	var user domain.AdminUser
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

const timeLayout = "2006-01-02 15:04:05"
