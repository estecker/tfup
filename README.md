# tfup

Find the latest version of every module and provider referenced in a Terraform
module and update the version constraints in place.

tfup combines [tfvc](https://github.com/tfverch/tfvc) (discover modules/providers
and their latest versions) and [tfupdate](https://github.com/minamijoyo/tfupdate)
(rewrite the Terraform code) into a single binary. tfvc's checker is vendored
under `internal/tfvc` (MIT, see `internal/tfvc/LICENSE`); tfupdate is a regular
Go module dependency.

## Install

```bash
go install github.com/estecker/tfup@latest
```

Or build from a checkout:

```bash
go build -o tfup .
```

## Usage

```bash
tfup path/to/terraform/module            # update in place
tfup --dry-run path/to/terraform/module  # only list available updates
```

| Flag                           | Description                                    |
| ------------------------------ | ---------------------------------------------- |
| `--directory`, `-d`            | Directory to check (default `.`)               |
| `--dry-run`, `-x`              | Show available updates without modifying files |
| `--verbose`, `-v`              | Verbose output                                 |
| `--include-prerelease`, `-e`   | Include prerelease versions                    |
| `--ssh-private-key-path`, `-s` | SSH private key for git module sources         |
| `--ssh-private-key-pwd`, `-w`  | Password for the SSH private key               |

## What gets updated

- Providers in `required_providers` (the `version` constraint).
- Registry modules (the `version` attribute).
- Git modules pinned with `?ref=vX.Y.Z`.

Afterwards run `terraform init -upgrade` to refresh `.terraform.lock.hcl`.
