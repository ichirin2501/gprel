package gprel

import (
	"os"
	"testing"
)

var eqErrFunc = func(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Error() == b.Error()
}

func resetEnv() {
	os.Unsetenv("MYSQL_HOST")
	os.Unsetenv("MYSQL_PWD")
	os.Unsetenv("MYSQL_TCP_PORT")
	os.Unsetenv("MYSQL_UNIX_PORT")
}

func TestParseOptions(t *testing.T) {
	patterns := []struct {
		args       []string
		env        map[string]string
		wantConfig *Configuration
		wantError  bool
		err        error
	}{
		{
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
			},
			false,
			nil,
		},
		{
			[]string{"test2", "-defaults-file", "./testdata/test1.cnf", "-delay", "5"},
			map[string]string{"MYSQL_HOST": "10.0.10.5", "MYSQL_PWD": "test"},
			&Configuration{
				Host:              "10.0.10.5",
				Socket:            "",
				User:              "root",
				Password:          "XXXYYY",
				Port:              3306,
				DatabaseName:      "",
				PurgeDelaySeconds: 5,
			},
			false,
			nil,
		},
	}

	for idx, p := range patterns {
		resetEnv()

		// set env
		for k, v := range p.env {
			if err := os.Setenv(k, v); err != nil {
				t.Fatal(err)
			}
		}
		gotConfig, gotErr := ParseOptions(p.args)
		if !p.wantError && gotErr != nil {
			t.Fatalf("pattern %d: want no err, but has error %+v", idx, gotErr)
		}
		if p.wantError && !eqErrFunc(p.err, gotErr) {
			t.Fatalf("pattern %d: want %+v, but %+v", idx, p.err, gotErr)
		}
		if *gotConfig != *p.wantConfig {
			t.Errorf("pattern %d: want (%+v), got (%+v)", idx, p.wantConfig, gotConfig)
		}
	}
}
