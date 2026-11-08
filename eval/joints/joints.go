// Package joints finds the mechanical joints a GSL document declares and proves each one's claims.
package joints

import (
	internaljoints "gradientcontrol.com/omega/core/internal/gsleval/joints"
)

// Joint is one joint call found in a document.
type Joint = internaljoints.Joint

// Claim is one predicate a joint requires to hold (Expr >= 0 over the joint's region).
type Claim = internaljoints.Claim

// Verdict is one certify outcome.
type Verdict = internaljoints.Verdict

// Prover supplies point evaluation and interval certification for Run.
type Prover = internaljoints.Prover

// ClaimResult is one claim and its verification outcome.
type ClaimResult = internaljoints.ClaimResult

// JointReport is one joint's claims and their verification outcomes.
type JointReport = internaljoints.JointReport

// Report is the verification report for every joint in a document.
type Report = internaljoints.Report

// Find returns the joint calls a document declares at top level, the expressions
// its top-level return lists, and a note for each joint call it cannot verify.
func Find(src string) (found []Joint, returns []string, skipped []string) {
	return internaljoints.Find(src)
}

// Others returns the returned parts the screw is held clear of.
func Others(j Joint, returns []string) []string {
	return internaljoints.Others(j, returns)
}

// Claims returns the claims required for joint j and the other assembly parts.
func Claims(j Joint, others []string) []Claim {
	return internaljoints.Claims(j, others)
}

// RegionExprs returns the expressions whose values Region needs.
func RegionExprs(j Joint) []string {
	return internaljoints.RegionExprs(j)
}

// Region computes the padded bounding box around a joint's axis.
func Region(seat, dir [3]float64, d, length, counterbore float64) ([6]float64, error) {
	return internaljoints.Region(seat, dir, d, length, counterbore)
}

// Run finds src's joints and proves each one's claims with pv.
func Run(src string, pv Prover) Report {
	return internaljoints.Run(src, pv)
}
