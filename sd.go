package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

var _error = fmt.Errorf
var print = fmt.Printf
var exit = os.Exit

var rAll, _ = regexp.Compile("{\\S+}")
var rDef, _ = regexp.Compile("{[0-9]+\\|.*}")
var rReg, _ = regexp.Compile("{[0-9]+}")

var keyFile = getHomeDir() + string(os.PathSeparator) + ".dial_keys"
var aliasFile = getHomeDir() + string(os.PathSeparator) + ".bash_aliases"

var (
	keyTableTitle        = "Key"
	valueTableTitle      = "Value"
	overflowIndicator    = "..."
	maxKey               = len(keyTableTitle)
	maxVal               = len(valueTableTitle)
	overflowIndicatorLen = len(overflowIndicator)
)

var (
	GET         = "get"
	KEYS        = "keys"
	VALUES      = "values"
	SAVE        = "save"
	DELETE      = "delete"
	EXPORT      = "export"
	LIST        = "list"
	HELP        = "help"
	HELPSHORT   = "-h"
	HELPSHORTER = "--help"
)

var helpText = map[string]string{
	SAVE:   "save\tSave/update a command as a speed dial key",
	DELETE: "delete\tDelete a saved speed dial key",
	GET:    "get\tGet speed dial entities (keys, values) as a whitespace separated list. Useful for the creation of helper functions (bash completion for ex).",
	EXPORT: "export\tExport your .dial_key file to another remote location",
	LIST:   "list\tList all dial keys",
	HELP:   "help\tPrint this help",
}

func getHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func fileExists() bool {
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func readFile() map[string]string {
	f, err := ioutil.ReadFile(keyFile)
	if err != nil {
		_error(err.Error())
	}
	speedDialStruct := map[string]string{}
	if err := json.Unmarshal(f, &speedDialStruct); err != nil {
		_error(err.Error())
	}
	return speedDialStruct
}

var writeFile = func(speedDialStruct map[string]string) {
	speedDialJSON, err := json.Marshal(speedDialStruct)
	if err != nil {
		_error(err.Error())
	}
	if err := ioutil.WriteFile(keyFile, speedDialJSON, 0644); err != nil {
		_error(err.Error())
	}
}

var transferFile = func(ip string, privateSSHKeyFile string, user string, sshAlias string) int {
	cmd := ""
	if sshAlias != "" {
		cmd = fmt.Sprintf("scp %s %s", keyFile, sshAlias)
	} else {
		cmd = fmt.Sprintf("scp -i %s %s %s:%s", privateSSHKeyFile, keyFile, user, ip)
	}
	return execCmd(cmd)
}

var exportToAlias = func() {
	if fileExists() {
		sdMap := readFile()
		str := ""
		for key, value := range sdMap {
			str += fmt.Sprintf("alias %s=\"%s\"", key, value)
		}
		if err := ioutil.WriteFile(aliasFile, []byte(str), 0644); err != nil {
			_error(err.Error())
		}
		print("Wrote speed-dial content to %s as BASH aliases\n", aliasFile)
	}
}

var execCmd = func(cmd string) int {
	binary, err := exec.LookPath("bash")
	if err != nil {
		_error(err.Error())
	}

	err = syscall.Exec(binary, []string{"bash", "-c", cmd}, os.Environ())
	if err != nil {
		_error(err.Error())
	}
	return 0
}

func printMainHelp() {
	print("Speed dial: a CLI intended to help you remember and faster execute commands you typically write, over and over again.\n")
	print("Commands:\n")
	print("%s\n", helpText[SAVE])
	print("%s\n", helpText[DELETE])
	print("%s\n", helpText[GET])
	print("%s\n", helpText[EXPORT])
	print("%s\n", helpText[LIST])
	print("%s\n", helpText[HELP])
}

func isHelpRequested(command *flag.FlagSet, args []string) bool {
	for _, v := range args {
		if v == HELP || v == HELPSHORT || v == HELPSHORTER {
			print("%s\n", helpText[args[1]])
			command.PrintDefaults()
			return true
		}
	}
	return false
}

func printEntity(sdMap map[string]string, entity string) {
	entityValues := make([]string, 0, len(sdMap))
	for k, v := range sdMap {
		if entity == KEYS {
			entityValues = append(entityValues, k)
		} else {
			entityValues = append(entityValues, v)
		}
	}
	sort.Strings(entityValues)
	print("%s\n", strings.Join(entityValues, " "))
}

func evalCmd(cmd string) string {
	cmdArgs := []string{"-c", cmd}
	cmdResult := exec.Command("bash", cmdArgs...)
	cmdResult.Stdin = os.Stdin
	out, err := cmdResult.Output()
	if err != nil {
		print("could not evaluate cmd \"%s\": %v\n", cmd, err.Error())
	}
	return string(out)
}

func parseCmd(val string, args []string) string {
	if rAll.MatchString(val) {
		matchedVals := rAll.FindAllString(val, -1)
		for _, matchedVal := range matchedVals {
			if strings.Contains(matchedVal, "|") {
				if len(args) == 0 {
					defaultVal := strings.Split(matchedVal, "|")
					defaultVal = strings.Split(defaultVal[1], "}")
					val = strings.Replace(val, matchedVal, defaultVal[0], 1)
				} else {
					val = strings.Replace(val, matchedVal, args[0], 1)
				}
			} else {
				val = strings.Replace(val, matchedVal, args[0], 1)
			}
			if len(args) > 0 {
				_, args = args[0], args[1:]
			}
		}
	}
	if len(args) > 0 {
		val = val + " " + strings.Join(args, " ")
	}
	return strings.Replace(val, "\\", "", -1)
}

func isValidSave(cmd string) bool {
	regularIdx := rReg.FindAllStringIndex(cmd, len(cmd))
	defaultIdx := rDef.FindAllStringIndex(cmd, len(cmd))
	if len(regularIdx) == 0 {
		return true
	}
	if len(defaultIdx) == 0 {
		return true
	}
	return defaultIdx[0][0] > regularIdx[len(regularIdx)-1][1]
}

func printAsTable(sdMap map[string]string, listLong bool) {
	padding := 5
	ellipsed := false
	windowSizes := evalCmd("stty size")
	windowSize := strings.Split(windowSizes, " ")
	windowWidth, _ := strconv.Atoi(strings.Replace(windowSize[1], "\n", "", 1))

	abs := func(val int) int {
		if val < 0 {
			return -val
		}
		return val
	}

	var sortedKeys []string
	for key := range sdMap {
		keyLen := len(key)
		if keyLen > maxKey {
			maxKey = keyLen
		}
		sortedKeys = append(sortedKeys, key)
	}

	for key, value := range sdMap {
		valueLen := len(value)
		if valueLen > maxVal {
			maxVal = valueLen
		}
		overflow := windowWidth - maxKey - valueLen - (4*padding + 3)
		if overflow < 0 && !listLong {
			ellipsed = true
			maxVal = valueLen - abs(overflow)
			sdMap[key] = value[0:maxVal-overflowIndicatorLen] + overflowIndicator
		}
	}
	sort.Strings(sortedKeys)

	printTableRow := func(key string, value string) {
		leftPaddingSpacing := strings.Repeat(" ", padding)
		keyRightPaddingSpacing := strings.Repeat(" ", abs(maxKey-len(key))+padding)
		valueRightPaddingSpacing := strings.Repeat(" ", abs(maxVal-len(value))+padding)
		print("|" + leftPaddingSpacing + key + keyRightPaddingSpacing + "|" + leftPaddingSpacing + value + valueRightPaddingSpacing + "|\n")

	}

	tableHorizontalBorder := strings.Repeat("-", maxVal+maxKey+(4*padding+3))

	print(tableHorizontalBorder + "\n")
	printTableRow(keyTableTitle, valueTableTitle)

	print(tableHorizontalBorder + "\n")
	for _, sKey := range sortedKeys {
		printTableRow(sKey, sdMap[sKey])
	}

	print(tableHorizontalBorder + "\n")
	if ellipsed {
		print("Note: some values have been ellipsed. Add \"-l\" to see values in full.\n")
	}
}

func readPrivateKeyFile(file string) []byte {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		_error(err.Error())
	}
	return content
}

func execute(key string, args []string) int {
	if !fileExists() {
		return 1
	}
	sdMap := readFile()
	val, exists := sdMap[key]
	if !exists {
		print("cannot execute command: unknown key \"%s\"\n", key)
		return 1
	}
	val = parseCmd(val, args)

	return execCmd(val)
}

func save(command *flag.FlagSet, key, val string) int {
	if key == "" || val == "" || strings.Contains(key, " ") {
		command.PrintDefaults()
		return 1
	}
	if !fileExists() {
		writeFile(map[string]string{})
	}
	sdMap := readFile()
	if isValidSave(val) {
		sdMap[key] = val
		writeFile(sdMap)
		print("Saved key %s as value: %s", key, val)
		return 0
	}
	print("cannot save key: \"%s\", value: \"%s\" contains default argument preceeding regular argument", key, val)
	return 1
}

func deleted(command *flag.FlagSet, key string) int {
	if key == "" {
		command.PrintDefaults()
		return 1
	}
	if !fileExists() {
		return 1
	}
	sdMap := readFile()
	if _, exists := sdMap[key]; exists {
		delete(sdMap, string(key))
		writeFile(sdMap)
		print("deleted the key: %s from speed dial keys", key)
		return 0
	}
	print("cannot execute command: %s, unknown key %s", DELETE, key)
	return 1
}

func get(command *flag.FlagSet, getKey, getVal bool) int {
	if getKey && getVal {
		command.PrintDefaults()
		return 1
	}
	if !fileExists() {
		return 1
	}
	sdMap := readFile()
	if getKey {
		printEntity(sdMap, KEYS)
	}
	if getVal {
		printEntity(sdMap, VALUES)
	}
	return 0
}

func export(command *flag.FlagSet, exportToAliasFormat bool, exportIP, exportPrivateKeyFile, exportUser, exportSSHAlias string) int {
	if exportToAliasFormat {
		exportToAlias()
		return 0
	}
	if (exportIP == "" && exportSSHAlias == "") || (exportIP != "" && exportSSHAlias != "") {
		command.PrintDefaults()
		return 1
	}
	return transferFile(exportIP, exportPrivateKeyFile, exportUser, exportSSHAlias)
}

func list(listLong bool) int {
	if !fileExists() {
		return 1
	}
	sdMap := readFile()
	printAsTable(sdMap, listLong)
	return 0
}

func sd(user *user.User) int {

	saveCommand := flag.NewFlagSet(SAVE, flag.ExitOnError)
	deleteCommand := flag.NewFlagSet(DELETE, flag.ExitOnError)
	exportCommand := flag.NewFlagSet(EXPORT, flag.ExitOnError)

	getCommand := flag.NewFlagSet(GET, flag.ExitOnError)
	getKeyPtr := getCommand.Bool("key", false, "Get keys as a whitespace separated list")
	getValPtr := getCommand.Bool("val", false, "Get values as whitespace separated list")

	listCommand := flag.NewFlagSet(LIST, flag.ExitOnError)
	listLongPtr := listCommand.Bool("l", false, "List saved commands in a non-truncated format independent of screen size")

	saveKeyPtr := saveCommand.String("key", "", "Key to save. (Required)")
	saveValPtr := saveCommand.String("val", "", "Val to map key to. (Required)\n\n"+
		"Note:\n"+
		"White space characters are not allowed in the key naming. \n"+
		"Special characters such as: $ - for variable reference or ' - single quoutes need to be escaped using the \\ character\n\t"+
		"Ex: sd save -key ex -val \"for i in {1,2,3}; do echo $\\i; done\"\n\t"+
		"or: sd save -key ex2 -val \"echo I\\'m home\"\n\t"+
		"or: sd save -key ex3 -val \"echo {1} {2}\", which can be expanded as: sd ex3 hello world -> hello world")

	deleteKeyPtr := deleteCommand.String("key", "", "Key to delete. (Required)")

	exportIP := exportCommand.String("ip", "", "Destination IP to transfer file to. (Required if no SSH alias)")
	exportPrivateKeyFile := exportCommand.String("id", user.HomeDir+"/.ssh/id_rsa", "Specific private key file to use. (Required if no SSH alias)")
	exportUser := exportCommand.String("user", user.Username, "User to connect with to remote machine. (Required if no SSH alias)")
	exportSSHAlias := exportCommand.String("ssh", "", "SSH alias - useful in case of multi-hop export")
	exportToAliasFormat := exportCommand.Bool("to-alias", false, "Export to alias format and update "+user.HomeDir+"/.bash_aliases")

	exitCode := 0

	if len(os.Args) < 2 {
		print("A subcommand or execution key is required\n")
		printMainHelp()
		return 1
	}

	switch os.Args[1] {

	case SAVE:
		saveCommand.Parse(os.Args[2:])
		if isHelpRequested(saveCommand, os.Args) {
			return 0
		}
	case DELETE:
		deleteCommand.Parse(os.Args[2:])
		if isHelpRequested(deleteCommand, os.Args) {
			return 0
		}
	case EXPORT:
		exportCommand.Parse(os.Args[2:])
		if isHelpRequested(exportCommand, os.Args) {
			return 0
		}
	case GET:
		getCommand.Parse(os.Args[2:])
		if isHelpRequested(getCommand, os.Args) {
			return 0
		}
	case LIST:
		listCommand.Parse(os.Args[2:])
		if isHelpRequested(listCommand, os.Args) {
			return 0
		}
	case HELP, HELPSHORT, HELPSHORTER:
		printMainHelp()
		return 0
	default:
		exitCode = execute(os.Args[1], os.Args[2:])
	}

	if saveCommand.Parsed() {
		exitCode = save(saveCommand, *saveKeyPtr, *saveValPtr)
	}

	if deleteCommand.Parsed() {
		exitCode = deleted(deleteCommand, *deleteKeyPtr)
	}

	if listCommand.Parsed() {
		exitCode = list(*listLongPtr)
	}

	if getCommand.Parsed() {
		exitCode = get(getCommand, *getKeyPtr, *getValPtr)
	}

	if exportCommand.Parsed() {
		exitCode = export(exportCommand, *exportToAliasFormat, *exportIP, *exportPrivateKeyFile, *exportUser, *exportSSHAlias)
	}
	return exitCode
}

func main() {
	user, err := user.Current()
	if err != nil {
		_error(err.Error())
	}
	exit(sd(user))
}
