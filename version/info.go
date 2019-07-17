package version

import (
	"fmt"
	"runtime"
	"strings"
)

// The git commit that was compiled. This will be filled in by the compiler.
var (
	GitCommit   string
	GitDescribe string

	// Version is main version number that is being run at the moment.
	Version = "1.2.1"
)

const (
	// ProductName is the name of the product
	ProductName = "DapperDox"

	// ShortName is a short condensed name of the product
	ShortName = "dapperdox"
)

// GetVersionDisplay composes the parts of the version in a way that's suitable
// for displaying to humans.
func GetVersionDisplay() string {
	return fmt.Sprintf("%s\n version\t%s\n Git commit\t%s\n Go Version\t%s\n OS/Arch\t%s\n",
		ProductName, getHumanVersion(), GitCommit, runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)
}

func getHumanVersion() string {
	version := Version
	if GitDescribe != "" {
		version = GitDescribe
	}

	// Strip off any single quotes added by the git information.
	return strings.ReplaceAll(version, "'", "")
}
