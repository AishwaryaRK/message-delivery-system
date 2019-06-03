package utility

import "os/exec"

func GenerateUUID() (string, error) {
	uuid, err := exec.Command("uuidgen").Output()
	return string(uuid), err
}
