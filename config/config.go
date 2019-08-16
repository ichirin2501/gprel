package config

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

var Config = newConfiguration()

func newConfiguration() *Configuration {
	return &Configuration{
		Host:              "127.0.0.1",
		Socket:            "",
		User:              "",
		Password:          "",
		Port:              3306,
		DatabaseName:      "",
		PurgeDelaySeconds: 7,
	}
}

func loadEnvironmentVariables() error {
	// ref. https://dev.mysql.com/doc/refman/5.6/ja/environment-variables.html
	if v := os.Getenv("MYSQL_HOST"); v != "" {
		Config.Host = v
	}
	if v := os.Getenv("MYSQL_PWD"); v != "" {
		Config.Password = v
	}
	if v := os.Getenv("MYSQL_TCP_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		Config.Port = i
	}
	if v := os.Getenv("MYSQL_UNIX_PORT"); v != "" {
		Config.Socket = v
	}
	return nil
}

func loadDefaultsFile(filepath string) error {
	file, err := ini.LoadFile(filepath)
	if err != nil {
		return err
	}
	section := file.Section("client")
	if user, ok := section["user"]; ok {
		Config.User = user
	}
	if password, ok := section["password"]; ok {
		Config.Password = password
	}
	if socket, ok := section["socket"]; ok {
		Config.Socket = socket
	}
	if host, ok := section["host"]; ok {
		Config.Host = host
	}
	if port, ok := section["port"]; ok {
		i, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		Config.Port = i
	}
	if database, ok := section["database"]; ok {
		Config.DatabaseName = database
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

func ParseOptions() error {
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
	flag.Var(&database, "d", "mysql database name")
	flag.Var(&host, "h", "mysql host")
	flag.Var(&socket, "S", "mysql unix socket")
	flag.Var(&user, "u", "mysql username")
	flag.Var(&password, "p", "mysql user password")
	flag.Var(&defaultsFile, "defaults-file", "Only read default options from the given file")
	flag.Var(&port, "P", "mysql port")
	flag.Var(&purgeDelaySeconds, "delay", "purge delay seconds")
	flag.Parse()

	if err := loadEnvironmentVariables(); err != nil {
		return err
	}

	if defaultsFile.fromCmdArg {
		if err := loadDefaultsFile(defaultsFile.value); err != nil {
			return err
		}
	}

	if database.fromCmdArg {
		Config.DatabaseName = database.value
	}
	if host.fromCmdArg {
		Config.Host = host.value
	}
	if socket.fromCmdArg {
		Config.Socket = socket.value
	}
	if user.fromCmdArg {
		Config.User = user.value
	}
	if password.fromCmdArg {
		Config.Password = password.value
	}
	if port.fromCmdArg {
		i, err := strconv.Atoi(port.value)
		if err != nil {
			return err
		}
		Config.Port = i
	}
	if purgeDelaySeconds.fromCmdArg {
		i, err := strconv.Atoi(purgeDelaySeconds.value)
		if err != nil {
			return err
		}
		Config.PurgeDelaySeconds = i
	}

	return nil
}
