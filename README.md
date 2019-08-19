# An (unofficial) Github command line client

This command line client is based on Github API V3.

## What it does

You can manipulate following Github resources:

- deployment
- deployment_status
- release

## What it will (probably) never do

Managing resources like pull requests, issues or repositories life cycle and settings.

Some very good tools like [hub](https://github.com/github/hub) or [terraform's github provider](https://www.terraform.io/docs/providers/github/index.html) are already great at doing that.

## Why another Github client?

The goal is to have a convenient, lightweight tool to use inside github [actions v2](https://github.com/features/actions) workflows.

Some use cases that motivated the creation of this tool were:

```shell
# Append some deploy button inside a release note
printf "%s\nDEPLOY_BUTTON_CODE" $(github release get ID | jq .body) | \
github release edit ID --body -

# Create a production deployment and corresponding status
DEPLOYMENT_ID=$(
  github deployment create \
  --environment production \
  --task deploy:migration \
  $GITHUB_REF
)
github deployment_status create $DEPLOYMENT_ID in_progress
```
