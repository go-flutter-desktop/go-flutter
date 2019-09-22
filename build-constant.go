package flutter

// Compile configuration constants persistent across all flutter.Application.
// The values of config(option.go) can change between flutter.Run calls, those
// values contains informations that needs to be access globally, without
// requiring an flutter.Application.
//
// Values overwritten by hover during the 'Compiling 'go-flutter' and
// plugins' phase.
var (
	// ProjectVersion contains the version of the build
	ProjectVersion = "unknown"
	// ProjectVersion contains the version of the go-flutter been used
	PlatformVersion = "unknown"
	// ProjectName contains the application name
	ProjectName = "unknown"
	// ProjectOrganizationName contains the package org name, (Can by set upon flutter create (--org flag))
	ProjectOrganizationName = "unknown"
)
