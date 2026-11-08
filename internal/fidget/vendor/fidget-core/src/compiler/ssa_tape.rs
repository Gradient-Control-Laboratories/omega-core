//use crate::vm::{RegisterAllocator, Tape as VmTape};
use crate::{
    Context,
    compiler::SsaOp,
    context::{BadNode, BinaryOpcode, Node, Op, UnaryOpcode},
    var::VarMap,
};
use serde::{Deserialize, Serialize};

use std::collections::{HashMap, HashSet};

/// Instruction tape, storing [opcodes in SSA form](crate::compiler::SsaOp)
///
/// Each operation has the following parameters
/// - 4-byte opcode (required)
/// - 4-byte output register (required)
/// - 4-byte LHS register
/// - 4-byte RHS register (or immediate `f32`)
///
/// All register addressing is absolute.
#[derive(Clone, Debug, Default, Serialize, Deserialize)]
pub struct SsaTape {
    /// The tape is stored in reverse order, such that the root of the tree is
    /// the first item in the tape.
    pub tape: Vec<SsaOp>,

    /// Number of choice operations in the tape
    pub choice_count: usize,

    /// Number of output operations in the tape
    pub output_count: usize,
}

impl SsaTape {
    /// Flattens a subtree of the graph into straight-line code.
    ///
    /// This should always succeed unless the `root` is from a different
    /// `Context`, in which case [`BadNode`] will be returned.
    pub fn new(
        ctx: &Context,
        roots: &[Node],
    ) -> Result<(Self, VarMap), BadNode> {
        let mut mapping = HashMap::new();
        let mut parent_count: HashMap<Node, usize> = HashMap::new();
        let mut slot_count = 0;

        // Get either a node or constant index
        #[derive(Copy, Clone, Debug)]
        enum Slot {
            Reg(u32),
            Immediate(f32),
        }

        // Accumulate parent counts and declare all nodes
        let mut seen = HashSet::new();
        let mut vars = VarMap::new();
        let mut todo = roots.to_vec();
        while let Some(node) = todo.pop() {
            if !seen.insert(node) {
                continue;
            }
            let op = ctx.get_op(node).ok_or(BadNode)?;
            let prev = match op {
                Op::Const(c) => mapping.insert(node, Slot::Immediate(c.0)),
                _ => {
                    if let Op::Input(v) = op {
                        vars.insert(*v);
                    }
                    let i = slot_count;
                    slot_count += 1;
                    mapping.insert(node, Slot::Reg(i))
                }
            };
            assert!(prev.is_none());
            for child in op.iter_children() {
                *parent_count.entry(child).or_default() += 1;
                todo.push(child);
            }
        }

        // Now that we've populated our parents, flatten the graph
        let mut seen = HashSet::new();
        let mut todo = roots.to_vec();
        let mut choice_count = 0;

        let mut tape = vec![];
        for (i, r) in roots.iter().enumerate() {
            let i = i as u32;
            match mapping[r] {
                Slot::Reg(out_reg) => tape.push(SsaOp::Output(out_reg, i)),
                Slot::Immediate(imm) => {
                    let o = slot_count;
                    slot_count += 1;
                    tape.push(SsaOp::Output(o, i));
                    tape.push(SsaOp::CopyImm(o, imm));
                }
            }
        }

        while let Some(node) = todo.pop() {
            if *parent_count.get(&node).unwrap_or(&0) > 0 || !seen.insert(node)
            {
                continue;
            }

            let op = ctx.get_op(node).unwrap();
            for child in op.iter_children() {
                todo.push(child);
                *parent_count.get_mut(&child).unwrap() -= 1;
            }

            let Slot::Reg(i) = mapping[&node] else {
                // Constants are skipped, because they become immediates
                continue;
            };
            let mut loads = vec![];
            let op = match op {
                Op::Input(v) => {
                    let arg = vars[v];
                    SsaOp::Input(i, arg.try_into().unwrap())
                }
                Op::Const(..) => {
                    unreachable!("skipped above")
                }
                Op::MeshOracle(asset_id, x, y, z) => {
                    // The oracle reads registers only. An input folded to a
                    // constant, as noiseHash3([cx, cy, 23]) folds its z, is
                    // loaded into a fresh slot, the way a constant output is
                    // above; pushed after the oracle, the load runs first.
                    let mut reg = |slot: Slot| match slot {
                        Slot::Reg(r) => r,
                        Slot::Immediate(imm) => {
                            let o = slot_count;
                            slot_count += 1;
                            loads.push(SsaOp::CopyImm(o, imm));
                            o
                        }
                    };
                    let x_slot = reg(mapping[x]);
                    let y_slot = reg(mapping[y]);
                    let z_slot = reg(mapping[z]);
                    SsaOp::MeshOracle(i, *asset_id, x_slot, y_slot, z_slot)
                }
                Op::Binary(op, lhs, rhs) => {
                    let lhs = mapping[lhs];
                    let rhs = mapping[rhs];

                    type RegFn = fn(u32, u32, u32) -> SsaOp;
                    type ImmFn = fn(u32, u32, f32) -> SsaOp;
                    let f: (RegFn, ImmFn, ImmFn) = match op {
                        BinaryOpcode::Add => (
                            SsaOp::AddRegReg,
                            SsaOp::AddRegImm,
                            SsaOp::AddRegImm,
                        ),
                        BinaryOpcode::Sub => (
                            SsaOp::SubRegReg,
                            SsaOp::SubRegImm,
                            SsaOp::SubImmReg,
                        ),
                        BinaryOpcode::Mul => (
                            SsaOp::MulRegReg,
                            SsaOp::MulRegImm,
                            SsaOp::MulRegImm,
                        ),
                        BinaryOpcode::Div => (
                            SsaOp::DivRegReg,
                            SsaOp::DivRegImm,
                            SsaOp::DivImmReg,
                        ),
                        BinaryOpcode::Atan => (
                            SsaOp::AtanRegReg,
                            SsaOp::AtanRegImm,
                            SsaOp::AtanImmReg,
                        ),
                        BinaryOpcode::Min => (
                            SsaOp::MinRegReg,
                            SsaOp::MinRegImm,
                            SsaOp::MinRegImm,
                        ),
                        BinaryOpcode::Max => (
                            SsaOp::MaxRegReg,
                            SsaOp::MaxRegImm,
                            SsaOp::MaxRegImm,
                        ),
                        BinaryOpcode::And => (
                            SsaOp::AndRegReg,
                            SsaOp::AndRegImm,
                            |_out, _lhs, _rhs| {
                                panic!("AndImmReg must be collapsed")
                            },
                        ),
                        BinaryOpcode::Or => (
                            SsaOp::OrRegReg,
                            SsaOp::OrRegImm,
                            |_out, _lhs, _rhs| {
                                panic!("OrImmReg must be collapsed")
                            },
                        ),
                        BinaryOpcode::Compare => (
                            SsaOp::CompareRegReg,
                            SsaOp::CompareRegImm,
                            SsaOp::CompareImmReg,
                        ),
                        BinaryOpcode::Mod => (
                            SsaOp::ModRegReg,
                            SsaOp::ModRegImm,
                            SsaOp::ModImmReg,
                        ),
                        BinaryOpcode::Mix => (
                            SsaOp::MixRegReg,
                            SsaOp::MixRegImm,
                            SsaOp::MixImmReg,
                        ),
                    };

                    if matches!(
                        op,
                        BinaryOpcode::Min
                            | BinaryOpcode::Max
                            | BinaryOpcode::And
                            | BinaryOpcode::Or
                    ) {
                        choice_count += 1;
                    }

                    match (lhs, rhs) {
                        (Slot::Reg(lhs), Slot::Reg(rhs)) => f.0(i, lhs, rhs),
                        (Slot::Reg(arg), Slot::Immediate(imm)) => {
                            f.1(i, arg, imm)
                        }
                        (Slot::Immediate(imm), Slot::Reg(arg)) => {
                            f.2(i, arg, imm)
                        }
                        (Slot::Immediate(..), Slot::Immediate(..)) => {
                            panic!("Cannot handle f(imm, imm)")
                        }
                    }
                }
                Op::Unary(op, lhs) => {
                    let lhs = match mapping[lhs] {
                        Slot::Reg(r) => r,
                        Slot::Immediate(..) => {
                            panic!("Cannot handle f(imm)")
                        }
                    };
                    let op = match op {
                        UnaryOpcode::Neg => SsaOp::NegReg,
                        UnaryOpcode::Abs => SsaOp::AbsReg,
                        UnaryOpcode::Recip => SsaOp::RecipReg,
                        UnaryOpcode::Sqrt => SsaOp::SqrtReg,
                        UnaryOpcode::Square => SsaOp::SquareReg,
                        UnaryOpcode::Floor => SsaOp::FloorReg,
                        UnaryOpcode::Ceil => SsaOp::CeilReg,
                        UnaryOpcode::Round => SsaOp::RoundReg,
                        UnaryOpcode::Sin => SsaOp::SinReg,
                        UnaryOpcode::Cos => SsaOp::CosReg,
                        UnaryOpcode::Tan => SsaOp::TanReg,
                        UnaryOpcode::Asin => SsaOp::AsinReg,
                        UnaryOpcode::Acos => SsaOp::AcosReg,
                        UnaryOpcode::Atan => SsaOp::AtanReg,
                        UnaryOpcode::Exp => SsaOp::ExpReg,
                        UnaryOpcode::Ln => SsaOp::LnReg,
                        UnaryOpcode::Not => SsaOp::NotReg,
                        UnaryOpcode::Rand => SsaOp::RandReg,
                    };
                    op(i, lhs)
                }
            };
            tape.push(op);
            tape.extend(loads);
        }

        Ok((
            SsaTape {
                tape,
                choice_count,
                output_count: roots.len(),
            },
            vars,
        ))
    }

    /// Checks whether the tape is empty
    pub fn is_empty(&self) -> bool {
        self.tape.is_empty()
    }

    /// Returns the length of the tape
    pub fn len(&self) -> usize {
        self.tape.len()
    }

    /// Iterates over clauses in the tape in reverse-evaluation order
    ///
    /// The root (output) of the tape will be first in the iterator
    pub fn iter(&self) -> impl DoubleEndedIterator<Item = &SsaOp> {
        self.tape.iter()
    }

    /// Resets to an empty tape, preserving allocations
    pub fn reset(&mut self) {
        self.tape.clear();
        self.choice_count = 0;
    }
    /// Pretty-prints the given tape to `stdout`
    pub fn pretty_print(&self) {
        for &op in self.tape.iter().rev() {
            match op {
                SsaOp::Output(arg, i) => {
                    println!("OUTPUT[{i}] = ${arg}");
                }
                SsaOp::Input(out, i) => {
                    println!("${out} = INPUT[{i}]");
                }
                SsaOp::NegReg(out, arg)
                | SsaOp::AbsReg(out, arg)
                | SsaOp::RecipReg(out, arg)
                | SsaOp::SqrtReg(out, arg)
                | SsaOp::CopyReg(out, arg)
                | SsaOp::SquareReg(out, arg)
                | SsaOp::FloorReg(out, arg)
                | SsaOp::CeilReg(out, arg)
                | SsaOp::RoundReg(out, arg)
                | SsaOp::SinReg(out, arg)
                | SsaOp::CosReg(out, arg)
                | SsaOp::TanReg(out, arg)
                | SsaOp::AsinReg(out, arg)
                | SsaOp::AcosReg(out, arg)
                | SsaOp::AtanReg(out, arg)
                | SsaOp::ExpReg(out, arg)
                | SsaOp::LnReg(out, arg)
                | SsaOp::NotReg(out, arg)
                | SsaOp::RandReg(out, arg) => {
                    let op = match op {
                        SsaOp::NegReg(..) => "NEG",
                        SsaOp::AbsReg(..) => "ABS",
                        SsaOp::RecipReg(..) => "RECIP",
                        SsaOp::SqrtReg(..) => "SQRT",
                        SsaOp::SquareReg(..) => "SQUARE",
                        SsaOp::FloorReg(..) => "FLOOR",
                        SsaOp::CeilReg(..) => "CEIL",
                        SsaOp::RoundReg(..) => "ROUND",
                        SsaOp::SinReg(..) => "SIN",
                        SsaOp::CosReg(..) => "COS",
                        SsaOp::TanReg(..) => "TAN",
                        SsaOp::AsinReg(..) => "ASIN",
                        SsaOp::AcosReg(..) => "ACOS",
                        SsaOp::AtanReg(..) => "ATAN",
                        SsaOp::ExpReg(..) => "EXP",
                        SsaOp::LnReg(..) => "LN",
                        SsaOp::NotReg(..) => "NOT",
                        SsaOp::CopyReg(..) => "COPY",
                        SsaOp::RandReg(..) => "RAND",
                        _ => unreachable!(),
                    };
                    println!("${out} = {op} ${arg}");
                }

                SsaOp::AddRegReg(out, lhs, rhs)
                | SsaOp::MulRegReg(out, lhs, rhs)
                | SsaOp::DivRegReg(out, lhs, rhs)
                | SsaOp::SubRegReg(out, lhs, rhs)
                | SsaOp::MinRegReg(out, lhs, rhs)
                | SsaOp::MaxRegReg(out, lhs, rhs)
                | SsaOp::ModRegReg(out, lhs, rhs)
                | SsaOp::AndRegReg(out, lhs, rhs)
                | SsaOp::AtanRegReg(out, lhs, rhs)
                | SsaOp::OrRegReg(out, lhs, rhs)
                | SsaOp::CompareRegReg(out, lhs, rhs)
                | SsaOp::MixRegReg(out, lhs, rhs) => {
                    let op = match op {
                        SsaOp::AddRegReg(..) => "ADD",
                        SsaOp::MulRegReg(..) => "MUL",
                        SsaOp::DivRegReg(..) => "DIV",
                        SsaOp::AtanRegReg(..) => "ATAN",
                        SsaOp::SubRegReg(..) => "SUB",
                        SsaOp::MinRegReg(..) => "MIN",
                        SsaOp::MaxRegReg(..) => "MAX",
                        SsaOp::ModRegReg(..) => "MAX",
                        SsaOp::AndRegReg(..) => "AND",
                        SsaOp::OrRegReg(..) => "OR",
                        SsaOp::CompareRegReg(..) => "COMPARE",
                        SsaOp::MixRegReg(..) => "MIX",
                        _ => unreachable!(),
                    };
                    println!("${out} = {op} ${lhs} ${rhs}");
                }

                SsaOp::AddRegImm(out, arg, imm)
                | SsaOp::MulRegImm(out, arg, imm)
                | SsaOp::DivRegImm(out, arg, imm)
                | SsaOp::DivImmReg(out, arg, imm)
                | SsaOp::SubImmReg(out, arg, imm)
                | SsaOp::SubRegImm(out, arg, imm)
                | SsaOp::AtanRegImm(out, arg, imm)
                | SsaOp::AtanImmReg(out, arg, imm)
                | SsaOp::MinRegImm(out, arg, imm)
                | SsaOp::MaxRegImm(out, arg, imm)
                | SsaOp::ModRegImm(out, arg, imm)
                | SsaOp::ModImmReg(out, arg, imm)
                | SsaOp::AndRegImm(out, arg, imm)
                | SsaOp::OrRegImm(out, arg, imm)
                | SsaOp::CompareRegImm(out, arg, imm)
                | SsaOp::CompareImmReg(out, arg, imm)
                | SsaOp::MixRegImm(out, arg, imm)
                | SsaOp::MixImmReg(out, arg, imm) => {
                    let (op, swap) = match op {
                        SsaOp::AddRegImm(..) => ("ADD", false),
                        SsaOp::MulRegImm(..) => ("MUL", false),
                        SsaOp::DivImmReg(..) => ("DIV", true),
                        SsaOp::DivRegImm(..) => ("DIV", false),
                        SsaOp::SubImmReg(..) => ("SUB", true),
                        SsaOp::SubRegImm(..) => ("SUB", false),
                        SsaOp::AtanImmReg(..) => ("ATAN", true),
                        SsaOp::AtanRegImm(..) => ("ATAN", false),
                        SsaOp::MinRegImm(..) => ("MIN", false),
                        SsaOp::MaxRegImm(..) => ("MAX", false),
                        SsaOp::ModRegImm(..) => ("MOD", false),
                        SsaOp::ModImmReg(..) => ("MOD", true),
                        SsaOp::AndRegImm(..) => ("AND", false),
                        SsaOp::OrRegImm(..) => ("OR", false),
                        SsaOp::CompareRegImm(..) => ("COMPARE", false),
                        SsaOp::CompareImmReg(..) => ("COMPARE", true),
                        SsaOp::MixRegImm(..) => ("MIX", false),
                        SsaOp::MixImmReg(..) => ("MIX", true),
                        _ => unreachable!(),
                    };
                    if swap {
                        println!("${out} = {op} {imm} ${arg}");
                    } else {
                        println!("${out} = {op} ${arg} {imm}");
                    }
                }
                SsaOp::CopyImm(out, imm) => {
                    println!("${out} = COPY {imm}");
                }
                SsaOp::MeshOracle(out, asset_id, x, y, z) => {
                    println!(
                        "${out} = MESH_ORACLE {asset_id} ${x} ${y} ${z}"
                    );
                }
            }
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_mesh_oracle_constant_coordinates() {
        use crate::{
            mesh_oracle::{self, MeshOracle},
            shape::EzShape,
            types::{Grad, Interval},
            vm::VmShape,
        };

        struct Plane;
        impl MeshOracle for Plane {
            fn point(&self, x: f32, y: f32, z: f32) -> f32 {
                x + 2.0 * y + 3.0 * z
            }
            fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
                x + y * 2.0 + z * 3.0
            }
            fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
                Grad::new(self.point(x, y, z), 1.0, 2.0, 3.0)
            }
        }

        mesh_oracle::register(0x7fff_ff01, std::sync::Arc::new(Plane));
        for mask in 1..8 {
            let mut ctx = Context::new();
            let axes = ctx.axes();
            let constants = [ctx.constant(2.0), ctx.constant(3.0), ctx.constant(4.0)];
            let [x, y, z] = std::array::from_fn(|i| {
                if mask & (1 << i) != 0 { constants[i] } else { axes[i] }
            });
            let node = ctx.mesh_oracle(0x7fff_ff01, x, y, z);
            let shape = VmShape::new(&ctx, node).unwrap();
            let tape = shape.ez_grad_slice_tape();
            let mut eval = VmShape::new_grad_slice_eval();
            let samples = eval.eval(&tape,
                &[Grad::new(2.0, 1.0, 0.0, 0.0)],
                &[Grad::new(3.0, 0.0, 1.0, 0.0)],
                &[Grad::new(4.0, 0.0, 0.0, 1.0)]).unwrap();
            let expected = |axis| if mask & (1 << axis) == 0 { (axis + 1) as f32 } else { 0.0 };
            assert_eq!(samples[0], Grad::new(20.0, expected(0), expected(1), expected(2)));
        }
    }

    #[test]
    fn test_ring() {
        let mut ctx = Context::new();
        let c0 = ctx.constant(0.5);
        let x = ctx.x();
        let y = ctx.y();
        let x2 = ctx.square(x).unwrap();
        let y2 = ctx.square(y).unwrap();
        let r = ctx.add(x2, y2).unwrap();
        let c6 = ctx.sub(r, c0).unwrap();
        let c7 = ctx.constant(0.25);
        let c8 = ctx.sub(c7, r).unwrap();
        let c9 = ctx.max(c8, c6).unwrap();

        let (tape, vs) = SsaTape::new(&ctx, &[c9]).unwrap();
        assert_eq!(tape.len(), 9);
        assert_eq!(vs.len(), 2);
    }

    #[test]
    fn test_dupe() {
        let mut ctx = Context::new();
        let x = ctx.x();
        let x_squared = ctx.mul(x, x).unwrap();

        let (tape, vs) = SsaTape::new(&ctx, &[x_squared]).unwrap();
        assert_eq!(tape.len(), 3); // x, square, output
        assert_eq!(vs.len(), 1);
    }

    #[test]
    fn test_constant() {
        let mut ctx = Context::new();
        let p = ctx.constant(1.5);
        let (tape, vs) = SsaTape::new(&ctx, &[p]).unwrap();
        assert_eq!(tape.len(), 2); // CopyImm, output
        assert_eq!(vs.len(), 0);
    }
}
