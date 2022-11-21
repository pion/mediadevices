package cmdsource

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/pion/mediadevices/internal/logging"
	"github.com/pion/mediadevices/pkg/prop"
)

var (
	errReadTimeout       = errors.New("read timeout")
	errInvalidCommand    = errors.New("invalid command")
	errUnsupportedFormat = errors.New("Unsupported frame format, no frame size function found")
)

var logger = logging.NewLogger("mediadevices/driver/cmdsource")

type cmdSource struct {
	cmdArgs     []string
	props       []prop.Media
	readTimeout uint32 // in seconds
	execCmd     *exec.Cmd
}

func init() {
	// No init. Call AddVideoCmdSource() or AddAudioCmdSource() to add a command source before calling getUserMedia().
}

func newCmdSource(command string, mediaProperties []prop.Media, readTimeout uint32) cmdSource {
	cmdArgs, err := shlex.Split(command) // split command string on whitespace, respecting quotes & comments
	if err != nil {
		panic(errInvalidCommand)
	}
	return cmdSource{
		cmdArgs:     cmdArgs,
		props:       mediaProperties,
		readTimeout: readTimeout,
	}
}

func (c *cmdSource) Open() error {
	c.execCmd = exec.Command(c.cmdArgs[0], c.cmdArgs[1:]...)
	return nil
}

func (c *cmdSource) Close() error {
	if c.execCmd == nil || c.execCmd.Process == nil {
		return nil
	}

	_ = c.execCmd.Process.Signal(os.Interrupt) // send SIGINT to process to stop it
	done := make(chan error)
	go func() { done <- c.execCmd.Wait() }()
	select {
	case err := <-done:
		return err // command exited normally or with an error code
	case <-time.After(3 * time.Second):
		err := c.execCmd.Process.Kill() // command timed out, kill it & return error
		return err
	}

}

func (c *cmdSource) Properties() []prop.Media {
	return c.props
}

// {BLOCKING GOROUTINE} logStdIoWithPrefix reads from the command's standard output or error, and prints it to the console as debug logs prefixed with the provided prefix
func (c *cmdSource) logStdIoWithPrefix(prefix string, stdIo io.ReadCloser) {
	reader := bufio.NewReader(stdIo)
	for {
		if line, err := reader.ReadBytes('\n'); err == nil {
			// logger.Debug(prefix + string(line))
			println(prefix + strings.Trim(string(line), " \r\n"))
		} else if err == io.EOF || err == io.ErrUnexpectedEOF {
			// logger.Debug(prefix + string(line))
			println(prefix + string(line))
			break
		} else if err != nil {
			logger.Error(err.Error())
			break
		}
	}
}

func (c *cmdSource) addEnvVarsFromStruct(props interface{}, logProps bool) {
	c.execCmd.Env = os.Environ() // inherit environment variables
	values := reflect.ValueOf(props)
	types := values.Type()
	if logProps {
		fmt.Print("Adding cmdsource environment variables: ")
	}
	for i := 0; i < values.NumField(); i++ {
		name := types.Field(i).Name
		value := values.Field(i)
		envVar := fmt.Sprintf("PION_MEDIA_%s=%v", name, value)
		if logProps {
			fmt.Print(envVar + ", ")
		}
		c.execCmd.Env = append(c.execCmd.Env, envVar)
	}
	if logProps {
		fmt.Println()
	}
}
