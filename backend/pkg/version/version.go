package version

import "fmt"

var (
	Version   string
	GoVersion string
	Commit    string
	BuildTime string
	CodeMode  string
)

type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"goVersion"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	CodeMode  string `json:"codeMode"`
}

func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   Version,
		GoVersion: GoVersion,
		Commit:    Commit,
		BuildTime: BuildTime,
		CodeMode:  CodeMode,
	}
}

func init() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("BuildTime: %s\n", BuildTime)
	fmt.Printf("GoVersion: %s\n", GoVersion)
	fmt.Printf("CodeMode: %s\n", CodeMode)
}
