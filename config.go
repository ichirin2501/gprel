package gprel

import (
	"flag"
	"os"
	"strconv"

	"github.com/vaughan0/go-ini"
)

type Configuration struct {
	Host              string
	Socket            string
	User              string
	Password          string
	Port              int
	DatabaseName      string
	PurgeDelaySeconds int
}

func (c *Configuration) loadEnvironmentVariables() error {
	// ref. https://dev.mysql.com/doc/refman/5.6/ja/environment-variables.html
	if v := os.Getenv("MYSQL_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("MYSQL_PWD"); v != "" {
		c.Password = v
	}
	if v := os.Getenv("MYSQL_TCP_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		c.Port = i
	}
	if v := os.Getenv("MYSQL_UNIX_PORT"); v != "" {
		c.Socket = v
	}
	return nil
}

func (c *Configuration) loadDefaultsFile(filepath string) error {
	file, err := ini.LoadFile(filepath)
	if err != nil {
		return err
	}
	section := file.Section("client")
	if user, ok := section["user"]; ok {
		c.User = user
	}
	if password, ok := section["password"]; ok {
		c.Password = password
	}
	if socket, ok := section["socket"]; ok {
		c.Socket = socket
	}
	if host, ok := section["host"]; ok {
		c.Host = host
	}
	if port, ok := section["port"]; ok {
		i, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		c.Port = i
	}
	if database, ok := section["database"]; ok {
		c.DatabaseName = database
	}
	return nil
}

type myArg struct {
	value      string
	fromCmdArg bool
}

func (v *myArg) String() string {
	return v.value
}
func (v *myArg) Set(s string) error {
	v.value = s
	v.fromCmdArg = true
	return nil
}

func ParseOptions(args []string) (*Configuration, error) {
	var (
		database          myArg
		host              myArg
		socket            myArg
		user              myArg
		password          myArg
		defaultsFile      myArg
		port              myArg
		purgeDelaySeconds myArg
	)
	f := flag.NewFlagSet(args[0], flag.ContinueOnError)

	f.Var(&database, "d", "mysql database name")
	f.Var(&host, "h", "mysql host")
	f.Var(&socket, "S", "mysql unix socket")
	f.Var(&user, "u", "mysql username")
	f.Var(&password, "p", "mysql user password")
	f.Var(&defaultsFile, "defaults-file", "Only read default options from the given file")
	f.Var(&port, "P", "mysql port")
	f.Var(&purgeDelaySeconds, "delay", "purge delay seconds")
	if err := f.Parse(args[1:]); err != nil {
		return nil, err
	}

	c := &Configuration{
		Host:              "127.0.0.1",
		Socket:            "",
		User:              "",
		Password:          "",
		Port:              3306,
		DatabaseName:      "",
		PurgeDelaySeconds: 7,
	}

	if err := c.loadEnvironmentVariables(); err != nil {
		return nil, err
	}

	if defaultsFile.fromCmdArg {
		if err := c.loadDefaultsFile(defaultsFile.value); err != nil {
			return nil, err
		}
	}

	if database.fromCmdArg {
		c.DatabaseName = database.value
	}
	if host.fromCmdArg {
		c.Host = host.value
	}
	if socket.fromCmdArg {
		c.Socket = socket.value
	}
	if user.fromCmdArg {
		c.User = user.value
	}
	if password.fromCmdArg {
		c.Password = password.value
	}
	if port.fromCmdArg {
		i, err := strconv.Atoi(port.value)
		if err != nil {
			return nil, err
		}
		c.Port = i
	}
	if purgeDelaySeconds.fromCmdArg {
		i, err := strconv.Atoi(purgeDelaySeconds.value)
		if err != nil {
			return nil, err
		}
		c.PurgeDelaySeconds = i
	}

	return c, nil
}
