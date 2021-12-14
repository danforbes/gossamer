// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

//go:build integration
// +build integration

package dot

import (
	"math/big"
	"os"
	"testing"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/utils"
	"github.com/ChainSafe/gossamer/pkg/scale"

	"github.com/stretchr/testify/require"
)

func TestNewTrieFromPairs(t *testing.T) {
	fp := setupStateFile(t)
	trie, err := newTrieFromPairs(fp)
	require.NoError(t, err)

	expectedRoot := common.MustHexToHash("0x09f9ca28df0560c2291aa16b56e15e07d1e1927088f51356d522722aa90ca7cb")
	require.Equal(t, expectedRoot, trie.MustHash())
}

func TestNewHeaderFromFile(t *testing.T) {
	fp := setupHeaderFile(t)
	header, err := newHeaderFromFile(fp)
	require.NoError(t, err)

	digestBytes := common.MustHexToBytes(
		"0x080642414245b501013c0000009659bd0f0000000070edad1c9064fff78cb18435223d8adaf5ea04c24b1a8766e3dc01eb03cc6" +
			"a0c11b79793d4e31cc0990838229c44fed1669a7c7c79e1e6d0a96374d6496728069d1ef739e290497a0e3b728fa88fcbdd3a550" +
			"4e0efde0242e7a806dd4fa9260c054241424501019e7f28dddcf27c1e6b328d5694c368d5b2ec5dbe0e412ae1c98f88d53be4d85" +
			"02fac571f3f19c9caaf281a673319241e0c5095a683ad34316204088a36a4bd86")
	digest := types.NewDigest()
	err = scale.Unmarshal(digestBytes, &digest)
	require.NoError(t, err)
	require.Equal(t, 2, len(digest.Types))

	expected := &types.Header{
		ParentHash:     common.MustHexToHash("0x3b45c9c22dcece75a30acc9c2968cb311e6b0557350f83b430f47559db786975"),
		Number:         big.NewInt(1482002),
		StateRoot:      common.MustHexToHash("0x09f9ca28df0560c2291aa16b56e15e07d1e1927088f51356d522722aa90ca7cb"),
		ExtrinsicsRoot: common.MustHexToHash("0xda26dc8c1455f8f81cae12e4fc59e23ce961b2c837f6d3f664283af906d344e0"),
		Digest:         digest,
	}

	require.Equal(t, expected, header)
}

func TestImportState_Integration(t *testing.T) {
	basepath := os.TempDir()

	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := NewTestGenesisRawFile(t, cfg)
	require.NotNil(t, genFile)

	defer utils.RemoveTestDir(t)

	cfg.Init.Genesis = genFile.Name()

	cfg.Global.BasePath = basepath
	err := InitNode(cfg)
	require.NoError(t, err)

	stateFP := setupStateFile(t)
	headerFP := setupHeaderFile(t)

	const firstSlot = uint64(262493679)
	err = ImportState(basepath, stateFP, headerFP, firstSlot)
	require.NoError(t, err)
}