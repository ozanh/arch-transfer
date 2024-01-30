package sftpclient

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client is a wrapper for sftp.Client that handles ssh connection.
// Close method closes both sftp and ssh connections.
type Client struct {
	ssh  *ssh.Client
	sftp *sftp.Client
}

// Connect connects to the address with ssh client config and constructs sftp client.
func (c *Client) Connect(ctx context.Context, address string, sshConf *ssh.ClientConfig) error {
	sshClient, err := dialSSH(ctx, address, sshConf)
	if err != nil {
		return fmt.Errorf("sftpclient: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		_ = sshClient.Close()
		return fmt.Errorf("sftpclient: %w", err)
	}
	c.ssh = sshClient
	c.sftp = sftpClient
	return nil
}

// Close closes both sftp and ssh connections. Must be called after Connect.
// After Close(), other methods should not be called.
func (c *Client) Close() error {
	err := c.sftp.Close()
	if e := c.ssh.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

// OpenFile opens the file with flags. See sftp.OpenFile for details.
// Must be called after Connect.
func (c *Client) OpenFile(path string, flags int) (*sftp.File, error) {
	return c.sftp.OpenFile(path, flags)
}

func dialSSH(ctx context.Context, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial for ssh: %w", err)
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create ssh connection: %w", err)
	}
	return ssh.NewClient(c, chans, reqs), nil
}
