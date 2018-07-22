package cmd

import (
	"fmt"

	"github.com/WillAbides/xqsmee/common/builddata"
)

const versionTemplate = `version: %v
built from commit: %v
built at: %v
`

type versionCmd struct{}

func (*versionCmd) Run() error {
	_, err := fmt.Printf(versionTemplate, builddata.Version(), builddata.Commit(), builddata.Date())
	return err
}
