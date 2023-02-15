package promptcmd

import (
	"reflect"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/shreyner/gophkeeper/internal/client/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("corrected create promptcmd", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		s := state.New()

		commands := []Command{
			{
				Command:     "login",
				Description: "description",
				Run:         nil,
				Auth:        CommandAuthNot,
			},
			{
				Command:     "logout",
				Description: "description",
				Run:         nil,
				Auth:        CommandAuthNeed,
			},
			{
				Command:     "sync",
				Description: "description",
				Run:         nil,
				Auth:        CommandAuthAny,
			},
		}

		pcmd := New(s, commands)

		require.Len(pcmd.suggestsBeforeAuth, 3)
		require.Len(pcmd.suggestsAfterAuth, 3)

		assert.Equal(pcmd.suggestsBeforeAuth[0].Text, "login")
		assert.Equal(pcmd.suggestsAfterAuth[len(pcmd.suggestsAfterAuth)-1].Text, "exit")

		assert.Equal(pcmd.suggestsAfterAuth[0].Text, "logout")
		assert.Equal(pcmd.suggestsAfterAuth[1].Text, "sync")

		assert.Equal(pcmd.suggestsAfterAuth[len(pcmd.suggestsAfterAuth)-1].Text, "exit")
	})

	t.Run("corrected create empty", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		s := state.New()

		commands := []Command{}

		pcmd := New(s, commands)

		require.Len(pcmd.suggestsBeforeAuth, 1)
		require.Len(pcmd.suggestsAfterAuth, 1)

		assert.Equal(pcmd.suggestsAfterAuth[len(pcmd.suggestsAfterAuth)-1].Text, "exit")
		assert.Equal(pcmd.suggestsAfterAuth[len(pcmd.suggestsAfterAuth)-1].Text, "exit")
	})
}

func TestPromptCMD_Completer(t *testing.T) {
	type fields struct {
		isAuth bool
	}
	type args struct {
		d prompt.Document
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []prompt.Suggest
	}{
		{
			name: "before auth with empty args",
			fields: fields{
				isAuth: false,
			},
			args: args{
				d: prompt.Document{
					Text: "",
				},
			},
			want: []prompt.Suggest{
				{
					Text:        "login",
					Description: "description",
				},
				{
					Text:        "sync",
					Description: "description",
				},
				{
					Text:        "exit",
					Description: "Exit program",
				},
			},
		},
		{
			name: "after auth with empty args",
			fields: fields{
				isAuth: true,
			},
			args: args{
				d: prompt.Document{
					Text: "",
				},
			},
			want: []prompt.Suggest{
				{
					Text:        "logout",
					Description: "description",
				},
				{
					Text:        "sync",
					Description: "description",
				},
				{
					Text:        "exit",
					Description: "Exit program",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := []Command{
				{
					Command:     "login",
					Description: "description",
					Run:         nil,
					Auth:        CommandAuthNot,
				},
				{
					Command:     "logout",
					Description: "description",
					Run:         nil,
					Auth:        CommandAuthNeed,
				},
				{
					Command:     "sync",
					Description: "description",
					Run:         nil,
					Auth:        CommandAuthAny,
				},
			}

			s := state.New()
			s.IsAuth = tt.fields.isAuth

			p := New(s, commands)

			if got := p.Completer(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Completer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptCMD_parseCommand(t *testing.T) {
	type args struct {
		arg string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []string
	}{
		{
			name:  "empty string",
			args:  args{""},
			want:  "",
			want1: []string{},
		},
		{
			name:  "empty string with spaces",
			args:  args{" "},
			want:  "",
			want1: []string{},
		},
		{
			name:  "with command string",
			args:  args{"login"},
			want:  "login",
			want1: []string{},
		},
		{
			name:  "with command string and spaces before",
			args:  args{" login"},
			want:  "login",
			want1: []string{},
		},
		{
			name:  "with command string and spaces after",
			args:  args{"login "},
			want:  "login",
			want1: []string{},
		},
		{
			name:  "with command string and spaces",
			args:  args{" login "},
			want:  "login",
			want1: []string{},
		},
		{
			name:  "with command and one arg",
			args:  args{"login alex"},
			want:  "login",
			want1: []string{"alex"},
		},
		{
			name:  "with command, spaces before command and one arg",
			args:  args{" login alex"},
			want:  "login",
			want1: []string{"alex"},
		},
		{
			name:  "with command, spaces after command and one arg",
			args:  args{"login alex"},
			want:  "login",
			want1: []string{"alex"},
		},
		{
			name:  "with command and two arg",
			args:  args{"login alex 123"},
			want:  "login",
			want1: []string{"alex", "123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromptCMD{}

			got, got1 := p.parseCommand(tt.args.arg)
			if got != tt.want {
				t.Errorf("parseCommand() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseCommand() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
