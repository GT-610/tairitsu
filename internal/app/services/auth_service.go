package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tairitsu/tairitsu/internal/app/models"
)

// AuthKeyType 认证密钥类型
type AuthKeyType string

const (
	// KeyTypeNormal 普通认证密钥
	KeyTypeNormal AuthKeyType = "normal"
	// KeyTypeSetupWizard 设置向导临时密钥
	KeyTypeSetupWizard AuthKeyType = "setup_wizard"
)

// AuthClaims 认证声明结构
type AuthClaims struct {
	UserID   string      `json:"user_id,omitempty"`
	Username string      `json:"username,omitempty"`
	Role     string      `json:"role,omitempty"`
	KeyType  AuthKeyType `json:"key_type"`
	jwt.RegisteredClaims
}

// AuthService 认证服务
type AuthService struct {
	keys          map[AuthKeyType]string
	tempKeys      map[string]time.Time // 临时密钥和过期时间
	keysMutex     sync.RWMutex
	tempKeysMutex sync.RWMutex
	normalExpiry  time.Duration
	tempExpiry    time.Duration
}

// NewAuthService 创建认证服务实例
func NewAuthService(normalSecret string) *AuthService {
	service := &AuthService{
		keys:         make(map[AuthKeyType]string),
		tempKeys:     make(map[string]time.Time),
		normalExpiry: time.Hour * 24,   // 普通令牌24小时过期
		tempExpiry:   time.Minute * 30, // 临时令牌30分钟过期
	}

	// 设置默认的普通密钥
	service.keys[KeyTypeNormal] = normalSecret

	return service
}

// SetKey 设置指定类型的密钥
func (s *AuthService) SetKey(keyType AuthKeyType, secret string) {
	s.keysMutex.Lock()
	defer s.keysMutex.Unlock()
	s.keys[keyType] = secret
}

// GetKey 获取指定类型的密钥
func (s *AuthService) GetKey(keyType AuthKeyType) (string, error) {
	s.keysMutex.RLock()
	defer s.keysMutex.RUnlock()

	key, exists := s.keys[keyType]
	if !exists {
		return "", fmt.Errorf("密钥类型不存在: %s", keyType)
	}
	return key, nil
}

// GenerateToken 生成令牌
func (s *AuthService) GenerateToken(keyType AuthKeyType, user *models.User) (string, error) {
	// 获取密钥
	secret, err := s.GetKey(keyType)
	if err != nil {
		return "", err
	}

	// 设置过期时间
	expiry := s.normalExpiry
	if keyType == KeyTypeSetupWizard {
		expiry = s.tempExpiry
	}

	// 创建声明
	claims := AuthClaims{
		KeyType: keyType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 如果提供了用户信息，则添加到声明中
	if user != nil {
		claims.UserID = user.ID
		claims.Username = user.Username
		claims.Role = user.Role
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

// GenerateTempToken 生成临时令牌
func (s *AuthService) GenerateTempToken() (string, error) {
	// 生成临时密钥
	tempKey := generateRandomString(32)

	// 保存临时密钥及其过期时间
	s.tempKeysMutex.Lock()
	s.tempKeys[tempKey] = time.Now().Add(s.tempExpiry)
	s.tempKeysMutex.Unlock()

	// 使用临时密钥生成令牌
	claims := AuthClaims{
		KeyType: KeyTypeSetupWizard,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tempExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tempKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(tokenString string) (*AuthClaims, error) {
	// 解析令牌以获取密钥类型
	parser := jwt.Parser{}
	token, _, err := parser.ParseUnverified(tokenString, &AuthClaims{})
	if err != nil {
		return nil, err
	}

	// 获取声明中的密钥类型
	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, errors.New("无效的令牌声明")
	}

	// 根据密钥类型选择验证密钥
	var secret string
	if claims.KeyType == KeyTypeSetupWizard {
		// 对于临时密钥，我们需要验证签名
		// 这里我们将尝试所有有效的临时密钥
		valid, err := s.validateWithTempKeys(tokenString)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, errors.New("无效的临时令牌")
		}
		// 验证通过，返回声明
		return claims, nil
	} else {
		// 对于普通密钥，使用配置的密钥
		secret, err = s.GetKey(claims.KeyType)
		if err != nil {
			return nil, err
		}
	}

	// 使用选定的密钥验证令牌
	verifiedToken, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证声明
	if verifiedClaims, ok := verifiedToken.Claims.(*AuthClaims); ok && verifiedToken.Valid {
		return verifiedClaims, nil
	}

	return nil, errors.New("无效的令牌")
}

// validateWithTempKeys 使用临时密钥验证令牌
func (s *AuthService) validateWithTempKeys(tokenString string) (bool, error) {
	s.tempKeysMutex.RLock()
	defer s.tempKeysMutex.RUnlock()

	// 清理过期的临时密钥
	s.cleanupExpiredTempKeys()

	// 尝试使用每个临时密钥验证令牌
	for key := range s.tempKeys {
		claims := &AuthClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(key), nil
		})

		if err == nil && token.Valid && claims.KeyType == KeyTypeSetupWizard {
			return true, nil
		}
	}

	return false, nil
}

// cleanupExpiredTempKeys 清理过期的临时密钥
func (s *AuthService) cleanupExpiredTempKeys() {
	now := time.Now()
	for key, expiry := range s.tempKeys {
		if now.After(expiry) {
			delete(s.tempKeys, key)
		}
	}
}

// RevokeTempTokens 撤销所有临时令牌
func (s *AuthService) RevokeTempTokens() {
	s.tempKeysMutex.Lock()
	defer s.tempKeysMutex.Unlock()

	// 清空所有临时密钥
	s.tempKeys = make(map[string]time.Time)
}

// RevokeAllTempKeys 撤销所有临时密钥（与RevokeTempTokens功能相同，提供兼容性接口）
func (s *AuthService) RevokeAllTempKeys() {
	s.RevokeTempTokens()
}

// GenerateSetupWizardToken 生成设置向导临时令牌
func (s *AuthService) GenerateSetupWizardToken() (string, error) {
	return s.GenerateTempToken()
}

// IsTempTokenValid 检查临时令牌是否有效
func (s *AuthService) IsTempTokenValid(tokenString string) bool {
	_, err := s.ValidateToken(tokenString)
	return err == nil
}
