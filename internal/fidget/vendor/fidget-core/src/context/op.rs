use crate::{
    context::{Node, indexed::Index},
    types::FloatExt,
    var::Var,
};
use ordered_float::OrderedFloat;

/// A one-argument math operation
#[allow(missing_docs)]
#[derive(Copy, Clone, Debug, Hash, Eq, PartialEq, Ord, PartialOrd)]
pub enum UnaryOpcode {
    Neg,
    Abs,
    Recip,
    Sqrt,
    Square,
    Floor,
    Ceil,
    Round,
    Sin,
    Cos,
    Tan,
    Asin,
    Acos,
    Atan,
    Exp,
    Ln,
    Not,
    Rand,
}

/// A two-argument math operation
#[allow(missing_docs)]
#[derive(Copy, Clone, Debug, Hash, Eq, PartialEq, Ord, PartialOrd)]
pub enum BinaryOpcode {
    Add,
    Sub,
    Mul,
    Div,
    Atan,
    Min,
    Max,
    Compare,
    Mod,
    And,
    Or,
    Mix,
}

impl BinaryOpcode {
    /// Evaluates the opcode
    pub fn eval(&self, a: f32, b: f32) -> f32 {
        match self {
            BinaryOpcode::Add => a + b,
            BinaryOpcode::Sub => a - b,
            BinaryOpcode::Mul => a * b,
            BinaryOpcode::Div => a / b,
            BinaryOpcode::Atan => a.atan2(b),
            BinaryOpcode::Min => a.min_choice(b).0,
            BinaryOpcode::Max => a.max_choice(b).0,
            BinaryOpcode::Compare => a.compare(b),
            BinaryOpcode::Mod => a.rem_euclid(b),
            BinaryOpcode::And => a.and_choice(b).0,
            BinaryOpcode::Or => a.or_choice(b).0,
            BinaryOpcode::Mix => a.mix(b),
        }
    }
}

impl UnaryOpcode {
    /// Evaluates the opcode
    pub fn eval(&self, a: f32) -> f32 {
        match self {
            UnaryOpcode::Neg => -a,
            UnaryOpcode::Abs => a.abs(),
            UnaryOpcode::Recip => 1.0 / a,
            UnaryOpcode::Sqrt => a.sqrt(),
            UnaryOpcode::Square => a * a,
            UnaryOpcode::Floor => a.floor(),
            UnaryOpcode::Ceil => a.ceil(),
            UnaryOpcode::Round => a.round(),
            UnaryOpcode::Sin => a.sin(),
            UnaryOpcode::Cos => a.cos(),
            UnaryOpcode::Tan => a.tan(),
            UnaryOpcode::Asin => a.asin(),
            UnaryOpcode::Acos => a.acos(),
            UnaryOpcode::Atan => a.atan(),
            UnaryOpcode::Exp => a.exp(),
            UnaryOpcode::Ln => a.ln(),
            UnaryOpcode::Not => a.not(),
            UnaryOpcode::Rand => a.rand(),
        }
    }
}

/// An operation in a math expression
///
/// `Op`s should be constructed by calling functions on
/// [`Context`](crate::context::Context), e.g.
/// [`Context::add`](crate::context::Context::add) will generate an
/// `Op::Binary(BinaryOpcode::Add, .., ..)` node and return an opaque handle.
///
/// Each `Op` is tightly coupled to the [`Context`](crate::context::Context)
/// which generated it, and will not be valid for a different `Context`.
#[allow(missing_docs)]
#[derive(Copy, Clone, Debug, Hash, Eq, PartialEq, Ord, PartialOrd)]
pub enum Op {
    Input(Var),
    Const(OrderedFloat<f32>),
    Binary(BinaryOpcode, Node, Node),
    Unary(UnaryOpcode, Node),
    /// Mesh-distance oracle leaf, parameterized by a host-managed asset id
    /// plus three child nodes giving the (x, y, z) world-space coordinates
    /// at which the oracle is sampled. The default lowering for a bare
    /// `Tree::mesh_oracle(id)` plugs in the canonical `Var::X / Y / Z`
    /// input nodes for x/y/z, so the oracle behaves like before; when a
    /// `RemapAxes` ancestor pushes substituted axes during `ctx.import`,
    /// those expressions become the x/y/z children — which is how
    /// `remap(vdbSdf(0), [p.x - 5, ...])` actually shifts the sampled
    /// coordinates instead of silently being a no-op.
    ///
    /// Programs that include this op cannot be JIT-compiled — the
    /// JIT/bytecode backends will refuse to compile, and the host must
    /// fall back to VM eval.
    MeshOracle(u32, Node, Node, Node),
}

fn dot_color_to_rgb(s: &str) -> &'static str {
    match s {
        "red" => "#FF0000",
        "green" => "#00FF00",
        "goldenrod" => "#DAA520",
        "dodgerblue" => "#1E90FF",
        s => panic!("Unknown X11 color '{s}'"),
    }
}

impl Op {
    /// Returns the color to be used in a GraphViz drawing for this node
    pub fn dot_node_color(&self) -> &str {
        match self {
            Op::Const(..) => "green",
            Op::Input(..) | Op::MeshOracle(..) => "red",
            Op::Binary(BinaryOpcode::Min | BinaryOpcode::Max, ..) => {
                "dodgerblue"
            }
            Op::Binary(..) | Op::Unary(..) => "goldenrod",
        }
    }

    /// Returns the shape to be used in a GraphViz drawing for this node
    pub fn dot_node_shape(&self) -> &str {
        match self {
            Op::Const(..) => "oval",
            Op::Input(..) | Op::MeshOracle(..) => "circle",
            Op::Binary(..) | Op::Unary(..) => "box",
        }
    }

    /// Iterates over children, producing 0, 1, 2, or 3 values
    pub fn iter_children(&self) -> impl Iterator<Item = Node> {
        let out = match self {
            Op::Binary(_, a, b) => [Some(*a), Some(*b), None],
            Op::Unary(_, a) => [Some(*a), None, None],
            Op::MeshOracle(_, x, y, z) => [Some(*x), Some(*y), Some(*z)],
            Op::Input(..) | Op::Const(..) => [None, None, None],
        };
        out.into_iter().flatten()
    }

    /// Returns a GraphViz string of edges from this node to its children
    pub fn dot_edges(&self, i: Node) -> String {
        let mut out = String::new();
        for c in self.iter_children() {
            out += &self.dot_edge(i, c, "FF");
        }
        out
    }

    /// Returns a single edge with user-specified transparency
    pub fn dot_edge(&self, a: Node, b: Node, alpha: &str) -> String {
        let color = dot_color_to_rgb(self.dot_node_color()).to_owned() + alpha;
        format!("n{} -> n{} [color = \"{color}\"]\n", a.get(), b.get(),)
    }
}
