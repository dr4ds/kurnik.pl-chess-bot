package uci

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type ChessEngine struct {
	cmd *exec.Cmd
	out *bufio.Reader
	in  *bufio.Writer
}

type Result struct {
	BestMove string
	Ponder   string
}

func (e *ChessEngine) LoadEngine(path string) error {
	e.cmd = exec.Command(path)
	in, err := e.cmd.StdinPipe()
	if err != nil {
		return err
	}

	out, err := e.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := e.cmd.Start(); err != nil {
		return err
	}

	e.in = bufio.NewWriter(in)
	e.out = bufio.NewReader(out)

	return nil
}

func (e *ChessEngine) writeString(str string) error {
	_, err := e.in.WriteString(str)
	if err != nil {
		return err
	}

	err = e.in.Flush()
	return err
}

func (e *ChessEngine) SetFEN(fen string) error {
	return e.writeString(fmt.Sprintf("position fen %s\n", fen))
}

func (e *ChessEngine) Depth(depth int) (Result, error) {
	r := Result{}
	err := e.writeString(fmt.Sprintf("go depth %d\n", depth))

	if err != nil {
		return r, err
	}

	for {
		line, err := e.out.ReadString('\n')
		if err != nil {
			return r, err
		}

		if strings.HasPrefix(line, "bestmove") {
			sp := strings.Split(strings.Trim(line, "\n"), " ")
			r.BestMove = sp[1]

			if len(sp) > 2 {
				r.Ponder = sp[3]
			}
			return r, err
		}
	}
}

func (e *ChessEngine) SetOptions(options map[string]string) error {
	for k, v := range options {
		err := e.writeString(fmt.Sprintf("setoption name %s value %v\n", k, v))

		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ChessEngine) Close() error {
	err := e.writeString("stop\n")
	if err != nil {
		return err
	}

	err = e.cmd.Process.Kill()
	if err != nil {
		return err
	}

	return e.cmd.Wait()
}
