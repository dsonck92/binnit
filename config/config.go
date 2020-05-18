/*
 *  This program is free software: you can redistribute it and/or
 *  modify it under the terms of the GNU Affero General Public License as
 *  published by the Free Software Foundation, either version 3 of the
 *  License, or (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 *  General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public
 *  License along with this program.  If not, see
 *  <http://www.gnu.org/licenses/>.
 *
 *  (c) Vincenzo "KatolaZ" Nicosia 2017 -- <katolaz@freaknet.org>
 *
 *
 *  This file is part of "binnit", a minimal no-fuss pastebin-like
 *  server written in golang
 *
 */

package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type options struct {
	confFile string
}

type Config struct {
	ServerName string
	BindAddr   string
	BindPort   string
	PasteDir   string
	TemplDir   string
	StaticDir  string
	MaxSize    uint16
	LogFile    string
	Storage    string
	Scheme     string
}

func (c Config) String() string {

	var s string

	s += "Server name: " + c.ServerName + "\n"
	s += "Listening on: " + c.BindAddr + ":" + c.BindPort + "\n"
	s += "paste_dir: " + c.PasteDir + "\n"
	s += "templ_dir: " + c.TemplDir + "\n"
	s += "static_dir: " + c.StaticDir + "\n"
	s += "storage: " + c.Storage + "\n"
	s += "max_size: " + string(c.MaxSize) + "\n"
	s += "log_file: " + c.LogFile + "\n"
	s += "scheme: " + c.Scheme + "\n"

	return s

}

func ParseConfig(fname string, c *Config) error {

	f, err := os.Open(fname)
	if err != nil {
		return err
	}

	r := bufio.NewScanner(f)

	line := 0
	for r.Scan() {
		s := r.Text()
		line++
		if matched, _ := regexp.MatchString("^([ \t]*)$", s); matched != true {
			// it's not a blank line
			if matched, _ := regexp.MatchString("^#", s); matched != true {
				// This is not a comment...
				if matched, _ := regexp.MatchString("^([a-z_ ]+)=.*", s); matched == true {
					// and contains an assignment
					fields := strings.Split(s, "=")
					switch strings.Trim(fields[0], " \t\"") {
					case "server_name":
						c.ServerName = strings.Trim(fields[1], " \t\"")
					case "bind_addr":
						c.BindAddr = strings.Trim(fields[1], " \t\"")
					case "bind_port":
						c.BindPort = strings.Trim(fields[1], " \t\"")
					case "paste_dir":
						c.PasteDir = strings.Trim(fields[1], " \t\"")
					case "templ_dir":
						c.TemplDir = strings.Trim(fields[1], " \t\"")
					case "static_dir":
						c.StaticDir = strings.Trim(fields[1], " \t\"")
					case "storage":
						c.Storage = strings.Trim(fields[1], " \t\"")
					case "log_file":
						c.LogFile = strings.Trim(fields[1], " \t\"")
					case "max_size":
						if mSize, err := strconv.ParseUint(fields[1], 10, 16); err == nil {
							c.MaxSize = uint16(mSize)
						} else {
							fmt.Fprintf(os.Stderr, "Invalid max_size value %s at line %d (max: 65535)\n",
								fields[1], line)
						}
					case "scheme":
						c.Scheme = strings.Trim(fields[1], " \t\"")
					default:
						fmt.Fprintf(os.Stderr, "Error reading config file %s at line %d: unknown variable '%s'\n",
							fname, line, fields[0])
					}
				} else {
					fmt.Fprintf(os.Stderr, "Error reading config file %s at line %d: unknown statement '%s'\n",
						fname, line, s)
				}
			}
		}
	}
	return nil
}
