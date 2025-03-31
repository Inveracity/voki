package cli

import (
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v8"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/inveracity/voki/internal/client"
	"github.com/inveracity/voki/internal/targets"
	"github.com/inveracity/voki/internal/targets/inline"
)

var (
	user     string
	parallel int
)

type CmdRun struct {
	Client *client.Client
}

func (h *CmdRun) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "run",
		Short:         "run a voki specification",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				log.Fatalln("expected 1 or more arguments")
			}

			// There should always be at least one worker running
			// If -p 0 or -p 1 is passed, we run in serial mode
			h.Client.Parallel = true
			if parallel < 1 || parallel == 1 {
				parallel = 1
				h.Client.Parallel = false
			}

			// Only progressbars will print when running in parallel mode
			h.Client.Printer.Silent = h.Client.Parallel

			h.Client.Bar = mpb.New(
				mpb.WithOutput(color.Output),
				mpb.WithAutoRefresh(),
			)

			// Start the worker(s)
			targetfiles := make(chan string, len(args))
			results := make(chan int, len(args))

			for range parallel {
				go worker(h.Client, user, targetfiles, results)
			}

			// Send the target files to the workers
			for _, arg := range args {
				targetfiles <- arg
			}
			close(targetfiles)

			// Wait for the workers to finish
			for range args {
				<-results
			}

			h.Client.Bar.Wait()

			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&user, "user", "u", "", "user")
	cmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "number of parallel runs")
	viper.BindPFlag("user", cmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("parallel", cmd.PersistentFlags().Lookup("parallel"))
	return cmd
}

func worker(client *client.Client, user string, targetfiles <-chan string, results chan<- int) {
	for targetfile := range targetfiles {
		content, err := os.ReadFile(targetfile)
		if err != nil {
			log.Fatalln(err)
		}

		// If there are variables in the target configuration, load those to add to ctx
		variables, err := targets.LoadVars(content)
		if err != nil {
			log.Fatalln(err)
		}

		client.EvalContext = &hcl.EvalContext{
			Functions: map[string]function.Function{
				"file":     inline.FileFunc,
				"template": inline.TemplateFunc,
				"vault":    inline.VaultFunc,
			},
			Variables: variables,
		}

		client.Run(string(content), user)
		results <- 0
	}
}
