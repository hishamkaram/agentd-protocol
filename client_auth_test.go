package protocol

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestClientAuthProofBytesGolden(t *testing.T) {
	t.Parallel()

	jti, err := hex.DecodeString("00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatalf("decode JTI: %v", err)
	}
	got, err := ClientAuthProofBytes(
		ClientAuthVersionV1,
		"session-123",
		strings.Repeat("0123456789abcdef", 4),
		1772671808,
		jti,
	)
	if err != nil {
		t.Fatalf("ClientAuthProofBytes() error = %v", err)
	}

	const want = "000000156167656e74642f636c69656e742d617574682f7631" +
		"0000000101" +
		"0000000b73657373696f6e2d313233" +
		"0000004030313233343536373839616263646566303132333435363738396162636465663031323334353637383961626364656630313233343536373839616263646566" +
		"000000080000000069a8d340" +
		"0000001000112233445566778899aabbccddeeff"
	if gotHex := hex.EncodeToString(got); gotHex != want {
		t.Fatalf("ClientAuthProofBytes() = %s, want %s", gotHex, want)
	}
}

func TestClientAuthProofBytesRejectsInvalidBindings(t *testing.T) {
	t.Parallel()

	validJTI := make([]byte, ClientAuthJTISize)
	tests := []struct {
		name      string
		version   uint8
		sessionID string
		deviceID  string
		timestamp int64
		jti       []byte
	}{
		{name: "version", version: 2, sessionID: "session", deviceID: strings.Repeat("a", 64), timestamp: 1, jti: validJTI},
		{name: "session", version: ClientAuthVersionV1, deviceID: strings.Repeat("a", 64), timestamp: 1, jti: validJTI},
		{name: "device", version: ClientAuthVersionV1, sessionID: "session", timestamp: 1, jti: validJTI},
		{name: "timestamp", version: ClientAuthVersionV1, sessionID: "session", deviceID: strings.Repeat("a", 64), jti: validJTI},
		{name: "jti", version: ClientAuthVersionV1, sessionID: "session", deviceID: strings.Repeat("a", 64), timestamp: 1, jti: validJTI[:15]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ClientAuthProofBytes(tt.version, tt.sessionID, tt.deviceID, tt.timestamp, tt.jti); err == nil {
				t.Fatal("ClientAuthProofBytes() error = nil, want validation error")
			}
		})
	}
}

func TestClientKeySyncAADGolden(t *testing.T) {
	t.Parallel()

	got, err := ClientKeySyncAAD(
		"session-123",
		strings.Repeat("a", 64),
		"client-456",
		"request-789",
		7,
		1772671808,
	)
	if err != nil {
		t.Fatalf("ClientKeySyncAAD() error = %v", err)
	}
	const want = "000000196167656e74642f6465766963652d6b65792d73796e632f7631" +
		"0000000b73657373696f6e2d313233" +
		"0000004061616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161" +
		"0000000a636c69656e742d343536" +
		"0000000b726571756573742d373839" +
		"000000080000000000000007" +
		"000000080000000069a8d340"
	if gotHex := hex.EncodeToString(got); gotHex != want {
		t.Fatalf("ClientKeySyncAAD() = %s, want %s", gotHex, want)
	}
}

func TestClientKeySyncAADRejectsInvalidBindings(t *testing.T) {
	t.Parallel()
	deviceID := strings.Repeat("a", 64)
	tests := []struct {
		name      string
		sessionID string
		deviceID  string
		clientID  string
		requestID string
		expiresAt int64
	}{
		{name: "session", deviceID: deviceID, clientID: "client", requestID: "request", expiresAt: 1},
		{name: "device", sessionID: "session", clientID: "client", requestID: "request", expiresAt: 1},
		{name: "client", sessionID: "session", deviceID: deviceID, requestID: "request", expiresAt: 1},
		{name: "request", sessionID: "session", deviceID: deviceID, clientID: "client", expiresAt: 1},
		{name: "expiry", sessionID: "session", deviceID: deviceID, clientID: "client", requestID: "request"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ClientKeySyncAAD(test.sessionID, test.deviceID, test.clientID, test.requestID, 0, test.expiresAt); err == nil {
				t.Fatal("ClientKeySyncAAD() error = nil, want validation error")
			}
		})
	}
}

func TestClientAuthDeviceIDGolden(t *testing.T) {
	t.Parallel()

	jwk := P256PublicJWK{
		Kty: "EC",
		Crv: "P-256",
		X:   "axfR8uEsQkf4vOblY6RA8ncDfYEt6zOg9KE5RdiYwpY",
		Y:   "T-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU",
		Ext: true,
	}
	deviceID, err := ClientAuthDeviceID(jwk)
	if err != nil {
		t.Fatalf("ClientAuthDeviceID() error = %v", err)
	}
	const want = "698bea63dc44a344663ff1429aea10842df27b6b991ef25866b2c6c02cdcc5be"
	if deviceID != want {
		t.Fatalf("ClientAuthDeviceID() = %q, want %q", deviceID, want)
	}
}

func TestParseP256PublicJWKRejectsPrivateOrMalformedKeys(t *testing.T) {
	t.Parallel()

	valid := P256PublicJWK{
		Kty: "EC",
		Crv: "P-256",
		X:   "axfR8uEsQkf4vOblY6RA8ncDfYEt6zOg9KE5RdiYwpY",
		Y:   "T-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU",
		Ext: true,
	}
	tests := []struct {
		name   string
		mutate func(*P256PublicJWK)
	}{
		{name: "wrong curve", mutate: func(jwk *P256PublicJWK) { jwk.Crv = "P-384" }},
		{name: "not extractable public", mutate: func(jwk *P256PublicJWK) { jwk.Ext = false }},
		{name: "short x", mutate: func(jwk *P256PublicJWK) { jwk.X = "AA" }},
		{name: "off curve", mutate: func(jwk *P256PublicJWK) { jwk.Y = jwk.X }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jwk := valid
			tc.mutate(&jwk)
			if _, err := ParseP256PublicJWK(jwk); err == nil {
				t.Fatal("ParseP256PublicJWK() error = nil")
			}
		})
	}
}

func TestClientAuthWirePayloadsRoundTrip(t *testing.T) {
	t.Parallel()

	epoch := uint64(9)
	join := JoinPayload{
		SessionID: "session-123",
		JWT:       "legacy-token",
		ClientAuth: &ClientAuthProof{
			Version:   ClientAuthVersionV1,
			DeviceID:  strings.Repeat("a", 64),
			Timestamp: 1772671808,
			JTI:       "ABEiM0RVZneImaq7zN3u_w",
			Signature: strings.Repeat("s", ClientAuthP1363SignatureEncodedSize),
		},
	}
	ack := AckPayload{
		SessionID:    join.SessionID,
		ClientID:     "connection-1",
		Capabilities: []string{CapabilityClientAuthV1, CapabilityClientKeySyncV1},
		AuthMode:     ClientAuthModeDevice,
		KeyEpoch:     &epoch,
	}
	enroll := ClientAuthEnrollPayload{
		Version:         ClientAuthVersionV1,
		SessionID:       join.SessionID,
		DeviceID:        join.ClientAuth.DeviceID,
		RequestID:       "request-1",
		ClientID:        ack.ClientID,
		SigningKey:      testP256PublicJWK("verify"),
		KeyAgreementKey: testP256PublicJWK(),
		ExpiresAt:       1772758208,
	}
	keySync := ClientKeySyncResponsePayload{
		Version:      ClientAuthVersionV1,
		SessionID:    join.SessionID,
		DeviceID:     join.ClientAuth.DeviceID,
		ClientID:     ack.ClientID,
		RequestID:    "request-2",
		Epoch:        epoch,
		ExpiresAt:    1772671838,
		EphemeralKey: testP256PublicJWK(),
		Nonce:        "ABEiM0RVZneImaq7",
		Ciphertext:   "opaque-ciphertext",
	}

	for name, value := range map[string]any{
		"join":     join,
		"ack":      ack,
		"enroll":   enroll,
		"key_sync": keySync,
	} {
		name, value := name, value
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if !json.Valid(encoded) {
				t.Fatalf("json.Marshal() produced invalid JSON: %q", encoded)
			}
		})
	}
}

func TestClientAuthFieldsRemainAdditive(t *testing.T) {
	t.Parallel()

	joinJSON, err := json.Marshal(JoinPayload{SessionID: "session", JWT: "token"})
	if err != nil {
		t.Fatalf("marshal join: %v", err)
	}
	if strings.Contains(string(joinJSON), "client_auth") {
		t.Fatalf("legacy join unexpectedly contains client_auth: %s", joinJSON)
	}

	ackJSON, err := json.Marshal(AckPayload{SessionID: "session"})
	if err != nil {
		t.Fatalf("marshal ack: %v", err)
	}
	for _, field := range []string{"capabilities", "auth_mode", "key_epoch"} {
		if strings.Contains(string(ackJSON), field) {
			t.Fatalf("legacy ack unexpectedly contains %s: %s", field, ackJSON)
		}
	}

	enrollJSON, err := json.Marshal(ClientAuthEnrollPayload{SessionID: "session"})
	if err != nil {
		t.Fatalf("marshal enrollment: %v", err)
	}
	if strings.Contains(string(enrollJSON), "client_id") {
		t.Fatalf("background enrollment unexpectedly contains client_id: %s", enrollJSON)
	}
}

func TestClientAuthControlAndCapabilityConstants(t *testing.T) {
	t.Parallel()

	controls := map[ControlType]string{
		CtrlClientAuthEnroll:      "client_auth_enroll",
		CtrlClientAuthEnrollAck:   "client_auth_enroll_ack",
		CtrlClientAuthRevoke:      "client_auth_revoke",
		CtrlClientAuthRevokeAck:   "client_auth_revoke_ack",
		CtrlClientKeySyncRequest:  "client_key_sync_request",
		CtrlClientKeySyncResponse: "client_key_sync_response",
	}
	for got, want := range controls {
		if string(got) != want {
			t.Fatalf("control constant = %q, want %q", got, want)
		}
	}

	offer := FullProtocolHelloOffer()
	v2Capabilities := offer.CapabilitiesByProtocol[ProtocolV2]
	for _, want := range []string{CapabilityClientAuthV1, CapabilityClientKeySyncV1} {
		if !containsClientAuthCapability(v2Capabilities, want) {
			t.Fatalf("ProtocolV2 capabilities = %v, missing %q", v2Capabilities, want)
		}
	}
}

func testP256PublicJWK(keyOps ...string) P256PublicJWK {
	return P256PublicJWK{
		Kty:    "EC",
		Crv:    "P-256",
		X:      strings.Repeat("x", 43),
		Y:      strings.Repeat("y", 43),
		KeyOps: keyOps,
		Ext:    true,
	}
}

func containsClientAuthCapability(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
