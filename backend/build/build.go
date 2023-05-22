package build

import "runtime/debug"

// version will be set via flag at build time
var version = "dev"

func GetVersion() string {
	return version
}

func GetRevision() string {
	info, _ := debug.ReadBuildInfo()

	for _, kv := range info.Settings {
		if kv.Key == "vcs.revision" {
			return kv.Value
		}
	}

	return ""
}
