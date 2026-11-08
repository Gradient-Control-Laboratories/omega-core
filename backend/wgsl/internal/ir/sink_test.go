package ir_test

import (
	"bytes"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

func TestSinkStatements_SingleGuard(t *testing.T) {
	block := ir.NewBlock()
	// Candidate let only read by guard
	block.Add(ir.NewLetStmt("noiseVal", "f32", "expensiveNoise(p)", []string{"p"}))
	// Unrelated let read by return
	block.Add(ir.NewLetStmt("baseVal", "f32", "1.0", nil))

	guardBody := ir.NewBlock()
	guardBody.Add(ir.NewAssignStmt("part", "noiseVal + 10.0", []string{"noiseVal"}))
	block.Add(ir.NewGuardStmt("detail > 0.0", guardBody, []string{"detail"}))

	block.Add(ir.NewReturnStmt("baseVal", []string{"baseVal"}))

	ir.SinkStatements(block)

	// After sinking: noiseVal should be inside guardBody, not in outer block!
	stmts := block.Statements()
	assert.Equal(t, 3, len(stmts))

	// Statement 0: baseVal
	let0, ok := stmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "baseVal", let0.Name)

	// Statement 1: guard
	guard, ok := stmts[1].(*ir.GuardStmt)
	assert.True(t, ok)

	// Statement 2: return
	_, ok = stmts[2].(*ir.ReturnStmt)
	assert.True(t, ok)

	// Inside guard: noiseVal should be first statement
	guardStmts := guard.Body.Statements()
	assert.Equal(t, 2, len(guardStmts))
	sunkLet, ok := guardStmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "noiseVal", sunkLet.Name)
}

func TestSinkStatements_InterveningWriteBlocksSinking(t *testing.T) {
	block := ir.NewBlock()
	// noiseVal reads p
	block.Add(ir.NewLetStmt("noiseVal", "f32", "expensiveNoise(p)", []string{"p"}))
	// Intervening write to p
	block.Add(ir.NewAssignStmt("p", "p * 2.0", []string{"p"}))

	guardBody := ir.NewBlock()
	guardBody.Add(ir.NewAssignStmt("part", "noiseVal", []string{"noiseVal"}))
	block.Add(ir.NewGuardStmt("detail > 0.0", guardBody, []string{"detail"}))
	block.Add(ir.NewReturnStmt("part", []string{"part"}))

	ir.SinkStatements(block)

	// Intervening write to p must prevent sinking noiseVal!
	stmts := block.Statements()
	assert.Equal(t, 4, len(stmts))
	let0, ok := stmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "noiseVal", let0.Name)
}

func TestSinkStatements_RecursiveGuards(t *testing.T) {
	block := ir.NewBlock()
	block.Add(ir.NewLetStmt("innerNoise", "f32", "deepNoise(p)", []string{"p"}))

	innerGuardBody := ir.NewBlock()
	innerGuardBody.Add(ir.NewAssignStmt("part", "innerNoise", []string{"innerNoise"}))
	innerGuard := ir.NewGuardStmt("subDetail > 0.0", innerGuardBody, []string{"subDetail"})

	outerGuardBody := ir.NewBlock()
	outerGuardBody.Add(innerGuard)

	block.Add(ir.NewGuardStmt("outerDetail > 0.0", outerGuardBody, []string{"outerDetail"}))
	block.Add(ir.NewReturnStmt("0.0", nil))

	ir.SinkStatements(block)

	// innerNoise should sink through outerGuard into innerGuardBody
	stmts := block.Statements()
	assert.Equal(t, 2, len(stmts))

	outerGuard, ok := stmts[0].(*ir.GuardStmt)
	assert.True(t, ok)

	var buf bytes.Buffer
	assert.NoError(t, outerGuard.Format(&buf, ""))
	rendered := buf.String()

	// Verify deepNoise is inside inner guard
	assert.True(t, strings.Contains(rendered, "if (subDetail > 0.0) {\n    let innerNoise: f32 = deepNoise(p);"))
}

func TestPruneDeadCode(t *testing.T) {
	block := ir.NewBlock()
	// Dead let: never read
	block.Add(ir.NewLetStmt("deadTemp", "f32", "123.0", nil))
	// Live let: read by return
	block.Add(ir.NewLetStmt("liveTemp", "f32", "456.0", nil))
	block.Add(ir.NewReturnStmt("liveTemp", []string{"liveTemp"}))

	ir.PruneDeadCode(block)

	stmts := block.Statements()
	assert.Equal(t, 2, len(stmts))

	let0, ok := stmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "liveTemp", let0.Name)

	_, ok = stmts[1].(*ir.ReturnStmt)
	assert.True(t, ok)
}

func TestPruneDeadCode_LoopBody(t *testing.T) {
	block := ir.NewBlock()
	loopBody := ir.NewBlock()
	// Dead let inside loop body
	loopBody.Add(ir.NewLetStmt("deadLoopLet", "f32", "999.0", nil))
	// Live let inside loop body read by assign
	loopBody.Add(ir.NewLetStmt("liveLoopLet", "f32", "1.0", nil))
	loopBody.Add(ir.NewAssignStmt("acc", "acc + liveLoopLet", []string{"acc", "liveLoopLet"}))

	loop := ir.NewLoopStmt("i", 10, loopBody)
	block.Add(loop)
	block.Add(ir.NewReturnStmt("acc", []string{"acc"}))

	ir.PruneDeadCode(block)

	// deadLoopLet must be pruned from loopBody!
	loopStmts := loop.Body.Statements()
	assert.Equal(t, 2, len(loopStmts))

	let0, ok := loopStmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "liveLoopLet", let0.Name)

	assign, ok := loopStmts[1].(*ir.AssignStmt)
	assert.True(t, ok)
	assert.Equal(t, "acc", assign.Target)
}

func TestSinkStatements_LoopBodyRecursion(t *testing.T) {
	block := ir.NewBlock()
	loopBody := ir.NewBlock()

	// Inside loop body: candidate let only read by guard
	loopBody.Add(ir.NewLetStmt("innerNoise", "f32", "sampleNoise(p)", []string{"p"}))
	// Unrelated let
	loopBody.Add(ir.NewLetStmt("baseVal", "f32", "1.0", nil))

	guardBody := ir.NewBlock()
	guardBody.Add(ir.NewAssignStmt("acc", "acc + innerNoise", []string{"acc", "innerNoise"}))
	loopBody.Add(ir.NewGuardStmt("detail > 0.0", guardBody, []string{"detail"}))
	loopBody.Add(ir.NewAssignStmt("acc", "acc + baseVal", []string{"acc", "baseVal"}))

	loop := ir.NewLoopStmt("i", 5, loopBody)
	block.Add(loop)
	block.Add(ir.NewReturnStmt("acc", []string{"acc"}))

	ir.SinkStatements(block)

	// Sinking should have recursed into loop.Body:
	// innerNoise should be sunk into the guard inside loopBody!
	loopStmts := loop.Body.Statements()
	assert.Equal(t, 3, len(loopStmts))

	let0, ok := loopStmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "baseVal", let0.Name)

	guard, ok := loopStmts[1].(*ir.GuardStmt)
	assert.True(t, ok)

	guardStmts := guard.Body.Statements()
	assert.Equal(t, 2, len(guardStmts))
	sunkLet, ok := guardStmts[0].(*ir.LetStmt)
	assert.True(t, ok)
	assert.Equal(t, "innerNoise", sunkLet.Name)
}
