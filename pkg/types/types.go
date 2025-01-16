package types

// Resource represents a network resource
type Resource struct {
	URL         string
	ContentType string
	Data        []byte
}

// DynamicContent represents dynamically loaded content
type DynamicContent struct {
	HTML      string
	Resources []Resource
}
