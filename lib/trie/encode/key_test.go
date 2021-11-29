// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExtraPartialKeyLength(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		publicKeyLength int
		encoding        []byte
		err             error
	}{
		"length equal to maximum": {
			publicKeyLength: int(maxPartialKeySize) + 63,
			err:             ErrPartialKeyTooBig,
		},
		"zero length": {
			encoding: []byte{0xc1},
		},
		"one length": {
			publicKeyLength: 1,
			encoding:        []byte{0xc2},
		},
		"length at maximum allowed": {
			publicKeyLength: int(maxPartialKeySize) + 62,
			encoding: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoding, err := ExtraPartialKeyLength(testCase.publicKeyLength)

			assert.ErrorIs(t, err, testCase.err)
			assert.Equal(t, testCase.encoding, encoding)
		})
	}
}

func Test_NibblesToKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		nibbles []byte
		key     []byte
	}{
		"nil nibbles": {
			key: []byte{},
		},
		"empty nibbles": {
			nibbles: []byte{},
			key:     []byte{},
		},
		"0xF 0xF": {
			nibbles: []byte{0xF, 0xF},
			key:     []byte{0xFF},
		},
		"0x3 0xa 0x0 0x5": {
			nibbles: []byte{0x3, 0xa, 0x0, 0x5},
			key:     []byte{0xa3, 0x50},
		},
		"0xa 0xa 0xf 0xf 0x0 0x1": {
			nibbles: []byte{0xa, 0xa, 0xf, 0xf, 0x0, 0x1},
			key:     []byte{0xaa, 0xff, 0x10},
		},
		"0xa 0xa 0xf 0xf 0x0 0x1 0xc 0x2": {
			nibbles: []byte{0xa, 0xa, 0xf, 0xf, 0x0, 0x1, 0xc, 0x2},
			key:     []byte{0xaa, 0xff, 0x10, 0x2c},
		},
		"0xa 0xa 0xf 0xf 0x0 0x1 0xc": {
			nibbles: []byte{0xa, 0xa, 0xf, 0xf, 0x0, 0x1, 0xc},
			key:     []byte{0xaa, 0xff, 0x10, 0x0c},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := NibblesToKey(testCase.nibbles)

			assert.Equal(t, testCase.key, key)
		})
	}
}

func Test_NibblesToKeyLE(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		nibbles []byte
		keyLE   []byte
	}{
		"nil nibbles": {
			keyLE: []byte{},
		},
		"empty nibbles": {
			nibbles: []byte{},
			keyLE:   []byte{},
		},
		"0xF 0xF": {
			nibbles: []byte{0xF, 0xF},
			keyLE:   []byte{0xFF},
		},
		"0x3 0xa 0x0 0x5": {
			nibbles: []byte{0x3, 0xa, 0x0, 0x5},
			keyLE:   []byte{0x3a, 0x05},
		},
		"0xa 0xa 0xf 0xf 0x0 0x1": {
			nibbles: []byte{0xa, 0xa, 0xf, 0xf, 0x0, 0x1},
			keyLE:   []byte{0xaa, 0xff, 0x01},
		},
		"0xa 0xa 0xf 0xf 0x0 0x1 0xc 0x2": {
			nibbles: []byte{0xa, 0xa, 0xf, 0xf, 0x0, 0x1, 0xc, 0x2},
			keyLE:   []byte{0xaa, 0xff, 0x01, 0xc2},
		},
		"0xa 0xa 0xf 0xf 0x0 0x1 0xc": {
			nibbles: []byte{0xa, 0xa, 0xf, 0xf, 0x0, 0x1, 0xc},
			keyLE:   []byte{0xa, 0xaf, 0xf0, 0x1c},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			keyLE := NibblesToKeyLE(testCase.nibbles)

			assert.Equal(t, testCase.keyLE, keyLE)
		})
	}
}
