package install

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh/spinner"
	ollama "github.com/jmorganca/ollama/api"
	"github.com/yusufcanb/tlm/shell"
)

func (i *Install) installModelfile(name, modelfile string) error {
	var err error
	err = i.api.Create(context.Background(), &ollama.CreateRequest{Model: name, Modelfile: modelfile}, func(res ollama.ProgressResponse) error {
		return nil
	})
	return err
}

func (i *Install) deployTlm(suggestModelfile, explainModelfile string) {
	var err error

	_ = spinner.New().Type(spinner.Line).Title(" Getting latest deepseek-coder").Action(func() {
		err = i.api.Pull(context.Background(), &ollama.PullRequest{Model: "deepseek-coder"}, func(res ollama.ProgressResponse) error {
			return nil
		})
		if err != nil {
			fmt.Println("- Installing deepseek-coder. " + shell.Err())
			os.Exit(-1)
		}
	}).Run()
	fmt.Println("- Getting latest deepseek-coder. " + shell.Ok())

	// 6. Install the modelfile (Suggest)
	_ = spinner.New().Type(spinner.Line).Title(" Creating Modelfile for suggestions").Action(func() {
		err = i.installModelfile("suggest:7b", suggestModelfile)
		time.Sleep(1 * time.Second)
		if err != nil {
			fmt.Println("- Creating Modelfile for suggestions. " + shell.Err())
			os.Exit(-1)
		}
	}).Run()
	fmt.Println("- Creating Modelfile for suggestions. " + shell.Ok())

	// 7. Install the modelfile (Suggest)
	_ = spinner.New().Type(spinner.Line).Title(" Creating Modelfile for explanations").Action(func() {
		err = i.installModelfile("explain:7b", explainModelfile)
		time.Sleep(1 * time.Second)
		if err != nil {
			fmt.Println("- Creating Modelfile for explanations. " + shell.Err())
			os.Exit(-1)
		}
	}).Run()
	fmt.Println("- Creating Modelfile for explanations. " + shell.Ok())
}
