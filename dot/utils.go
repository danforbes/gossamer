// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package dot

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ctoml "github.com/ChainSafe/gossamer/dot/config/toml"
	"github.com/ChainSafe/gossamer/internal/log"
	"github.com/ChainSafe/gossamer/lib/genesis"
	"github.com/ChainSafe/gossamer/lib/runtime"
	"github.com/ChainSafe/gossamer/lib/runtime/wasmer"
	"github.com/ChainSafe/gossamer/lib/utils"
	"github.com/cosmos/go-bip39"
	"github.com/naoina/toml"
	"github.com/stretchr/testify/require"
)

// NewTestGenesis returns a test genesis instance using "gssmr" raw data
func NewTestGenesis(t *testing.T) *genesis.Genesis {
	fp := utils.GetGssmrGenesisRawPath()

	gssmrGen, err := genesis.NewGenesisFromJSONRaw(fp)
	require.NoError(t, err)

	return &genesis.Genesis{
		Name:       "test",
		ID:         "test",
		Bootnodes:  []string(nil),
		ProtocolID: "/gossamer/test/0",
		Genesis:    gssmrGen.GenesisFields(),
	}
}

// NewTestGenesisRawFile returns a test genesis file using "gssmr" raw data
func NewTestGenesisRawFile(t *testing.T, cfg *Config) *os.File {
	dir := utils.NewTestDir(t)

	file, err := os.CreateTemp(dir, "genesis-")
	require.Nil(t, err)

	fp := utils.GetGssmrGenesisRawPath()

	gssmrGen, err := genesis.NewGenesisFromJSONRaw(fp)
	require.Nil(t, err)

	gen := &genesis.Genesis{
		Name:       cfg.Global.Name,
		ID:         cfg.Global.ID,
		Bootnodes:  cfg.Network.Bootnodes,
		ProtocolID: cfg.Network.ProtocolID,
		Genesis:    gssmrGen.GenesisFields(),
	}

	b, err := json.Marshal(gen)
	require.Nil(t, err)

	_, err = file.Write(b)
	require.Nil(t, err)

	return file
}

// NewTestGenesisFile returns a human-readable test genesis file using "gssmr" human readable data
func NewTestGenesisFile(t *testing.T, cfg *Config) *os.File {
	dir := utils.NewTestDir(t)

	file, err := os.CreateTemp(dir, "genesis-")
	require.Nil(t, err)

	fp := utils.GetGssmrGenesisPath()

	gssmrGen, err := genesis.NewGenesisFromJSON(fp, 0)
	require.Nil(t, err)

	gen := &genesis.Genesis{
		Name:       cfg.Global.Name,
		ID:         cfg.Global.ID,
		Bootnodes:  cfg.Network.Bootnodes,
		ProtocolID: cfg.Network.ProtocolID,
		Genesis:    gssmrGen.GenesisFields(),
	}

	b, err := json.Marshal(gen)
	require.Nil(t, err)

	_, err = file.Write(b)
	require.Nil(t, err)

	return file
}

// NewTestGenesisAndRuntime create a new test runtime and a new test genesis
// file with the test runtime stored in raw data and returns the genesis file
func NewTestGenesisAndRuntime(t *testing.T) string {
	dir := utils.NewTestDir(t)

	_ = wasmer.NewTestInstance(t, runtime.NODE_RUNTIME)
	runtimeFilePath := runtime.GetAbsolutePath(runtime.NODE_RUNTIME_FP)

	runtimeData, err := os.ReadFile(filepath.Clean(runtimeFilePath))
	require.Nil(t, err)

	gen := NewTestGenesis(t)
	hex := hex.EncodeToString(runtimeData)

	gen.Genesis.Raw = map[string]map[string]string{}
	if gen.Genesis.Raw["top"] == nil {
		gen.Genesis.Raw["top"] = make(map[string]string)
	}
	gen.Genesis.Raw["top"]["0x3a636f6465"] = "0x" + hex
	gen.Genesis.Raw["top"]["0xcf722c0832b5231d35e29f319ff27389f5032bfc7bfc3ba5ed7839f2042fb99f"] = "0x0000000000000001"

	genFile, err := os.CreateTemp(dir, "genesis-")
	require.Nil(t, err)

	genData, err := json.Marshal(gen)
	require.Nil(t, err)

	_, err = genFile.Write(genData)
	require.Nil(t, err)

	return genFile.Name()
}

// NewTestConfig returns a new test configuration using the provided basepath
func NewTestConfig(t *testing.T) *Config {
	dir := utils.NewTestDir(t)

	cfg := &Config{
		Global: GlobalConfig{
			Name:     GssmrConfig().Global.Name,
			ID:       GssmrConfig().Global.ID,
			BasePath: dir,
			LogLvl:   log.Info,
		},
		Log:     GssmrConfig().Log,
		Init:    GssmrConfig().Init,
		Account: GssmrConfig().Account,
		Core:    GssmrConfig().Core,
		Network: GssmrConfig().Network,
		RPC:     GssmrConfig().RPC,
	}

	return cfg
}

// NewTestConfigWithFile returns a new test configuration and a temporary configuration file
func NewTestConfigWithFile(t *testing.T) (*Config, *os.File) {
	cfg := NewTestConfig(t)

	file, err := os.CreateTemp(cfg.Global.BasePath, "config-")
	require.NoError(t, err)

	cfgFile := ExportConfig(cfg, file.Name())
	return cfg, cfgFile
}

// ExportConfig exports a dot configuration to a toml configuration file
func ExportConfig(cfg *Config, fp string) *os.File {
	raw, err := toml.Marshal(*cfg)
	if err != nil {
		logger.Errorf("failed to marshal configuration: %s", err)
		os.Exit(1)
	}
	return WriteConfig(raw, fp)
}

// ExportTomlConfig exports a dot configuration to a toml configuration file
func ExportTomlConfig(cfg *ctoml.Config, fp string) *os.File {
	raw, err := toml.Marshal(*cfg)
	if err != nil {
		logger.Errorf("failed to marshal configuration: %s", err)
		os.Exit(1)
	}
	return WriteConfig(raw, fp)
}

// WriteConfig writes the config `data` in the file 'fp'.
func WriteConfig(data []byte, fp string) *os.File {
	newFile, err := os.Create(filepath.Clean(fp))
	if err != nil {
		logger.Errorf("failed to create configuration file: %s", err)
		os.Exit(1)
	}

	_, err = newFile.Write(data)
	if err != nil {
		logger.Errorf("failed to write to configuration file: %s", err)
		os.Exit(1)
	}

	if err := newFile.Close(); err != nil {
		logger.Errorf("failed to close configuration file: %s", err)
		os.Exit(1)
	}

	return newFile
}

// CreateJSONRawFile will generate a JSON genesis file with raw storage
func CreateJSONRawFile(bs *BuildSpec, fp string) *os.File {
	data, err := bs.ToJSONRaw()
	if err != nil {
		logger.Errorf("failed to convert into raw json: %s", err)
		os.Exit(1)
	}
	return WriteConfig(data, fp)
}

// RandomNodeName generates a new random name if there is no name configured for the node
func RandomNodeName() string {
	entropy, _ := bip39.NewEntropy(128)
	randomNamesString, _ := bip39.NewMnemonic(entropy)
	randomNames := strings.Split(randomNamesString, " ")
	number := binary.BigEndian.Uint16(entropy)
	return randomNames[0] + "-" + randomNames[1] + "-" + fmt.Sprint(number)
}
