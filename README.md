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
go install
```

To compile for all platforms using [`gox`](https://github.com/mitchellh/gox), run the following:
```
VERSION=v1.0.0 gox -output "dist/$VERSION/{{.OS}}-{{.Arch}}/{{.Dir}}"
```

## Testing licensing

Below is a list of license keys you can use to test program functionality.

| Key    | Validity  |
|:-------|:----------|
| 000001 | Valid     |
| 000010 | Suspended |
…

## Testing updates

Enter a valid license key and when available, press `ctrl+u` to install the
update from v1.0.0 to v1.0.1. It will download the update, and then replace
the current executable with the updated one.

## Questions?

Reach out at [support@keygen.sh](mailto:support@keygen.sh) if you have any
questions or concerns!
