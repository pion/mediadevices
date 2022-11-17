package cmdsource

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
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

// var cmdSourceLabelCounts map[string]uint = make(map[string]uint)

type cmdSource struct {
	cmdArgs     []string
	props       []prop.Media
	readTimeout uint32 // in seconds
	execCmd     *exec.Cmd
}

func init() {
	// No init, call AddVideoCmdSource() or AddAudioCmdSource() to add a command source before getUserMedia().
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

func (c *cmdSource) addEnvVarsFromStruct(props interface{}) {
	c.execCmd.Env = os.Environ() // inherit environment variables
	values := reflect.ValueOf(props)
	types := values.Type()
	fmt.Println("Added Environment Variables: ")
	for i := 0; i < values.NumField(); i++ {
		name := types.Field(i).Name
		value := values.Field(i)
		envVar := fmt.Sprintf("MEDIA_DEVICES_%s=%v", name, value)
		fmt.Println(envVar + ", ")
		c.execCmd.Env = append(c.execCmd.Env, envVar)
	}
}

// func (c *cmdSource) getCmdLabel() string {
// 	programName := c.cmdArgs[0]
// 	if _, ok := cmdSourceLabelCounts[programName]; ok {
// 		cmdSourceLabelCounts[programName]++
// 	} else {
// 		cmdSourceLabelCounts[programName] = 0
// 	}
// 	return programName + "_" + fmt.Sprintf("%d", cmdSourceLabelCounts[programName])
// }
