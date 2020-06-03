package gprel

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetEnv() {
	os.Unsetenv("MYSQL_HOST")
	os.Unsetenv("MYSQL_PWD")
	os.Unsetenv("MYSQL_TCP_PORT")
	os.Unsetenv("MYSQL_UNIX_PORT")
}

func TestParseOptions(t *testing.T) {
	patterns := []struct {
		name       string
		args       []string
		env        map[string]string
		wantConfig *Configuration
		wantError  bool
		err        error
	}{
		{
			"1",
			[]string{"test1", "-u", "root", "-p", "pass", "-h", "localhost", "-P", "13306"},
			map[string]string{},
			&Configuration{
				Host:              "localhost",
				Socket:            "",
				User:              "root",
				Password:          "pass",
				Port:              13306,
				DatabaseName:      "",
				PurgeDelaySeconds: 7,
				DryRun:            true,
			},
			false,
			nil,
		},
		{
			"2",
			[]string{"test2", "-defaults-file", "./testdata/test1.cnf", "-delay", "5", "-go"},
			map[string]string{"MYSQL_HOST": "10.0.10.5", "MYSQL_PWD": "test"},
			&Configuration{
				Host:              "10.0.10.5",
				Socket:            "",
				User:              "root",
				Password:          "XXXYYY",
				Port:              3306,
				DatabaseName:      "",
				PurgeDelaySeconds: 5,
				DryRun:            false,
			},
			false,
			nil,
		},
		{
			"3",
			[]string{"test3", "-g", "nihaha"},
			map[string]string{},
			&Configuration{},
			true,
			errors.New("flag provided but not defined: -g"),
		},
		{
			"4",
			[]string{"test4", "-defaults-file", "./testdata/test100.cnf", "-delay", "5"},
			map[string]string{},
			&Configuration{},
			true,
			errors.New("open ./testdata/test100.cnf: no such file or directory"),
		},
		{
			"5",
			[]string{"test5", "-defaults-file", "./testdata/test2.cnf"},
			map[string]string{},
			&Configuration{},
			true,
			errors.New("invalid INI syntax on line 1: //// [client]"),
		},
	}

	for _, p := range patterns {
		p := p
		t.Run(p.name, func(t *testing.T) {
			resetEnv()

			// set env
			for k, v := range p.env {
				if err := os.Setenv(k, v); err != nil {
					t.Fatal(err)
				}
			}
			gotConfig, gotErr := ParseOptions(p.args)
			if p.wantError {
				assert.Nil(t, gotConfig)
				assert.EqualError(t, gotErr, p.err.Error())
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, p.wantConfig, gotConfig)
			}
		})
	}
}
