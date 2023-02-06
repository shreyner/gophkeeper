package promptcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/shreyner/gophkeeper/internal/client/state"
)

const (
	CommandAuthAny = iota
	CommandAuthNeed
	CommandAuthNot
)

type RunnerFunc func(ctx context.Context, args []string)

type Command struct {
	Command     string
	Description string
	Run         RunnerFunc
	Auth        int
}

var exitCommand = "exit"
var exitCommandDescription = "Exit program"

//type AppState struct{}

type PromptCMD struct {
	appState *state.State
	commands Command

	suggestsBeforeAuth []prompt.Suggest
	suggestsAfterAuth  []prompt.Suggest

	mapRunnersCommand map[string]RunnerFunc
}

func New(appState *state.State, commands []Command) *PromptCMD {
	promptcmd := PromptCMD{
		appState: appState,
	}

	suggestsBeforeAuth := make([]prompt.Suggest, 0)
	suggestsAfterAuth := make([]prompt.Suggest, 0)

	mapRunnersCommand := make(map[string]RunnerFunc, len(commands))

	for _, command := range commands {
		suggest := prompt.Suggest{
			Text:        command.Command,
			Description: command.Description,
		}

		if command.Auth == CommandAuthNeed {
			suggestsAfterAuth = append(suggestsAfterAuth, suggest)
		}

		if command.Auth == CommandAuthNot {
			suggestsBeforeAuth = append(suggestsBeforeAuth, suggest)
		}

		if command.Auth == CommandAuthAny {
			suggestsAfterAuth = append(suggestsAfterAuth, suggest)
			suggestsBeforeAuth = append(suggestsBeforeAuth, suggest)
		}

		mapRunnersCommand[command.Command] = command.Run
	}

	suggestsBeforeAuth = append(suggestsBeforeAuth, prompt.Suggest{Text: exitCommand, Description: exitCommandDescription})
	suggestsAfterAuth = append(suggestsAfterAuth, prompt.Suggest{Text: exitCommand, Description: exitCommandDescription})

	promptcmd.suggestsBeforeAuth = suggestsBeforeAuth
	promptcmd.suggestsAfterAuth = suggestsAfterAuth
	promptcmd.mapRunnersCommand = mapRunnersCommand

	return &promptcmd
}

func (p *PromptCMD) Completer(d prompt.Document) []prompt.Suggest {
	if p.appState.IsAuth {
		return prompt.FilterHasPrefix(p.suggestsAfterAuth, d.GetWordBeforeCursor(), true)
	}

	return prompt.FilterHasPrefix(p.suggestsBeforeAuth, d.GetWordBeforeCursor(), true)
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
