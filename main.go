package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Main(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("target file is required. `stackprof-wrap TARGET`")
	}
	for _, path := range args[1:] {
		v, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s, err := os.Stat(path)
		if err != nil {
			return err
		}
		content := Wrap(v)
		ioutil.WriteFile(path, []byte(content), s.Mode())
	}
	return nil
}

func Wrap(content []byte) string {
	reBlank := regexp.MustCompile(`^\s+$`)
	r := bufio.NewScanner(bytes.NewBuffer(content))
	r.Split(bufio.ScanLines)
	lines := []string{}

	for r.Scan() {
		line := r.Text()
		// Doesn't wrap shebang, comment, and empty line
		if !(strings.HasPrefix(line, "#") || reBlank.MatchString(line)) {
			lines = append(lines, `require 'stackprof'
StackProf.run(mode: :cpu, out: 'stackprof-out') do
  begin`)
			lines = append(lines, line)
			break
		}
		lines = append(lines, line)
	}
	for r.Scan() {
		lines = append(lines, r.Text())
	}
	lines = append(lines, `  rescue Exception
    puts $!
  end
end
`)

	return strings.Join(lines, "\n")
}
