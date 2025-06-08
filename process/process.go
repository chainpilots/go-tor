// Package process is the low-level abstraction for a Tor instance.
//
// The standard use is to create a Creator with NewCreator and the path to the
// Tor executable. The child package 'embedded' can be used if Tor is statically
// linked in the binary. Most developers will prefer the tor package adjacent to
// this one for a higher level abstraction over the process and control port
// connection.
package process

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/chainpilots/go-tor/torutil"
)

/*
#cgo CFLAGS: -I${SRCDIR}/libs
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/linux_amd64 -ltor
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/linux_amd64 -levent
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/linux_amd64 -lz
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/linux_amd64 -lssl -lcrypto
#cgo windows LDFLAGS: -lws2_32 -lcrypt32 -lgdi32 -liphlpapi -lshlwapi -Wl,-Bstatic -lpthread
#cgo !windows LDFLAGS: -lm
#include <stdlib.h>
#ifdef _WIN32
	#include <winsock2.h>
#endif
#include <tor_api.h>
// Ref: https://stackoverflow.com/questions/45997786/passing-array-of-string-as-parameter-from-go-to-c-function
static char** makeCharArray(int size) {
	return calloc(sizeof(char*), size);
}
static void setArrayString(char **a, char *s, int n) {
	a[n] = s;
}
static void freeCharArray(char **a, int size) {
	int i;
	for (i = 0; i < size; i++)
		free(a[i]);
	free(a);
}
*/
import "C"

type Creator struct{}

// ProviderVersion returns the Tor provider name and version exposed from the
// Tor embedded API.
func ProviderVersion() string {
	return C.GoString(C.tor_api_get_provider_version())
}

type Process struct {
	ctx      context.Context
	mainConf *C.struct_tor_main_configuration_t
	args     []string
	doneCh   chan int
}

// New implements process.Creator.New
func New(ctx context.Context, args ...string) (*Process, error) {
	return &Process{
		ctx: ctx,
		// TODO: mem leak if they never call Start; consider adding a Close()
		mainConf: C.tor_main_configuration_new(),
		args:     args,
	}, nil
}

func (e *Process) Start() error {
	if e.doneCh != nil {
		return fmt.Errorf("already started")
	}
	// Create the char array for the args
	args := append([]string{"tor"}, e.args...)
	charArray := C.makeCharArray(C.int(len(args)))
	for i, a := range args {
		C.setArrayString(charArray, C.CString(a), C.int(i))
	}
	// Build the conf
	if code := C.tor_main_configuration_set_command_line(e.mainConf, C.int(len(args)), charArray); code != 0 {
		C.tor_main_configuration_free(e.mainConf)
		C.freeCharArray(charArray, C.int(len(args)))
		return fmt.Errorf("failed to set command line args, code: %v", int(code))
	}
	// Run it async
	e.doneCh = make(chan int, 1)
	go func() {
		defer C.freeCharArray(charArray, C.int(len(args)))
		defer C.tor_main_configuration_free(e.mainConf)
		e.doneCh <- int(C.tor_run_main(e.mainConf))
	}()
	return nil
}

func (e *Process) Wait() error {
	if e.doneCh == nil {
		return fmt.Errorf("not started")
	}
	ctx := e.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case code := <-e.doneCh:
		if code == 0 {
			return nil
		}
		return fmt.Errorf("command completed with error exit code: %v", code)
	}
}

func (e *Process) EmbeddedControlConn() (net.Conn, error) {
	file := os.NewFile(uintptr(C.tor_main_configuration_setup_control_socket(e.mainConf)), "")
	conn, err := net.FileConn(file)
	if err != nil {
		err = fmt.Errorf("unable to create conn from control socket: %v", err)
	}
	return conn, err
}

// ControlPortFromFileContents reads a control port file that is written by Tor
// when ControlPortWriteToFile is set.
func ControlPortFromFileContents(contents string) (int, error) {
	contents = strings.TrimSpace(contents)
	_, port, ok := torutil.PartitionString(contents, ':')
	if !ok || !strings.HasPrefix(contents, "PORT=") {
		return 0, fmt.Errorf("invalid port format: %v", contents)
	}
	return strconv.Atoi(port)
}
