# Self-Updating Go Program
The following is a self-updating program written in Go that showcases how
to use [Keygen](https://keygen.sh) for licensing and auto-updates.

## Compiling

First up, install dependencies with [`dep`](https://github.com/golang/dep):
```
dep ensure
```

To compile for your operating system, simply run the following from the root of the project directory:
```
go build
```

To compile for all platforms using [`gox`](https://github.com/mitchellh/gox), run the following:
```
gox -output "dist/v1.x.x/{{.OS}}-{{.Arch}}/{{.Dir}}"
```

## Prebuilt-binaries

You can download pre-built binaries for 64bit operating systems. The pre-built binaries are of v1.0.0, so you will still be able to test auto-update functionality.

- [`darwin/amd64`](https://dist.keygen.sh/v1/1fddcec8-8dd3-4d8d-9b16-215cac0f9b52/2d130468-27aa-4c41-b064-18fc6b3046d9/releases/darwin-amd64/v1.0.0.zip?key=val-key)
- [`windows/amd64`](https://dist.keygen.sh/v1/1fddcec8-8dd3-4d8d-9b16-215cac0f9b52/2d130468-27aa-4c41-b064-18fc6b3046d9/releases/windows-amd64/v1.0.0.zip?key=val-key)
- [`linux/amd64`](https://dist.keygen.sh/v1/1fddcec8-8dd3-4d8d-9b16-215cac0f9b52/2d130468-27aa-4c41-b064-18fc6b3046d9/releases/linux-amd64/v1.0.0.zip?key=val-key)
- [`freebsd/amd64`](https://dist.keygen.sh/v1/1fddcec8-8dd3-4d8d-9b16-215cac0f9b52/2d130468-27aa-4c41-b064-18fc6b3046d9/releases/freebsd-amd64/v1.0.0.zip?key=val-key)

## Testing licensing

Below is a list of license keys you can use to test program functionality.

| Key       | Validity  |
|:----------|:----------|
| `val-key` | Valid     |
| `sus-key` | Suspended |
| `exp-key` | Expired   |

## Testing updates

Enter a valid license key and when available, press `ctrl+u` to install the
update from v1.0.0 to v1.0.1. It will download the update, and then replace
the current executable with the updated one. You can then rerun the program
to use the newly updated version.

## Questions?

Reach out at [support@keygen.sh](mailto:support@keygen.sh) if you have any
questions or concerns!
