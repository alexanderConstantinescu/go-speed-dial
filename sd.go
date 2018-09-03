package main

import (
	"flag"
	"fmt"
	"strings"
	"sort"
	"log"
	"os"
	"os/exec"
	"os/user"
	"syscall"
	"io/ioutil"
	"encoding/json"
	"runtime"
	"strconv"
)

var keyFile string = getHomeDir() + string(os.PathSeparator) + ".dial_keys"

var (
	keyTableTitle string = "Key"
	valueTableTitle string = "Value"
	overflowIndicator string = "..."
	maxKey int = len(keyTableTitle)
	maxVal int = len(valueTableTitle)
	overflowIndicatorLen int = len(overflowIndicator)
)

var (
	get string = "get"
	keys string = "keys"
	save string  = "save"
	update string = "update"
	del string = "delete"
	export string = "export"
	list string = "list"
	help string = "help"
	helpShort string = "-h"
	helpShorter string = "--help"
)

var helpText map[string]string = map[string]string{
	save : "save\tSave a command as a speed dial key",
	update : "update\tUpdate a saved speed dial key",
	del : "delete\tDelete a saved speed dial key",
	get : "get\tGet speed dial entities (keys, values) as a whitespace separated list. Useful for the creation of helper functions (bash completion for ex).",
	export : "export\tExport your .dial_key file to another remote location",
	list : "list\tList all dial keys",
	help : "help\tPrint this help",
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

func readFile(checkRequired bool) map[string]string {

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		if checkRequired {
			fmt.Println("No key has been saved, yet. Go ahead and save a command first.")
			os.Exit(1)
		 }
		if err := ioutil.WriteFile(keyFile, []byte(`{}`),  0644); err != nil  {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	f, err := ioutil.ReadFile(keyFile)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var speedDialStruct map[string]string

	if err := json.Unmarshal(f, &speedDialStruct); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return speedDialStruct

}

func writeFile(speedDialStruct map[string]string) {

	speedDialJson, err := json.Marshal(speedDialStruct)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(keyFile, speedDialJson,  0644); err != nil  {
		log.Fatal(err)
		os.Exit(1)
	}

}

func printMainHelp() {
	fmt.Println("Speed dial: a CLI intended to help you remember and faster execute commands you typically write, over and over again.")
	fmt.Println("\nCommands:\n")
	fmt.Println(helpText[save])
	fmt.Println(helpText[update])
	fmt.Println(helpText[del])
	fmt.Println(helpText[get])
	fmt.Println(helpText[export])
	fmt.Println(helpText[list])
	fmt.Println(helpText[help])
}

func isHelpRequested(command *flag.FlagSet, args []string) bool {
	for _, v := range args {
		if v == help || v == helpShort || v == helpShorter {
			fmt.Println(helpText[args[1]])
			command.PrintDefaults()
			return true
		}
	}
	return false
}

func verifyKey(sdMap map[string]string, key string) (string, bool) {
	val, exists := sdMap[key]
	return val, exists
}

func printEntity(sdMap map[string]string, entity string) {
	entityValues := make([]string, 0, len(sdMap))
	for k, v := range sdMap {
		if entity == keys {
			entityValues = append(entityValues, k)
		} else {
			entityValues = append(entityValues, v)
		}
	}
	fmt.Printf("%s", strings.Join(entityValues, " "))
}

func evalCmd(cmd string, args []string) string {
	cmdResult := exec.Command(cmd, args...)
	cmdResult.Stdin = os.Stdin
	out, err := cmdResult.Output()
	if err != nil {
		log.Fatal("Could not evaluate cmd: ", err)
		return ""
	}
	return string(out)
}

func execCmd(cmd string) {
	binary, err := exec.LookPath("bash")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = syscall.Exec(binary, []string{"bash", "-c", cmd}, os.Environ())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func parseCmd(val string) string {
	if len(os.Args) > 2 {
		val = strings.Replace(val, "{}", os.Args[2], -1)
	}
	return strings.Replace(val, "\\", "", -1)
}

func printAsTable(sdMap map[string]string, listLong bool) {
	padding := 5

	windowSizes := evalCmd("stty", []string{"size"})
	windowSize := strings.Split(windowSizes, " ")
	windowWidth, _ := strconv.Atoi(strings.Replace(windowSize[1], "\n", "", 1))


        abs := func(val int) int {
                if val < 0 {
                         return -val
                }
                return val
        }

	var sortedKeys []string
	for key, _ := range sdMap {
		keyLen := len(key)
		if keyLen > maxKey{
			maxKey = keyLen
		}
		sortedKeys = append(sortedKeys, key)
	}

	for key, value := range sdMap {
		valueLen := len(value)
		if valueLen > maxVal {
			maxVal = valueLen
		}
		overflow := windowWidth - maxKey - valueLen - (4 * padding + 3)
		if overflow < 0 && !listLong {
			maxVal = valueLen - abs(overflow)
			sdMap[key] = value[0: maxVal - overflowIndicatorLen] + overflowIndicator
		}
	}
	sort.Strings(sortedKeys)

	printTableRow := func(key string, value string) {
		leftPaddingSpacing := strings.Repeat(" ", padding)
		keyRightPaddingSpacing := strings.Repeat(" ", abs(maxKey - len(key)) + padding)
		valueRightPaddingSpacing := strings.Repeat(" ", abs(maxVal - len(value)) + padding)
		fmt.Println("|" + leftPaddingSpacing + key + keyRightPaddingSpacing + "|" + leftPaddingSpacing +  value + valueRightPaddingSpacing + "|")

	}

	func() {
		tableHorizontalBorder := strings.Repeat("-", maxVal + maxKey + (4 * padding + 3))
		fmt.Println(tableHorizontalBorder)
		printTableRow(keyTableTitle, valueTableTitle)
		fmt.Println(tableHorizontalBorder)
		for _, sKey := range sortedKeys {
			printTableRow(sKey, sdMap[sKey])
		}
		fmt.Println(tableHorizontalBorder)
	}()
}

func readPrivateKeyFile(file string) []byte {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return content
}

func transferFile(ip string, privateKeyFile string, user string, sshAlias string) {
	cmd := ""
	if sshAlias != "" {
		cmd = "scp " + keyFile + " " + sshAlias + ":"
	} else {
		cmd = "scp -i " + privateKeyFile + " " + keyFile + " " + user + "@" + ip + ":"
	}
	execCmd(cmd)
}

func main() {

	user,err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	saveCommand := flag.NewFlagSet(save, flag.ExitOnError)
	updateCommand := flag.NewFlagSet(update, flag.ExitOnError)
	deleteCommand := flag.NewFlagSet(del, flag.ExitOnError)
	exportCommand := flag.NewFlagSet(export, flag.ExitOnError)

	getCommand := flag.NewFlagSet(get, flag.ExitOnError)
	getKeyPtr := getCommand.Bool("key", false, "Get keys as a whitespace separated list")
	getValuePtr := getCommand.Bool("value", false, "Get values as whitespace separated list")

	listCommand := flag.NewFlagSet(list, flag.ExitOnError)
	listLongPtr := listCommand.Bool("l", false, "List saved commands in a non-truncated format independent of screen size")

	saveKeyPtr := saveCommand.String("key", "", "Key to save. (Required)")
	saveValPtr := saveCommand.String("val", "", "Val to map key to. (Required)\n\n" +
						    "Note:\n" +
						    "The key naming: \"keys\" is reserved and white space characters are not allowed in the key naming. \n" +
						    "Special characters such as: $ - for variable reference or ' - single quoutes need to be escaped using the \\ character\n\t" +
						    "Ex: speedial save -key ex -val \"for i in {1,2,3}; do echo $\\i; done\"\n\t" +
						    "or: speedial save -key ex2 -val \"echo I\\'m home\"\n\n" +
						    "Save is implemented to save non-existing keys. To update a key use command: update\n")

	updateKeyPtr := updateCommand.String("key", "", "Key to update. (Required)")
	updateValPtr := updateCommand.String("val", "", "Value to update key with. (Required)\n\n" +
							"Note:\n" +
							"Update is implemented to update existing keys. To save a new use command: save")

	deleteKeyPtr := deleteCommand.String("key", "", "Key to delete. (Required)")

	exportIp := exportCommand.String("ip", "", "Destination IP to transfer file to. (Required if no SSH alias)")
	exportPrivateKeyFile := exportCommand.String("id", user.HomeDir + "/.ssh/id_rsa", "Specific private key file to use. (Required if no SSH alias)")
	exportUser := exportCommand.String("user", user.Username, "User to connect with to remote machine. (Required if no SSH alias)")
	exportSshAlias := exportCommand.String("ssh", "", "SSH alias - useful in case of multi-hop export")

	if (len(os.Args) < 2) {
		fmt.Println("A subcommand or execution key is required")
		printMainHelp()
		os.Exit(1)
	}

	switch os.Args[1] {

		case save:
			saveCommand.Parse(os.Args[2:])
			if isHelpRequested(saveCommand, os.Args) {
				os.Exit(0)
			}
		case update:
			updateCommand.Parse(os.Args[2:])
			if isHelpRequested(updateCommand, os.Args) {
				os.Exit(0)
			}
		case del:
			deleteCommand.Parse(os.Args[2:])
			if isHelpRequested(deleteCommand, os.Args) {
				os.Exit(0)
			}
		case export:
			exportCommand.Parse(os.Args[2:])
			if isHelpRequested(exportCommand, os.Args) {
				os.Exit(0)
			}
		case get:
			getCommand.Parse(os.Args[2:])
			if isHelpRequested(getCommand, os.Args) {
				os.Exit(0)
			}
		case list:
			listCommand.Parse(os.Args[2:])
			if isHelpRequested(listCommand, os.Args) {
				os.Exit(0)
			}
		case help, helpShort, helpShorter:
			printMainHelp()
			os.Exit(0)
		default:
			sdMap := readFile(true)
			val,_ := verifyKey(sdMap, os.Args[1])
			val = parseCmd(val)
			execCmd(val)
	}


	if saveCommand.Parsed() {

		if *saveKeyPtr == "" || *saveValPtr == "" || *saveKeyPtr == "keys" || strings.Contains(*saveKeyPtr, " ") {
			saveCommand.PrintDefaults()
			os.Exit(1)
		}
		sdMap := readFile(false)
		_, exists := verifyKey(sdMap, *saveKeyPtr)
		if !exists {
			sdMap[*saveKeyPtr] = *saveValPtr
			writeFile(sdMap)

			fmt.Printf("Saved key %s as value: %s\n", *saveKeyPtr, *saveValPtr)
			os.Exit(0)
		}
		fmt.Printf("Cannot execute command %s, key: %s exists. Use update instead.\n", save, *saveKeyPtr)
		os.Exit(1)
	}

	if updateCommand.Parsed() {

		if *updateKeyPtr == "" || *updateValPtr == "" {
			updateCommand.PrintDefaults()
			os.Exit(1)
		}

		sdMap := readFile(true)
		_, exists := verifyKey(sdMap, *updateKeyPtr)
		if exists {
			sdMap[*updateKeyPtr] = *updateValPtr
			writeFile(sdMap)

			fmt.Printf("Updated key %s as value: %s\n", *updateKeyPtr, *updateValPtr)
			os.Exit(0)
		}
		fmt.Printf("Cannot execute command: %s, unknown key %s\n", update, *updateKeyPtr)
		os.Exit(1)
	}

	if deleteCommand.Parsed() {

		if *deleteKeyPtr == "" {
			deleteCommand.PrintDefaults()
			os.Exit(1)
		}

		sdMap := readFile(true)
		_, exists := verifyKey(sdMap, *deleteKeyPtr)
		if exists {
			delete(sdMap, string(*deleteKeyPtr))
			writeFile(sdMap)

			fmt.Printf("Deleted the key %s from speed dial keys\n", *deleteKeyPtr)
			os.Exit(0)
		}
		fmt.Printf("Cannot execute command: %s, unknown key %s\n", del, *deleteKeyPtr)
		os.Exit(1)
	}

	if listCommand.Parsed() {

		sdMap := readFile(true)
		printAsTable(sdMap, *listLongPtr)
		os.Exit(0)
	}

	if getCommand.Parsed() {

		if len(os.Args) != 3 {
			getCommand.PrintDefaults()
			os.Exit(1)
		}

		sdMap := readFile(false)
		if *getKeyPtr {
			printEntity(sdMap, "keys")
		}
		if *getValuePtr {
			printEntity(sdMap, "values")
		}
		os.Exit(0)
	}

	if exportCommand.Parsed() {

		if *exportIp == "" && *exportSshAlias == "" {
			exportCommand.PrintDefaults()
			os.Exit(1)
		}

		transferFile(*exportIp, *exportPrivateKeyFile, *exportUser, *exportSshAlias)
		os.Exit(0)
	}
}
