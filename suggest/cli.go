package suggest

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"github.com/yusufcanb/tlm/explain"
	"github.com/yusufcanb/tlm/shell"
)

func (s *Suggest) before(_ *cli.Context) error {
	_, err := s.api.Version(context.Background())
	if err != nil {
		fmt.Println(shell.Err() + " " + err.Error())
		fmt.Println(shell.Err() + " Ollama connection failed. Please check your Ollama if it's running or configured correctly.")
		os.Exit(-1)
	}

	list, err := s.api.List(context.Background())
	if err != nil {
		fmt.Println(shell.Err() + " " + err.Error())
		fmt.Println(shell.Err() + " Ollama connection failed. Please check your Ollama if it's running or configured correctly.")
		os.Exit(-1)
	}

	found := false
	for _, model := range list.Models {
		if model.Name == s.modelfileName {
			found = true
			break
		}
	}

	if !found {
		fmt.Println(shell.Err() + " " + "tlm's suggest model not found.\n\nPlease run `tlm deploy` to deploy tlm models first.")
		os.Exit(-1)
	}

	return nil
}

func (s *Suggest) action(c *cli.Context) error {
	var responseText string
	var err error

	var t1, t2 time.Time

	prompt := c.Args().Get(0)
	spinner.New().
		Type(spinner.Line).
		Title(" Thinking...").
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))).
		Action(func() {
			t1 = time.Now()
			fmt.Println("using open ai 1")
			responseText, err = s.getCommandSuggestionFor(Stable, viper.GetString("shell"), prompt)
			t2 = time.Now()
		}).
		Run()

	if err != nil {
		fmt.Println(shell.Err()+" error getting suggestion:", err)
	}

	fmt.Printf(shell.SuccessMessage("┃ >")+" Thinking... (%s)\n", t2.Sub(t1).String())
	if len(s.extractCommandsFromResponse(responseText)) == 0 {
		fmt.Println(shell.WarnMessage("┃ >") + " No command found for given prompt..." + "\n")
		return nil
	}

	form := NewCommandForm(s.extractCommandsFromResponse(responseText)[0])
	err = form.Run()

	fmt.Println(shell.SuccessMessage("┃ > ") + form.command)
	if err != nil {
		fmt.Println(shell.WarnMessage("┃ > ") + "Aborted..." + "\n")
		return nil
	}

	if form.action == Execute {
		fmt.Println(shell.SuccessMessage("┃ > ") + "Executing..." + "\n")
		cmd, stdout, stderr := shell.Exec2(form.command)
		err = cmd.Run()
		if err != nil {
			fmt.Println(stderr.String())
			return nil
		}

		if stderr.String() != "" {
			fmt.Println(stderr.String())
			return nil
		}

		fmt.Println(stdout.String())
		return nil
	}

	if form.action == Explain {
		fmt.Println(shell.SuccessMessage("┃ > ") + "Explaining..." + "\n")

		exp := explain.New(s.api)
		fmt.Println("using open ai")
		err = exp.StreamExplanationFor(Stable, form.command)
		if err != nil {
			return err
		}

	} else {
		fmt.Println(shell.WarnMessage("┃ > ") + "Aborted..." + "\n")
	}

	return nil
}

func (s *Suggest) Command() *cli.Command {
	return &cli.Command{
		Name:        "suggest",
		Aliases:     []string{"s"},
		Usage:       "Suggests a command.",
		UsageText:   "tlm suggest <prompt>",
		Description: "suggests a command for given prompt.",
		Before:      s.before,
		Action:      s.action,
	}
}
