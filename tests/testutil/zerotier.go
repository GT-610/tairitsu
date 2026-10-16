package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/zerotier"
)

// FakeController is an in-process stand-in for a ZeroTier controller's HTTP API,
// covering the endpoints the client reads: the controller status, the network
// list, a network, its member list, and a single member.
//
// The original tests each spun up a bespoke handler for one response shape.
// Folding them into one server keeps a single description of the API surface,
// so a change to how the client reads it is exercised everywhere at once.
type FakeController struct {
	// Networks are returned by the network list and by ID.
	Networks map[string]zerotier.Network
	// Members is the member list per network ID, in the order the controller
	// should report them. A network absent from this map answers the member
	// list with 404, which is what a controller does for a network it does not
	// know.
	Members map[string][]zerotier.Member
	// MemberObjects is a member list keyed by member ID, which some controller
	// builds answer with. Each key becomes the member's ID.
	MemberObjects map[string]map[string]zerotier.Member
	// MemberIndexes is an index response keyed by member ID. The member objects
	// are looked up in Members by the index's position.
	MemberIndexes map[string]map[string]int
	// Status is served by /status. TCPFallback is reported through both field
	// names a controller build may use, so the caller's preference is exercised
	// whichever of the two it reads.
	Status zerotier.Status
	// StatusCode, when non-zero, makes /status fail with that status.
	StatusCode int
	// BrokenNetworkID is announced in the network list but fails on read,
	// covering a caller that has to survive one unreadable network.
	BrokenNetworkID string

	// Delay is slept before answering a member list request, for tests that
	// assert requests overlap.
	Delay time.Duration
	// Hooks run around member list requests.
	Hooks MemberHooks
	// MemberResponse, when set, replaces the rendered member list body. It
	// exists for the shapes the encoders above do not model, such as a response
	// that decodes but does not look like a member collection.
	MemberResponse func() any

	// Server is the listener serving this controller.
	Server *httptest.Server
}

// MemberHooks observes member list requests. Both may be nil.
type MemberHooks struct {
	// OnMemberRequest runs when a member list request arrives, before Delay.
	OnMemberRequest func()
	// OnMemberResponse runs after the response is written.
	OnMemberResponse func()
}

// NewFakeController starts a controller serving the supplied configuration and
// stops it when the caller's test finishes.
func NewFakeController(t *testing.T, controller *FakeController) *zerotier.Client {
	t.Helper()

	controller.Server = httptest.NewServer(controller)
	t.Cleanup(controller.Server.Close)

	return &zerotier.Client{
		BaseURL:    controller.Server.URL,
		Token:      "test-token",
		HTTPClient: controller.Server.Client(),
	}
}

// ServeHTTP implements the slice of the controller API the client reads.
func (f *FakeController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.URL.Path == "/status":
		f.serveStatus(w)
	case r.URL.Path == "/controller/network":
		f.serveNetworkList(w)
	case strings.HasPrefix(r.URL.Path, "/controller/network/"):
		f.serveNetwork(w, strings.TrimPrefix(r.URL.Path, "/controller/network/"))
	default:
		http.NotFound(w, r)
	}
}

func (f *FakeController) serveStatus(w http.ResponseWriter) {
	if f.StatusCode != 0 {
		http.Error(w, "broken status", f.StatusCode)
		return
	}

	writeJSON(w, map[string]any{
		"version":              f.Status.Version,
		"address":              f.Status.Address,
		"online":               f.Status.Online,
		"tcpFallbackAvailable": f.Status.TCPFallback,
		"tcpFallbackActive":    f.Status.TCPFallback,
		"apiReady":             f.Status.APIReady,
	})
}

func (f *FakeController) serveNetworkList(w http.ResponseWriter) {
	ids := make([]string, 0, len(f.Networks)+1)
	for id := range f.Networks {
		ids = append(ids, id)
	}
	if f.BrokenNetworkID != "" {
		ids = append(ids, f.BrokenNetworkID)
	}

	writeJSON(w, ids)
}

func (f *FakeController) serveNetwork(w http.ResponseWriter, rest string) {
	networkID, sub, hasSub := strings.Cut(rest, "/")

	switch {
	case hasSub && sub == "member":
		f.serveMembers(w, networkID)
	case hasSub && strings.HasPrefix(sub, "member/"):
		f.serveMember(w, networkID, strings.TrimPrefix(sub, "member/"))
	default:
		f.serveNetworkDetail(w, networkID)
	}
}

func (f *FakeController) serveNetworkDetail(w http.ResponseWriter, networkID string) {
	if networkID == f.BrokenNetworkID {
		http.Error(w, "broken", http.StatusInternalServerError)
		return
	}

	network, ok := f.Networks[networkID]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, network)
}

func (f *FakeController) serveMember(w http.ResponseWriter, networkID, memberID string) {
	if object, ok := f.MemberObjects[networkID]; ok {
		if member, ok := object[memberID]; ok {
			member.ID = memberID
			writeJSON(w, member)
			return
		}
	}

	for _, member := range f.Members[networkID] {
		if member.ID == memberID {
			writeJSON(w, member)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (f *FakeController) serveMembers(w http.ResponseWriter, networkID string) {
	if f.Hooks.OnMemberRequest != nil {
		f.Hooks.OnMemberRequest()
	}
	if f.Delay > 0 {
		time.Sleep(f.Delay)
	}

	if f.MemberResponse != nil {
		writeJSON(w, f.MemberResponse())
		if f.Hooks.OnMemberResponse != nil {
			f.Hooks.OnMemberResponse()
		}
		return
	}

	f.writeMembers(w, networkID)

	if f.Hooks.OnMemberResponse != nil {
		f.Hooks.OnMemberResponse()
	}
}

func (f *FakeController) writeMembers(w http.ResponseWriter, networkID string) {
	if indexes, ok := f.MemberIndexes[networkID]; ok {
		writeJSON(w, indexes)
		return
	}

	if objects, ok := f.MemberObjects[networkID]; ok {
		// The key is the member ID, so the stored values do not have to repeat
		// it.
		byID := make(map[string]zerotier.Member, len(objects))
		for memberID, member := range objects {
			member.ID = memberID
			byID[memberID] = member
		}
		writeJSON(w, byID)
		return
	}

	members, ok := f.Members[networkID]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, members)
}

func writeJSON(w http.ResponseWriter, payload any) {
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
