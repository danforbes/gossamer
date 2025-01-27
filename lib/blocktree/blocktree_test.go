// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package blocktree

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/stretchr/testify/require"
)

var zeroHash, _ = common.HexToHash("0x00")
var testHeader = &types.Header{
	ParentHash: zeroHash,
	Number:     big.NewInt(0),
	Digest:     types.NewDigest(),
}

type testBranch struct {
	hash        Hash
	number      *big.Int
	arrivalTime int64
}

func newBlockTreeFromNode(root *node) *BlockTree {
	return &BlockTree{
		root:   root,
		leaves: newLeafMap(root),
	}
}

func createTestBlockTree(t *testing.T, header *types.Header, number int) (*BlockTree, []testBranch) {
	bt := NewBlockTreeFromRoot(header)
	previousHash := header.Hash()

	// branch tree randomly
	branches := []testBranch{}
	r := *rand.New(rand.NewSource(rand.Int63()))

	at := int64(0)

	// create base tree
	for i := 1; i <= number; i++ {
		header := &types.Header{
			ParentHash: previousHash,
			Number:     big.NewInt(int64(i)),
			Digest:     types.NewDigest(),
		}

		hash := header.Hash()
		err := bt.AddBlock(header, time.Unix(0, at))
		require.NoError(t, err)

		previousHash = hash

		isBranch := r.Intn(2)
		if isBranch == 1 {
			branches = append(branches, testBranch{
				hash:        hash,
				number:      bt.getNode(hash).number,
				arrivalTime: at,
			})
		}

		at += int64(r.Intn(8))
	}

	// create tree branches
	for _, branch := range branches {
		at := branch.arrivalTime
		previousHash = branch.hash

		for i := int(branch.number.Uint64()); i <= number; i++ {
			header := &types.Header{
				ParentHash: previousHash,
				Number:     big.NewInt(int64(i) + 1),
				StateRoot:  common.Hash{0x1},
				Digest:     types.NewDigest(),
			}

			hash := header.Hash()
			err := bt.AddBlock(header, time.Unix(0, at))
			require.NoError(t, err)

			previousHash = hash
			at += int64(r.Intn(8))

		}
	}

	return bt, branches
}

func createFlatTree(t *testing.T, number int) (*BlockTree, []common.Hash) {
	bt := NewBlockTreeFromRoot(testHeader)
	require.NotNil(t, bt)

	previousHash := bt.root.hash

	hashes := []common.Hash{bt.root.hash}
	for i := 1; i <= number; i++ {
		header := &types.Header{
			ParentHash: previousHash,
			Number:     big.NewInt(int64(i)),
			Digest:     types.NewDigest(),
		}

		hash := header.Hash()
		hashes = append(hashes, hash)

		err := bt.AddBlock(header, time.Unix(0, 0))
		require.Nil(t, err)
		previousHash = hash
	}

	return bt, hashes
}

func TestNewBlockTreeFromNode(t *testing.T) {
	var bt *BlockTree
	var branches []testBranch

	for {
		bt, branches = createTestBlockTree(t, testHeader, 5)
		if len(branches) > 0 && len(bt.getNode(branches[0].hash).children) > 0 {
			break
		}
	}

	testNode := bt.getNode(branches[0].hash).children[0]
	leaves := testNode.getLeaves(nil)

	newBt := newBlockTreeFromNode(testNode)
	require.ElementsMatch(t, leaves, newBt.leaves.nodes())
}

func TestBlockTree_GetBlock(t *testing.T) {
	bt, hashes := createFlatTree(t, 2)

	n := bt.getNode(hashes[2])
	if n == nil {
		t.Fatal("node is nil")
	}

	if !bytes.Equal(hashes[2][:], n.hash[:]) {
		t.Fatalf("Fail: got %x expected %x", n.hash, hashes[2])
	}

}

func TestBlockTree_AddBlock(t *testing.T) {
	bt, hashes := createFlatTree(t, 1)

	header := &types.Header{
		ParentHash: hashes[1],
		Number:     big.NewInt(2),
	}

	hash := header.Hash()
	err := bt.AddBlock(header, time.Unix(0, 0))
	require.Nil(t, err)

	node := bt.getNode(hash)

	if n, err := bt.leaves.load(node.hash); n == nil || err != nil {
		t.Errorf("expected %x to be a leaf", n.hash)
	}

	oldHash := common.Hash{0x01}

	if n, err := bt.leaves.load(oldHash); n != nil || err == nil {
		t.Errorf("expected %x to no longer be a leaf", oldHash)
	}
}

func TestNode_isDecendantOf(t *testing.T) {
	// Create tree with number 4 (with 4 nodes)
	bt, hashes := createFlatTree(t, 4)

	// Check leaf is descendant of root
	leaf := bt.getNode(hashes[3])
	if !leaf.isDescendantOf(bt.root) {
		t.Error("failed to verify leaf is descendant of root")
	}

	// Verify the inverse relationship does not hold
	if bt.root.isDescendantOf(leaf) {
		t.Error("root should not be descendant of anything")
	}
}

func TestBlockTree_LongestPath(t *testing.T) {
	bt, hashes := createFlatTree(t, 3)

	// Insert a block to create a competing path
	header := &types.Header{
		ParentHash: hashes[0],
		Number:     big.NewInt(1),
	}

	header.Hash()
	err := bt.AddBlock(header, time.Unix(0, 0))
	require.NotNil(t, err)

	longestPath := bt.longestPath()

	for i, n := range longestPath {
		if n.hash != hashes[i] {
			t.Errorf("expected Hash: 0x%X got: 0x%X\n", hashes[i], n.hash)
		}
	}
}

func TestBlockTree_Subchain(t *testing.T) {
	bt, hashes := createFlatTree(t, 4)
	expectedPath := hashes[1:]

	// Insert a block to create a competing path
	extraBlock := &types.Header{
		ParentHash: hashes[0],
		Number:     big.NewInt(1),
		Digest:     types.NewDigest(),
	}

	extraBlock.Hash()
	err := bt.AddBlock(extraBlock, time.Unix(0, 0))
	require.NotNil(t, err)

	subChain, err := bt.subChain(hashes[1], hashes[3])
	if err != nil {
		t.Fatal(err)
	}

	for i, n := range subChain {
		if n.hash != expectedPath[i] {
			t.Errorf("expected Hash: 0x%X got: 0x%X\n", expectedPath[i], n.hash)
		}
	}
}

func TestBlockTree_DeepestLeaf(t *testing.T) {
	arrivalTime := int64(256)
	var expected Hash

	bt, _ := createTestBlockTree(t, testHeader, 8)

	deepest := big.NewInt(0)

	for leaf, node := range bt.leaves.toMap() {
		node.arrivalTime = time.Unix(arrivalTime, 0)
		arrivalTime--
		if node.number.Cmp(deepest) >= 0 {
			deepest = node.number
			expected = leaf
		}

		t.Logf("leaf=%s number=%d arrivalTime=%s", leaf, node.number, node.arrivalTime)
	}

	deepestLeaf := bt.deepestLeaf()
	if deepestLeaf.hash != expected {
		t.Fatalf("Fail: got %s expected %s", deepestLeaf.hash, expected)
	}
}

func TestBlockTree_GetNode(t *testing.T) {
	bt, branches := createTestBlockTree(t, testHeader, 16)

	for _, branch := range branches {
		header := &types.Header{
			ParentHash: branch.hash,
			Number:     big.NewInt(0).Add(branch.number, big.NewInt(1)),
			StateRoot:  Hash{0x2},
		}

		err := bt.AddBlock(header, time.Unix(0, 0))
		require.NoError(t, err)
	}

	block := bt.getNode(branches[0].hash)
	require.NotNil(t, block)
}

func TestBlockTree_GetAllBlocksAtNumber(t *testing.T) {
	bt, _ := createTestBlockTree(t, testHeader, 8)
	hashes := bt.root.getNodesWithNumber(big.NewInt(10), []common.Hash{})

	expected := []common.Hash{}
	require.Equal(t, expected, hashes)

	// create one-path tree
	btNumber := 8
	desiredNumber := 6
	bt, btHashes := createFlatTree(t, btNumber)

	expected = []common.Hash{btHashes[desiredNumber]}

	// add branch
	previousHash := btHashes[4]

	for i := 4; i <= btNumber; i++ {
		digest := types.NewDigest()
		err := digest.Add(types.ConsensusDigest{
			ConsensusEngineID: types.BabeEngineID,
			Data:              common.MustHexToBytes("0x0118ca239392960473fe1bc65f94ee27d890a49c1b200c006ff5dcc525330ecc16770100000000000000b46f01874ce7abbb5220e8fd89bede0adad14c73039d91e28e881823433e723f0100000000000000d684d9176d6eb69887540c9a89fa6097adea82fc4b0ff26d1062b488f352e179010000000000000068195a71bdde49117a616424bdc60a1733e96acb1da5aeab5d268cf2a572e94101000000000000001a0575ef4ae24bdfd31f4cb5bd61239ae67c12d4e64ae51ac756044aa6ad8200010000000000000018168f2aad0081a25728961ee00627cfe35e39833c805016632bf7c14da5800901000000000000000000000000000000000000000000000000000000000000000000000000000000"), //nolint:lll
		})
		require.NoError(t, err)
		header := &types.Header{
			ParentHash: previousHash,
			Number:     big.NewInt(int64(i) + 1),
			Digest:     digest,
		}

		hash := header.Hash()
		err = bt.AddBlock(header, time.Unix(0, 0))
		require.NoError(t, err)
		previousHash = hash

		if i == desiredNumber-1 {
			expected = append(expected, hash)
		}
	}

	// add another branch
	previousHash = btHashes[2]

	for i := 2; i <= btNumber; i++ {
		digest := types.NewDigest()
		err := digest.Add(types.SealDigest{
			ConsensusEngineID: types.BabeEngineID,
			Data:              common.MustHexToBytes("0x4625284883e564bc1e4063f5ea2b49846cdddaa3761d04f543b698c1c3ee935c40d25b869247c36c6b8a8cbbd7bb2768f560ab7c276df3c62df357a7e3b1ec8d"), //nolint:lll
		})
		require.NoError(t, err)
		header := &types.Header{
			ParentHash: previousHash,
			Number:     big.NewInt(int64(i) + 1),
			Digest:     digest,
		}

		hash := header.Hash()
		err = bt.AddBlock(header, time.Unix(0, 0))
		require.NoError(t, err)
		previousHash = hash

		if i == desiredNumber-1 {
			expected = append(expected, hash)
		}
	}

	hashes = bt.root.getNodesWithNumber(big.NewInt(int64(desiredNumber)), []common.Hash{})
	if !reflect.DeepEqual(hashes, expected) {
		t.Fatalf("Fail: did not get all expected hashes got %v expected %v", hashes, expected)
	}
}

func TestBlockTree_IsDecendantOf(t *testing.T) {
	// Create tree with number 4 (with 4 nodes)
	bt, hashes := createFlatTree(t, 4)

	isDescendant, err := bt.IsDescendantOf(bt.root.hash, hashes[3])
	require.NoError(t, err)
	require.True(t, isDescendant)

	isDescendant, err = bt.IsDescendantOf(hashes[3], bt.root.hash)
	require.NoError(t, err)
	require.False(t, isDescendant)
}

func TestBlockTree_HighestCommonAncestor(t *testing.T) {
	var bt *BlockTree
	var leaves []common.Hash
	var branches []testBranch

	for {
		bt, branches = createTestBlockTree(t, testHeader, 8)
		leaves = bt.Leaves()
		if len(leaves) == 2 {
			break
		}
	}

	expected := branches[0].hash

	a := leaves[0]
	b := leaves[1]

	p, err := bt.HighestCommonAncestor(a, b)
	require.NoError(t, err)
	require.Equal(t, expected, p)
}

func TestBlockTree_HighestCommonAncestor_SameNode(t *testing.T) {
	bt, _ := createTestBlockTree(t, testHeader, 8)
	leaves := bt.Leaves()

	a := leaves[0]

	p, err := bt.HighestCommonAncestor(a, a)
	require.NoError(t, err)
	require.Equal(t, a, p)
}

func TestBlockTree_HighestCommonAncestor_SameChain(t *testing.T) {
	bt, _ := createTestBlockTree(t, testHeader, 8)
	leaves := bt.Leaves()

	a := leaves[0]
	b := bt.getNode(a).parent.hash

	// b is a's parent, so their highest common Ancestor is b.
	p, err := bt.HighestCommonAncestor(a, b)
	require.NoError(t, err)
	require.Equal(t, b, p)
}

func TestBlockTree_Prune(t *testing.T) {
	var bt *BlockTree
	var branches []testBranch

	for {
		bt, branches = createTestBlockTree(t, testHeader, 5)
		if len(branches) > 0 && len(bt.getNode(branches[0].hash).children) > 1 {
			break
		}
	}

	copy := bt.DeepCopy()

	// pick some block to finalise
	finalised := bt.root.children[0].children[0].children[0]
	pruned := bt.Prune(finalised.hash)

	for _, prunedHash := range pruned {
		prunedNode := copy.getNode(prunedHash)
		if prunedNode.isDescendantOf(finalised) {
			t.Fatal("pruned node that's descendant of finalised node!!")
		}

		if finalised.isDescendantOf(prunedNode) {
			t.Fatal("pruned an ancestor of the finalised node!!")
		}
	}

	require.NotEqual(t, 0, len(bt.leaves.nodes()))
	for _, leaf := range bt.leaves.nodes() {
		require.NotEqual(t, leaf.hash, finalised.hash)
		require.True(t, leaf.isDescendantOf(finalised))
	}
}

func TestBlockTree_GetHashByNumber(t *testing.T) {
	bt, _ := createTestBlockTree(t, testHeader, 8)
	best := bt.DeepestBlockHash()
	bn := bt.getNode(best)

	for i := int64(0); i < bn.number.Int64(); i++ {
		hash, err := bt.GetHashByNumber(big.NewInt(i))
		require.NoError(t, err)
		require.Equal(t, big.NewInt(i), bt.getNode(hash).number)
		desc, err := bt.IsDescendantOf(hash, best)
		require.NoError(t, err)
		require.True(t, desc, fmt.Sprintf("index %d failed, got hash=%s", i, hash))
	}

	_, err := bt.GetHashByNumber(big.NewInt(-1))
	require.Error(t, err)

	_, err = bt.GetHashByNumber(big.NewInt(0).Add(bn.number, big.NewInt(1)))
	require.Error(t, err)
}

func TestBlockTree_AllLeavesHasSameNumberAndArrivalTime_DeepestBlockHash_ShouldHasConsistentOutput(t *testing.T) {
	bt := NewBlockTreeFromRoot(testHeader)
	previousHash := testHeader.Hash()

	branches := []testBranch{}

	const fixedArrivalTime = 99
	const deep = 8

	// create a base tree with a fixed amount of blocks
	// and all block with the same arrival time

	/**
	base tree and nodes representation, all with the same arrival time and all
	the leaves has the same number (8) the numbers in the right represents the order
	the nodes are inserted into the blocktree.

	a -> b -> c -> d -> e -> f -> g -> h (1)
		|    |    |    |    |    |> h (7)
		|    |    |    |    |> g -> h (6)
		|    |    |    |> f -> g -> h (5)
		|    |    |> e -> f -> g -> h (4)
		|    |> d -> e -> f -> g -> h (3)
		|> c -> d -> e -> f -> g -> h (2)
	**/

	for i := 1; i <= deep; i++ {
		header := &types.Header{
			ParentHash: previousHash,
			Number:     big.NewInt(int64(i)),
			Digest:     types.NewDigest(),
		}

		hash := header.Hash()

		err := bt.AddBlock(header, time.Unix(0, fixedArrivalTime))
		require.NoError(t, err)

		previousHash = hash

		// the last block on the base tree should not generates a branch
		if i < deep {
			branches = append(branches, testBranch{
				hash:   hash,
				number: bt.getNode(hash).number,
			})
		}
	}

	// create all the branch nodes with the same arrival time
	for _, branch := range branches {
		previousHash = branch.hash

		for i := int(branch.number.Uint64()); i < deep; i++ {
			header := &types.Header{
				ParentHash: previousHash,
				Number:     big.NewInt(int64(i) + 1),
				StateRoot:  common.Hash{0x1},
				Digest:     types.NewDigest(),
			}

			hash := header.Hash()
			err := bt.AddBlock(header, time.Unix(0, fixedArrivalTime))
			require.NoError(t, err)

			previousHash = hash
		}
	}

	// check all leaves has the same number and timestamps
	leaves := bt.leaves.nodes()
	for idx := 0; idx < len(leaves)-2; idx++ {
		curr := leaves[idx]
		next := leaves[idx+1]

		require.Equal(t, curr.number, next.number)
		require.Equal(t, curr.arrivalTime, next.arrivalTime)
	}

	require.Len(t, leaves, deep)

	// expects currentDeepestLeaf nil till call deepestLeaf() function
	require.Nil(t, bt.leaves.currentDeepestLeaf)
	deepestLeaf := bt.deepestLeaf()

	require.Equal(t, deepestLeaf, bt.leaves.currentDeepestLeaf)
	require.Contains(t, leaves, deepestLeaf)

	// adding a new node with a greater number, should update the currentDeepestLeaf
	header := &types.Header{
		ParentHash: previousHash,
		Number:     big.NewInt(int64(deepestLeaf.number.Uint64() + 1)),
		StateRoot:  common.Hash{0x1},
		Digest:     types.NewDigest(),
	}

	hash := header.Hash()
	err := bt.AddBlock(header, time.Unix(0, fixedArrivalTime))
	require.NoError(t, err)

	deepestLeaf = bt.deepestLeaf()
	require.Equal(t, hash, deepestLeaf.hash)
	require.Equal(t, hash, bt.leaves.currentDeepestLeaf.hash)
}

func TestBlockTree_DeepCopy(t *testing.T) {
	bt, _ := createFlatTree(t, 8)

	btCopy := bt.DeepCopy()
	require.True(t, equalNodeValue(bt.root, btCopy.root), "BlockTree heads not equal")
	require.True(t, equalLeaves(bt.leaves, btCopy.leaves), "BlockTree leaves not equal")

	btCopy.root = &node{}
	require.NotEqual(t, bt.root, btCopy.root)
}

func equalNodeValue(nd *node, ndCopy *node) bool {
	if nd.hash != ndCopy.hash {
		return false
	}
	if nd.number.Cmp(ndCopy.number) != 0 {
		return false
	}
	if nd.arrivalTime != ndCopy.arrivalTime {
		return false
	}
	for i, child := range nd.children {
		return equalNodeValue(child, ndCopy.children[i])
	}
	if nd.parent.hash != ndCopy.parent.hash {
		return false
	}
	if nd.parent.arrivalTime != ndCopy.parent.arrivalTime {
		return false
	}
	if nd.parent.number.Cmp(ndCopy.parent.number) != 0 {
		return false
	}
	return true
}

func equalLeaves(lm *leafMap, lmCopy *leafMap) bool {
	lmm := lm.toMap()
	lmCopyM := lmCopy.toMap()
	for key, val := range lmm {
		lmCopyVal := lmCopyM[key]
		return equalNodeValue(val, lmCopyVal)
	}
	return true
}
