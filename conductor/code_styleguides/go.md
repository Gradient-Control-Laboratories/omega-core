# Effective Go Style Guide Summary

This document summarizes key rules and best practices from the official
"Effective Go" guide for writing idiomatic Go code.

## 1. Formatting

-   **`gofmt`:** All Go code **must** be formatted with `gofmt` (or `go fmt`).
    This is a non-negotiable, automated standard.
-   **Indentation:** Use tabs for indentation (`gofmt` handles this).
-   **Line Length:** Go has no strict line length limit. Let `gofmt` handle line
    wrapping.

## 2. Naming

-   **`MixedCaps`:** Use `MixedCaps` or `mixedCaps` for multi-word names. Do not
    use underscores.
-   **Exported vs. Unexported:** Names starting with an uppercase letter are
    exported (public). Names starting with a lowercase letter are not exported
    (private).
-   **Package Names:** Short, concise, single-word, lowercase names.
-   **Getters:** Do not name getters with a `Get` prefix. A getter for a field
    named `owner` should be named `Owner()`.
-   **Interface Names:** One-method interfaces are named by the method name plus
    an `-er` suffix (e.g., `Reader`, `Writer`).

## 3. Control Structures

-   **`if`:** No parentheses around the condition. Braces are mandatory. Can
    include an initialization statement (e.g., `if err := file.Chmod(0664); err
    != nil`).
-   **`for`:** Go's only looping construct. Unifies `for` and `while`. Use
    `for...range` to iterate over slices, maps, strings, and channels.
-   **`switch`:** More general than in C. Cases do not fall through by default
    (use `fallthrough` explicitly). Can be used without an expression to
    function as a cleaner `if-else-if` chain.

## 4. Functions

-   **Multiple Returns:** Functions can return multiple values. This is the
    standard way to return a result and an error (e.g., `value, err`).
-   **Named Result Parameters:** Return parameters can be named. This can make
    code clearer and more concise.
-   **`defer`:** Schedules a function call to be run immediately before the
    function executing `defer` returns. Use it for cleanup tasks like closing
    files.

## 5. Data

-   **`new` vs. `make`:**
    -   `new(T)`: Allocates memory for a new item of type `T`, zeroes it, and
        returns a pointer (`*T`).
    -   `make(T, ...)`: Creates and initializes slices, maps, and channels only.
        Returns an initialized value of type `T` (not a pointer).
-   **Slices:** The preferred way to work with sequences. They are more flexible
    than arrays.
-   **Maps:** Use the "comma ok" idiom to check for the existence of a key:
    `value, ok := myMap[key]`.

## 6. Interfaces

-   **Implicit Implementation:** A type implements an interface by implementing
    its methods. No `implements` keyword is needed.
-   **Small Interfaces:** Prefer many small interfaces over one large one. The
    standard library is full of single-method interfaces (e.g., `io.Reader`).

## 7. Concurrency & Synchronization

-   **Share Memory By Communicating:** This is the core philosophy. Do not
    communicate by sharing memory; instead, share memory by communicating.
-   **Goroutines:** Lightweight, concurrently executing functions. Start one
    with the `go` keyword.
-   **Channels:** Typed conduits for communication between goroutines. Use
    `make` to create them.
-   **Struct Field Layout & Synchronization:** Struct fields MUST be ordered by lock requirements:
    1.  Read-only/immutable configuration fields set at construction MUST be listed first (above
        `sync.Mutex`), preceded by a comment (e.g.,
        `// Read-only immutable fields set at construction`).
    2.  The mutex (`mu sync.Mutex`) MUST be listed next, preceded by a comment (e.g.,
        `// Lock for accessing the fields below`).
    3.  Fields that require `mu` to be locked before access MUST be listed immediately after `mu`.

## 8. Errors

-   **`error` type:** The built-in `error` interface is the standard way to
    handle errors.
-   **Explicit Error Handling:** Do not discard errors with the blank identifier
    (`_`). Check for errors explicitly.
-   **`panic`:** Reserved for truly exceptional, unrecoverable situations.
    Generally, libraries should not panic.

## 9. Context Lifecycles & Graceful Shutdown

-   **Root Contexts:** Agent entry points (`main.go`) must create a cancellable
    root context via `ctx, cancel := context.WithCancel(context.Background())`
    and `defer cancel()` after calling ParseFlags.
    Pass `ctx` to client connections (`app.Connect`, `dispatcher.Connect`, `storage.Connect`).
-   **No Redundant Teardown:** Do NOT use `defer appClient.Disconnect()`.
    Client stream lifecycles are bound to `<-ctx.Done()`.
-   **Graceful Server Shutdown:** When stopping HTTP or gRPC servers on context
    cancellation, use a bounded timeout context
    (e.g., `context.WithTimeout(context.Background(), 5*time.Second)`) rather than
    an unbounded `context.Background()`.

## 10. Agent Architecture & In-Process Start Conventions

-   **Package Structure & Visibility:**
    -   `agent/<name>/internal/config`: Unexported package providing `Config`
        struct and `ParseFlags() (*Config, error)` for `run/main.go`.
    -   `agent/<name>/agent`: Public package exposing `start.go` with
        `Start(ctx context.Context, appClient *app.Client, ...params) chan error`
        or `Start(...) (error, chan error)`.
    -   `agent/<name>/lib`: Optional public package exposing pure domain functions
        and types independent of the agent framework or dispatcher transport.
-   **Configuration Decoupling:** The `agent` package maintains zero dependencies
    on `internal/config`. `run/main.go` unpacks configuration values from `cfg`
    and passes them individually to `agent.Start`.
-   **Caller Responsibility for Passed-In Collaborators:** The caller (`run/main.go`
    or test harness) manages the lifecycle and external listeners of passed-in
    collaborators (e.g., `appClient.ListenForExit()`, `srcMgr.StartListener()`).
    `agent.Start` registers only the agent's own event handling paths.
-   **Error Channel Consolidation:** Agent-managed event listeners and background routine
    error channels must be consolidated inside `Start` and merged into a single exit stream
    using `event.CombineErrChan` (or returned directly if single).
-   **No Test-Induced Helpers:** Do not create test-only flag parsing functions
    (such as `parseFlagSet`). Tests in `config_test.go` configure `flag.CommandLine`
    and `os.Args` directly.

*Source: [Effective Go](https://go.dev/doc/effective_go)*
