package services

import (
	"sync"
	"time"

	"github.com/GT-610/tempest-zt/internal/app/config"
	"github.com/GT-610/tempest-zt/internal/app/database"
	"github.com/GT-610/tempest-zt/internal/app/logger"
	"github.com/GT-610/tempest-zt/internal/zerotier"
	"go.uber.org/zap"
)

type SetupStatus struct {
	Initialized             bool             `json:"initialized"`
	DatabaseConfigured      bool             `json:"databaseConfigured"`
	SetupAuthorized         bool             `json:"setupAuthorized"`
	HasAdmin                bool             `json:"hasAdmin"`
	ZeroTierConfigured      bool             `json:"zerotierConfigured"`
	AdminCreationPrepared   bool             `json:"adminCreationPrepared"`
	AdminUsername           string           `json:"adminUsername,omitempty"`
	DatabaseConfig          *SetupDatabase   `json:"databaseConfig,omitempty"`
	ZeroTierConfig          *SetupZeroTier   `json:"zeroTierConfig,omitempty"`
	AllowPublicRegistration bool             `json:"allowPublicRegistration"`
	ZTStatus                *zerotier.Status `json:"ztStatus,omitempty"`
}

type SetupDatabase struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

type SetupZeroTier struct {
	ControllerURL string `json:"controllerUrl"`
	TokenPath     string `json:"tokenPath"`
}

type RuntimeSettings struct {
	AllowPublicRegistration bool `json:"allow_public_registration"`
}

// ztProbeTTL is how long a controller reachability answer is reused.
//
// The setup status endpoint is public and polled, and what it reports about the
// controller is a reachability answer rather than live data, so asking the
// controller once per poll would turn public traffic straight into controller
// traffic. Concurrent callers that miss the cache may probe more than once; the
// bound is one probe per caller rather than one per unit of time, which is
// enough to stop a poll from fanning out.
const ztProbeTTL = 5 * time.Second

// ztProbeResult is a cached controller reachability answer.
type ztProbeResult struct {
	status    *zerotier.Status
	expiresAt time.Time
	// generation is the configuration this answer was produced under. A probe can
	// outlive the configuration it started with, and publishing its answer then
	// would report the previous controller for the life of the cache entry.
	generation uint64
}

type StateService struct {
	cfg *config.Config

	// cacheMutex guards ztClient and ztProbe.
	//
	// Both exist to keep the public setup status endpoint off two expensive
	// paths: building the controller client decrypts the stored token with
	// argon2id (64 MiB, tens of milliseconds per call), and probing the
	// controller is an outbound request with a multi-second timeout.
	cacheMutex sync.Mutex
	ztClient   *zerotier.Client
	ztProbe    *ztProbeResult
	// cacheGeneration advances whenever the configuration the caches were built
	// from changes, so an in-flight probe can tell that its answer is stale.
	cacheGeneration uint64
}

func NewStateServiceWithConfig(cfg *config.Config) *StateService {
	return &StateService{cfg: cfg}
}

func (s *StateService) Config() *config.Config {
	if s.cfg != nil {
		return s.cfg
	}
	return nil
}

func (s *StateService) SetInitialized(initialized bool) error {
	cfg := s.ensureConfig()
	cfg.Initialized = initialized
	return config.SaveConfig(cfg)
}

func (s *StateService) IsInitialized() bool {
	cfg := s.Config()
	return cfg != nil && cfg.Initialized
}

func (s *StateService) DatabaseConfigured() bool {
	cfg := s.Config()
	return cfg != nil && cfg.Database.Type != ""
}

func (s *StateService) DatabaseConfig() database.Config {
	return database.LoadConfigFromApp(s.Config())
}

func (s *StateService) SaveDatabaseConfig(dbCfg database.Config) error {
	cfg := s.ensureConfig()
	if err := database.SaveConfigToApp(cfg, dbCfg); err != nil {
		return err
	}
	return nil
}

func (s *StateService) SaveZeroTierConfig(url, tokenPath string) error {
	cfg := s.ensureConfig()
	err := config.SetZTConfigOn(cfg, url, tokenPath)
	// Dropped whether or not the write succeeded. SetZTConfigOn assigns the new URL
	// and token before it saves, so on a failed save the in-memory configuration is
	// already the new one while the cached client is still the old one; keeping it
	// would leave the two disagreeing.
	s.InvalidateZTClient()
	return err
}

func (s *StateService) RuntimeSettings() RuntimeSettings {
	return RuntimeSettings{
		AllowPublicRegistration: config.AllowPublicRegistration(s.Config()),
	}
}

func (s *StateService) SaveRuntimeSettings(settings RuntimeSettings) error {
	cfg := s.ensureConfig()
	return config.SetAllowPublicRegistrationOn(cfg, settings.AllowPublicRegistration)
}

// CreateZTClient returns the controller client, building it at most once.
//
// Building it decrypts the stored token with argon2id, which costs about 64 MiB
// and tens of milliseconds. Rebuilding that per request made the public setup
// status endpoint a memory amplifier: enough concurrent callers could push the
// process into the kernel's OOM killer while the controller was merely
// unreachable. The result is cached for the life of the process and dropped only
// when the configuration it was built from changes.
//
// A failure is not cached, so a deployment that fixes a bad token path recovers
// without a restart.
func (s *StateService) CreateZTClient() (*zerotier.Client, error) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if s.ztClient != nil {
		return s.ztClient, nil
	}

	client, err := zerotier.NewClientWithConfig(s.Config())
	if err != nil {
		return nil, err
	}

	s.ztClient = client
	return client, nil
}

// InvalidateZTClient drops the cached client and the cached reachability answer
// so the next caller rebuilds both from the current configuration.
func (s *StateService) InvalidateZTClient() {
	s.cacheMutex.Lock()
	s.ztClient = nil
	s.ztProbe = nil
	s.cacheGeneration++
	s.cacheMutex.Unlock()
}

// GetSetupStatus reports how far the first-run wizard has progressed.
//
// setupCaller reports whether the request presented the setup token. The
// endpoint that serves this is public, because the client has to know whether
// to show the wizard before anyone can sign in, so that token is what decides
// whether the reply carries the deployment layout: the administrator name, the
// SQLite path, and the controller URL and token path.
//
// The details are withheld once the instance is initialized even from a caller
// holding the token. The token exists to close the first-run window, and this
// process keeps the one it issued until it restarts, so honouring it afterwards
// would leave the deployment layout readable from a stale log line forever.
func (s *StateService) GetSetupStatus(userService *UserService, networkService *NetworkService, setupCaller bool) SetupStatus {
	cfg := s.Config()
	initialized := s.IsInitialized()
	includeSensitive := setupCaller && !initialized
	databaseConfigured := s.DatabaseConfigured()
	zeroTierConfigured := cfg != nil && cfg.ZeroTier.URL != "" && cfg.ZeroTier.TokenPath != ""

	status := SetupStatus{
		Initialized:             initialized,
		DatabaseConfigured:      databaseConfigured,
		SetupAuthorized:         includeSensitive,
		ZeroTierConfigured:      zeroTierConfigured,
		AdminCreationPrepared:   config.GetTempSetting("admin_creation_reset_done") == "true",
		AllowPublicRegistration: config.AllowPublicRegistration(cfg),
	}

	if includeSensitive && databaseConfigured && cfg != nil {
		status.DatabaseConfig = &SetupDatabase{
			Type: string(cfg.Database.Type),
			Path: cfg.Database.Path,
		}
	}

	if includeSensitive && zeroTierConfigured && cfg != nil {
		status.ZeroTierConfig = &SetupZeroTier{
			ControllerURL: cfg.ZeroTier.URL,
			TokenPath:     cfg.ZeroTier.TokenPath,
		}
	}

	if databaseConfigured && userService != nil {
		if includeSensitive {
			// The wizard shows which account was created, so the name is needed.
			users, err := userService.GetAllUsers()
			if err != nil {
				logger.Warn("GetSetupStatus: GetAllUsers failed", zap.Error(err))
			}
			for _, user := range users {
				if user.Role == "admin" {
					status.HasAdmin = true
					status.AdminUsername = user.Username
					break
				}
			}
		} else {
			// The public answer only needs to know whether anyone can sign in,
			// which is a count rather than a full table read.
			hasAdmin, err := userService.HasAdminUser()
			if err != nil {
				logger.Warn("GetSetupStatus: HasAdminUser failed", zap.Error(err))
			}
			status.HasAdmin = hasAdmin
		}
	}

	if zeroTierConfigured {
		status.ZTStatus = s.probeZeroTier(networkService)
	}

	// An initialized instance always reports a status object. After a restart
	// there is no last-seen answer to give, so it says the version is unknown
	// and the controller is offline rather than omitting the field.
	if status.Initialized && status.ZTStatus == nil {
		status.ZTStatus = &zerotier.Status{
			Version: "unknown",
			Address: "",
			Online:  false,
		}
	}

	return status
}

// probeZeroTier reports whether the controller is reachable, reusing a recent
// answer so that polling this endpoint does not become controller traffic.
func (s *StateService) probeZeroTier(networkService *NetworkService) *zerotier.Status {
	s.cacheMutex.Lock()
	if s.ztProbe != nil && time.Now().Before(s.ztProbe.expiresAt) {
		cached := s.ztProbe.status
		s.cacheMutex.Unlock()
		return cached
	}
	generation := s.cacheGeneration
	s.cacheMutex.Unlock()

	// The probe itself runs without the cache lock: it can block for the
	// client's whole timeout, and holding the lock across it would serialise
	// every caller behind one unreachable controller.
	status := s.probeZeroTierNow(networkService)

	s.cacheMutex.Lock()
	if s.cacheGeneration == generation {
		// Published only while the configuration is still the one it describes.
		// A probe that overlapped a reconfiguration reports its own answer to its
		// own caller and leaves the cache for the new configuration to fill.
		s.ztProbe = &ztProbeResult{
			status:     status,
			expiresAt:  time.Now().Add(ztProbeTTL),
			generation: generation,
		}
	}
	s.cacheMutex.Unlock()

	return status
}

func (s *StateService) probeZeroTierNow(networkService *NetworkService) *zerotier.Status {
	// The bound client is asked first: it is the one real operations use, so a
	// reachable answer here means they will work too.
	if networkService != nil {
		if status, err := networkService.GetStatus(); err == nil {
			return status
		}
	}

	// Fall back to a client built from the stored configuration, which covers
	// the window between saving the controller details and binding a client.
	client, err := s.CreateZTClient()
	if err != nil {
		return nil
	}

	status, err := client.GetStatus()
	if err != nil {
		return nil
	}

	return status
}

func (s *StateService) ensureConfig() *config.Config {
	cfg := s.Config()
	if cfg == nil {
		cfg = &config.Config{}
	}
	s.cfg = cfg
	return cfg
}
