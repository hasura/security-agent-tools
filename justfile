# Load environment variables from .env file if it exists
set dotenv-load := true

# Setup environment file from example
setup-env:
    #!/usr/bin/env bash
    set -euo pipefail

    if [ -f "./.env" ]; then
        echo "Environment file ./.env already exists."
        echo "Remove it first if you want to recreate it from the example."
        exit 1
    fi

    echo "Creating ./.env from example..."
    cp ./.env.example ./.env
    echo "Environment file created at ./.env"
    echo "Please edit this file and add your actual API key and configuration."

# Build the upload-file binary locally
build-upload-file:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "--- :: Building upload-file binary"
    cd upload-file
    go build -o upload-file .
    echo "Binary built: ./upload-file/upload-file"

# Test upload-file binary locally (requires .env file)
test-upload-file-local:
    #!/usr/bin/env bash
    set -euo pipefail

    # Check if .env file exists
    if [ ! -f "./.env" ]; then
        echo "Error: ./.env file not found!"
        echo "Please copy ./.env.example to ./.env and configure your values."
        exit 1
    fi

    # Load environment variables
    echo "--- :: Loading environment variables from ./.env"
    set -a
    source ./.env
    set +a

    # Build the binary first
    just build-upload-file

    echo "--- :: Creating test file"
    echo '{"test": "data", "timestamp": "'$(date -Iseconds)'"}' > test-file.json

    echo "--- :: Testing upload-file binary locally"
    cd upload-file
    INPUT_FILE_PATH="../test-file.json" \
    INPUT_DESTINATION="${INPUT_DESTINATION:-temp-data/local-test-$(date +%s).json}" \
    INPUT_SECURITY_AGENT_API_ENDPOINT="${INPUT_SECURITY_AGENT_API_ENDPOINT}" \
    INPUT_SECURITY_AGENT_API_KEY="${INPUT_SECURITY_AGENT_API_KEY}" \
    ./upload-file

    echo "--- :: Cleaning up test file"
    rm -f ../test-file.json

    echo "Local upload test completed successfully!"

# Test the upload-file Docker image
test-upload-file-docker:
    #!/usr/bin/env bash
    set -euo pipefail

    # Load environment variables from .env file if it exists
    if [ -f "./.env" ]; then
        echo "--- :: Loading environment variables from ./.env"
        set -a
        source ./.env
        set +a
    else
        echo "Warning: ./.env not found. Using default values."
        echo "Copy ./.env.example to ./.env and configure your values."
    fi

    echo "--- :: Building upload-file Docker image"
    docker build -t upload-file-test ./upload-file

    echo "--- :: Creating test file"
    echo '{"test": "data", "timestamp": "'$(date -Iseconds)'", "message": "Docker test file"}' > test-file.json

    echo "--- :: Testing upload-file Docker image"
    docker run --rm \
        -v "$(pwd)/test-file.json:/test-file.json:ro" \
        -e INPUT_FILE_PATH="/test-file.json" \
        -e INPUT_DESTINATION="${INPUT_DESTINATION:-temp-data/test-$(date +%s).json}" \
        -e INPUT_SECURITY_AGENT_API_ENDPOINT="${INPUT_SECURITY_AGENT_API_ENDPOINT:-}" \
        -e INPUT_SECURITY_AGENT_API_KEY="${INPUT_SECURITY_AGENT_API_KEY:-}" \
        upload-file-test

    echo "--- :: Cleaning up test file"
    rm -f test-file.json

    echo "Upload-file Docker test completed successfully!"

# Show available recipes
help:
    @echo "security-agent-tools:"
    @echo "  Tools for working with the PromptQL Security Agent."
    @echo "  All tools are packaged as Docker image and GitHub Action."
    @echo ""
    @echo "Common commands:"
    @echo "  setup-env                - Create .env file from example"
    @echo ""
    @echo "upload-file:"
    @echo "  build-upload-file        - Build the upload-file binary locally"
    @echo "  test-upload-file-local   - Test binary locally (requires .env)"
    @echo "  test-upload-file-docker  - Test with Docker (uses .env if available)"