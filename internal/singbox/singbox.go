package singbox

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/utils"
)

func Start() apperr.Err {
	return execCommand("start")
}

func Stop() apperr.Err {
	return execCommand("stop")
}

func Restart() apperr.Err {
	return execCommand("restart")
}

func execCommand(action string) apperr.Err {
	disabled, err := utils.GetEnvBool("DISABLE_SINGBOX_INTERACTION", false)
	if err != nil {
		apperr.NewFatalErr("Singbox_EnvReadingFailed", err.Error())
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
		apperr.NewFatalErr("Singbox_CommandFailed", fmt.Sprintf("failed to execute the command '%s' with error '%s', output: '%s'", action, err, string(output)))
	}
	return nil
}
