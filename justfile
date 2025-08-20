# Test the upload-file Docker image
test-upload-file:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Building upload-file Docker image..."
    docker build -t upload-file-test ./upload-file

    echo "Creating test file..."
    echo "This is a test file for Docker testing" > test-file.txt
    echo "Created at: $(date)" >> test-file.txt

    echo "Testing upload-file Docker image..."
    docker run --rm \
        -v "$(pwd)/test-file.txt:/test-file.txt:ro" \
        -e INPUT_FILE_PATH="/test-file.txt" \
        -e INPUT_DESTINATION="https://httpbin.org/post" \
        upload-file-test

    echo "Cleaning up test file..."
    rm -f test-file.txt

    echo "Upload-file Docker test completed successfully!"