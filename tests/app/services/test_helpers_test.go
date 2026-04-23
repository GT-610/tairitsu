package services

import (
	"github.com/GT-610/tairitsu/internal/app/models"
)

type stateServiceDBStub struct {
	users []*models.User
}

func (s *stateServiceDBStub) Init() error { return nil }
func (s *stateServiceDBStub) CreateUser(user *models.User) error {
	s.users = append(s.users, user)
	return nil
}
func (s *stateServiceDBStub) GetUserByID(id string) (*models.User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}
func (s *stateServiceDBStub) GetUserByUsername(username string) (*models.User, error) {
	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}
func (s *stateServiceDBStub) GetAllUsers() ([]*models.User, error) { return s.users, nil }
func (s *stateServiceDBStub) UpdateUser(user *models.User) error   { return nil }
func (s *stateServiceDBStub) DeleteUser(id string) error           { return nil }
func (s *stateServiceDBStub) CreateSession(session *models.Session) error {
	return nil
}
func (s *stateServiceDBStub) GetSessionByID(id string) (*models.Session, error) {
	return nil, nil
}
func (s *stateServiceDBStub) GetSessionsByUserID(userID string) ([]*models.Session, error) {
	return []*models.Session{}, nil
}
func (s *stateServiceDBStub) UpdateSession(session *models.Session) error { return nil }
func (s *stateServiceDBStub) CreateNetwork(network *models.Network) error {
	return nil
}
func (s *stateServiceDBStub) GetNetworkByID(id string) (*models.Network, error) {
	return nil, nil
}
func (s *stateServiceDBStub) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	return []*models.Network{}, nil
}
func (s *stateServiceDBStub) GetAllNetworks() ([]*models.Network, error) {
	return []*models.Network{}, nil
}
func (s *stateServiceDBStub) UpdateNetwork(network *models.Network) error { return nil }
func (s *stateServiceDBStub) DeleteNetwork(id string) error               { return nil }
func (s *stateServiceDBStub) HasAdminUser() (bool, error) {
	for _, user := range s.users {
		if user.Role == "admin" {
			return true, nil
		}
	}
	return false, nil
}
func (s *stateServiceDBStub) Close() error { return nil }
