package incognitomail

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/valyala/gorpc"
	"golang.org/x/net/websocket"
	"gopkg.in/tylerb/graceful.v1"
)

// Server holds all the data required to run the server.
type Server struct {
	lockFileHandle *os.File

	persistence      *IncognitoData
	mailSystemWriter MailSystemHandleWriter
	commandCh        chan interface{}
	signalCh         chan os.Signal

	httpServer *graceful.Server
	rpcServer  *gorpc.Server

	started  bool
	finishCh chan bool
}

// MailSystemHandleWriter has methods for adding and removing mappings from the mail system.
type MailSystemHandleWriter interface {
	AddHandle(string, string) (string, error)
	RemoveHandle(string) error
}

type newHandleCommand struct {
	source        string
	accountSecret string
	resultCh      chan string
	errorCh       chan error
}

type newAccountCommand struct {
	source   string
	target   string
	resultCh chan string
	errorCh  chan error
}

type deleteHandleCommand struct {
	source   string
	handle   string
	secret   string
	resultCh chan string
	errorCh  chan error
}

type deleteAccountCommand struct {
	source   string
	secret   string
	resultCh chan string
	errorCh  chan error
}

type terminateCommand struct{}

const (
	accountSecretSize = 64
	handleSize        = 18

	commandQueue                  = 10
	httpServerTimeout             = 10 * time.Second
	httpServerTCPKeepAliveTimeout = 3 * time.Minute
)

var (
	// ErrServerNotStarted is used whenever an action is taken that expects a started server, but the server is actually not started.
	ErrServerNotStarted = errors.New("server not started")

	// ErrLockFileAlreadyExists is used only when trying to acquire a lock file, but the lock file already exists.
	ErrLockFileAlreadyExists = errors.New("lock file already exists")

	// ErrLockFileNotFound is used whenever an action requires a lock file to be used, but the lock file handle is nil.
	ErrLockFileNotFound = errors.New("could not find lock file")

	// ErrEmptyCommand is used when the server receives an empty command.
	ErrEmptyCommand = errors.New("empty command received")

	// ErrUnknownCommand is used when the server receives an unknown/garbage command.
	ErrUnknownCommand = errors.New("unknown command received")

	// ErrWrongCommand is used when a known command is received, but is malformed.
	ErrWrongCommand = errors.New("wrong command usage")

	// ErrInvalidPermission is used when a command has been received from the websocket, but the server shouldn't execute it.
	ErrInvalidPermission = errors.New("invalid permission to do this")
)

func mailSystemWriterFromConfig() MailSystemHandleWriter {
	switch Config.General.MailSystem {
	case "postfix":
		return NewPostfixWriter()
	}

	return nil
}

// NewServer returns an IncognitoMailServer object ready for use.
func NewServer() (*Server, error) {
	server := &Server{
		mailSystemWriter: mailSystemWriterFromConfig(),
		commandCh:        make(chan interface{}, commandQueue),
		signalCh:         make(chan os.Signal, 1),
	}

	data, err := OpenIncognitoData()
	if err != nil {
		return nil, err
	}

	server.persistence = data
	err = server.getLockFile()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) getLockFile() error {
	if s.lockFileHandle != nil {
		return ErrLockFileAlreadyExists
	}

	err := os.MkdirAll(filepath.Dir(Config.General.LockFilePath), os.FileMode(0755))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(Config.General.LockFilePath, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		return err
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return err
	}

	// Writing the PID in the lock file according to File System Hierarchy (FHS) standards
	fmt.Fprintf(file, "%10d\n", os.Getpid())
	s.lockFileHandle = file

	return nil
}

func (s *Server) removeLockFile() error {
	if s.lockFileHandle == nil {
		return ErrLockFileNotFound
	}

	err := syscall.Flock(int(s.lockFileHandle.Fd()), syscall.LOCK_UN|syscall.LOCK_NB)
	if err != nil {
		return err
	}

	s.lockFileHandle.Close()

	err = os.Remove(Config.General.LockFilePath)
	if err != nil {
		// If the lock file stays in the system, we won't have a problem when executing the program again, so just log the occurrence.
		log.Printf("[DEBUG] Could not remove lock file in %s\n", Config.General.LockFilePath)
	}

	s.lockFileHandle = nil
	return nil
}

// startRPCListener starts the RPC service for communication between the running incognito server and any other processes.
func (s *Server) startRPCListener() error {
	d := gorpc.NewDispatcher()
	d.AddService("IncognitoRPCService", s)

	server := gorpc.NewUnixServer(Config.General.UnixSockPath, d.NewHandlerFunc())
	err := server.Start()
	if err != nil {
		return err
	}

	s.rpcServer = server
	return nil
}

// Start begins listening for websocket connections (external requests) and RPC calls (internal requests).
func (s *Server) Start() {
	if s.started {
		return
	}

	s.started = true

	signal.Notify(s.signalCh, syscall.SIGINT, syscall.SIGTERM)

	go handleSignals(s)

	err := s.startRPCListener()
	if err != nil {
		log.Fatal(err)
	}

	go handleCommands(s)

	mux := http.NewServeMux()

	// We listen for websocket connection this way to avoid receiving an 403 when connecting from the localhost (or anything that passes a "null" Origin header)
	mux.HandleFunc(Config.General.ListenPath, func(w http.ResponseWriter, req *http.Request) {
		server := websocket.Server{Handler: websocket.Handler(func(ws *websocket.Conn) {
			var args string

			err := websocket.Message.Receive(ws, &args)
			if err != nil {
				log.Printf("[DEBUG] Error receiving command from websocket: %s\n", err)
				websocket.Message.Send(ws, "error receiving command")
				return
			}

			result, err := s.SendCommand("websocket", args)
			if err != nil {
				websocket.Message.Send(ws, "error "+err.Error())
				return
			}

			websocket.Message.Send(ws, result)
		})}

		server.ServeHTTP(w, req)
	})

	srv := &graceful.Server{
		Timeout:      httpServerTimeout,
		TCPKeepAlive: httpServerTCPKeepAliveTimeout,
		Server: &http.Server{
			Addr:    Config.General.ListenAddress,
			Handler: mux,
		},
	}

	s.httpServer = srv

	if Config.General.TLSCertFile != "" && Config.General.TLSKeyFile != "" {
		err = srv.ListenAndServeTLS(Config.General.TLSCertFile, Config.General.TLSKeyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		log.Fatal(err)
	}
}

// stopAllButHTTPServer is an utility for stopping everything else besides the http server (read the docs for handleSignals for an explanation).
func (s *Server) stopAllButHTTPServer() {
	s.rpcServer.Stop()
	s.commandCh <- terminateCommand{}
	s.persistence.Close()
	s.removeLockFile()
}

// Stop will stop everything from a running server
func (s *Server) Stop() {
	s.stopAllButHTTPServer()
	s.httpServer.Stop(httpServerTimeout)

	// Waiting for the http server to stop
	<-s.httpServer.StopChan()
}

// Wait blocks until the server has finished executing. If the server wasn't started, it returns an error instead.
func (s *Server) Wait() error {
	if s.finishCh == nil {
		return ErrServerNotStarted
	}

	<-s.finishCh
	return nil
}

// SendCommand is executed for every message received either by the websocket or RPC interface. Builds a well-defined command to send to the goroutine listening for commands to execute.
func (s *Server) SendCommand(source, args string) (string, error) {
	c := strings.Fields(args)
	if len(c) == 0 {
		return "", ErrEmptyCommand
	}

	command := c[0]
	// c[1:] works even if len(c) == 1, and in this case it's just an empty slice
	extra := c[1:]

	// It is important to make buffered channels, because we'll send values and close them afterwards
	resultCh := make(chan string, 1)
	errorCh := make(chan error, 1)

	switch command {
	case "new":
		if len(extra) != 2 {
			return "", ErrWrongCommand
		}

		switch extra[0] {
		case "handle":
			s.commandCh <- newHandleCommand{
				source:        source,
				accountSecret: extra[1],
				resultCh:      resultCh,
				errorCh:       errorCh,
			}
		case "account":
			s.commandCh <- newAccountCommand{
				source:   source,
				target:   extra[1],
				resultCh: resultCh,
				errorCh:  errorCh,
			}
		default:
			log.Printf("[DEBUG] received unknown 'new' option: %s\n", args)
			return "", ErrWrongCommand
		}
	case "delete":
		if len(extra) < 1 {
			return "", ErrWrongCommand
		}

		switch extra[0] {
		case "handle":
			if len(extra) != 3 {
				return "", ErrWrongCommand
			}

			s.commandCh <- deleteHandleCommand{
				source:   source,
				handle:   extra[1],
				secret:   extra[2],
				resultCh: resultCh,
				errorCh:  errorCh,
			}
		case "account":
			if len(extra) != 2 {
				return "", ErrWrongCommand
			}

			s.commandCh <- deleteAccountCommand{
				source:   source,
				secret:   extra[1],
				resultCh: resultCh,
				errorCh:  errorCh,
			}
		}
	default:
		log.Printf("[DEBUG] received unknown command %s\n", args)
		return "", ErrUnknownCommand
	}

	return <-resultCh, <-errorCh
}

// Receives any command that needs to be executed, and executes them.
func handleCommands(s *Server) {
	for {
		command := <-s.commandCh

		var res string
		var err error
		var resCh chan string
		var errCh chan error

		switch t := command.(type) {
		case terminateCommand:
			log.Println("[INFO] Terminating server")
			return
		case newHandleCommand:
			res, err = s.NewHandle(t.accountSecret)
			resCh = t.resultCh
			errCh = t.errorCh
		case newAccountCommand:
			// New accounts should be created locally only, at least at first
			if t.source == "websocket" {
				err = ErrInvalidPermission
				res = ""
			} else {
				res, err = s.NewAccount(t.target)
			}

			resCh = t.resultCh
			errCh = t.errorCh
		case deleteHandleCommand:
			res = ""
			if t.source == "websocket" {
				err = ErrInvalidPermission
			} else {
				err = s.DeleteHandle(t.secret, t.handle)
				if err == nil {
					res = "success"
				}
			}

			resCh = t.resultCh
			errCh = t.errorCh
		case deleteAccountCommand:
			res = ""
			if t.source == "websocket" {
				err = ErrInvalidPermission
			} else {
				err = s.DeleteAccount(t.secret)
				if err == nil {
					res = "success"
				}
			}

			resCh = t.resultCh
			errCh = t.errorCh
		default:
			log.Printf("[DEBUG] unrecognized command %v\n", t)
			continue
		}

		errCh <- err
		resCh <- res
		close(errCh)
		close(resCh)
	}
}

func handleSignals(s *Server) {
	<-s.signalCh

	// Upon receiving a signal, just stop. s.httpServer will also receive the signal, so stop everything but s.httpServer
	s.stopAllButHTTPServer()
}

// NewHandle creates a new handle for the account with the given secret.
func (s *Server) NewHandle(accountSecret string) (string, error) {
	target, err := s.persistence.GetAccountTarget(accountSecret)
	if err != nil {
		return "", err
	}

	var newHandle string

	// We'll keep looping until we find a handle that hasn't been used
	for {
		newHandle, err = generateRandomString(handleSize)
		if err != nil {
			return "", err
		}

		if !s.persistence.HasHandleGlobal(newHandle) {
			break
		}
	}

	err = s.persistence.NewAccountHandle(accountSecret, newHandle)
	if err != nil {
		return "", err
	}

	// fullHandle will have the domain attached, so it's the complete incognito email
	fullHandle, err := s.mailSystemWriter.AddHandle(newHandle, target)
	if err != nil {
		return "", err
	}

	return fullHandle, nil
}

// NewAccount creates a new account with the given target email address and returns the secret.
func (s *Server) NewAccount(target string) (string, error) {
	var secret string
	var err error

	// We'll keep looping until we find an unused secret
	for {
		secret, err = generateRandomString(accountSecretSize)
		if err != nil {
			return "", err
		}

		if !s.persistence.HasAccount(secret) {
			break
		}
	}

	err = s.persistence.NewAccount(secret, target)
	if err != nil {
		return "", err
	}

	return secret, nil
}

// DeleteHandle deletes the given handle from the account with the given secret. If the account does not exist, it returns an error.
func (s *Server) DeleteHandle(secret, handle string) error {
	exists := s.persistence.HasAccount(secret)

	if !exists {
		return ErrAccountNotFound
	}

	s.persistence.DeleteAccountHandle(secret, handle)
	s.mailSystemWriter.RemoveHandle(handle)

	return nil
}

// DeleteAccount deletes all data from the account with the given secret. If the account does not exist, it returns an error.
func (s *Server) DeleteAccount(secret string) error {
	exists := s.persistence.HasAccount(secret)

	if !exists {
		return ErrAccountNotFound
	}

	// Listing all handles for this account and removing them from the mail system
	handles, err := s.persistence.ListAccountHandles(secret)
	if err != nil {
		return err
	}

	for _, handle := range handles {
		err := s.mailSystemWriter.RemoveHandle(handle)
		if err != nil {
			return err
		}
	}

	// Only after removing all handles from the mail system, delete from persistence system
	s.persistence.DeleteAccount(secret)

	return nil
}

// ListHandles returns all handles from the account with the given secret.
func (s *Server) ListHandles(secret string) ([]string, error) {
	exists := s.persistence.HasAccount(secret)

	if !exists {
		return nil, ErrAccountNotFound
	}

	// Listing all handles for this account and removing them from the mail system
	handles, err := s.persistence.ListAccountHandles(secret)
	if err != nil {
		return nil, err
	}

	return handles, nil
}

// CreateRPCServiceClient creates and returns a reasy to use RPC dispatcher client.
func CreateRPCServiceClient() *gorpc.DispatcherClient {
	// Using an empty server struct is not a problem, we only want the methods
	s := &Server{}
	d := gorpc.NewDispatcher()
	d.AddService("IncognitoRPCService", s)

	c := gorpc.NewUnixClient(Config.General.UnixSockPath)
	c.Start()

	dc := d.NewServiceClient("IncognitoRPCService", c)
	return dc
}
