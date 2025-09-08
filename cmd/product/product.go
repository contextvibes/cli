package product

import (
	"github.com/contextvibes/cli/cmd/product/build"
	"github.com/contextvibes/cli/cmd/product/clean"
	"github.com/contextvibes/cli/cmd/product/codemod"
	"github.com/contextvibes/cli/cmd/product/format"
	"github.com/contextvibes/cli/cmd/product/quality"
	"github.com/contextvibes/cli/cmd/product/run"
	"github.com/contextvibes/cli/cmd/product/test"
	"github.com/spf13/cobra"
)

var ProductCmd = &cobra.Command{
	Use:   "product",
	Short: "Commands for the product you are building (the 'what').",
}

func init() {
	ProductCmd.AddCommand(build.BuildCmd)
	ProductCmd.AddCommand(test.TestCmd)
	ProductCmd.AddCommand(quality.QualityCmd)
	ProductCmd.AddCommand(format.FormatCmd)
	ProductCmd.AddCommand(clean.CleanCmd)
	ProductCmd.AddCommand(run.RunCmd)
	ProductCmd.AddCommand(codemod.CodemodCmd)
}
