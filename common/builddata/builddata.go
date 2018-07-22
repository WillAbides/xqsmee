//Package builddata contains read-only data about the current build set with ldflags
package builddata

var (
	commit  string
	date    string
	version string
)

//Version is semver formatted version
func Version() string {
	return version
}

//Commit is the git commit of the build
func Commit() string {
	return commit
}

//Date is the date of the build in RFC3339 format
func Date() string {
	return date
}
