# Github Integration

## Overview

**GitHub** integration for all GitHub's data available through GitHub's APIs.

Provided connectors:
- Importer

*Note: when creating a GitHub token for this integration,
make sure it has READ rights for all data you want to backup.*

## Configuration

- `token` (**required**): The GitHub token that will be used to authenticate to GitHub's API.
- `orgs`: Comma separated organizations to backup, use the login value (the value in github's URL for the org page). **Default**: lists all orgs.

## Examples

```sh
# Create a github configuration
$ plakar source add my-github location=gh:// orgs=org1,org2 token=$GITHUB_TOKEN
# Backup my github
$ plakar backup "@my-github"
```