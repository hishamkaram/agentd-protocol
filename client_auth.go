package protocol

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const ClientAuthVersionV1 uint8 = 1

const (
	ClientAuthJTISize                   = 16
	ClientAuthP1363SignatureSize        = 64
	ClientAuthP1363SignatureEncodedSize = 86

	ClientAuthModeLegacy = "legacy_hmac"
	ClientAuthModeDevice = "device_proof"

	ClientAuthProofDomain = "agentd/client-auth/v1"
	ClientKeySyncHKDFInfo = "agentd/device-key-sync/v1"
)

var (
	errClientAuthVersion   = errors.New("version must be client_auth_v1")
	errClientAuthSession   = errors.New("session ID must be 1-64 alphanumeric or hyphen characters")
	errClientAuthDevice    = errors.New("device ID must be a 64-character lowercase SHA-256 hex fingerprint")
	errClientAuthTimestamp = errors.New("timestamp must be a positive Unix second")
	errClientAuthJTI       = errors.New("JTI must contain exactly 16 bytes")
)

// P256PublicJWK is the public subset exported by Web Crypto for ECDSA or ECDH.
// Private JWK members are intentionally not representable by this wire type.
type P256PublicJWK struct {
	Kty    string   `json:"kty"`
	Crv    string   `json:"crv"`
	X      string   `json:"x"`
	Y      string   `json:"y"`
	KeyOps []string `json:"key_ops,omitempty"`
	Ext    bool     `json:"ext"`
}

// ClientAuthProof sender-constrains a legacy join token to an enrolled device.
type ClientAuthProof struct {
	Version   uint8  `json:"version"`
	DeviceID  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`
	JTI       string `json:"jti"`
	Signature string `json:"signature"`
}

// ClientAuthEnrollPayload enrolls public device keys after an E2E-authenticated
// pairing. Repeating the same request and keys is idempotent.
type ClientAuthEnrollPayload struct {
	Version         uint8         `json:"version"`
	SessionID       string        `json:"sid"`
	DeviceID        string        `json:"device_id"`
	RequestID       string        `json:"request_id"`
	SigningKey      P256PublicJWK `json:"signing_key"`
	KeyAgreementKey P256PublicJWK `json:"key_agreement_key"`
	ExpiresAt       int64         `json:"expires_at"`
}

// ClientAuthEnrollAckPayload correlates relay enrollment with the daemon and
// the encrypted browser request that initiated it.
type ClientAuthEnrollAckPayload struct {
	Version   uint8  `json:"version"`
	SessionID string `json:"sid"`
	DeviceID  string `json:"device_id"`
	RequestID string `json:"request_id"`
	Enrolled  bool   `json:"enrolled"`
	Code      string `json:"code,omitempty"`
}

// ClientAuthRevokePayload removes an enrolled device without rotating or
// exposing the session traffic key.
type ClientAuthRevokePayload struct {
	Version   uint8  `json:"version"`
	SessionID string `json:"sid"`
	DeviceID  string `json:"device_id"`
	RequestID string `json:"request_id"`
}

// ClientAuthRevokeAckPayload confirms a correlated, idempotent revocation.
type ClientAuthRevokeAckPayload struct {
	Version   uint8  `json:"version"`
	SessionID string `json:"sid"`
	DeviceID  string `json:"device_id"`
	RequestID string `json:"request_id"`
	Revoked   bool   `json:"revoked"`
	Code      string `json:"code,omitempty"`
}

// ClientKeySyncRequestPayload asks the daemon to wrap its current traffic key
// for the device that authenticated the target relay connection.
type ClientKeySyncRequestPayload struct {
	Version    uint8   `json:"version"`
	SessionID  string  `json:"sid"`
	DeviceID   string  `json:"device_id"`
	ClientID   string  `json:"client_id"`
	RequestID  string  `json:"request_id"`
	KnownEpoch *uint64 `json:"known_epoch,omitempty"`
}

// ClientKeySyncResponsePayload contains only public ECDH material and an opaque
// AES-256-GCM frame. The relay never receives the recovered traffic key.
type ClientKeySyncResponsePayload struct {
	Version      uint8         `json:"version"`
	SessionID    string        `json:"sid"`
	DeviceID     string        `json:"device_id"`
	ClientID     string        `json:"client_id"`
	RequestID    string        `json:"request_id"`
	Epoch        uint64        `json:"epoch"`
	ExpiresAt    int64         `json:"expires_at"`
	EphemeralKey P256PublicJWK `json:"ephemeral_key"`
	Nonce        string        `json:"nonce"`
	Ciphertext   string        `json:"ciphertext"`
}

// ParseP256PublicJWK validates and parses the exact public JWK subset accepted
// for device signing and key-agreement credentials.
func ParseP256PublicJWK(jwk P256PublicJWK) (*ecdsa.PublicKey, error) {
	if jwk.Kty != "EC" || jwk.Crv != "P-256" || !jwk.Ext {
		return nil, errors.New("P-256 public JWK metadata is invalid")
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil || len(xBytes) != 32 {
		return nil, errors.New("P-256 public JWK x coordinate is invalid")
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil || len(yBytes) != 32 {
		return nil, errors.New("P-256 public JWK y coordinate is invalid")
	}
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)
	encoded := make([]byte, 65)
	encoded[0] = 4
	copy(encoded[1:33], xBytes)
	copy(encoded[33:65], yBytes)
	if _, err := ecdh.P256().NewPublicKey(encoded); err != nil {
		return nil, errors.New("P-256 public JWK point is not on the curve")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// ClientAuthDeviceID returns the lowercase SHA-256 fingerprint of the
// uncompressed P-256 signing public key.
func ClientAuthDeviceID(jwk P256PublicJWK) (string, error) {
	encoded, err := P256PublicJWKBytes(jwk)
	if err != nil {
		return "", fmt.Errorf("protocol.ClientAuthDeviceID: %w", err)
	}
	fingerprint := sha256.Sum256(encoded)
	return hex.EncodeToString(fingerprint[:]), nil
}

// P256PublicJWKBytes returns the canonical uncompressed SEC1 point encoding.
func P256PublicJWKBytes(jwk P256PublicJWK) ([]byte, error) {
	publicKey, err := ParseP256PublicJWK(jwk)
	if err != nil {
		return nil, err
	}
	encoded := make([]byte, 65)
	encoded[0] = 4
	publicKey.X.FillBytes(encoded[1:33])
	publicKey.Y.FillBytes(encoded[33:65])
	return encoded, nil
}

// ClientAuthProofBytes returns the canonical bytes signed with ECDSA P-256.
// Each field, including the domain separator, has a uint32 big-endian length.
func ClientAuthProofBytes(version uint8, sessionID, deviceID string, timestamp int64, jti []byte) ([]byte, error) {
	if version != ClientAuthVersionV1 {
		return nil, fmt.Errorf("protocol.ClientAuthProofBytes: %w", errClientAuthVersion)
	}
	if !validClientAuthSessionID(sessionID) {
		return nil, fmt.Errorf("protocol.ClientAuthProofBytes: %w", errClientAuthSession)
	}
	if !validClientAuthDeviceID(deviceID) {
		return nil, fmt.Errorf("protocol.ClientAuthProofBytes: %w", errClientAuthDevice)
	}
	if timestamp <= 0 {
		return nil, fmt.Errorf("protocol.ClientAuthProofBytes: %w", errClientAuthTimestamp)
	}
	if len(jti) != ClientAuthJTISize {
		return nil, fmt.Errorf("protocol.ClientAuthProofBytes: %w", errClientAuthJTI)
	}

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	return joinLengthPrefixed(
		[]byte(ClientAuthProofDomain),
		[]byte{version},
		[]byte(sessionID),
		[]byte(deviceID),
		timestampBytes,
		jti,
	), nil
}

// ClientKeySyncAAD returns the authenticated-data binding for a key-sync frame.
func ClientKeySyncAAD(sessionID, deviceID, clientID, requestID string, epoch uint64, expiresAt int64) ([]byte, error) {
	if !validClientAuthSessionID(sessionID) {
		return nil, fmt.Errorf("protocol.ClientKeySyncAAD: %w", errClientAuthSession)
	}
	if !validClientAuthDeviceID(deviceID) {
		return nil, fmt.Errorf("protocol.ClientKeySyncAAD: %w", errClientAuthDevice)
	}
	if clientID == "" || requestID == "" {
		return nil, errors.New("protocol.ClientKeySyncAAD: client and request IDs are required")
	}
	if expiresAt <= 0 {
		return nil, errors.New("protocol.ClientKeySyncAAD: expiry must be a positive Unix second")
	}

	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, epoch)
	expiryBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expiryBytes, uint64(expiresAt))
	return joinLengthPrefixed(
		[]byte(ClientKeySyncHKDFInfo),
		[]byte(sessionID),
		[]byte(deviceID),
		[]byte(clientID),
		[]byte(requestID),
		epochBytes,
		expiryBytes,
	), nil
}

func joinLengthPrefixed(fields ...[]byte) []byte {
	total := 0
	for _, field := range fields {
		total += 4 + len(field)
	}
	out := make([]byte, 0, total)
	var length [4]byte
	for _, field := range fields {
		binary.BigEndian.PutUint32(length[:], uint32(len(field)))
		out = append(out, length[:]...)
		out = append(out, field...)
	}
	return out
}

func validClientAuthSessionID(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for i := range len(value) {
		c := value[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

func validClientAuthDeviceID(value string) bool {
	if len(value) != 64 {
		return false
	}
	for i := range len(value) {
		c := value[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
