package helloworld

import (
	"github.com/spf13/cobra"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var helloWorldLong = templates.LongDesc(i18n.T(`
	Print "Hello World".`))

// NewCmdHelloWorld returns the hello-world Cobra command
func NewCmdHelloWorld() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "hello-world",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T(`Print "Hello World".`),
		Long:                  helloWorldLong,

		Run: RunHelloWorld,
	}

	return cmd
}

// RunHelloWorld checks given arguments and executes command
func RunHelloWorld(cmd *cobra.Command, args []string) {
	cmd.Println("Hello World")
}
