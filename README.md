# An (unofficial) Github command line client

This command line client is based on Github API V3.

## What it does

You can manipulate following Github resources:

- [deployments](https://developer.github.com/v3/repos/deployments/)
- [deployment_status](https://developer.github.com/v3/repos/deployments/#list-deployment-statuses)

## What it will (probably) never do

Managing resources like pull requests, issues or repositories life cycle and
settings for example.

Some very good tools like [hub](https://github.com/github/hub) or
[terraform provider](https://www.terraform.io/docs/providers/github/index.html)
are already great at doing that.

## How to install it?

```shell
go get github.com/inextensodigital/github
```

or you can use the install script:

```shell
curl -qs https://raw.githubusercontent.com/inextensodigital/github/master/install.sh | bash -
```

or a simplified version (feel free to replace `linux` by either `windows` or `darwin`)

```shell
curl -s https://api.github.com/repos/inextensodigital/github/releases/latest |
jq -r '.assets[] | select(.name | test("linux-amd64$")) | .browser_download_url' |
wget -qi - -O github && chmod +x github
```

## Why another Github client?

The goal is to have a convenient, lightweight tool to use inside github
[actions v2](https://github.com/features/actions) workflows.

Some use cases that motivated the creation of this tool were:

### How it can help you in github actions v2?

#### Continuous deployment on staging

```yaml
name: trigger staging deployment on pull request merged
on:
  push:
    branches: [master, dev]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: echo build app
      - run: curl -qs https://raw.githubusercontent.com/inextensodigital/github/master/install.sh | bash -
      - id: deployment
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          id=$(github deployment create --productionEnvironment=false --environment staging $GITHUB_REF)
          github deployment_status create $id in_progress
          echo "##[set-output name=deployment_id;]$id"
      - env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          ID: ${{ steps.deployment.outputs.deployment_id }}
        run: |
          echo deploy app
          github deployment_status create $ID success
      - if: failure()
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          ID: ${{ steps.deployment.outputs.deployment_id }}
        run: github deployment_status create $ID failure
```

#### Continuous delivery on production

```yaml
name: create production deployment on release published
on: release
jobs:
  is-published:
    runs-on: ubuntu-latest
    steps:
      - if: github.event.action != 'published'
        run: exit 1
  build:
    runs-on: ubuntu-latest
    needs: is-published
    steps:
      - run: echo build app
  create-deployment:
    runs-on: ubuntu-latest
    needs: build
    steps:
      # # optional: install hub
      # - run: |
      #     curl -s https://api.github.com/repos/github/hub/releases/latest |
      #     jq -r '.assets[] | select(.name | contains("linux-amd64")) | .browser_download_url' |
      #     wget -qi - -O - | sudo tar xzpf - -C / --strip-components=1
      # - uses: actions/checkout@master
      - run: curl -qs https://raw.githubusercontent.com/inextensodigital/github/master/install.sh | bash -
      - id: deployment
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          id=$(github deployment create $GITHUB_REF)
          echo "##[set-output name=deployment_id;]$id"
      # #-----------------------------8<----------------------------------------
      # # here: some job(s) to add for example a deploy button/process
      # - uses: some-deploy-button-action
      #   id: button
      #   with:
      #     deployment_id: ${{ steps.deployment.outputs.deployment_id }}
      # - name: append button to release note
      #   env:
      #     DEPLOY_BUTTON: ${{ steps.button.outputs.release-button }}
      #     TAG_NAME: ${{ github.event.release.tag_name }}
      #     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      #     GITHUB_USER: ${{ github.actor }}
      #   run: |
      #     button=$(printf '\n## Deploy to production :rocket:\n%s\n' "$DEPLOY_BUTTON")
      #     hub release edit -m "" -m "$(hub release show $TAG_NAME -f %b)" -m "$button" $TAG_NAME
      # #-----------------------------8<----------------------------------------
```

Then, some external approbation process put the deployment status to `in_progress`
and you effectively deploy on production :tada:

```yaml
name: deploy on production
on: deployment_status
jobs:
  should-deploy:
    runs-on: ubuntu-latest
    steps:
      - if: github.event.deployment.original_environment != 'production' || github.event.deployment_status.state != 'in_progress'
        run: exit 1
  deploy:
    needs: should-deploy
    runs-on: ubuntu-latest
    steps:
      - run: curl -qs https://raw.githubusercontent.com/inextensodigital/github/master/install.sh | bash -
      - env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          ID: ${{ github.event.deployment.id }}
        run: |
          echo deploy app
          github deployment_status create $ID success
      - if: failure()
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          ID: ${{ github.event.deployment.id }}
        run: github deployment_status create $ID failure
```
