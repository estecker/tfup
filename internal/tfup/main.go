package tfup

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/estecker/tfup/internal/tfvc/checker"
	"github.com/estecker/tfup/internal/tfvc/output"
	"github.com/minamijoyo/tfupdate/tfregistry"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/spf13/afero"
)

// gitRefRegexp matches the version ref suffix that tfupdate rewrites on git module sources.
var gitRefRegexp = regexp.MustCompile(`\?ref=v[0-9]+(\.[0-9]+)*(-.*)*$`)

// Options controls how tfup checks and updates Terraform code.
type Options struct {
	DryRun            bool
	Verbose           bool
	IncludePrerelease bool
	SSHKeyPath        string
	SSHKeyPwd         string
}

// RootCommand finds the latest versions of all modules and providers in
// directory and rewrites the Terraform code to use them.
func RootCommand(directory string, opts Options) error {
	absPath, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("getting absolute path for %s: %w", directory, err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("directory %s does not exist: %w", directory, err)
	}
	fmt.Println("Checking for updates in", absPath)

	// tfupdate logs via the standard logger; only show it in verbose mode.
	if !opts.Verbose {
		log.SetOutput(io.Discard)
	}

	// 1. Find modules/providers and their latest versions (tfvc).
	results, err := checker.Main(absPath, opts.IncludePrerelease, opts.SSHKeyPath, opts.SSHKeyPwd)
	if err != nil {
		return fmt.Errorf("checking versions: %w", err)
	}

	// 2. Rewrite the Terraform code (tfupdate).
	// tfupdate's tfconfig/afero bridge does not support absolute paths,
	// so root the filesystem at the module directory and operate on ".".
	fs := afero.NewBasePathFs(afero.NewOsFs(), absPath)
	ctx := context.Background()

	var failed int
	for _, r := range results {
		latest := r.LatestOverall.String()
		if latest == "" || !needsUpdate(r) {
			if opts.Verbose {
				fmt.Printf("  %s %s: up to date\n", r.Type, r.Name)
			}
			continue
		}

		var updateType, name string
		switch r.Type {
		case "module":
			// tfupdate matches git modules by source without the "?ref=vX.Y.Z" suffix.
			updateType, name = "module", gitRefRegexp.ReplaceAllString(r.Source, "")
		case "provider":
			updateType, name = "provider", r.Name
		default:
			continue
		}

		fmt.Printf("  %s %s: %s -> %s\n", updateType, name, r.VersionConstraints, latest)
		if opts.DryRun {
			continue
		}
		if err := apply(ctx, fs, updateType, name, latest); err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "  failed to update %s %s: %v\n", updateType, name, err)
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d update(s) failed", failed)
	}
	return nil
}

// needsUpdate reports whether a newer version is available. tfvc's Status
// alone is not enough: it reports git modules pinned to "?ref=vX.Y.Z" as
// PASSED even when a newer tag exists.
func needsUpdate(r output.Update) bool {
	if r.Status != "PASSED" {
		return true
	}
	return r.Version.String() != "" && r.LatestOverall.GreaterThan(&r.Version)
}

func apply(ctx context.Context, fs afero.Fs, updateType, name, version string) error {
	opt, err := tfupdate.NewOption(updateType, name, version, nil, false, nil, "full", tfregistry.Config{})
	if err != nil {
		return err
	}
	gc, err := tfupdate.NewGlobalContext(fs, opt)
	if err != nil {
		return err
	}
	return tfupdate.UpdateFileOrDir(ctx, gc, ".")
}
