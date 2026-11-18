package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GT-610/tempest-zt/internal/app/models"
	"github.com/GT-610/tempest-zt/internal/app/services"
	"github.com/GT-610/tempest-zt/internal/zerotier"
	"github.com/GT-610/tempest-zt/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturingController records the body of the update request the service sends to
// the controller, which is the only place the "did this request change that
// field" question can be answered.
type capturingController struct {
	updates []map[string]any
}

func (c *capturingController) start(t *testing.T, networkID string) *zerotier.Client {
	t.Helper()

	network := zerotier.Network{ID: networkID, Name: "alpha"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/controller/network/"+networkID:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var decoded map[string]any
			if err := json.Unmarshal(body, &decoded); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			c.updates = append(c.updates, decoded)
			_ = json.NewEncoder(w).Encode(network)
		case r.URL.Path == "/controller/network/"+networkID:
			_ = json.NewEncoder(w).Encode(network)
		case r.URL.Path == "/controller/network/"+networkID+"/member":
			_ = json.NewEncoder(w).Encode([]zerotier.Member{})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return &zerotier.Client{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}
}

func (c *capturingController) lastUpdate(t *testing.T) map[string]any {
	t.Helper()
	require.NotEmpty(t, c.updates, "the service did not send an update")
	return c.updates[len(c.updates)-1]
}

func newUpdatableNetwork(t *testing.T, controller *capturingController, networkID, ownerID string) *services.NetworkService {
	t.Helper()

	db := testutil.NewSQLiteDB(t)
	require.NoError(t, db.CreateUser(&models.User{
		ID: ownerID, Username: ownerID, Password: "hashed", Role: "user",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}))
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: networkID, Name: "alpha", OwnerID: ownerID,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}))

	return services.NewNetworkService(controller.start(t, networkID), db)
}

// TestNetworkServiceMetadataUpdateLeavesBroadcastAlone is the regression test for
// the silent data loss: renaming a network used to send enableBroadcast:false as
// well, turning broadcast off. Measured before the fix — a network created with
// broadcast on reported it off after a rename.
func TestNetworkServiceMetadataUpdateLeavesBroadcastAlone(t *testing.T) {
	controller := &capturingController{}
	service := newUpdatableNetwork(t, controller, "8056c2e21c0000b1", "owner-1")

	_, err := service.UpdateNetworkMetadata("8056c2e21c0000b1", "renamed", "a description", "owner-1")
	require.NoError(t, err)

	update := controller.lastUpdate(t)
	assert.Equal(t, "renamed", update["name"])
	assert.NotContains(t, update, "enableBroadcast", "renaming must not touch broadcast")
	assert.Equal(t, true, update["private"], "managed networks stay private")
}

// TestNetworkServiceDNSSaveLeavesBroadcastAlone covers the same loss on the other
// settings path.
func TestNetworkServiceDNSSaveLeavesBroadcastAlone(t *testing.T) {
	controller := &capturingController{}
	service := newUpdatableNetwork(t, controller, "8056c2e21c0000b2", "owner-1")

	_, err := service.UpdateNetwork("8056c2e21c0000b2", &zerotier.NetworkUpdateRequest{
		DNS: &zerotier.DNSConfig{Domain: "example.test"},
	}, "owner-1")
	require.NoError(t, err)

	update := controller.lastUpdate(t)
	assert.NotContains(t, update, "enableBroadcast", "saving DNS must not touch broadcast")
	assert.Contains(t, update, "dns")
}

// TestNetworkServiceMulticastSaveDoesChangeBroadcast is the other direction: a
// request that names the flag must carry it, including an explicit false.
func TestNetworkServiceMulticastSaveDoesChangeBroadcast(t *testing.T) {
	controller := &capturingController{}
	service := newUpdatableNetwork(t, controller, "8056c2e21c0000b3", "owner-1")

	disabled := false
	limit := 64
	_, err := service.UpdateNetwork("8056c2e21c0000b3", &zerotier.NetworkUpdateRequest{
		MulticastLimit:  &limit,
		EnableBroadcast: &disabled,
	}, "owner-1")
	require.NoError(t, err)

	update := controller.lastUpdate(t)
	assert.Equal(t, false, update["enableBroadcast"], "an explicit false must reach the controller")
	assert.Equal(t, float64(64), update["multicastLimit"])
}
