// apcmetrics - APC UPS metrics exporter for Prometheus
//
// Copyright 2021 Nick Pillitteri
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package apcmetrics

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/go-kit/log"
)

const readBufferSize = 255

type ApcClient struct {
	address string
	logger  log.Logger
}

func NewApcClient(address string, logger log.Logger) *ApcClient {
	return &ApcClient{
		address: address,
		logger:  logger,
	}
}

func (a *ApcClient) connect(ctx context.Context) (net.Conn, error) {
	var d net.Dialer

	// use the provided context for making the connection
	conn, err := d.DialContext(ctx, "tcp", a.address)
	if err != nil {
		return nil, err
	}

	// if the context had a deadline set, propagate it for all reads and writes
	// on this connection. connections are not long-lived and so it's reasonable
	// to use the deadline / context provided to the .Status() or .Event() methods
	if deadline, ok := ctx.Deadline(); ok {
		err = conn.SetDeadline(deadline)
		if err != nil {
			// if we couldn't set the deadline successfully, close the connection
			// since we're going to short-circuit the rest of the calls and return
			// the error
			_ = conn.Close()
			return nil, err
		}
	}

	return conn, nil
}

func (a *ApcClient) formatCommand(cmd string) []byte {
	cmdLen := len(cmd)
	buf := make([]byte, 2+cmdLen)
	binary.BigEndian.PutUint16(buf, uint16(cmdLen))
	copy(buf[2:], cmd)
	return buf
}

func (a *ApcClient) send(ctx context.Context, cmd string) ([]string, error) {
	conn, err := a.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()
	cmdBytes := a.formatCommand(cmd)
	cmdLen := len(cmdBytes)

	n, err := conn.Write(cmdBytes)
	if err != nil {
		return nil, err
	}

	if n != cmdLen {
		return nil, fmt.Errorf("short write cmd=%s expected=%d got=%d", cmd, cmdLen, n)
	}

	lineBuf := make([]byte, readBufferSize)
	var out []string
	for {
		n, err = conn.Read(lineBuf[0:2])
		if err != nil {
			return nil, err
		}

		if n == 0 {
			break
		}

		sz := int(binary.BigEndian.Uint16(lineBuf[0:2]))
		if readBufferSize < sz {
			sz = readBufferSize
		}

		if sz == 0 {
			break
		}

		n, err = conn.Read(lineBuf[0:sz])
		if err != nil {
			return nil, err
		}

		if n == 0 || n != sz {
			break
		}

		s := strings.TrimSpace(string(lineBuf[0:n]))
		out = append(out, s)
	}

	return out, nil
}

func (a *ApcClient) Status(ctx context.Context) (*ApcStatus, error) {
	status, err := a.StatusRaw(ctx)
	if err != nil {
		return nil, err
	}

	return ParseStatusFromLines(status)
}

func (a *ApcClient) Events(ctx context.Context) ([]ApcEvent, error) {
	events, err := a.EventsRaw(ctx)
	if err != nil {
		return nil, err
	}

	return ParseEventsFromLines(events)
}

func (a *ApcClient) StatusRaw(ctx context.Context) ([]string, error) {
	r, err := a.send(ctx, "status")
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (a *ApcClient) EventsRaw(ctx context.Context) ([]string, error) {
	r, err := a.send(ctx, "events")
	if err != nil {
		return nil, err
	}

	return r, nil
}
