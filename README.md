# Security Agent Tools

> Tools for working with the PromptQL Security Agent.

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

## GitHub Actions

An example workflow to build a Docker image and scan it with Trivy, then upload the results to the Security Agent.

```yaml
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