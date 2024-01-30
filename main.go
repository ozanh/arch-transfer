package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"

	"github.com/ozanh/arch-transfer/common"
	"github.com/ozanh/arch-transfer/sftpclient"
)

// TODO: Add SMB support.
// TODO: Handle existing destination file.
// TODO: Add tar.gzip support.
// TODO: Add progress bar.

var (
	loginf = log.New(os.Stdout, "INFO : ", log.Lshortfile|log.LstdFlags)
	logerr = log.New(os.Stderr, "ERROR: ", log.Lshortfile|log.LstdFlags)
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := App()
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		logerr.Print(err)
		os.Exit(1)
	}
}

func App() *cli.App {
	return &cli.App{
		Name:            "arch-transfer",
		Version:         "0.1.0",
		Copyright:       "MIT License (c) 2024 Ozan Hacıbekiroğlu",
		Usage:           "Transfer files and directories as archive.",
		Description:     AppDescription,
		HideHelpCommand: true,
		Action: func(cCtx *cli.Context) error {
			cli.ShowAppHelp(cCtx)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:   CommandLocal,
				Action: LocalAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     FlagSource,
						Aliases:  []string{FlagSourceShort},
						Required: true,
						Action:   checkFlagValueNotEmpty[string],
						Usage:    "Source directory or file to archive.",
					},
					&cli.StringFlag{
						Name:     FlagDestination,
						Aliases:  []string{FlagDestinationShort},
						Required: true,
						Action:   checkFlagValueNotEmpty[string],
						Usage:    "Destination archive file.",
					},
					&cli.StringFlag{
						Name:    FlagArchiveType,
						Aliases: []string{FlagArchiveTypeShort},
						Value:   ArchiveTypeZip,
						Action: func(_ *cli.Context, s string) error {
							if s != ArchiveTypeZip {
								return fmt.Errorf("invalid archive type: %s", s)
							}
							return nil
						},
					},
				},
			},
			{
				Name:   CommandSFTP,
				Action: SFTPAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     FlagAddress,
						Aliases:  []string{FlagAddressShort},
						Required: true,
						Usage:    "SFTP server address e.g. IP[:PORT].",
					},
					&cli.StringFlag{
						Name:     FlagUsername,
						Aliases:  []string{FlagUsernameShort},
						EnvVars:  []string{EnvSFTPUsername},
						Usage:    "SFTP username.",
						Required: true,
						Action:   checkFlagValueNotEmpty[string],
					},
					&cli.StringFlag{
						Name:     FlagPassword,
						Aliases:  []string{FlagPasswordShort},
						EnvVars:  []string{EnvSFTPPassword},
						Usage:    "SFTP password.",
						Required: true,
						Action:   checkFlagValueNotEmpty[string],
					},
					&cli.StringFlag{
						Name:     FlagSource,
						Aliases:  []string{FlagSourceShort},
						Required: true,
						Action:   checkFlagValueNotEmpty[string],
						Usage:    "Source directory or file to archive.",
					},
					&cli.StringFlag{
						Name:     FlagDestination,
						Aliases:  []string{FlagDestinationShort},
						Required: true,
						Action:   checkFlagValueNotEmpty[string],
						Usage:    "Destination archive file.",
					},
					&cli.StringFlag{
						Name:    FlagArchiveType,
						Aliases: []string{FlagArchiveTypeShort},
						Value:   ArchiveTypeZip,
						Action: func(_ *cli.Context, s string) error {
							if s != ArchiveTypeZip {
								return fmt.Errorf("invalid archive type: %s", s)
							}
							return nil
						},
					},
				},
			},
		},
	}
}

func LocalAction(cCtx *cli.Context) error {
	destination := cCtx.String(FlagDestination)

	destf, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	source := cCtx.String(FlagSource)
	zipWriter := zipArchiveWriteTo(source)

	return transfer(cCtx.Context, zipWriter, destf)
}

func SFTPAction(cCtx *cli.Context) error {
	address := cCtx.String(FlagAddress)
	if strings.Contains(address, ":") {
		if _, _, err := net.SplitHostPort(address); err != nil {
			// TODO: Handle IPv6
			return fmt.Errorf("invalid or unsupported address: %w", err)
		}
	} else {
		address = address + ":22"
	}

	sshConf := &ssh.ClientConfig{
		User: cCtx.String(FlagUsername),
		Auth: []ssh.AuthMethod{
			ssh.Password(cCtx.String(FlagPassword)),
		},
		// TODO: Improve security.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshConnectionTimeout,
	}

	sftpClient := new(sftpclient.Client)

	err := sftpClient.Connect(cCtx.Context, address, sshConf)
	if err != nil {
		return fmt.Errorf("failed to connect to sftp server: %w", err)
	}
	defer sftpClient.Close()

	destination := cCtx.String(FlagDestination)
	destf, err := sftpClient.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_EXCL)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	source := cCtx.String(FlagSource)
	zipWriter := zipArchiveWriteTo(source)

	return transfer(cCtx.Context, zipWriter, destf)
}

func transfer(ctx context.Context, writeTo common.ArchiveWriteToFunc, dest io.WriteCloser) error {
	var written int64
	start := time.Now()

	defer func() {
		elapsed := time.Since(start).Round(100 * time.Millisecond)
		loginf.Printf("Transfered %d bytes in %s", written, elapsed)
	}()

	bufdest := bufio.NewWriterSize(dest, bufioWriterSize)

	written, err := writeTo(ctx, bufdest)
	if err == nil {
		err = bufdest.Flush()
	}

	errClose := dest.Close()
	if err != nil {
		return err
	}
	if errClose != nil {
		return fmt.Errorf("failed to close destination file: %w", errClose)
	}
	return nil
}

func checkFlagValueNotEmpty[T comparable](cCtx *cli.Context, value T) error {
	var zero T
	if value == zero {
		return errors.New("flag value cannot be empty")
	}
	return nil
}
