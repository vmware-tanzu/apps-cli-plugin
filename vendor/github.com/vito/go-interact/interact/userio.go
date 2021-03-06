package interact

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type userIO interface {
	WriteLine(line string) error

	ReadLine(prompt string) (string, error)
	ReadPassword(prompt string) (string, error)
}

type ttyUser struct {
	*term.Terminal
}

func newTTYUser(input io.Reader, output *os.File) (ttyUser, error) {
	t := term.NewTerminal(readWriter{input, output}, "")

	width, height, err := term.GetSize(int(output.Fd()))
	if err != nil {
		return ttyUser{}, err
	}

	err = t.SetSize(width, height)
	if err != nil {
		return ttyUser{}, err
	}

	return ttyUser{
		Terminal: t,
	}, nil
}

func (u ttyUser) WriteLine(line string) error {
	_, err := fmt.Fprintf(u.Terminal, "%s\r\n", line)
	return err
}

func (u ttyUser) ReadLine(prompt string) (string, error) {
	u.Terminal.SetPrompt(prompt)
	return u.Terminal.ReadLine()
}

type nonTTYUser struct {
	io.Reader
	io.Writer
}

func newNonTTYUser(input io.Reader, output io.Writer) nonTTYUser {
	return nonTTYUser{
		Reader: input,
		Writer: output,
	}
}

func (u nonTTYUser) WriteLine(line string) error {
	_, err := fmt.Fprintf(u.Writer, "%s\n", line)
	return err
}

func (u nonTTYUser) ReadLine(prompt string) (string, error) {
	_, err := fmt.Fprintf(u.Writer, "%s", prompt)
	if err != nil {
		return "", err
	}

	line, err := u.readLine()
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(u.Writer, "%s\n", line)
	if err != nil {
		return "", err
	}

	return line, nil
}

func (u nonTTYUser) ReadPassword(prompt string) (string, error) {
	_, err := fmt.Fprintf(u.Writer, "%s", prompt)
	if err != nil {
		return "", err
	}

	line, err := u.readLine()
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(u.Writer, "\n")
	if err != nil {
		return "", err
	}

	return line, nil
}

func (u nonTTYUser) readLine() (string, error) {
	var line string

	for {
		chr := make([]byte, 1)
		n, err := u.Reader.Read(chr)

		if n == 1 {
			if chr[0] == '\n' {
				return line, nil
			} else if chr[0] == '\r' {
				continue
			}

			line += string(chr)
		}

		if err != nil {
			return "", err
		}
	}
}

type readWriter struct {
	io.Reader
	io.Writer
}
