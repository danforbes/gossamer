// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package pprof

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	settings := Settings{}
	logger := NewMockLogger(ctrl)

	service := NewService(settings, logger)

	expectedSettings := Settings{
		ListeningAddress: "localhost:6060",
	}
	assert.Equal(t, expectedSettings, service.settings)
	assert.NotNil(t, service.server)
	assert.NotNil(t, service.done)
}

//go:generate mockgen -destination=runner_mock_test.go -package $GOPACKAGE github.com/ChainSafe/gossamer/internal/httpserver Runner

func Test_Service_StartStop_success(t *testing.T) {
	t.Parallel()

	errDummy := errors.New("dummy")

	testCases := map[string]struct {
		startDone    bool
		startDoneErr error
		startErr     error
		stopDoneErr  error
		stopErr      error
	}{
		"start nil error": {
			startDone: true,
			startErr:  ErrServerDoneBeforeReady,
		},
		"start error": {
			startDone:    true,
			startDoneErr: errDummy,
			startErr:     errDummy,
		},
		"stop error": {
			stopDoneErr: errDummy,
			stopErr:     errDummy,
		},
		"success": {},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			server := NewMockRunner(ctrl)
			ctxType, cancelType := context.WithCancel(context.Background())
			defer cancelType()
			server.EXPECT().Run(
				gomock.AssignableToTypeOf(ctxType),
				gomock.AssignableToTypeOf(make(chan<- struct{})),
				gomock.AssignableToTypeOf(make(chan<- error)),
			).Do(func(ctx context.Context, ready chan<- struct{}, done chan<- error) {
				if testCase.startDone {
					done <- testCase.startDoneErr
					return // start failure
				}
				close(ready)
				<-ctx.Done()
				done <- testCase.startDoneErr
			})

			service := &Service{
				server: server,
				done:   make(chan error),
			}

			err := service.Start()

			if testCase.startErr != nil {
				require.EqualError(t, err, testCase.startErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if testCase.startDone {
				return // start failed, we won't stop
			}

			err = service.Stop()
			if testCase.startErr != nil {
				require.EqualError(t, err, testCase.stopErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}