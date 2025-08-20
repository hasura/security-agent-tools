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