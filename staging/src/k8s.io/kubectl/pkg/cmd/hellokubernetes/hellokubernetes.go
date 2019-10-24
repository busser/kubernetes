package hellokubernetes

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	helloKubernetesLong = templates.LongDesc(i18n.T(`
		Print information about a resource, either existing or from a file.`))

	helloKubernetesExample = templates.Examples(i18n.T(`
		# Get the type and name of a resource specified in "foo.yaml".
		kubectl hello-kubernetes -f foo.yaml

		# Get the type, name, and creation time of a pod named 'foo'.
		kubectl hello-kubernetes pod/foo
		
		# Get the type, name, and creation time of all resources with label app=hello.
		kubectl hello-kubernetes all -l app=hello`))
)

// HelloKubernetesOptions contains all command options.
type HelloKubernetesOptions struct {
	FilenameOptions resource.FilenameOptions
	RecordFlags     *genericclioptions.RecordFlags
	PrintFlags      *genericclioptions.PrintFlags
	PrintObj        printers.ResourcePrinterFunc

	Selector string
	All      bool

	Recorder         genericclioptions.Recorder
	builder          *resource.Builder
	namespace        string
	enforceNamespace bool
	args             []string

	genericclioptions.IOStreams
}

// NewHelloKubernetesOptions returns a new instance of HelloKubernetesOptions.
func NewHelloKubernetesOptions(ioStreams genericclioptions.IOStreams) *HelloKubernetesOptions {
	return &HelloKubernetesOptions{
		PrintFlags:  genericclioptions.NewPrintFlags(""),
		RecordFlags: genericclioptions.NewRecordFlags(),
		Recorder:    genericclioptions.NoopRecorder{},
		IOStreams:   ioStreams,
	}
}

// NewCmdHelloKubernetes returns a cobra command with the appropriate
// configuration and flags to run hello-kubernetes.
func NewCmdHelloKubernetes(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewHelloKubernetesOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "hello-kubernetes (-f FILENAME | TYPE NAME)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Print information about a resource"),
		Long:                  helloKubernetesLong,
		Example:               helloKubernetesExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd))
			cmdutil.CheckErr(o.RunHelloKubernetes(cmd))
		},
	}

	o.RecordFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.All, "all", o.All, "Select all resources in the namespace of the specified resource types")
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, "identifying the resource to print information about")

	return cmd
}

// Complete performs completion of command options.
func (o *HelloKubernetesOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error

	o.RecordFlags.Complete(cmd)
	o.Recorder, err = o.RecordFlags.ToRecorder()
	if err != nil {
		return err
	}

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObj = printer.PrintObj

	o.namespace, o.enforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.builder = f.NewBuilder()

	o.args = args

	return nil
}

// Validate performs validation of command options.
func (o *HelloKubernetesOptions) Validate(cmd *cobra.Command) error {
	return nil
}

// RunHelloKubernetes fetches the information.
func (o *HelloKubernetesOptions) RunHelloKubernetes(cmd *cobra.Command) error {
	r := o.builder.
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().
		FilenameParam(o.enforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(o.All, o.args...).
		Flatten().
		LabelSelectorParam(o.Selector).
		Do()
	err := r.Err()
	if err != nil {
		return err
	}

	var infos []*resource.Info
	err = r.Visit(func(info *resource.Info, err error) error {
		if err == nil {
			infos = append(infos, info)
		}
		return nil
	})

	var counter int
	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		metadata := meta.NewAccessor()

		kind, err := metadata.Kind(info.Object)
		if err != nil {
			return err
		}

		name, err := metadata.Name(info.Object)
		if err != nil {
			return err
		}

		obj, err := meta.Accessor(info.Object)
		if err != nil {
			return err
		}

		creationTimestamp := obj.GetCreationTimestamp()

		var zeroTime metav1.Time
		if creationTimestamp == zeroTime { // object not created
			cmd.Printf("Hello %s %s\n", kind, name)
		} else {
			cmd.Printf("Hello %s %s %v\n", kind, name, creationTimestamp)
		}

		counter++

		return nil
	})
	if err != nil {
		return err
	}
	if counter == 0 {
		return fmt.Errorf("no objects passed to kubernetes-hello")
	}

	return nil
}
