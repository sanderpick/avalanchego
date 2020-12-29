// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package conflicts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
)

func TestProcessing(t *testing.T) {
	c := New()

	tx := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: ids.GenerateTestID(),
		},
	}

	processing := c.Processing(tx.Transition().ID())
	assert.False(t, processing)

	err := c.Add(tx)
	assert.NoError(t, err)

	processing = c.Processing(tx.Transition().ID())
	assert.True(t, processing)
}

func TestNoConflicts(t *testing.T) {
	c := New()

	tx := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: ids.GenerateTestID(),
		},
	}

	virtuous, err := c.IsVirtuous(tx)
	assert.NoError(t, err)
	assert.True(t, virtuous)

	conflicts, err := c.Conflicts(tx)
	assert.NoError(t, err)
	assert.Empty(t, conflicts)
}

func TestInputConflicts(t *testing.T) {
	c := New()

	inputIDs := []ids.ID{ids.GenerateTestID()}
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV:       ids.GenerateTestID(),
			InputIDsV: inputIDs,
		},
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV:       ids.GenerateTestID(),
			InputIDsV: inputIDs,
		},
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	virtuous, err := c.IsVirtuous(tx1)
	assert.NoError(t, err)
	assert.False(t, virtuous)

	conflicts, err := c.Conflicts(tx1)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
}

func TestOuterRestrictionConflicts(t *testing.T) {
	c := New()

	transitionID := ids.GenerateTestID()
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: transitionID,
		},
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: ids.GenerateTestID(),
		},
		EpochV:        1,
		RestrictionsV: []ids.ID{transitionID},
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	virtuous, err := c.IsVirtuous(tx1)
	assert.NoError(t, err)
	assert.False(t, virtuous)

	conflicts, err := c.Conflicts(tx1)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
}

func TestInnerRestrictionConflicts(t *testing.T) {
	c := New()

	transitionID := ids.GenerateTestID()
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: transitionID,
		},
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: ids.GenerateTestID(),
		},
		EpochV:        1,
		RestrictionsV: []ids.ID{transitionID},
	}

	err := c.Add(tx1)
	assert.NoError(t, err)

	virtuous, err := c.IsVirtuous(tx0)
	assert.NoError(t, err)
	assert.False(t, virtuous)

	conflicts, err := c.Conflicts(tx0)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
}

func TestAcceptNoConflicts(t *testing.T) {
	c := New()

	tx := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV: ids.GenerateTestID(),
		},
	}

	err := c.Add(tx)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)

	toAccept := toAccepts[0]
	assert.Equal(t, tx.ID(), toAccept.ID())
}

func TestAcceptNoConflictsWithDependency(t *testing.T) {
	c := New()

	transitionID := ids.GenerateTestID()
	tr0 := &TestTransition{
		IDV: transitionID,
	}
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr0,
	}
	tr1 := &TestTransition{
		IDV:           ids.GenerateTestID(),
		DependenciesV: []Transition{tr0},
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr1,
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	err = c.Add(tx1)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx0.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept := toAccepts[0]
	assert.Equal(t, tx0.ID(), toAccept.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)

	toAccept = toAccepts[0]
	assert.Equal(t, tx1.ID(), toAccept.ID())
}

func TestNoConflictsNoEarlyAcceptDependency(t *testing.T) {
	c := New()

	tr0 := &TestTransition{
		IDV: ids.GenerateTestID(),
	}
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr0,
	}
	tr1 := &TestTransition{
		IDV:           ids.GenerateTestID(),
		DependenciesV: []Transition{tr0},
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr1,
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	err = c.Add(tx1)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx0.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept := toAccepts[0]
	assert.Equal(t, tx0.ID(), toAccept.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)

	toAccept = toAccepts[0]
	assert.Equal(t, tx1.ID(), toAccept.ID())
}

func TestAcceptNoConflictsWithDependenciesAcrossMultipleRounds(t *testing.T) {
	c := New()

	tr0 := &TestTransition{
		IDV: ids.GenerateTestID(),
	}
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr0,
	}
	tr1 := &TestTransition{
		IDV: ids.GenerateTestID(),
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr1,
	}
	tr2 := &TestTransition{
		IDV:           ids.GenerateTestID(),
		DependenciesV: []Transition{tr0, tr1},
	}
	tx2 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr2,
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	err = c.Add(tx1)
	assert.NoError(t, err)

	err = c.Add(tx2)
	assert.NoError(t, err)

	// Check that no transactions are mistakenly marked
	// as accepted/rejected
	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	// Accept tx2 and ensure that it is marked
	// as conditionally accepted pending its
	// dependencies.
	c.Accept(tx2.ID())

	assert.Equal(t, c.conditionallyAccepted.Len(), 1)

	toAccepts, toRejects = c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)
	assert.Equal(t, c.conditionallyAccepted.Len(), 1)

	// Accept tx1 and ensure that it is the only
	// transaction marked as accepted. Note: tx2
	// still requires tx0 to be accepted.
	c.Accept(tx1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept := toAccepts[0]
	assert.Equal(t, tx1.ID(), toAccept.ID())

	// Ensure that additional call to updateable
	// does not return any new accepted/rejected txs.
	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 0)
	assert.Len(t, toRejects, 0)

	// Accept tx0 and ensure that it is
	// returned from Updateable
	c.Accept(tx0.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept = toAccepts[0]
	assert.Equal(t, tx0.ID(), toAccept.ID())

	// tx2 should be returned by the subseqeuent call
	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept = toAccepts[0]
	assert.Equal(t, tx2.ID(), toAccept.ID())

	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)
}
func TestAcceptRejectedDependency(t *testing.T) {
	c := New()

	inputIDs := []ids.ID{ids.GenerateTestID()}
	tr0 := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDs,
	}
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr0,
	}
	tr1 := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDs,
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr1,
	}
	tx2 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV:           ids.GenerateTestID(),
			DependenciesV: []Transition{tr0},
		},
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	err = c.Add(tx1)
	assert.NoError(t, err)

	err = c.Add(tx2)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 1)

	toAccept := toAccepts[0]
	assert.Equal(t, tx1.ID(), toAccept.ID())

	toReject := toRejects[0]
	assert.Equal(t, tx0.ID(), toReject.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Len(t, toRejects, 1)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)

	toReject = toRejects[0]
	assert.Equal(t, tx2.ID(), toReject.ID())
}

func TestAcceptRejectedEpochDependency(t *testing.T) {
	c := New()

	inputIDs := []ids.ID{ids.GenerateTestID()}
	tr := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDs,
	}
	tx0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr,
	}
	tx1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr,
	}
	tx2 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: tr,
		EpochV:      1,
	}
	tx3 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: &TestTransition{
			IDV:           ids.GenerateTestID(),
			DependenciesV: []Transition{tr},
		},
	}

	err := c.Add(tx0)
	assert.NoError(t, err)

	err = c.Add(tx1)
	assert.NoError(t, err)

	err = c.Add(tx2)
	assert.NoError(t, err)

	err = c.Add(tx3)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(tx2.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 3)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)
}

func TestAcceptRestrictedDependency(t *testing.T) {
	c := New()

	inputIDs := []ids.ID{ids.GenerateTestID()}
	trA := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDs,
	}
	trB := &TestTransition{
		IDV:           ids.GenerateTestID(),
		InputIDsV:     []ids.ID{ids.GenerateTestID()},
		DependenciesV: []Transition{trA},
	}
	trC := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: []ids.ID{ids.GenerateTestID()},
	}

	txA0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trA,
	}
	txA1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trA,
		EpochV:      1,
	}
	txB0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trB,
	}
	txB1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trB,
		EpochV:      1,
	}
	txC0 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV:   trC,
		RestrictionsV: []ids.ID{trA.ID()},
	}
	txC1 := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV:   trC,
		EpochV:        1,
		RestrictionsV: []ids.ID{trA.ID()},
	}

	err := c.Add(txA0)
	assert.NoError(t, err)

	err = c.Add(txA1)
	assert.NoError(t, err)

	err = c.Add(txB0)
	assert.NoError(t, err)

	err = c.Add(txB1)
	assert.NoError(t, err)

	err = c.Add(txC0)
	assert.NoError(t, err)

	err = c.Add(txC1)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	// Accepting tx1 should restrict trA to epoch 1 rejecting
	// txA0 and txB0 as a result.
	c.Accept(txC1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 2)

	toAccept := toAccepts[0]
	assert.Equal(t, txC1.ID(), toAccept.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Len(t, toRejects, 1)

	toReject := toRejects[0]
	assert.Equal(t, txB0.ID(), toReject.ID())

	c.Accept(txA1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept = toAccepts[0]
	assert.Equal(t, txA1.ID(), toAccept.ID())

	c.Accept(txB1.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Empty(t, toRejects)

	toAccept = toAccepts[0]
	assert.Equal(t, txB1.ID(), toAccept.ID())

	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)
}

func TestRejectedRejectedDependency(t *testing.T) {
	c := New()

	inputIDA := ids.GenerateTestID()
	inputIDB := ids.GenerateTestID()

	//   A.X - A.Y
	//          |
	//   B.X - B.Y
	trAX := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: []ids.ID{inputIDA, ids.GenerateTestID()},
	}
	txAX := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trAX,
	}
	trAY := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: []ids.ID{inputIDA},
	}
	txAY := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trAY,
	}
	trBX := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: []ids.ID{inputIDB},
	}
	txBX := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trBX,
	}
	trBY := &TestTransition{
		IDV:           ids.GenerateTestID(),
		DependenciesV: []Transition{trAY},
		InputIDsV:     []ids.ID{inputIDB},
	}
	txBY := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trBY,
	}

	err := c.Add(txAY)
	assert.NoError(t, err)

	err = c.Add(txAX)
	assert.NoError(t, err)

	err = c.Add(txBY)
	assert.NoError(t, err)

	err = c.Add(txBX)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(txBX.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 1)

	c.Accept(txAY.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 1)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)
}

func TestAcceptVirtuousRejectedDependency(t *testing.T) {
	c := New()

	inputIDsA := []ids.ID{ids.GenerateTestID()}
	inputIDsB := []ids.ID{ids.GenerateTestID()}

	//   A.X - A.Y
	//          |
	//   B.X - B.Y
	trAX := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDsA,
	}
	txAX := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trAX,
	}
	trAY := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDsA,
	}
	txAY := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trAY,
	}
	trBX := &TestTransition{
		IDV:       ids.GenerateTestID(),
		InputIDsV: inputIDsB,
	}
	txBX := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trBX,
	}
	trV := &TestTransition{
		IDV:           ids.GenerateTestID(),
		DependenciesV: []Transition{trAY},
		InputIDsV:     inputIDsB,
	}
	txBY := &TestTx{
		TestDecidable: choices.TestDecidable{
			IDV:     ids.GenerateTestID(),
			StatusV: choices.Processing,
		},
		TransitionV: trV,
	}

	err := c.Add(txAX)
	assert.NoError(t, err)

	err = c.Add(txAY)
	assert.NoError(t, err)

	err = c.Add(txBX)
	assert.NoError(t, err)

	err = c.Add(txBY)
	assert.NoError(t, err)

	toAccepts, toRejects := c.Updateable()
	assert.Empty(t, toAccepts)
	assert.Empty(t, toRejects)

	c.Accept(txAX.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 1)

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 0)
	assert.Len(t, toRejects, 1)

	c.Accept(txBX.ID())

	toAccepts, toRejects = c.Updateable()
	assert.Len(t, toAccepts, 1)
	assert.Len(t, toRejects, 0)
	assert.Empty(t, c.txs)
	assert.Empty(t, c.utxos)
	assert.Empty(t, c.transitionNodes)
}
