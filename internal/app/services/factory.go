package services

import (
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/zerotier"
)

// ServiceFactory is a factory for creating service instances with consistent dependencies
// It centralizes the creation of all services, ensuring they share the same dependencies
// and follow a consistent initialization pattern.
type ServiceFactory struct {
	db        database.Database // Database interface for data operations
	ztClient  *zerotier.Client  // ZeroTier client for API interactions
	jwtSecret string            // JWT secret for token generation and validation
}

// NewServiceFactory creates a new service factory instance with the given dependencies
// Parameters:
//   - db: Database interface for data operations
//   - ztClient: ZeroTier client for API interactions
//   - jwtSecret: JWT secret for token generation and validation
//
// Returns:
//   - *ServiceFactory: A new service factory instance
func NewServiceFactory(db database.Database, ztClient *zerotier.Client, jwtSecret string) *ServiceFactory {
	return &ServiceFactory{
		db:        db,
		ztClient:  ztClient,
		jwtSecret: jwtSecret,
	}
}

// CreateUserService creates a new UserService instance with the factory's database dependency
// Returns:
//   - *UserService: A new UserService instance
func (f *ServiceFactory) CreateUserService() *UserService {
	return NewUserServiceWithDB(f.db)
}

// CreateNetworkService creates a new NetworkService instance with the factory's dependencies
// Returns:
//   - *NetworkService: A new NetworkService instance
func (f *ServiceFactory) CreateNetworkService() *NetworkService {
	return NewNetworkService(f.ztClient, f.db)
}

// CreateJWTService creates a new JWTService instance with the factory's JWT secret
// Returns:
//   - *JWTService: A new JWTService instance
func (f *ServiceFactory) CreateJWTService() *JWTService {
	return NewJWTService(f.jwtSecret)
}
