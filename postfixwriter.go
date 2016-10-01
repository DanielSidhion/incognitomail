package incognitomail

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// PostfixWriter holds all the information required to add or remove handles to a postfix system.
type PostfixWriter struct {
	mapFilename string
	domain      string
}

// NewPostfixWriter returns a PostfixWriter object initialized with values from the config.
func NewPostfixWriter() *PostfixWriter {
	return &PostfixWriter{
		mapFilename: Config.PostfixConfig.MapFilePath,
		domain:      Config.PostfixConfig.Domain,
	}
}

// AddHandle adds a handle to the map file.
func (p *PostfixWriter) AddHandle(h string, t string) (string, error) {
	f, err := os.OpenFile(p.mapFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fullHandle := fmt.Sprintf("%s%s", h, p.domain)

	_, err = f.WriteString(fmt.Sprintf("%s %s\n", fullHandle, t))
	if err != nil {
		return "", err
	}

	f.Close()
	err = p.invokePostmap()
	if err != nil {
		return "", err
	}

	return fullHandle, nil
}

// RemoveHandle scans a map file for a line starting with the handle and removes it.
func (p *PostfixWriter) RemoveHandle(h string) error {
	f, err := os.OpenFile(p.mapFilename, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), h) {
			t.WriteString(fmt.Sprintf("%s\n", scanner.Text()))
		}
	}

	t.Close()
	f.Close()
	os.Rename(t.Name(), f.Name())

	err = p.invokePostmap()
	if err != nil {
		return err
	}

	return nil
}

// invokePostmap runs the 'postmap' command in the shell to update the map file in postfix.
func (p *PostfixWriter) invokePostmap() error {
	cmd := "postmap"
	args := []string{Config.PostfixConfig.MapFilePath}

	err := exec.Command(cmd, args...).Run()
	if err != nil {
		return err
	}

	return nil
}
