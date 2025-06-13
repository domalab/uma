package lib

import (
	"bufio"
	"log"
	"os/exec"

	"github.com/domalab/omniraid/daemon/logger"
)

// Callback - shell output handler
type Callback func(line string)

// Shell - shell executor
func Shell(command string, callback Callback) {
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Unable to stdoutpipe %s: %s", command, err)
	}

	scanner := bufio.NewScanner(out)

	if err = cmd.Start(); err != nil {
		log.Fatal("Unable to start command: ", err)
	}

	for scanner.Scan() {
		callback(scanner.Text())
	}

	// Wait for the result of the command; also closes our end of the pipe
	err = cmd.Wait()
	if err != nil {
		log.Fatal("Unable to wait for process to finish: ", err)
	}
}

// ShellEx -
func ShellEx(command string, callback Callback, arg ...string) {
	cmd := exec.Command(command, arg...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Unable to stdoutpipe %s: %s", command, err)
	}

	scanner := bufio.NewScanner(out)

	if err = cmd.Start(); err != nil {
		log.Fatal("Unable to start command: ", err)
	}

	for scanner.Scan() {
		callback(scanner.Text())
	}

	// Wait for the result of the command; also closes our end of the pipe
	err = cmd.Wait()
	if err != nil {
		log.Fatal("Unable to wait for process to finish: ", err)
	}
}

// // Run - simple shell execution
// func Run(name string, arg ...string) (string, error) {
// 	cmd := exec.Command(name, arg...)
// 	cmd.Env = os.Environ()
// 	cmd.Stderr = os.Stderr
// 	result, err := cmd.Output()
// 	return string(result), err
// }

// SafeShell - shell executor that doesn't crash on errors
func SafeShell(command string, callback Callback) {
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.StdoutPipe()
	if err != nil {
		logger.Yellow("Unable to stdoutpipe %s: %s", command, err)
		return
	}

	scanner := bufio.NewScanner(out)

	if err = cmd.Start(); err != nil {
		logger.Yellow("Unable to start command %s: %s", command, err)
		return
	}

	for scanner.Scan() {
		callback(scanner.Text())
	}

	// Wait for the result of the command; also closes our end of the pipe
	err = cmd.Wait()
	if err != nil {
		logger.Yellow("Command %s finished with error: %s", command, err)
	}
}

// SafeShellEx - safe shell executor with arguments
func SafeShellEx(command string, callback Callback, arg ...string) {
	cmd := exec.Command(command, arg...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		logger.Yellow("Unable to stdoutpipe %s: %s", command, err)
		return
	}

	scanner := bufio.NewScanner(out)

	if err = cmd.Start(); err != nil {
		logger.Yellow("Unable to start command %s: %s", command, err)
		return
	}

	for scanner.Scan() {
		callback(scanner.Text())
	}

	// Wait for the result of the command; also closes our end of the pipe
	err = cmd.Wait()
	if err != nil {
		logger.Yellow("Command %s finished with error: %s", command, err)
	}
}
