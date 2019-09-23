package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type T interface {
}

type ttFStruct struct {
	tName       string
	tInput      []T
	tFunc       T
	tOutput     T
	tPipeOutput T
	tCleanup    T
}

func (tt ttFStruct) tPipe(t *testing.T) func(format string, a ...interface{}) (int, error) {
	return func(format string, a ...interface{}) (int, error) {
		s1 := fmt.Sprintf(format, a...)
		s2 := fmt.Sprintf("%v", tt.tPipeOutput)
		if strings.Compare(s1, s2) != 0 {
			t.Fatalf("expected string: \"%s\", got: \"%s\"", s2, s1)
		}
		return 0, nil
	}
}

func testPackageMethod(tt []ttFStruct, t *testing.T) {
	for _, tc := range tt {
		print = tc.tPipe(t)
		inputs := make([]reflect.Value, len(tc.tInput))
		for i := range tc.tInput {
			inputs[i] = reflect.ValueOf(tc.tInput[i])
		}
		for _, v := range reflect.ValueOf(tc.tFunc).Call(inputs) {
			t1 := fmt.Sprintf("%v", reflect.ValueOf(tc.tOutput).Interface())
			t2 := fmt.Sprintf("%v", reflect.ValueOf(v).Interface())
			if t1 != t2 {
				t.Fatalf("test: \"%s\" failed! Expected: '%v', got: '%v'", tc.tName, t1, t2)
			}
		}
		if tc.tCleanup != nil {
			os.Remove(tc.tCleanup.(string))
		}
	}
}

func TestEvalCMD(t *testing.T) {
	tt := []ttFStruct{
		{
			tName: "Test eval cmd does not work",
			tInput: []T{
				"this will fail",
			},
			tFunc:       evalCmd,
			tOutput:     "",
			tPipeOutput: "could not evaluate cmd \"this will fail\": exit status 127\n",
		},
		{
			tName: "Test eval cmd works",
			tInput: []T{
				"echo hello world!",
			},
			tFunc:   evalCmd,
			tOutput: "hello world!\n",
		},
	}
	testPackageMethod(tt, t)
}

func TestPrintEntity(t *testing.T) {
	tt := []ttFStruct{
		{
			tName: "Test print entity keys",
			tInput: []T{
				map[string]string{
					"test1": "hello",
					"test2": "world",
				},
				KEYS,
			},
			tFunc:       printEntity,
			tPipeOutput: "test1 test2\n",
		},
		{
			tName: "Test print entity values",
			tInput: []T{
				map[string]string{
					"test1": "hello",
					"test2": "world",
				},
				VALUES,
			},
			tFunc:       printEntity,
			tPipeOutput: "hello world\n",
		},
	}
	testPackageMethod(tt, t)
}
func TestIsHelpRequested(t *testing.T) {
	tt := []ttFStruct{
		{
			tName: "Test help is not requested",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				[]string{"sd", "this"},
			},
			tFunc:   isHelpRequested,
			tOutput: false,
		},
		{
			tName: "Test help is requested using \"help\"",
			tInput: []T{
				flag.NewFlagSet(DELETE, flag.ExitOnError),
				[]string{"sd", "save", "help"},
			},
			tFunc:       isHelpRequested,
			tOutput:     true,
			tPipeOutput: "save\tSave/update a command as a speed dial key\n",
		},
		{
			tName: "Test help is requested using \"-h\"",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				[]string{"sd", "save", "-h"},
			},
			tFunc:       isHelpRequested,
			tOutput:     true,
			tPipeOutput: "save\tSave/update a command as a speed dial key\n",
		},
		{
			tName: "Test help is requested using \"--help\"",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				[]string{"sd", "save", "--help"},
			},
			tFunc:       isHelpRequested,
			tOutput:     true,
			tPipeOutput: "save\tSave/update a command as a speed dial key\n",
		},
		{
			tName: "Test help is requested using \"--help\" following an argument",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				[]string{"save", "save", "sok", "--help"},
			},
			tFunc:       isHelpRequested,
			tOutput:     true,
			tPipeOutput: "save\tSave/update a command as a speed dial key\n",
		},
	}
	testPackageMethod(tt, t)
}

func TestReadFile(t *testing.T) {
	keyFile = "./test/.dial_keys_valid"
	tt := []ttFStruct{
		{
			tName:  "Test read file which exists and is JSON valid",
			tInput: []T{},
			tFunc:  readFile,
			tOutput: map[string]string{
				"something": "new",
				"hello":     "echo world",
			},
		},
	}
	testPackageMethod(tt, t)

	keyFile = "./test/.dial_keys_invalid"
	tt = []ttFStruct{
		{
			tName:   "Test read file which exists but is JSON invalid",
			tInput:  []T{},
			tFunc:   readFile,
			tOutput: map[string]string{},
		},
	}
	testPackageMethod(tt, t)

	keyFile = "./test/.dial_keys_does_not_exist"
	tt = []ttFStruct{
		{
			tName:   "Test read file which does not exists",
			tInput:  []T{},
			tFunc:   readFile,
			tOutput: map[string]string{},
		},
	}
	testPackageMethod(tt, t)
}

func TestParseCMD(t *testing.T) {
	tt := []ttFStruct{
		{
			tName: "Test parse CMD",
			tInput: []T{
				"echo {1} {2} from sd",
				[]string{
					"hello",
					"world",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo hello world from sd",
		},
		{
			tName: "Test parse CMD with one default arg",
			tInput: []T{
				"echo {1|test} from sd",
				[]string{},
			},
			tFunc:   parseCmd,
			tOutput: "echo test from sd",
		},
		{
			tName: "Test parse CMD with multiple same default arg",
			tInput: []T{
				"echo {1|test} {1|test} {1|test} from sd",
				[]string{},
			},
			tFunc:   parseCmd,
			tOutput: "echo test test test from sd",
		},
		{
			tName: "Test parse CMD with multiple same default arg",
			tInput: []T{
				"echo {1|test} {1|test} {1|test} from sd",
				[]string{
					"something",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo something something something from sd",
		},
		{
			tName: "Test parse CMD with more args than parameters",
			tInput: []T{
				"echo {1} {2} from sd",
				[]string{
					"something",
				},
			},
			tFunc:       parseCmd,
			tOutput:     "",
			tPipeOutput: "Cannot parse cmd: echo {1} {2} from sd, not enough arguments: [something]",
		},
		{
			tName: "Test parse without variable expansion",
			tInput: []T{
				"echo this is something",
				[]string{
					"test",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo this is something test",
		},
		{
			tName: "Test parse without variable expansion but added arguments",
			tInput: []T{
				"echo this is something",
				[]string{
					"and",
					"something",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo this is something and something",
		},
		{
			tName: "Test parse with variable expansion",
			tInput: []T{
				"echo {1} is {2} {3|crazy} something",
				[]string{
					"something",
					"new",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo something is new crazy something",
		},
		{
			tName: "Test parse with variable expansion replacement",
			tInput: []T{
				"echo {1} is {2} {3|crazy} something",
				[]string{
					"something",
					"new",
					"not",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo something is new not something",
		},
		{
			tName: "Test parse with variable expansion replacement and added arguments",
			tInput: []T{
				"echo {1} is {2} {3|crazy} something",
				[]string{
					"something",
					"new",
					"not",
					"whatever",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo something is new not something whatever",
		},
		{
			tName: "Test parse with multiple variable expansion replacement",
			tInput: []T{
				"echo {1} is {2} {3|crazy} {4|ok} something",
				[]string{
					"something",
					"new",
					"not",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo something is new not ok something",
		},
		{
			tName: "Test parse with all variable expansion replacement",
			tInput: []T{
				"echo {1} is {2} {3|crazy} {4|ok} something",
				[]string{
					"something",
					"new",
					"not",
					"whatever",
				},
			},
			tFunc:   parseCmd,
			tOutput: "echo something is new not whatever something",
		},
	}
	testPackageMethod(tt, t)
}

func TestIsValidSave(t *testing.T) {
	tt := []ttFStruct{
		{
			tName: "Test parse CMD valid",
			tInput: []T{
				"echo {1} {2|test} from sd",
			},
			tFunc:   isValidSave,
			tOutput: true,
		},
		{
			tName: "Test parse CMD not valid",
			tInput: []T{
				"echo {1|test} {2} from sd",
			},
			tFunc:   isValidSave,
			tOutput: false,
		},
		{
			tName: "Test parse multiple args CMD invalid",
			tInput: []T{
				"echo {1} {2} {3|test} {4} from sd",
			},
			tFunc:   isValidSave,
			tOutput: false,
		},
		{
			tName: "Test parse multiple args CMD valid",
			tInput: []T{
				"echo {1} {2} {3|test} {4|ok} from sd",
			},
			tFunc:   isValidSave,
			tOutput: true,
		},
		{
			tName: "Test simple command",
			tInput: []T{
				"echo from sd",
			},
			tFunc:   isValidSave,
			tOutput: true,
		},
		{
			tName: "Test simple command",
			tInput: []T{
				"echo {1} sd",
			},
			tFunc:   isValidSave,
			tOutput: true,
		},
	}
	testPackageMethod(tt, t)
}

func TestDelete(t *testing.T) {
	keyFile = "./test/.dial_keys_valid"
	writeFile = func(sdMap map[string]string) {}
	tt := []ttFStruct{
		{
			tName: "Test delete command with insufficient args",
			tInput: []T{
				flag.NewFlagSet(DELETE, flag.ExitOnError),
				"",
			},
			tFunc:   deleted,
			tOutput: 1,
		},
		{
			tName: "Test delete unknown command",
			tInput: []T{
				flag.NewFlagSet(DELETE, flag.ExitOnError),
				"does_not_exists",
			},
			tFunc:       deleted,
			tOutput:     1,
			tPipeOutput: "cannot execute command: " + DELETE + ", unknown key " + "does_not_exists",
		},
		{
			tName: "Test delete command which exists",
			tInput: []T{
				flag.NewFlagSet(DELETE, flag.ExitOnError),
				"hello",
			},
			tFunc:       deleted,
			tOutput:     0,
			tPipeOutput: "deleted the key: hello from speed dial keys",
		},
	}
	testPackageMethod(tt, t)
}

func TestSave(t *testing.T) {
	keyFile = "./test/.dial_keys_valid"
	writeFile = func(sdMap map[string]string) {}
	tt := []ttFStruct{
		{
			tName: "Test save command with bad key 1",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				"this wont work",
				"echo hello world",
			},
			tFunc:   save,
			tOutput: 1,
		},
		{
			tName: "Test save command with bad val",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				"this wont work",
				"",
			},
			tFunc:   save,
			tOutput: 1,
		},
		{
			tName: "Test save command with bad key 2",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				"",
				"this wont work",
			},
			tFunc:   save,
			tOutput: 1,
		},
		{
			tName: "Test save command with valid content",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				"test",
				"echo hello world",
			},
			tFunc:       save,
			tPipeOutput: "Saved key test as value: echo hello world",
			tOutput:     0,
		},
		{
			tName: "Test save command with invalid content",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				"test",
				"echo {1|test} {2}",
			},
			tFunc:       save,
			tPipeOutput: "cannot save key: \"test\", value: \"echo {1|test} {2}\" contains default argument preceeding regular argument",
			tOutput:     1,
		},
	}
	testPackageMethod(tt, t)
}

func TestCommandsWithNoExistingFile(t *testing.T) {
	keyFile = "./test/.dial_keys_does_not_exist"
	tt := []ttFStruct{
		{
			tName: "Test save command with file does not exist",
			tInput: []T{
				flag.NewFlagSet(SAVE, flag.ExitOnError),
				"this",
				"echo hello world",
			},
			tFunc:       save,
			tPipeOutput: "Saved key this as value: echo hello world",
			tOutput:     0,
			tCleanup:    "./test/.dial_keys_does_not_exist",
		},
		{
			tName: "Test delete command with file which does not exist",
			tInput: []T{
				flag.NewFlagSet(DELETE, flag.ExitOnError),
				"test",
			},
			tFunc:   deleted,
			tOutput: 1,
		},
		{
			tName: "Test get command with file which does not exist",
			tInput: []T{
				flag.NewFlagSet(GET, flag.ExitOnError),
				true,
				false,
			},
			tFunc:   get,
			tOutput: 1,
		},
		{
			tName: "Test execute command with file which does not exist",
			tInput: []T{
				"test",
				[]string{"whet", "sij"},
			},
			tFunc:   execute,
			tOutput: 1,
		},
	}
	testPackageMethod(tt, t)
}

func TestGet(t *testing.T) {
	keyFile = "./test/.dial_keys_valid"
	tt := []ttFStruct{
		{
			tName: "Test get command with both key and val",
			tInput: []T{
				flag.NewFlagSet(GET, flag.ExitOnError),
				true,
				true,
			},
			tFunc:   get,
			tOutput: 1,
		},
		{
			tName: "Test get command with key",
			tInput: []T{
				flag.NewFlagSet(GET, flag.ExitOnError),
				true,
				false,
			},
			tFunc:       get,
			tOutput:     0,
			tPipeOutput: "hello something\n",
		},
		{
			tName: "Test get command with val",
			tInput: []T{
				flag.NewFlagSet(GET, flag.ExitOnError),
				false,
				true,
			},
			tFunc:       get,
			tOutput:     0,
			tPipeOutput: "echo world new\n",
		},
	}
	testPackageMethod(tt, t)
}

func TestExecute(t *testing.T) {
	keyFile = "./test/.dial_keys_valid"
	execCmd = func(cmd string) int {
		return 0
	}
	tt := []ttFStruct{
		{
			tName: "Test execute that does not exist",
			tInput: []T{
				"not_exists",
				[]string{},
			},
			tFunc:       execute,
			tOutput:     1,
			tPipeOutput: "cannot execute command: unknown key \"not_exists\"\n",
		},
		{
			tName: "Test execute command that does exist",
			tInput: []T{
				"hello",
				[]string{},
			},
			tFunc:   execute,
			tOutput: 0,
		},
	}
	testPackageMethod(tt, t)
}

func TestExport(t *testing.T) {
	keyFile = "./test/.dial_keys_valid"
	transferFile = func(ip string, privateKeyFile string, user string, sshAlias string) int {
		return 0
	}
	exportToAlias = func() {}
	tt := []ttFStruct{
		{
			tName: "Test export command without destination",
			tInput: []T{
				flag.NewFlagSet(EXPORT, flag.ExitOnError),
				false,
				"",
				"",
				"",
				"",
			},
			tFunc:   export,
			tOutput: 1,
		},
		{
			tName: "Test export command with destination alias",
			tInput: []T{
				flag.NewFlagSet(EXPORT, flag.ExitOnError),
				false,
				"",
				"",
				"",
				"myAlias",
			},
			tFunc:   export,
			tOutput: 0,
		},
		{
			tName: "Test export command with destination ip",
			tInput: []T{
				flag.NewFlagSet(EXPORT, flag.ExitOnError),
				false,
				"127.0.0.1",
				"",
				"",
				"",
			},
			tFunc:   export,
			tOutput: 0,
		},
		{
			tName: "Test export command with destination ip and alias",
			tInput: []T{
				flag.NewFlagSet(EXPORT, flag.ExitOnError),
				false,
				"127.0.0.1",
				"",
				"",
				"myAlias",
			},
			tFunc:   export,
			tOutput: 1,
		},
		{
			tName: "Test export command with local export to alias format",
			tInput: []T{
				flag.NewFlagSet(EXPORT, flag.ExitOnError),
				true,
				"127.0.0.1",
				"",
				"",
				"",
			},
			tFunc:   export,
			tOutput: 0,
		},
	}
	testPackageMethod(tt, t)
}
