package cmd

import (
	"fmt"
	"os"

	"github.com/estecker/tfup/internal/tfup"
	"github.com/spf13/cobra"
)

var verbose bool
var dryRun bool
var directory string
var includePrerelease bool
var sshKeyPath string
var sshKeyPwd string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tfup [directory]",
	Short: "Update Terraform modules and providers",
	Long: `Find the latest version of every module and provider referenced in a
Terraform module and update the version constraints in place. For example:

  tfup .
  tfup --dry-run ./infra`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := directory
		if len(args) > 0 {
			dir = args[0]
		}
		err := tfup.RootCommand(dir, tfup.Options{
			DryRun:            dryRun,
			Verbose:           verbose,
			IncludePrerelease: includePrerelease,
			SSHKeyPath:        sshKeyPath,
			SSHKeyPwd:         sshKeyPwd,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "x", false, "show available updates without modifying files")
	rootCmd.PersistentFlags().StringVarP(&directory, "directory", "d", ".", "directory to check")
	rootCmd.PersistentFlags().BoolVarP(&includePrerelease, "include-prerelease", "e", false, "include prerelease versions")
	rootCmd.PersistentFlags().StringVarP(&sshKeyPath, "ssh-private-key-path", "s", "", "path to SSH private key for git module sources")
	rootCmd.PersistentFlags().StringVarP(&sshKeyPwd, "ssh-private-key-pwd", "w", "", "password for the SSH private key")
}
