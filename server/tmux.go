package server

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/kahgeh/lemonade/param"
)

type Tmux struct{}

func (_ *Tmux) SendKeys(p *param.TmuxSendKeysParam, _ *struct{}) error {
	<-connCh
	log.Printf("SendKeys: %v", p)
	tmuxPath := "/opt/homebrew/bin/tmux"

	cmd := exec.Command(tmuxPath, "send-keys", "-t", p.Target, p.Keys)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running tmux command: %v", err)
		log.Printf("Command output: %s", string(output))
		return fmt.Errorf("tmux command failed: %v (output: %s)", err, string(output))
	}

	if p.SelectTargetPane {
		cmd := exec.Command(tmuxPath, "select-pane", "-t", p.Target)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to select-pane", err)
		}
	}

	log.Println("Tmux send-keys command executed successfully")

	return nil

}
