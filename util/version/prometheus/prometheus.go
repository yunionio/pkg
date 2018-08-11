// Package prometheus registers version information as a
// prometheus metric.
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"yunion.io/x/pkg/util/version"
)

func init() {
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yunionyun_build_info",
			Help: "A metric with a constant '1' value labeled by major, minor, git version, git commit, git tree state, build date, Go version, and compiler from which Kubernetes was built, and platform on which it is running.",
		},
		[]string{"major", "minor", "gitVersion", "gitCommit", "gitTreeState", "buildDate", "goVersion", "compiler", "platform"},
	)
	info := version.Get()
	buildInfo.WithLabelValues(info.Major, info.Minor, info.GitVersion, info.GitCommit, info.GitTreeState, info.BuildDate, info.GoVersion, info.Compiler, info.Platform).Set(1)

	prometheus.MustRegister(buildInfo)
}
