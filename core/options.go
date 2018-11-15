package core

// DevMode puts application in to dev mode
var DevMode = false

// ShowDebugMessages allows for log.Debug to print to console.
var ShowDebugMessages = false

// ProtectedMode forces Tile38 to default in protected mode.
var ProtectedMode = "no"

// AppendOnly allows for disabling the appendonly file.
var AppendOnly = true

// AppendFileName allows for custom appendonly file path
var AppendFileName = ""

// QueueFileName allows for custom queue.db file path
var QueueFileName = ""

// NumThreads is the number of network threads to use.
var NumThreads int

// Evio set the networking to use the evio package.
var Evio = false
