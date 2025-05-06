package project

// Type represents the detected project type.
type Type string

const (
	Terraform Type = "Terraform"
	Pulumi    Type = "Pulumi"
	Go        Type = "Go"
	Python    Type = "Python"
	Unknown   Type = "Unknown"
)
