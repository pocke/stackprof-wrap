package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/ogier/pflag"
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
	outFile := ""
	fs := pflag.NewFlagSet(args[0], pflag.ExitOnError)
	fs.StringVarP(&outFile, "out", "o", "stackprof-out", "output file name for stackprof")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	for _, path := range fs.Args() {
		v, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s, err := os.Stat(path)
		if err != nil {
			return err
		}
		content := Wrap(v, outFile)
		ioutil.WriteFile(path, []byte(content), s.Mode())
	}
	return nil
}

func Wrap(content []byte, outFile string) string {
	reBlank := regexp.MustCompile(`^\s+$`)
	r := bufio.NewScanner(bytes.NewBuffer(content))
	r.Split(bufio.ScanLines)
	lines := []string{}

	for r.Scan() {
		line := r.Text()
		// Doesn't wrap shebang, comment, and empty line
		if !(strings.HasPrefix(line, "#") || reBlank.MatchString(line)) {
			lines = append(lines, fmt.Sprintf(`require 'stackprof'
StackProf.run(mode: :cpu, out: '%s') do
  begin`, outFile))
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
