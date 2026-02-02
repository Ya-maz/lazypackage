package msg

// Msg types
type NodeLine string // Still useful for line-based updates
type NodeData string // Useful for direct PTY data chunks
type NodeDone struct{}
type NodeErr string

