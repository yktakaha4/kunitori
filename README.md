# kunitori

Aggregate SLOCs from repositories and plot contributors of now `living code` on a map of Japan.

![Example](https://user-images.githubusercontent.com/20282867/209167916-3dd01384-f8e4-401c-b9cc-0f9947e01b51.png)

## Install

[Download binary](https://github.com/yktakaha4/kunitori/releases)

## Usage

```
$ kunitori generate -h
Usage of generate:
  -authors value
        target file author regex (multiple specified, format: author=regex)
  -filters value
        target file filter regex (multiple specified)
  -interval duration
        commit pick interval (default 720h0m0s)
  -json
        export as json format
  -limit int
        commit pick limit (default 12)
  -out string
        out directory path (default ".")
  -path string
        repository path
  -region string
        chart region (default "JP")
  -since string
        filter commit since date (format: 2006-01-02T15:04:05Z07:00)
  -until string
        filter commit until date (format: 2006-01-02T15:04:05Z07:00)
  -url string
        repository url
```

## Example

```
# Kunitori with Python and Frontend Contributors
$ kunitori generate -path /path-to/your-org/your-repo -filters '.+\.py$' -filters 'test_.+\.py$' -filters '\.(vue|ts)$' -filters '\.(spec|test)\.(vue|ts)$
```

## Development

```
# Test
$ make test

# Build binary
$make build
```
