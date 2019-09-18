package flutter

// Compile configuration constants persistent across all flutter.Application.
// The values of config(option.go) can change between flutter.Run calls, this
// struct contains informations that needs to be access globally, without
// requiring an flutter.Application.
//
// Those value may be setted by hover during the 'Compiling 'go-flutter' and plugins' phase.
type buildConstant struct {
	projectVersion   string
	projectName      string
	goFlutterVersion string
	organizationName string
}

var buildConfig = buildConstant{}

// SetProjectVersion a string used as the project projectVersion number.
//
// This value is setted by hover with the name attribute available in the
// pubspec.yaml file.
func SetProjectVersion(projectVersion string) {
	buildConfig.projectVersion = projectVersion
}

// ProjectVersion return the project version number.
func ProjectVersion() string {
	return buildConfig.projectVersion
}

// AppName return the project projectVersion number.
func AppName() string {
	return buildConfig.projectName
}

// SetAppName a string used as the project app name.
//
// This value is setted by hover with the name attribute available in the
// pubspec.yaml file.
func SetAppName(projectName string) {
	buildConfig.projectName = projectName
}

// PlatformVersion return the go-flutter projectVersion number.
func PlatformVersion() string {
	return buildConfig.goFlutterVersion
}

// SetPlatformVersion a string used as the go-flutter projectVersion number.
//
// This value is setted by hover with the version attribute available in the
// pubspec.yaml file.
func SetPlatformVersion(version string) {
	buildConfig.goFlutterVersion = version
}

// OrganizationName return the project Organization name, (Using hover, it's
// equal the flag org value `flutter create --org tld.domain`).
// Default to 'com.example'
func OrganizationName() string {
	return buildConfig.organizationName
}

// SetOrganizationName a string used to represent the organization responsible
// for your Flutter project, in reverse domain name notation.
//
// This value is setted by hover with the xml attribute /manifest[@package]
// available in the android/app/src/main/AndroidManifest.xml file.
// Default to 'com.example'
func SetOrganizationName(name string) {
	buildConfig.organizationName = name
}
