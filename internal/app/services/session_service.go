package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/google/uuid"
)

const sessionTouchInterval = 5 * time.Minute

type SessionCreateInput struct {
	UserID     string
	UserAgent  string
	IPAddress  string
	RememberMe bool
	ExpiresAt  time.Time
}

type SessionService struct {
	db    database.DBInterface
	mutex sync.RWMutex
}

func NewSessionService(db database.DBInterface) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) SetDB(db database.DBInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.db = db
}

func (s *SessionService) getDB() database.DBInterface {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.db
}

func (s *SessionService) CreateSession(input SessionCreateInput) (*models.Session, error) {
	db := s.getDB()
	if db == nil {
		return nil, ErrUserDBUnavailable
	}

	now := time.Now()
	session := &models.Session{
		ID:         uuid.New().String(),
		UserID:     input.UserID,
		UserAgent:  input.UserAgent,
		IPAddress:  input.IPAddress,
		RememberMe: input.RememberMe,
		LastSeenAt: now,
		ExpiresAt:  input.ExpiresAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := db.CreateSession(session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

func (s *SessionService) GetSessionByID(sessionID string) (*models.Session, error) {
	db := s.getDB()
	if db == nil {
		return nil, ErrUserDBUnavailable
	}

	session, err := db.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("读取会话失败: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func (s *SessionService) GetUserSessions(userID string) ([]*models.Session, error) {
	db := s.getDB()
	if db == nil {
		return nil, ErrUserDBUnavailable
	}

	sessions, err := db.GetSessionsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("读取会话列表失败: %w", err)
	}

	return sessions, nil
}

func (s *SessionService) ValidateSession(userID, sessionID string) (*models.Session, error) {
	session, err := s.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, ErrSessionAccessDenied
	}
	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

func (s *SessionService) TouchSession(session *models.Session) error {
	if session == nil {
		return nil
	}
	if time.Since(session.LastSeenAt) < sessionTouchInterval {
		return nil
	}

	db := s.getDB()
	if db == nil {
		return ErrUserDBUnavailable
	}

	session.LastSeenAt = time.Now()
	session.UpdatedAt = session.LastSeenAt
	if err := db.UpdateSession(session); err != nil {
		return fmt.Errorf("更新会话活跃时间失败: %w", err)
	}

	return nil
}

func (s *SessionService) RevokeSession(userID, sessionID string) error {
	db := s.getDB()
	if db == nil {
		return ErrUserDBUnavailable
	}

	session, err := s.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return ErrSessionAccessDenied
	}
	if session.RevokedAt != nil {
		return nil
	}

	now := time.Now()
	session.RevokedAt = &now
	session.UpdatedAt = now
	if err := db.UpdateSession(session); err != nil {
		return fmt.Errorf("吊销会话失败: %w", err)
	}

	return nil
}

func (s *SessionService) RevokeOtherSessions(userID, currentSessionID string) (int, error) {
	sessions, err := s.GetUserSessions(userID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, session := range sessions {
		if session.ID == currentSessionID || session.RevokedAt != nil {
			continue
		}
		if err := s.RevokeSession(userID, session.ID); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}
