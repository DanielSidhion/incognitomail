package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/danielsidhion/incognitomail"
	"github.com/hashicorp/logutils"
)

type arguments struct {
	configPath string
}

var (
	errWrongUsage = errors.New("wrong usage") // When the command has been invoked with wrong arguments or parameters

	cliArguments arguments
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-c|--config <path>] [command [arguments]]\n", os.Args[0])
		fmt.Printf("\n")
		fmt.Printf("if command is ommitted, will act as a server listening for connections\n\n")
		fmt.Printf("commands:\n")
		fmt.Printf("  new account <address>            \tcreates a new account with the given address\n")
		fmt.Printf("  new handle <secret>              \tcreates a new handle for the account with the given secret\n")
		fmt.Printf("  delete account <secret>          \tdeletes the account registered with the given secret\n")
		fmt.Printf("  delete handle <handle> <secret>  \tdeletes the given handle. Uses the given secret to confirm account ownership\n")
		fmt.Printf("  list <secret>                    \tlists all handles registered for the account with the given secret\n")
		fmt.Printf("  stop                             \tstops the current server process\n\n")
		fmt.Printf("options:\n")

		flag.PrintDefaults()
	}

	flag.StringVar(&cliArguments.configPath, "config", "", "path to a configuration file")
	flag.StringVar(&cliArguments.configPath, "c", "", "path to a configuration file (shorthand)")

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
}

func main() {
	flag.Parse()

	// Checking for config
	if cliArguments.configPath != "" {
		err := incognitomail.ReadConfigFromFile(cliArguments.configPath)

		if err != nil {
			log.Printf("[DEBUG] %s\n", err)
			fmt.Println("The program was unsuccessful due to an error.")
			os.Exit(1)
		}
	}

	success, err := parseAndExecuteCommand()
	if err == errWrongUsage {
		flag.Usage()
		os.Exit(2)
	}

	if !success {
		log.Printf("[DEBUG] %s\n", err)
		fmt.Println("The program was unsuccessful due to an error.")
		os.Exit(1)
	}
}

func parseAndExecuteCommand() (bool, error) {
	numCommands := flag.NArg()

	if numCommands == 0 {
		// Start server
		server, err := incognitomail.NewServer()
		if err != nil {
			return false, err
		}

		server.Start()
		server.Wait()
		return true, nil
	}

	c := incognitomail.CreateRPCServiceClient()

	switch flag.Arg(0) {
	case "stop":
		_, err := c.Call("Stop", nil)
		if err != nil {
			return false, err
		}

		fmt.Println("Stopped server")
	case "list":
		if flag.NArg() != 2 {
			return false, errWrongUsage
		}

		res, err := c.Call("ListHandles", flag.Arg(1))
		if err != nil {
			return false, err
		}

		handles := res.([]string)

		for _, handle := range handles {
			fmt.Println(handle)
		}
	default:
		res, err := c.Call("SendCommand", strings.Join(flag.Args(), " "))
		if err != nil {
			return false, err
		}

		fmt.Println(res)
	}

	return true, nil
}
