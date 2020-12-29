package constants

// PaddleService Key
const (
	PaddleService = "paddleService"
)

// PaddleService configuration name and namespce
var (
	PaddleServiceConfigName      = "paddleservice-config"
	PaddleServiceConfigNamespace = "paddleservice-system"
)

// PaddleService resource defaults
var (
	PaddleServiceDefaultCPU      = "0.1"
	PaddleServiceDefaultMemory   = "128Mi"
	PaddleServiceDefaultMinScale = 1 // 0 if scale-to-zero is desired
	PaddleServiceDefaultMaxScale = 0 // 0 means limitless
)
