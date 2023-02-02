package promptcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type RunnerFunc func(ctx context.Context, args []string)

type Command struct {
	Command     string
	Description string
	Run         RunnerFunc
}

var exitCommand = "exit"
var exitCommandDescription = "Exit program"

//type AppState struct{}

type PromptCMD struct {
	commands Command

	suggests          []prompt.Suggest
	mapRunnersCommand map[string]RunnerFunc
}

func New(commands []Command) *PromptCMD {
	promptcmd := PromptCMD{}

	suggests := make([]prompt.Suggest, 0, len(commands))
	mapRunnersCommand := make(map[string]RunnerFunc, len(commands))
	for _, command := range commands {
		suggest := prompt.Suggest{
			Text:        command.Command,
			Description: command.Description,
		}

		suggests = append(suggests, suggest)
		mapRunnersCommand[command.Command] = command.Run
	}

	suggests = append(suggests, prompt.Suggest{Text: exitCommand, Description: exitCommandDescription})

	promptcmd.suggests = suggests
	promptcmd.mapRunnersCommand = mapRunnersCommand

	return &promptcmd
}

func (p *PromptCMD) Completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(p.suggests, d.GetWordBeforeCursor(), true)
}

func (p *PromptCMD) Executor(s string) {
	command, args := p.parseCommand(s)
	ctx := context.Background()

	if command == "" {
		return
	}

	if command == "exit" {
		return
	}

	runner, ok := p.mapRunnersCommand[command]

	if !ok {
		fmt.Println("Command not found")
		return
	}

	runner(ctx, args)
}

func (p *PromptCMD) ExitChecker(in string, breakline bool) bool {
	return strings.TrimSpace(in) == exitCommand && breakline
}

func (p *PromptCMD) parseCommand(arg string) (string, []string) {
	args := strings.Split(strings.TrimSpace(arg), " ")

	if len(args) == 0 {
		return "", []string{}
	}

	command := args[0]
	params := args[1:]

	return command, params
}
