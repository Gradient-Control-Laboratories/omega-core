package ir

// SinkStatements performs deterministic in-memory statement sinking and local dead-code pruning.
// It sinks LetStmt nodes into GuardStmt.Body when read exclusively within that guard and
// unaffected by intervening writes, recurses through nested guard bodies, and prunes unused
// synthetic LetStmt nodes. liveExtra specifies optional external identifiers (such as from a return
// expression) that must be preserved as live.
func SinkStatements(block *Block, liveExtra ...string) {
	if block.Len() == 0 {
		return
	}

	liveExtraMap := make(map[string]bool, len(liveExtra))
	for _, id := range liveExtra {
		liveExtraMap[id] = true
	}

	stmts := block.stmts
	n := len(stmts)
	into := make([]int, n)
	for i := range into {
		into[i] = -1
	}

	// 1. Settle destinations backwards so transitive lets settle into the same guard.
	for i := n - 1; i >= 0; i-- {
		letStmt, ok := stmts[i].(*LetStmt)
		if !ok {
			continue
		}

		if liveExtraMap[letStmt.Name] {
			// Read by external expression / return; cannot sink into an inner guard.
			continue
		}

		targetGuardIdx := -1
		varReadElsewhere := false
		letReads := letStmt.Reads()
		letReadsMap := make(map[string]bool, len(letReads))
		for _, r := range letReads {
			letReadsMap[r] = true
		}

		for j := i + 1; j < n && !varReadElsewhere; j++ {
			if guard, isGuard := stmts[j].(*GuardStmt); isGuard {
				for _, hr := range guard.HeaderReads() {
					if hr == letStmt.Name {
						varReadElsewhere = true
						break
					}
				}
				if varReadElsewhere {
					break
				}
				guardBodyReadsLet := false
				for _, br := range guard.BodyReads() {
					if br == letStmt.Name {
						guardBodyReadsLet = true
						break
					}
				}
				if guardBodyReadsLet {
					if targetGuardIdx >= 0 && targetGuardIdx != j {
						varReadElsewhere = true
					} else {
						targetGuardIdx = j
					}
				}
			} else if otherLet, isOtherLet := stmts[j].(*LetStmt); isOtherLet {
				otherReadsLet := false
				for _, r := range otherLet.Reads() {
					if r == letStmt.Name {
						otherReadsLet = true
						break
					}
				}
				if otherReadsLet {
					if into[j] >= 0 {
						if targetGuardIdx >= 0 && targetGuardIdx != into[j] {
							varReadElsewhere = true
						} else {
							targetGuardIdx = into[j]
						}
					} else {
						varReadElsewhere = true
					}
				}
			} else {
				for _, r := range stmts[j].Reads() {
					if r == letStmt.Name {
						varReadElsewhere = true
						break
					}
				}
			}
		}

		// Check for intervening writes between i and targetGuardIdx
		if targetGuardIdx >= 0 && !varReadElsewhere {
			for k := i + 1; k < targetGuardIdx; k++ {
				if _, isGuard := stmts[k].(*GuardStmt); isGuard {
					continue
				}
				for _, w := range stmts[k].Writes() {
					if letReadsMap[w] {
						varReadElsewhere = true
						break
					}
				}
				if varReadElsewhere {
					break
				}
			}
		}

		if targetGuardIdx >= 0 && !varReadElsewhere {
			into[i] = targetGuardIdx
		}
	}

	// 2. Move sunk statements in increasing order of i into their respective guards.
	// For each guard, keep count of statements inserted so far at the front of its body.
	sunkCount := make(map[*GuardStmt]int)
	newStmts := make([]Stmt, 0, n)
	for i := 0; i < n; i++ {
		if into[i] >= 0 {
			guard := stmts[into[i]].(*GuardStmt)
			insertPos := sunkCount[guard]
			guard.Body.Insert(insertPos, stmts[i])
			sunkCount[guard]++
		} else {
			newStmts = append(newStmts, stmts[i])
		}
	}
	block.stmts = newStmts

	// 2b. Re-align guard accumulator variables immediately before their guard statements.
	for i := 0; i < len(block.stmts); i++ {
		guard, ok := block.stmts[i].(*GuardStmt)
		if !ok {
			continue
		}
		var assignedVar string
		for _, s := range guard.Body.stmts {
			if assign, ok := s.(*AssignStmt); ok {
				assignedVar = assign.Target
				break
			}
		}
		if assignedVar == "" {
			continue
		}
		varIdx := -1
		for k := i - 1; k >= 0; k-- {
			if v, ok := block.stmts[k].(*VarStmt); ok && v.Name == assignedVar {
				varIdx = k
				break
			}
		}
		if varIdx >= 0 && varIdx < i-1 {
			vStmt := block.stmts[varIdx]
			copy(block.stmts[varIdx:i-1], block.stmts[varIdx+1:i])
			block.stmts[i-1] = vStmt
		}
	}

	// 3. Recursive sinking on child guard and loop bodies.
	for _, s := range block.stmts {
		switch node := s.(type) {
		case *GuardStmt:
			SinkStatements(node.Body)
		case *LoopStmt:
			SinkStatements(node.Body)
		}
	}

	// 4. Local dead-code pruning: eliminate LetStmt whose Name is never read.
	PruneDeadCode(block, liveExtra...)
}

// PruneDeadCode eliminates LetStmt instances that are never read by subsequent statements or returns.
func PruneDeadCode(block *Block, liveExtra ...string) {
	if block.Len() == 0 {
		return
	}

	// Recurse into child guard and loop bodies first.
	for _, s := range block.stmts {
		switch node := s.(type) {
		case *GuardStmt:
			PruneDeadCode(node.Body)
		case *LoopStmt:
			PruneDeadCode(node.Body)
		}
	}

	live := make(map[string]bool, len(liveExtra))
	for _, id := range liveExtra {
		live[id] = true
	}

	i := len(block.stmts) - 1
	for i >= 0 {
		s := block.stmts[i]
		if let, ok := s.(*LetStmt); ok {
			if !live[let.Name] {
				block.Remove(i)
				i--
				continue
			}
			for _, r := range let.Reads() {
				live[r] = true
			}
		} else {
			for _, r := range s.Reads() {
				live[r] = true
			}
		}
		i--
	}
}
