# Upload File GitHub Action

A GitHub Action to upload files to a specified destination URL.

## Features

- Upload files via HTTP POST requests
- Support for authentication via API keys
- Configurable timeout settings
- Outputs upload URL and status for further workflow steps

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `file-path` | Path to the file to upload | Yes | - |
| `destination` | Destination URL for the upload | Yes | - |
| `api-key` | API key for authentication | No | - |
| `timeout` | Timeout for the upload operation in seconds | No | `300` |

## Outputs

| Output | Description |
|--------|-------------|
| `upload-url` | URL of the uploaded file |
| `upload-status` | Status of the upload operation (`success` or `failed`) |

## Usage

### Basic Usage

```yaml
- name: Upload file
  uses: ./upload-file
  with:
    file-path: './my-file.txt'
    destination: 'https://api.example.com/upload'
```

### With Authentication

```yaml
- name: Upload file with API key
  uses: ./upload-file
  with:
    file-path: './my-file.txt'
    destination: 'https://api.example.com/upload'
    api-key: ${{ secrets.API_KEY }}
    timeout: '600'
```

### Using Outputs

```yaml
- name: Upload file
  id: upload
  uses: ./upload-file
  with:
    file-path: './my-file.txt'
    destination: 'https://api.example.com/upload'

- name: Use upload results
  run: |
    echo "Upload URL: ${{ steps.upload.outputs.upload-url }}"
    echo "Upload Status: ${{ steps.upload.outputs.upload-status }}"
```

## Example Workflow

```yaml
name: Upload Build Artifacts

on:
  push:
    branches: [ main ]

jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Build application
      run: |
        # Your build commands here
        echo "Building application..."
        
    - name: Upload build artifact
      id: upload
      uses: ./upload-file
      with:
        file-path: './dist/app.zip'
        destination: 'https://storage.example.com/artifacts'
        api-key: ${{ secrets.STORAGE_API_KEY }}
        timeout: '300'
        
    - name: Notify on success
      if: steps.upload.outputs.upload-status == 'success'
      run: echo "File uploaded successfully to ${{ steps.upload.outputs.upload-url }}"
```

## Development

### Building Locally

```bash
cd upload-file
go build -o upload-file main.go
```

### Testing

```bash
# Set environment variables
export INPUT_FILE-PATH="./test-file.txt"
export INPUT_DESTINATION="https://httpbin.org/post"
export INPUT_TIMEOUT="60"

# Run the action
./upload-file
```

### Docker Build

```bash
docker build -t upload-file-action .
```

## License

This project is licensed under the MIT License.
