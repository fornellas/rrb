module github.com/fornellas/rrb

go 1.27.3

require (
	al.essio.dev/pkg/shellescape v1.6.1
	github.com/bmatcuk/doublestar/v4 v4.10.4
	github.com/client9/misspell v0.3.4
	github.com/fatih/color v1.19.0
	github.com/fsnotify/fsnotify v1.10.1
	github.com/fzipp/gocyclo v0.6.0
	github.com/prometheus/procfs v0.23.0
	github.com/rakyll/gotest v0.0.7
	github.com/sirupsen/logrus v1.10.2
	github.com/spf13/cobra v1.10.2
	github.com/williammartin/subreaper v0.0.0-20181101193406-731d9ece6883
	golang.org/x/tools v0.52.0
	honnef.co/go/tools v0.8.2
)

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/gordonklaus/ineffassign v0.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jandelgado/gcov2lcov v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.24.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/exp/typeparams v0.0.0-20261014143641-f0251eb4aa3e // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/telemetry v0.0.0-20261021172232-6cd571d43fbb // indirect
	golang.org/x/vuln v1.9.0 // indirect
)

tool (
	github.com/client9/misspell/cmd/misspell
	github.com/fornellas/rrb
	github.com/fzipp/gocyclo/cmd/gocyclo
	github.com/gordonklaus/ineffassign
	github.com/jandelgado/gcov2lcov
	github.com/rakyll/gotest
	golang.org/x/tools/cmd/goimports
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
)
