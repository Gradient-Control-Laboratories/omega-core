# GSL Language Specification Changes: v1 to v2

This document summarizes the changes, additions, deprecations, and design decisions between the **GSL v1 Language Specification** ([gsl_lang_spec_1.md](gsl_lang_spec_1.md)) and the **GSL v2 Language Specification** ([gsl_lang_spec_2.md](gsl_lang_spec_2.md)).

---

## 1. Syntax & Grammar Changes

### Introduction of Annotations
GSL v2 introduces a compile-time-only annotation syntax (`@annotationName` or `@annotationName(...)`), replacing dot-property accessors for sliders and wrapper functions for viewport manipulators.

* **Overview & Semantics**:
  - Annotations precede top-level variable declarations, named functions, or parameters in a function declaration.
  - Parsed and stored as metadata in the AST, but ignored by code generation backends.
  - By default, each type of annotation may only be specified at most once per target declaration unless explicitly allowed by the annotation definition.
  - May define specific restrictions (e.g., requiring arguments to evaluate to compile-time constants or restrictions on target variable types/values).
* **Grammar Addition**:
  ```ebnf
  Declaration        = Annotation* ( NamedFunction | PropertyAssignment | Assignment )
  PropertyAssignment = Ident "." Ident "=" Expression ";"
  Annotation         = "@" Ident ( "(" ( Expression ( "," Expression )* ","? )? ")" )?
  Parameters         = "(" ( Parameter ( "," Parameter )* )? ")" "{"
  Parameter          = Annotation* Ident
  ```

### Standardized Slider & Discrete Properties (`@range` & `@values`)
* **GSL v1**: Configured via dot-property assignments on top-level slider variables:
  ```gsl
  radius = 5.00;
  radius.min = 0.5;
  radius.max = 10.0;
  radius.period = 4.0;
  ```
* **GSL v2**:
  - Replaced slider dot-properties with the `@range(min, max, [period])` annotation:
    ```gsl
    @range(0.5, 10.0, 4.0)
    radius = 5.00;
    ```
  - Introduced the `@values(v1, v2, ..., vn)` annotation for discrete scalar choices:
    ```gsl
    @values(1.0, 2.0, 4.0, 8.0)
    stepSize = 2.0;
    ```
  - Removed support for `.min`, `.max`, and `.period` property assignments.
  - Arguments must be scalar expressions that evaluate to compile-time numeric constants.
  - `@values` requires at least 2 value parameters.
  - Violating these conditions raises a compile error.

### Viewport Manipulator Annotations (`@point` & `@transform`)
* **GSL v1**: Used wrapper functions `uiPoint(...)` and `uiTransform(...)`:
  ```gsl
  pos = uiPoint([1.0, 2.0, 3.0]);
  m = uiTransform(matrix.translation([1.0, 2.0, 3.0]));
  ```
* **GSL v2**: Supported via annotations placed directly above top-level variable declarations
  or before inline expressions:
  ```gsl
  @point
  pos = [1.0, 2.0, 3.0];

  @transform
  m = matrix.translation([1.0, 2.0, 3.0]);
  ```
  - **Backward-compatible parser translation**: The `parser` agent parses `@point`
    and `@transform` annotations and converts them internally into existing `uiPoint(...)`
    and `uiTransform(...)` AST call node model representations.
  - Target variable/expression for `@point` must evaluate to a 3-component list.
  - Target variable/expression for `@transform` must evaluate to a 4x4 matrix.
  - Model representation (`common/model` and `common/program`)
    and backend engines (`wgsl`, `fidget`) remain unchanged.
  - Direct function call expressions `uiPoint(v)` and `uiTransform(m)`
    in GSL v1 remain fully supported.

### Trailing Commas (`","?`)
- GSL v2 permits optional trailing commas (`","?`) across all comma-separated multi-item lists,
  including function parameters, call argument lists, return expressions,
  variable assignment lists, and collection/list literals (`[...]`).

---

## 2. Terminology & Type System Updates

### Terminology Alignment
* **"Vector" / "Collection" $\rightarrow$ "List"**: Replaced the terms "vector" and "collection" with **"list"** throughout the specification to accurately describe ordered scalar component sequences.
* **Matrix Structure**: Explicitly defines a **matrix** as a list of 2 or more column lists (column-major order), where each inner list represents a column vector.

### Explicit Section `4. Types` & Type Promotion
* **Type Promotion**: Explicitly documents automatic type promotion rules:
  - Scalars promote to lists and matrices as needed (e.g. `3` $\rightarrow$ `[3, 3, 3]`).
  - Lists promote to matrices as needed (e.g. `[1, 2, 3]` $\rightarrow$ `[[1, 2, 3], [1, 2, 3], [1, 2, 3], [1, 2, 3]]`).
* **Matrix Indexing**: Explicitly documents matrix indexing (e.g. `m[0]` for the first column vector, `m[0][1]` for the second element of the first column) alongside list zero-indexing (`v[0]`).

---

## 3. Removed Language Constructs & Compiler Error Diagnostics

* **Removed Constructs**:
  - `uiPoint`: Replaced by the `@point` annotation and removed in GSL v2.
  - `uiTransform`: Replaced by the `@transform` annotation and removed in GSL v2.
  - `min`, `max`, and `period` property assignments on sliders (e.g., `s.min = 0.5;`): Replaced by the `@range(...)` annotation and removed in GSL v2. (Note: `min`/`max` or similar properties may be added as properties if decided to be useful in the future).
* **Compiler Error Diagnostic Guidance**:
  - For language elements of GSL v1 that are removed in GSL v2, the compiler should report a compile error along with a suggestion to change the language element to its GSL v2 equivalent.

---

## 4. Language Semantics & Compiler Behaviors

### Ternary Operator Associativity & Nesting
* **GSL v1**: Inline nesting of unparenthesized ternary expressions (e.g. `a ? b : c ? d : e`) was rejected by the compiler.
* **GSL v2**: Unparenthesized nested ternaries are now supported. The parser generates right-associative ASTs, and the compiler executes an AST rotation pass (`fixTernaryAssociativity`) to correct stolen branches to their parent scope.

### Expression Recursive Depth Limit
* **GSL v1**: Hard-coded parsing recursive depth limit of 127 expressions.
* **GSL v2**: Recognizes the 127-expression limit as an implementation constraint rather than a language specification limit. GSL v2 relies on compiler/parser capabilities for arbitrary depth expression trees.

### Standardized `reduce` Signature & Matrix Indexing
* **Standardized `reduce` Signature**: GSL v1 placed `initial` first (`reduce(initial, fn, collection)`). GSL v2 standardizes `reduce` to always place the callback function `fn` first (`reduce(fn, numeric [, initial])`), making `initial` an optional 3rd parameter.
* **Matrix Element Indexing**: When `map` or `reduce` is invoked over a matrix, the `index` parameter passed to the callback function is a 2-element list `[col, row]` (representing 0-indexed column and row numbers), whereas list operations pass a scalar integer (`0.0`, `1.0`, etc.).
* **`length(matrix)` Type Checking**: Calling `length(v)` on a matrix is explicitly prohibited and produces a compile-time error.
* **Globally Reserved Identifiers**: `length` and `auto` are added to the reserved keywords list.
  Implicit globals `p` and `time` are globally reserved identifiers.
  Annotation names exist in their own `@`-prefixed namespace and are not reserved words.

### Formalized Property Assignment Validation & Lifecycle Semantics
* **Scope & Pairing Enforcement**: Slider properties (`.min`, `.max`, `.period`) are restricted
  strictly to top-level file scope scalar numeric variables.
  A variable declaring any slider property must declare both `.min` and `.max`, where `min < max`.
  `.period` must be strictly positive (`period > 0`).
* **Metadata Lifecycle & Expressions**: Properties are declaration metadata and cannot be read
  in general expressions (except `.box` which queries the bounding box).
* **Bounding Box Validation & Keyword `auto`**: `.box` requires 3D corner coordinates `[lo, hi]`
  where constant extents satisfy `lo < hi` across all axes, without dependency on `p`.
  The keyword `auto` may be assigned or used within coordinates to infer analytical bounds.
* **Transitive DAG Cycle Detection**: Circular dependencies among declarations and properties
  are strictly forbidden and detected at compile-time across the package AST.
