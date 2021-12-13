// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package network

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/utils"

	"github.com/stretchr/testify/require"
)

func TestDecodeTransactionHandshake(t *testing.T) {
	t.Parallel()

	testHandshake := &transactionHandshake{}

	enc, err := testHandshake.Encode()
	require.NoError(t, err)

	msg, err := decodeTransactionHandshake(enc)
	require.NoError(t, err)
	require.Equal(t, testHandshake, msg)
}

func TestHandleTransactionMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	expectedMsgArg := &TransactionMessage{
		Extrinsics: []types.Extrinsic{{1, 1}, {2, 2}},
	}

	th := NewMockTransactionHandler(ctrl)
	th.EXPECT().
		HandleTransactionMessage(gomock.Any(), expectedMsgArg).
		Return(true, nil).AnyTimes()
	th.EXPECT().TransactionsCount().Return(0).AnyTimes()

	basePath := utils.NewTestBasePath(t, "nodeA")

	config := &Config{
		BasePath:           basePath,
		Port:               availablePort(t),
		NoBootstrap:        true,
		NoMDNS:             true,
		TransactionHandler: th,
	}

	s := createTestService(t, config)
	s.handleTransactionMessage(peer.ID(""), expectedMsgArg)
}
