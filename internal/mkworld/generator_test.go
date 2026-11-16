package mkworld

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validIdentityPublic       = "f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715"
	secondValidIdentityPublic = "6a4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9726"
)

func testRootNode(identity string, endpoints ...string) RootNodeConfig {
	return RootNodeConfig{
		IdentityPublic: identity,
		Comments:       "test root",
		Endpoints:      endpoints,
	}
}

func assertPlanetBinaryMetadata(t *testing.T, data []byte, planetID uint64, birthTime int64, rootNodeCount int) {
	t.Helper()

	nodeCountOffset := 1 + 8 + 8 + ZT_C25519_PUBLIC_KEY_LEN + ZT_C25519_SIGNATURE_LEN
	if len(data) <= nodeCountOffset {
		t.Fatalf("PlanetData length = %d, want more than header length %d", len(data), nodeCountOffset)
	}
	if data[0] != byte(ZT_WORLD_TYPE_PLANET) {
		t.Fatalf("world type = %d, want %d", data[0], ZT_WORLD_TYPE_PLANET)
	}
	if got := binary.BigEndian.Uint64(data[1:9]); got != planetID {
		t.Fatalf("serialized planet ID = %d, want %d", got, planetID)
	}
	if got := binary.BigEndian.Uint64(data[9:17]); got != uint64(birthTime) {
		t.Fatalf("serialized birth time = %d, want %d", got, birthTime)
	}
	if got := int(data[nodeCountOffset]); got != rootNodeCount {
		t.Fatalf("serialized root node count = %d, want %d", got, rootNodeCount)
	}
}

func TestGeneratePlanet_RejectsInvalidInput(t *testing.T) {
	testCases := []struct {
		name    string
		options *GenerateOptions
		target  error
	}{
		{
			name:    "missing root nodes",
			options: &GenerateOptions{},
			target:  ErrNoRootNodes,
		},
		{
			name: "missing identity",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{{Endpoints: []string{"203.0.113.1/9993"}}},
				RecommendValues: true,
			},
			target: ErrIdentityPublicRequired,
		},
		{
			name: "missing endpoints",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{{IdentityPublic: validIdentityPublic}},
				RecommendValues: true,
			},
			target: ErrNoEndpoints,
		},
		{
			name: "invalid identity",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode("invalid", "203.0.113.1/9993")},
				RecommendValues: true,
			},
			target: ErrInvalidIdentity,
		},
		{
			name: "invalid endpoint",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "not-an-endpoint")},
				RecommendValues: true,
			},
			target: ErrInvalidEndpoint,
		},
		{
			name: "duplicate endpoint",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993", "203.0.113.1/9993")},
				RecommendValues: true,
			},
			target: ErrDuplicateEndpoint,
		},
		{
			name: "duplicate identity",
			options: &GenerateOptions{
				RootNodes: []RootNodeConfig{
					testRootNode(validIdentityPublic, "203.0.113.1/9993"),
					testRootNode(validIdentityPublic, "203.0.113.2/9993"),
				},
				RecommendValues: true,
			},
			target: ErrDuplicateIdentity,
		},
		{
			name: "reserved planet id",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
				PlanetID:        uint64(ZT_WORLD_ID_EARTH),
				BirthTime:       time.Now().UnixMilli(),
				RecommendValues: false,
			},
			target: ErrReservedPlanetID,
		},
		{
			name: "invalid birth time",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
				PlanetID:        123456789,
				BirthTime:       1567191349589,
				RecommendValues: false,
			},
			target: ErrInvalidBirthTime,
		},
		{
			name: "too many root nodes",
			options: &GenerateOptions{
				RootNodes: []RootNodeConfig{
					testRootNode(validIdentityPublic, "203.0.113.1/9993"),
					testRootNode(secondValidIdentityPublic, "203.0.113.2/9993"),
					testRootNode("7b4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9727", "203.0.113.3/9993"),
					testRootNode("8c4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9728", "203.0.113.4/9993"),
					testRootNode("9d4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9729", "203.0.113.5/9993"),
				},
				RecommendValues: true,
			},
			target: ErrMaxRootNodesExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GeneratePlanet(tc.options)
			if !errors.Is(err, tc.target) {
				t.Fatalf("GeneratePlanet() error = %v, want %v", err, tc.target)
			}
		})
	}
}

func TestGeneratePlanet_SerializesMetadataAndRootNodes(t *testing.T) {
	result, err := GeneratePlanet(&GenerateOptions{
		RootNodes: []RootNodeConfig{
			testRootNode(validIdentityPublic, "203.0.113.1/9993", "2001:db8::1/9993"),
			testRootNode(secondValidIdentityPublic, "203.0.113.2/9993"),
		},
		RecommendValues: true,
		DownloadName:    "planet.custom",
	})
	if err != nil {
		t.Fatalf("GeneratePlanet() error = %v", err)
	}

	if result.PlanetID == 0 {
		t.Fatalf("PlanetID = 0, want non-zero")
	}
	if result.BirthTime <= 0 {
		t.Fatalf("BirthTime = %d, want positive value", result.BirthTime)
	}
	assertPlanetBinaryMetadata(t, result.PlanetData, result.PlanetID, result.BirthTime, result.RootNodeCount)
	if result.RootNodeCount != 2 {
		t.Fatalf("RootNodeCount = %d, want 2", result.RootNodeCount)
	}
	if result.EndpointCount != 3 {
		t.Fatalf("EndpointCount = %d, want 3", result.EndpointCount)
	}
	if result.DownloadName != "planet.custom" {
		t.Fatalf("DownloadName = %q, want planet.custom", result.DownloadName)
	}
	if !result.UsedRecommendedValues {
		t.Fatal("UsedRecommendedValues = false, want true")
	}
}

func TestGeneratePlanet_UsesCustomMetadataAndSigningKeys(t *testing.T) {
	const planetID = 123456789
	birthTime := time.Now().UnixMilli()
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	if err := CreateSigningKeys(prevPath, curPath); err != nil {
		t.Fatalf("CreateSigningKeys() error = %v", err)
	}

	result, err := GeneratePlanet(&GenerateOptions{
		RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
		SigningKeyPath:  tempDir,
		PlanetID:        planetID,
		BirthTime:       birthTime,
		RecommendValues: false,
	})
	if err != nil {
		t.Fatalf("GeneratePlanet() error = %v", err)
	}

	if result.UsedRecommendedValues {
		t.Fatal("UsedRecommendedValues = true, want false")
	}
	if result.PlanetID != planetID {
		t.Fatalf("PlanetID = %d, want %d", result.PlanetID, planetID)
	}
	if result.BirthTime != birthTime {
		t.Fatalf("BirthTime = %d, want %d", result.BirthTime, birthTime)
	}
	assertPlanetBinaryMetadata(t, result.PlanetData, result.PlanetID, result.BirthTime, result.RootNodeCount)
}

func TestReadSigningKeys_RejectsInvalidLength(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	if err := os.WriteFile(prevPath, []byte("bad"), 0644); err != nil {
		t.Fatalf("write previous key: %v", err)
	}
	if err := os.WriteFile(curPath, []byte("bad"), 0644); err != nil {
		t.Fatalf("write current key: %v", err)
	}

	if _, _, _, _, err := ReadSigningKeys(prevPath, curPath); err == nil {
		t.Fatal("ReadSigningKeys() error = nil, want error")
	}
}

// TestCreateSigningKeysWritesDistinctPairs is the regression test for a rotation
// pair that was really one key twice: previous.c25519 signs the published world
// and current.c25519 is the successor, so they must differ or the rollover the
// model exists for can never happen.
func TestCreateSigningKeysWritesDistinctPairs(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	if err := CreateSigningKeys(prevPath, curPath); err != nil {
		t.Fatalf("CreateSigningKeys() error = %v", err)
	}

	prevFile, err := os.ReadFile(prevPath)
	if err != nil {
		t.Fatalf("read previous key: %v", err)
	}
	curFile, err := os.ReadFile(curPath)
	if err != nil {
		t.Fatalf("read current key: %v", err)
	}

	if string(prevFile) == string(curFile) {
		t.Fatal("the rotation pair is the same key twice; current.c25519 must be a successor, not a copy")
	}

	// Both halves have to be readable as a pair, which is what publishing a
	// planet relies on.
	if _, _, _, _, err := ReadSigningKeys(prevPath, curPath); err != nil {
		t.Fatalf("ReadSigningKeys() error = %v", err)
	}
}

// TestGeneratePlanetAdvancesTheSigningRing is the regression test for consecutive
// publications that cannot both be accepted. A planet advertises current.c25519 as
// the key its own next update must be signed by, so signing the second publication
// with the same key as the first makes every node that took the first reject it.
//
// The assertion is the chain itself: after each publication, previous.c25519 has to
// be the key that publication advertised, which is what current.c25519 held before
// it ran.
func TestGeneratePlanetAdvancesTheSigningRing(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	require.NoError(t, CreateSigningKeys(prevPath, curPath))

	readKey := func(path string) []byte {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		return content
	}
	generate := func() {
		_, err := GeneratePlanet(&GenerateOptions{
			RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
			SigningKeyPath:  tempDir,
			RecommendValues: true,
		})
		require.NoError(t, err)
	}

	// First publication signs with the stored previous key and advertises the
	// stored current key.
	advertisedByFirst := readKey(curPath)
	generate()

	assert.Equal(t, advertisedByFirst, readKey(prevPath),
		"the second publication must be signed by the key the first one advertised")
	assert.NotEqual(t, advertisedByFirst, readKey(curPath),
		"a fresh successor must be generated")

	// Second publication: same rule, one step along the ring.
	advertisedBySecond := readKey(curPath)
	generate()

	assert.Equal(t, advertisedBySecond, readKey(prevPath),
		"the third publication must be signed by the key the second one advertised")
	assert.NotEqual(t, advertisedBySecond, readKey(curPath))

	// And the pair is still readable, so a third publication is possible at all.
	_, _, _, _, err := ReadSigningKeys(prevPath, curPath)
	require.NoError(t, err)
}

// TestCreateSigningKeysRefusesToReplaceAnExistingPair covers the other half: a
// deployed planet names the stored current key, so overwriting it silently would
// break that planet's update chain.
func TestCreateSigningKeysRefusesToReplaceAnExistingPair(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	require.NoError(t, CreateSigningKeys(prevPath, curPath))

	before, err := os.ReadFile(prevPath)
	require.NoError(t, err)

	err = CreateSigningKeys(prevPath, curPath)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSigningKeysExist), "got %v", err)

	after, err := os.ReadFile(prevPath)
	require.NoError(t, err)
	assert.Equal(t, before, after, "a refused call must not have touched the stored keys")
}

// TestLockSigningKeyDirectorySerialisesPerDirectory covers the primitive that keeps
// two generations from advancing one ring from the same snapshot: holders of the
// same directory never overlap, and different directories do not wait for each other.
func TestLockSigningKeyDirectorySerialisesPerDirectory(t *testing.T) {
	var active, peak int32
	var waitGroup sync.WaitGroup

	for worker := 0; worker < 8; worker++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			unlock := lockSigningKeyDirectory("/keys/shared")
			defer unlock()

			current := atomic.AddInt32(&active, 1)
			for {
				observed := atomic.LoadInt32(&peak)
				if current <= observed || atomic.CompareAndSwapInt32(&peak, observed, current) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&active, -1)
		}()
	}
	waitGroup.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&peak),
		"two holders of one directory must never overlap")

	// A different directory has its own lock, so the two do not block each other.
	first := lockSigningKeyDirectory("/keys/first")
	second := lockSigningKeyDirectory("/keys/second")
	second()
	first()
}

// TestCreateSigningKeysAllowsExactlyOneConcurrentCreator covers two requests racing
// to create a pair: the second must lose rather than replace the first, and what it
// wrote must be a complete, readable pair.
func TestCreateSigningKeysAllowsExactlyOneConcurrentCreator(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	var waitGroup sync.WaitGroup
	var mutex sync.Mutex
	var succeeded int
	for worker := 0; worker < 8; worker++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := CreateSigningKeys(prevPath, curPath); err == nil {
				mutex.Lock()
				succeeded++
				mutex.Unlock()
			}
		}()
	}
	waitGroup.Wait()

	assert.Equal(t, 1, succeeded, "only one creator may win")
	_, _, _, _, err := ReadSigningKeys(prevPath, curPath)
	require.NoError(t, err, "the winning pair must be complete")
}

// TestCreateKeyFileRefusesToReplaceAnExistingFile is the primitive behind the above:
// creation is exclusive, so the only copy of a key cannot be truncated by a second
// writer — including one in another process, which the in-process lock cannot stop.
func TestCreateKeyFileRefusesToReplaceAnExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "current.c25519")

	pub, priv, err := GenerateDualPair()
	require.NoError(t, err)
	require.NoError(t, createKeyFile(path, pub, priv))

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	otherPub, otherPriv, err := GenerateDualPair()
	require.NoError(t, err)
	err = createKeyFile(path, otherPub, otherPriv)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSigningKeysExist), "got %v", err)

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, before, after, "a refused create must not have touched the file")
}

// TestKeyRotationLeavesNoTemporaryFiles covers the atomic replacement: the new key is
// written beside the target and renamed over it, which must not leave debris in the
// directory an operator is told to keep.
func TestKeyRotationLeavesNoTemporaryFiles(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	require.NoError(t, CreateSigningKeys(prevPath, curPath))
	_, err := GeneratePlanet(&GenerateOptions{
		RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
		SigningKeyPath:  tempDir,
		RecommendValues: true,
	})
	require.NoError(t, err)

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	assert.ElementsMatch(t, []string{"previous.c25519", "current.c25519"}, names,
		"rotation must leave exactly the two key files")

	// And the rotated pair is readable, which a partial write would prevent.
	_, _, _, _, err = ReadSigningKeys(prevPath, curPath)
	require.NoError(t, err)
}

// TestFillKeyFileRemovesAFileItCouldNotWrite covers the recovery path: a key file that
// was created but could not be filled must not be left behind, because the existence
// check would then refuse every later attempt and only a manual deletion would clear it.
func TestFillKeyFileRemovesAFileItCouldNotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "previous.c25519")

	// Created exclusively, as createKeyFile does, then closed so the content write
	// fails the way a storage error would.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	require.NoError(t, err)
	require.NoError(t, file.Close())

	pub, priv, err := GenerateDualPair()
	require.NoError(t, err)
	require.Error(t, fillKeyFile(path, file, pub, priv))

	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr),
		"a key file whose content could not be written must be removed, not left to block the next attempt")
}

// TestAdvanceSigningKeysWritesTheSuccessorFirst covers the ordering that keeps a failed
// generation recoverable.
//
// A generation that returns an error never hands a planet to the caller, so the stored
// pair has to remain usable for a retry. The signer is what the *deployed* world expects
// to verify, so it is promoted last: with the writer failing on the successor, nothing
// has changed; with it failing on the promotion, the successor is already in place and
// the signer is untouched, so the retry is accepted.
func TestAdvanceSigningKeysWritesTheSuccessorFirst(t *testing.T) {
	dir := t.TempDir()
	prevPath := filepath.Join(dir, "previous.c25519")
	curPath := filepath.Join(dir, "current.c25519")
	require.NoError(t, CreateSigningKeys(prevPath, curPath))

	signerBefore, err := os.ReadFile(prevPath)
	require.NoError(t, err)
	pub, priv, err := GenerateDualPair()
	require.NoError(t, err)

	// The successor write fails: the signer must not be touched at all.
	var attempted []string
	failingSuccessor := func(path string, pub [ZT_C25519_PUBLIC_KEY_LEN]byte, priv [ZT_C25519_PRIVATE_KEY_LEN]byte) error {
		attempted = append(attempted, path)
		if path == curPath {
			return errors.New("simulated storage failure")
		}
		return writeKeyFile(path, pub, priv)
	}
	require.Error(t, advanceSigningKeys(prevPath, curPath, pub, priv, failingSuccessor))
	assert.Equal(t, []string{curPath}, attempted,
		"the successor must be attempted before the signer, so nothing is consumed")

	signerAfter, err := os.ReadFile(prevPath)
	require.NoError(t, err)
	assert.Equal(t, signerBefore, signerAfter)

	// The promotion write fails: the successor is already on disk and the signer still
	// holds what the deployed world expects, so a retry remains valid.
	successorBefore, err := os.ReadFile(curPath)
	require.NoError(t, err)
	failingPromotion := func(path string, pub [ZT_C25519_PUBLIC_KEY_LEN]byte, priv [ZT_C25519_PRIVATE_KEY_LEN]byte) error {
		if path == prevPath {
			return errors.New("simulated storage failure")
		}
		return writeKeyFile(path, pub, priv)
	}
	require.Error(t, advanceSigningKeys(prevPath, curPath, pub, priv, failingPromotion))

	signerAfterPromotion, err := os.ReadFile(prevPath)
	require.NoError(t, err)
	assert.Equal(t, signerBefore, signerAfterPromotion,
		"a failed promotion must leave the signer the deployed world expects")

	successorAfter, err := os.ReadFile(curPath)
	require.NoError(t, err)
	assert.NotEqual(t, successorBefore, successorAfter,
		"the successor was written before the promotion was attempted")
}
