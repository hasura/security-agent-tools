# Security Agent Tools

> Tools for working with the PromptQL Security Agent.

## GitHub Action

### Upload File

The `upload-file` action uploads a file to the Security Agent. It can be used to upload scan results from various security scanners to the Security Agent.

```yaml
- uses: hasura/security-agent-tools/upload-file@v1
  with:
    # Local file system path for the file to be uploaded
    # Required
    file_path: ''

    # Security Agent API key
    # Required
    security_agent_api_key: ${{ secrets.SECURITY_AGENT_API_KEY }}

    # Security Agent API endpoint
    # Optional. If not provided, the default public endpoint will be used.
    security_agent_api_endpoint: ''

    # Key value pairs of tags separated by new line.
    # Optional. If not provided, no tags will be uploaded.
    tags: |
      scanner=trivy

    # Path to upload metadata.
    # Optional. If not provided, the action uses `service` value in `tags` to construct the path.
    metadata_upload_path: ''
```

#### Usage

This is compatible to run with all event types in GitHub Actions: `push`, `pull_request`, `schedule`, `workflow_dispatch`. That means that, you can use the GitHub action to store security scan reports from pull request builds, branch builds, cron builds, and manual builds.

```yaml
name: Scan Sample Service

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
  schedule:
    - cron: "0 0 * * *" # Daily midnight
  workflow_dispatch:

jobs:
  test-and-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      # docker build

      - name: Run Trivy vulnerability scanner (json output)
        uses: aquasecurity/trivy-action@0.32.0
        with:
          image-ref: ${{ steps.docker-build.outputs.IMG_NAME }}
          format: json
          output: trivy-results.json
          severity: CRITICAL,HIGH
          scanners: vuln

      - name: Upload Trivy scan results to PromptQL Security Agent
        uses: hasura/security-agent-tools/upload-file@main
        with:
          file_path: trivy-results.json
          security_agent_api_key: ${{ secrets.SECURITY_AGENT_API_KEY }}
          tags: |
            service=sample-service
            source_code_path=path/to/source-code
            docker_file_path=path/to/source-code/Dockerfile
            scanner=trivy
```


## Buildkite

The `upload-file` action can also be used in Buildkite pipelines. It will automatically detect that it is running in Buildkite and upload the metadata file to the appropriate path.

You will need to use it via Docker:

```sh
trivy_scan() {
  echo "--- :shield: Scanning container image"
  trivy image --format json -o trivy-results.json --scanners vuln "$IMG_NAME"
}

upload_report_to_security_agent() {
  echo "--- :shield: Uploading report to security agent"
  TAGS=$(cat << EOF
service=$SERVICE_NAME
scanner=trivy
image_name=$IMG_NAME
EOF
)
  echo "Tags: $TAGS"
  docker run --rm -v $(pwd):/workdir \
    -e INPUT_FILE_PATH=/workdir/trivy-results.json \
    -e INPUT_SECURITY_AGENT_API_KEY=$SECURITY_AGENT_API_KEY \
    -e INPUT_TAGS="$TAGS" \
    -e BUILDKITE \
    -e BUILDKITE_COMMIT \
    -e BUILDKITE_PIPELINE_SLUG \
    -e BUILDKITE_BRANCH \
    -e BUILDKITE_TAG \
    -e BUILDKITE_PULL_REQUEST \
    -e GITHUB_REPOSITORY \
    ghcr.io/hasura/security-agent-tools/upload-file:v1
}

trivy_scan
upload_report_to_security_agent
```

## Development

You can use [just](https://github.com/casey/just) to build and test the tools.

```
$ just help

security-agent-tools:
  Tools for working with the PromptQL Security Agent.
  All tools are packaged as Docker image and GitHub Action.

Common commands:
  setup-env                - Create .env file from example

upload-file:
  build-upload-file        - Build the upload-file binary locally
  test-upload-file-local   - Test binary locally (requires .env)
  test-upload-file-docker  - Test with Docker (uses .env if available)
```
