// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ChainSafe/chaindb"
	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/scale"
)

var (
	genesisSetID      = uint64(0)
	grandpaPrefix     = "grandpa"
	authoritiesPrefix = []byte("auth")
	setIDChangePrefix = []byte("change")
	currentSetIDKey   = []byte("setID")
)

// GrandpaState tracks information related to grandpa
type GrandpaState struct {
	baseDB chaindb.Database
	db     chaindb.Database
}

// NewGrandpaStateFromGenesis returns a new GrandpaState given the grandpa genesis authorities
func NewGrandpaStateFromGenesis(db chaindb.Database, genesisAuthorities []*types.GrandpaVoter) (*GrandpaState, error) {
	grandpaDB := chaindb.NewTable(db, grandpaPrefix)
	s := &GrandpaState{
		baseDB: db,
		db:     grandpaDB,
	}

	err := s.setCurrentSetID(genesisSetID)
	if err != nil {
		return nil, err
	}

	err = s.setAuthorities(genesisSetID, genesisAuthorities)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// NewGrandpaState returns a new GrandpaState
func NewGrandpaState(db chaindb.Database) (*GrandpaState, error) {
	return &GrandpaState{
		baseDB: db,
		db:     chaindb.NewTable(db, grandpaPrefix),
	}, nil
}

func authoritiesKey(setID uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, setID)
	return append(authoritiesPrefix, buf...)
}

func setIDChangeKey(setID uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, setID)
	return append(setIDChangePrefix, buf...)
}

// setAuthorities sets the authorities for a given setID
func (s *GrandpaState) setAuthorities(setID uint64, authorities []*types.GrandpaVoter) error {
	enc, err := scale.Encode(authorities)
	if err != nil {
		return err
	}

	return s.db.Put(authoritiesKey(setID), enc)
}

// GetAuthorities returns the authorities for the given setID
func (s *GrandpaState) GetAuthorities(setID uint64) ([]*types.GrandpaVoter, error) {
	enc, err := s.db.Get(authoritiesKey(setID))
	if err != nil {
		return nil, err
	}

	r := &bytes.Buffer{}
	_, _ = r.Write(enc)
	v, err := types.DecodeGrandpaVoters(r)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// setCurrentSetID sets the current set ID
func (s *GrandpaState) setCurrentSetID(setID uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, setID)
	return s.db.Put(currentSetIDKey, buf[:])
}

// GetCurrentSetID retrieves the current set ID
func (s *GrandpaState) GetCurrentSetID() (uint64, error) {
	id, err := s.db.Get(currentSetIDKey)
	if err != nil {
		return 0, err
	}

	if len(id) < 8 {
		return 0, errors.New("invalid setID")
	}

	return binary.LittleEndian.Uint64(id), nil
}

// SetNextChange sets the next authority change
func (s *GrandpaState) SetNextChange(authorities []*types.GrandpaVoter, number *big.Int) error {
	currSetID, err := s.GetCurrentSetID()
	if err != nil {
		return err
	}

	nextSetID := currSetID + 1
	err = s.setAuthorities(nextSetID, authorities)
	if err != nil {
		return err
	}

	err = s.setSetIDChangeAtBlock(nextSetID, number)
	if err != nil {
		return err
	}

	return nil
}

// IncrementSetID increments the set ID
func (s *GrandpaState) IncrementSetID() error {
	currSetID, err := s.GetCurrentSetID()
	if err != nil {
		return err
	}

	nextSetID := currSetID + 1
	return s.setCurrentSetID(nextSetID)
}

// setSetIDChangeAtBlock sets a set ID change at a certain block
func (s *GrandpaState) setSetIDChangeAtBlock(setID uint64, number *big.Int) error {
	return s.db.Put(setIDChangeKey(setID), number.Bytes())
}

// GetSetIDChange returs the block number where the set ID was updated
func (s *GrandpaState) GetSetIDChange(setID uint64) (*big.Int, error) {
	num, err := s.db.Get(setIDChangeKey(setID))
	if err != nil {
		return nil, err
	}

	return big.NewInt(0).SetBytes(num), nil
}
