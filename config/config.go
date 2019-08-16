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
	// ref. https://dev.mysql.com/doc/refman/5.6/ja/environment-variables.html
	c := &Configuration{
		Host:              "127.0.0.1",
		Port:              3306,
		PurgeDelaySeconds: 7,
	}
	if v := os.Getenv("MYSQL_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("MYSQL_PWD"); v != "" {
		c.Password = v
	}
	if v := os.Getenv("MYSQL_TCP_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.Port = i
		}
	}
	if v := os.Getenv("MYSQL_UNIX_PORT"); v != "" {
		c.Socket = v
	}
	return c
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

func ParseOptions() error {
	var (
		database          string
		host              string
		socket            string
		user              string
		password          string
		defaultsFile      string
		port              int
		purgeDelaySeconds int
	)
	flag.StringVar(&database, "d", "_default_", "mysql database name")
	flag.StringVar(&host, "h", "_default_", "mysql host")
	flag.StringVar(&socket, "S", "_default_", "mysql unix socket")
	flag.StringVar(&user, "u", "_default_", "mysql username")
	flag.StringVar(&password, "p", "_default_", "mysql user password")
	flag.StringVar(&defaultsFile, "defaults-file", "_default_", "Only read default options from the given file")
	flag.IntVar(&port, "P", -1, "mysql port")
	flag.IntVar(&purgeDelaySeconds, "delay", -1, "purge delay seconds")
	flag.Parse()

	if defaultsFile != "_default_" {
		if err := loadDefaultsFile(defaultsFile); err != nil {
			return err
		}
	}

	if database != "_default_" {
		Config.DatabaseName = database
	}
	if host != "_default_" {
		Config.Host = host
	}
	if socket != "_default_" {
		Config.Socket = socket
	}
	if user != "_default_" {
		Config.User = user
	}
	if password != "_default_" {
		Config.Password = password
	}
	if port != -1 {
		Config.Port = port
	}
	if purgeDelaySeconds != -1 {
		Config.PurgeDelaySeconds = purgeDelaySeconds
	}

	return nil
}
