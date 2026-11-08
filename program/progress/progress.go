package progress

// Callback reports compilation, assembly, or backend progress milestones.
type Callback func(stage string, completed, total uint32)

