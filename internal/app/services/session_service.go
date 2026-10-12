package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const sessionTouchInterval = 5 * time.Minute

// sessionCredentialBytes is the entropy of the credential stored in the
// browser's session cookie.
const sessionCredentialBytes = 32

type SessionCreateInput struct {
	UserID     string
	UserAgent  string
	IPAddress  string
	RememberMe bool
	ExpiresAt  time.Time
}

// SessionService manages server-side sessions and the credentials that map to
// them. Credentials are returned to the caller exactly once, at creation; only
// their hash is persisted.
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

// hashSessionCredential derives the stored lookup value for a credential.
func hashSessionCredential(credential string) string {
	sum := sha256.Sum256([]byte(credential))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// CreateSession starts a session and returns it along with the credential the
// browser must present. The credential is not recoverable afterwards.
func (s *SessionService) CreateSession(input SessionCreateInput) (*models.Session, string, error) {
	db := s.getDB()
	if db == nil {
		return nil, "", ErrUserDBUnavailable
	}

	rawCredential := make([]byte, sessionCredentialBytes)
	if _, err := rand.Read(rawCredential); err != nil {
		return nil, "", fmt.Errorf("failed to generate session credential: %w", err)
	}
	credential := base64.RawURLEncoding.EncodeToString(rawCredential)
	tokenHash := hashSessionCredential(credential)

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
		TokenHash:  &tokenHash,
	}

	if err := db.CreateSession(session); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return session, credential, nil
}

// ResolveSession looks up the session bound to a browser credential and
// verifies it is still usable. An unknown, revoked, or expired credential
// yields an error so the caller can reject the request.
func (s *SessionService) ResolveSession(credential string) (*models.Session, error) {
	if credential == "" {
		return nil, ErrSessionNotFound
	}

	db := s.getDB()
	if db == nil {
		return nil, ErrUserDBUnavailable
	}

	session, err := db.GetSessionByTokenHash(hashSessionCredential(credential))
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
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
		return nil, fmt.Errorf("failed to read session: %w", err)
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
		return nil, fmt.Errorf("failed to read session list: %w", err)
	}

	return sessions, nil
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
		return fmt.Errorf("failed to update session activity time: %w", err)
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
		return fmt.Errorf("failed to revoke session: %w", err)
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

// PurgeUnusableSessions removes session rows that cannot authenticate anything
// because they carry no session credential. They are leftovers from the
// bearer-token era and would otherwise be listed as active sessions.
func (s *SessionService) PurgeUnusableSessions() {
	db := s.getDB()
	if db == nil {
		return
	}

	removed, err := db.DeleteSessionsWithoutCredential()
	if err != nil {
		logger.Warn("failed to purge sessions without credentials", zap.Error(err))
		return
	}
	if removed > 0 {
		logger.Info("removed sessions that carried no credential", zap.Int64("count", removed))
	}
}

func (s *SessionService) StartCleanup(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cleanupExpired()
			}
		}
	}()
	return done
}

func (s *SessionService) cleanupExpired() {
	db := s.getDB()
	if db == nil {
		return
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	if err := db.DeleteExpiredSessions(cutoff); err != nil {
		logger.Warn("session cleanup failed", zap.Error(err))
	}
}
