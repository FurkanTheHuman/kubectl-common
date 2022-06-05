package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

type Command struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func (c Command) getDecodedCommand() string {
	command_byte, err := base64.StdEncoding.DecodeString(c.Command)
	if err != nil {
		log.Fatalln(err)
	}

	return string(command_byte)
}

func getConfigDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return dirname + "/.config/commands.json"
}
func createFileIfNotExist(file string) error { //create file if does not exists
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}
func getCommandsFromFile() ([]Command, error) {
	command_file := getConfigDir()
	err := createFileIfNotExist(command_file)
	if err != nil {
		return []Command{}, err
	}

	commands := []Command{}
	contents_byte, err := os.ReadFile(command_file)
	if err != nil {
		return []Command{}, err
	}

	json.Unmarshal(contents_byte, &commands)
	return commands, nil
}

func saveCommandToFile(command Command) error {
	command_file := getConfigDir()
	commands, err := getCommandsFromFile()
	if err != nil {
		return err
	}

	isFound := false
	for _, c := range commands {
		if c.Name == command.Name {
			fmt.Println("Command already exists, Updating the old command")
			c.Command = command.Command
			isFound = true
			break
		}
	}
	if !isFound {
		commands = append(commands, command)
	}
	contents_byte, err := json.Marshal(commands)
	if err != nil {
		return err
	}

	err = os.Truncate(command_file, 0)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(command_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(contents_byte)
	if err != nil {
		return err
	}

	return nil
}

func removeCommandFromFile(command_name string) error {
	command_file := getConfigDir()
	commands, err := getCommandsFromFile()
	if err != nil {
		return err
	}
	// filter out the command
	filtered_commands := []Command{}
	for _, c := range commands {
		if c.Name != command_name {
			filtered_commands = append(filtered_commands, c)
		}
	}
	// save the new commands
	contents_byte, err := json.Marshal(filtered_commands)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(command_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(contents_byte)
	if err != nil {
		return err
	}
	return nil
}

func listCommands() error {
	commands, err := getCommandsFromFile()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME\tCOMMAND")

	for _, command := range commands {
		formatted := fmt.Sprintf("%s\t%s", command.Name, command.getDecodedCommand())

		fmt.Fprintln(w, formatted)

	}
	w.Flush()

	return nil
}

func main() {
	app := &cli.App{
		Name:  "kubectl-common",
		Usage: "save common commands for kubectl and others",
		Action: func(c *cli.Context) error {
			var command string
			if c.NArg() > 0 {
				command = c.Args().Get(0)
				commands, err := getCommandsFromFile()
				if err != nil {
					return err
				}
				isFound := false
				for _, cmd := range commands {
					if cmd.Name == command {
						isFound = true
						//decode base64 command
						decoded, err := base64.StdEncoding.DecodeString(cmd.Command)
						if err != nil {
							return err
						}
						exec := exec.Command("sh", "-c", string(decoded))
						exec.Stdout = os.Stdout
						exec.Stderr = os.Stderr
						err = exec.Run()
						if err != nil {
							log.Fatal(err)
						}
					}
				}
				if !isFound {
					fmt.Println("command not found")
				}
			} else {
				listCommands()
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "add",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "name/alias of the command",
					},
					&cli.StringFlag{
						Name:    "command",
						Aliases: []string{"c"},
						Usage:   "the command to execute",
					},
				},
				Action: func(c *cli.Context) error {
					//note: command should base64 for security
					command := Command{
						Name:    c.String("name"),
						Command: base64.StdEncoding.EncodeToString([]byte(c.String("command"))),
					}
					fmt.Println("added command:", command.getDecodedCommand())
					if saveCommandToFile(command) != nil {
						log.Fatal("failed to save command:", command)

					}
					return nil
				},
			},
			{
				Name: "list",
				Action: func(c *cli.Context) error {
					listCommands()
					return nil
				},
			},
			{
				Name: "remove",
				Action: func(c *cli.Context) error {
					removeCommandFromFile(c.Args().Get(0))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
