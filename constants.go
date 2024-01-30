package main

import "time"

const AppDescription = `Transfer files and directories as zip or tar.gzip archive to local or a remote server without creating intermediate file.`

// CLI Commands
const (
	CommandLocal = "local"
	CommandSFTP  = "sftp"
)

// CLI Flags
const (
	FlagAddress          = "address"
	FlagAddressShort     = "a"
	FlagUsername         = "username"
	FlagUsernameShort    = "u"
	FlagPassword         = "password"
	FlagPasswordShort    = "p"
	FlagSource           = "source"
	FlagSourceShort      = "s"
	FlagDestination      = "destination"
	FlagDestinationShort = "d"
	FlagArchiveType      = "type"
	FlagArchiveTypeShort = "t"
)

// Archive Types
const (
	ArchiveTypeZip = "zip"
)

// Environment Variables
const (
	EnvSFTPUsername = "AT_SFTP_USERNAME"
	EnvSFTPPassword = "AT_SFTP_PASSWORD"
)

const (
	sshConnectionTimeout = 30 * time.Second
)

const (
	bufioWriterSize = 1024 * 1024
)
