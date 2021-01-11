package constants

// PaddleService Key
const (
	PaddleService               = "paddleService"
	PaddleServiceDefaultPodName = "http1"
)

// Ingress Key
const (
	Ingress = "ingress"
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
	PaddleServiceDefaultMinScale = 0 // 0 if scale-to-zero is desired
	PaddleServiceDefaultMaxScale = 0 // 0 means limitless
)

func DefaultServiceName(name string) string {
	return name + "-service"
}
