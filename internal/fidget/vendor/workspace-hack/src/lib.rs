//! Empty stand-in for the `workspace-hack` crate of upstream Fidget.
//!
//! Upstream generates that crate with cargo-hakari, to unify dependency
//! features across its own workspace, and `fidget-core` names it with
//! `use workspace_hack as _;` (Fidget #517). This repository has its own
//! workspace and lockfile and needs no such unification, so the vendored
//! manifests do not depend on upstream's copy. This crate only gives that one
//! line something to name, and it is not part of what `tool/vendor-fidget`
//! copies.
