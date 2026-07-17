# Github Integration

## Overview

**GitHub** integration for all GitHub's data available through GitHub's APIs.

Provided connectors:
- Importer

*Note: when creating a GitHub token for this integration,
make sure it has READ rights for all data you want to backup.*

## Configuration

- `token` (required) The GitHub token that will be used to authenticate to GitHub's API.

## Examples

```sh
# Create a github configuration
$ plakar source add my-github location=gh:// token=$GITHUB_TOKEN
# Backup my github
$ plakar backup "@my-github"
```