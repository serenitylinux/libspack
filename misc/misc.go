package misc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/cam72cam/go-lumberjack/color"
	"github.com/cam72cam/go-lumberjack/log"
)

func GetWidth() int {
	var buf bytes.Buffer
	cmd := exec.Command("tput", "cols")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	str := buf.String()
	if err == nil && len(str) > 2 {
		width, _ := strconv.Atoi(str[:len(str)-1])
		return width
	} else {
		return 60
	}
}

var Bar = strings.Repeat("=", GetWidth()) + "\n"

func LogBar(l io.Writer, c color.Code) {
	l.Write([]byte(c.String(Bar)))
}

func WithFileReader(filename string, action func(io.Reader)) error {
	file, ioerr := os.Open(filename)
	if ioerr != nil {
		return ioerr
	}

	action(bufio.NewReader(file))

	return file.Close()
}

func WithFileWriter(filename string, create bool, action func(io.Writer)) error {
	var file *os.File
	var err error
	if create {
		file, err = os.Create(filename)
	} else {
		file, err = os.Open(filename)
	}
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)
	action(writer)

	err = writer.Flush()
	if err != nil {
		return err
	}

	return file.Close()
}

//I think this will work
//Christian Mesh 2014
func CopyFile(src, dest string) error {
	var err, e error

	e = WithFileWriter(dest, true, func(writer io.Writer) {
		e = WithFileReader(src, func(reader io.Reader) {
			_, e = io.Copy(writer, reader)
			if e != nil {
				err = e
			}
		})
		if e != nil {
			err = e
		}
	})
	if e != nil {
		err = e
	}

	return err
}

func InDir(path string, action func()) error {
	prevDir, _ := os.Getwd()
	if err := os.Chdir(path); err != nil {
		return err
	}
	action()
	if err := os.Chdir(prevDir); err != nil {
		return err
	}
	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsSymlink(f os.FileInfo) bool {
	return f.Mode()&os.ModeSymlink != 0
}

func GetUidGid(f os.FileInfo) (int, int) {
	st := f.Sys().(*syscall.Stat_t)
	return int(st.Uid), int(st.Gid)
}

func RunCommandToString(cmd *exec.Cmd) (string, error) {
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ReaderToString(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String()
}

func RunCommandToStdOutErr(cmd *exec.Cmd) (err error) {
	return RunCommand(cmd, os.Stdout, os.Stderr)
}

func RunCommand(cmd *exec.Cmd, stdout io.Writer, stderr io.Writer) error {
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

func Header(str string) {
	log.Info.Print(str + ": ")
	log.Debug.Println()
	LogBar(log.Debug, color.Brown)
}
func HeaderFormat(str string, extra ...interface{}) {
	Header(fmt.Sprintf(str, extra...))
}

func PrintSuccess() {
	log.Info.Println(color.Green.String("Success"))
	log.Debug.Println()
}

func AskYesNo(question string, def bool) bool {
	yn := "[Y/n]"
	if !def {
		yn = "[y/N]"
	}
	fmt.Printf("%s: %s ", question, yn)

	var answer string
	fmt.Scanf("%s", &answer)

	yesRgx := regexp.MustCompile("(y|Y|yes|Yes)")
	noRgx := regexp.MustCompile("(n|N|no|No)")
	switch {
	case answer == "":
		return def
	case yesRgx.MatchString(answer):
		return true
	case noRgx.MatchString(answer):
		return false
	default:
		fmt.Println("Please enter Y or N")
		return AskYesNo(question, def)
	}
}

var GitRegex = regexp.MustCompile(".*\\.git")
var RsyncRegex = regexp.MustCompile("rsync://.*")
var HttpRegex = regexp.MustCompile("(http|https|ftp)://.*")
