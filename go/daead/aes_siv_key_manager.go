// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////

package daead

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"github.com/google/tink/go/core/registry"
	"github.com/google/tink/go/daead/subtle"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/subtle/random"

	aspb "github.com/google/tink/go/proto/aes_siv_go_proto"
	tpb "github.com/google/tink/go/proto/tink_go_proto"
)

const (
	aesSIVKeyVersion = 0
	aesSIVTypeURL    = "type.googleapis.com/google.crypto.tink.AesSivKey"
)

var (
	errInvalidAESSIVKeyFormat = errors.New("aes_siv_key_manager: invalid key format")
	errInvalidAESSIVKeySize   = fmt.Errorf("aes_siv_key_manager: key size != %d", subtle.AESSIVKeySize)
)

// aesSIVKeyManager generates AES-SIV keys and produces instances of AES-SIV.
type aesSIVKeyManager struct{}

// Assert that aesSIVKeyManager implements the KeyManager interface.
var _ registry.KeyManager = (*aesSIVKeyManager)(nil)

// Primitive constructs an AES-SIV for the given serialized AesSivKey.
func (km *aesSIVKeyManager) Primitive(serializedKey []byte) (interface{}, error) {
	if len(serializedKey) == 0 {
		return nil, errors.New("aes_siv_key_manager: invalid key")
	}
	key := new(aspb.AesSivKey)
	if err := proto.Unmarshal(serializedKey, key); err != nil {
		return nil, err
	}
	if err := km.validateKey(key); err != nil {
		return nil, err
	}
	ret, err := subtle.NewAESSIV(key.KeyValue)
	if err != nil {
		return nil, fmt.Errorf("aes_siv_key_manager: cannot create new primitive: %v", err)
	}
	return ret, nil
}

// NewKey generates a new AesSivKey. serializedKeyFormat is optional because
// there is only one valid key format.
func (km *aesSIVKeyManager) NewKey(serializedKeyFormat []byte) (proto.Message, error) {
	if serializedKeyFormat != nil {
		keyFormat := new(aspb.AesSivKeyFormat)
		if err := proto.Unmarshal(serializedKeyFormat, keyFormat); err != nil {
			return nil, errInvalidAESSIVKeyFormat
		}
		if keyFormat.KeySize != subtle.AESSIVKeySize {
			return nil, errInvalidAESSIVKeySize
		}
	}
	return &aspb.AesSivKey{
		Version:  aesSIVKeyVersion,
		KeyValue: random.GetRandomBytes(subtle.AESSIVKeySize),
	}, nil
}

// NewKeyData generates a new KeyData. serializedKeyFormat is optional because
// there is only one valid key format. This should be used solely by the key
// management API.
func (km *aesSIVKeyManager) NewKeyData(serializedKeyFormat []byte) (*tpb.KeyData, error) {
	key, err := km.NewKey(serializedKeyFormat)
	if err != nil {
		return nil, err
	}
	serializedKey, err := proto.Marshal(key)
	if err != nil {
		return nil, fmt.Errorf("aes_siv_key_manager: %v", err)
	}
	return &tpb.KeyData{
		TypeUrl:         aesSIVTypeURL,
		Value:           serializedKey,
		KeyMaterialType: tpb.KeyData_SYMMETRIC,
	}, nil
}

// DoesSupport checks whether this key manager supports the given key type.
func (km *aesSIVKeyManager) DoesSupport(typeURL string) bool {
	return typeURL == aesSIVTypeURL
}

// TypeURL returns the type URL of keys managed by this key manager.
func (km *aesSIVKeyManager) TypeURL() string {
	return aesSIVTypeURL
}

// validateKey validates the given AesSivKey.
func (km *aesSIVKeyManager) validateKey(key *aspb.AesSivKey) error {
	err := keyset.ValidateKeyVersion(key.Version, aesSIVKeyVersion)
	if err != nil {
		return fmt.Errorf("aes_siv_key_manager: %v", err)
	}
	keySize := uint32(len(key.KeyValue))
	if keySize != subtle.AESSIVKeySize {
		return errInvalidAESSIVKeySize
	}
	return nil
}
