# Arch-Transfer

Arch-Transfer is a simple tool to transfer files from a local machine to local
or to a remote machine via SFTP or SMB without creating an intermediate file.

The project is written in Golang and is under active development.

## Installation

> go install github.com/ozanh/arch-transfer

## Usage

### Help
> ./arch-transfer --help

### Local transfer

> ./arch-transfer local --source /path/to/source --destination /path/to/destination.zip --type zip

### SFTP transfer

> ./arch-transfer sftp --address localhost --username root --password mypassword --source /path/to/source --destination /path/to/destination.zip --type zip

## TODO
- Add SMB support.
- Handle existing destination file.
- Improve SFTP Host Key check.
- Add tar.gzip support.
- Add progress bar.
- Add tests.