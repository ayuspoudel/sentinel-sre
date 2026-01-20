package version

// These variables are set at build time using -ldflags.
// Do not assign defaults here.
/*
go build \
  -ldflags "\
    -X github.com/ayuspoudel/sentinel-sre/controlplane/api-server/version.Version=0.1.0 \
    -X github.com/ayuspoudel/sentinel-sre/controlplane/api-server/version.GitCommit=$(git rev-parse HEAD) \
    -X github.com/ayuspoudel/sentinel-sre/controlplane/api-server/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  ./cmd/sentinel-api-server
*/
var (
	Version   string
	GitCommit string
	BuildDate string
)

// Info represents Sentinel API server build metadata.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
}

// Get returns build information in a structured form.
func Get() Info {
	return Info{
		Version:   valueOrUnknown(Version),
		GitCommit: valueOrUnknown(GitCommit),
		BuildDate: valueOrUnknown(BuildDate),
	}
}

func valueOrUnknown(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}
