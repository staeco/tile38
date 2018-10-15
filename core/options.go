package core

// DevMode puts application in to dev mode
var DevMode = false

// ShowDebugMessages allows for log.Debug to print to console.
var ShowDebugMessages = false

// ProtectedMode forces Tile38 to default in protected mode.
var ProtectedMode = "yes"

// AppendOnly allows for disabling the appendonly file.
var AppendOnly = "yes"

// AppendFileName allows for custom appendonly file path
var AppendFileName string

// QueueFileName allows for custom queue.db file path
var QueueFileName string
