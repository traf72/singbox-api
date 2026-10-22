package singbox

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/utils"
)

type ServiceStatus struct {
	Description         string `json:"description,omitempty"`
	LoadState           string `json:"load_state,omitempty"`
	ActiveState         string `json:"active_state,omitempty"`
	SubState            string `json:"sub_state,omitempty"`
	UnitFileState       string `json:"unit_file_state,omitempty"`
	Result              string `json:"result,omitempty"`
	MainPID             int    `json:"main_pid"`
	ExecMainStatus      int    `json:"exec_main_status"`
	Active              bool   `json:"active"`
	InteractionDisabled bool   `json:"interaction_disabled"`
}

func Start() apperr.Err {
	return execCommand("start")
}

func Stop() apperr.Err {
	return execCommand("stop")
}

func Restart() apperr.Err {
	return execCommand("restart")
}

func Status() (*ServiceStatus, apperr.Err) {
	disabled, err := utils.GetEnvBool("DISABLE_SINGBOX_INTERACTION", false)
	if err != nil {
		return nil, apperr.NewFatalErr("Singbox_EnvReadingFailed", err.Error())
	}
	if disabled {
		return &ServiceStatus{InteractionDisabled: true}, nil
	}
	if runtime.GOOS != "linux" {
		return nil, apperr.NewFatalErr("Singbox_InvalidOS", fmt.Sprintf("invalid OS '%s'", runtime.GOOS))
	}

	properties := strings.Join([]string{
		"Description",
		"LoadState",
		"ActiveState",
		"SubState",
		"UnitFileState",
		"Result",
		"MainPID",
		"ExecMainStatus",
	}, ",")
	cmd := exec.Command("systemctl", "show", "sing-box", "--no-pager", "--property="+properties)
	output, commandErr := cmd.CombinedOutput()
	if commandErr != nil {
		return nil, apperr.NewFatalErr("Singbox_CommandFailed", fmt.Sprintf("failed to get sing-box status: %s, output: '%s'", commandErr, string(output)))
	}

	status, parseErr := parseServiceStatus(output)
	if parseErr != nil {
		return nil, apperr.NewFatalErr("Singbox_StatusParseFailed", parseErr.Error())
	}
	return status, nil
}

func execCommand(action string) apperr.Err {
	disabled, err := utils.GetEnvBool("DISABLE_SINGBOX_INTERACTION", false)
	if err != nil {
		return apperr.NewFatalErr("Singbox_EnvReadingFailed", err.Error())
	}

	if disabled {
		log.Println("Singbox interaction is disabled")
		return nil
	}

	if runtime.GOOS != "linux" {
		return apperr.NewFatalErr("Singbox_InvalidOS", fmt.Sprintf("invalid OS '%s'", runtime.GOOS))
	}

	cmd := exec.Command("sudo", "systemctl", action, "sing-box")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return apperr.NewFatalErr("Singbox_CommandFailed", fmt.Sprintf("failed to execute the command '%s' with error '%s', output: '%s'", action, err, string(output)))
	}
	return nil
}

func parseServiceStatus(output []byte) (*ServiceStatus, error) {
	status := new(ServiceStatus)
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		key, value, found := bytes.Cut(line, []byte{'='})
		if !found {
			continue
		}
		switch string(key) {
		case "Description":
			status.Description = string(value)
		case "LoadState":
			status.LoadState = string(value)
		case "ActiveState":
			status.ActiveState = string(value)
		case "SubState":
			status.SubState = string(value)
		case "UnitFileState":
			status.UnitFileState = string(value)
		case "Result":
			status.Result = string(value)
		case "MainPID":
			pid, err := strconv.Atoi(string(value))
			if err != nil {
				return nil, fmt.Errorf("invalid MainPID value %q", value)
			}
			status.MainPID = pid
		case "ExecMainStatus":
			exitStatus, err := strconv.Atoi(string(value))
			if err != nil {
				return nil, fmt.Errorf("invalid ExecMainStatus value %q", value)
			}
			status.ExecMainStatus = exitStatus
		}
	}
	status.Active = status.ActiveState == "active"

	return status, nil
}
