package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type argError struct {
	arg  int
	prob string
}

func (e *argError) Error() string {
	return fmt.Sprintf("%d - %s", e.arg, e.prob)
}

// func main() {
// 	res := operateFile(8, 1, 1000)
// 	fmt.Println(res)
// }

func operateFile(rid uint64, ftype int, pid int) (int) {
	err := compileFile(rid, ftype, pid)
	if err != nil {
		return err.(*argError).arg
	}

	err = execFile(rid, ftype, pid)
	if err != nil {
		return err.(*argError).arg
	}

	err = compareFile(rid, ftype, pid)
	return err.(*argError).arg
}

func compileFile(rid uint64, ftype int, pid int) error {
	cmd := exec.Command("")
	switch ftype {
	case 0:
		cmd = exec.Command("gcc", "-o", strconv.FormatUint(rid, 10), strconv.FormatUint(rid, 10)+".c")
	case 1:
		cmd = exec.Command("gcc", "-o", strconv.FormatUint(rid, 10), strconv.FormatUint(rid, 10)+".cpp", "-lstdc++")
	}
	cmd.Dir = "../../filesystem/submissions"

	out, err := cmd.CombinedOutput()
	if err != nil {
		return &argError{3, string(out)}
	}

	return nil
}

func execFile(rid uint64, ftype int, pid int) error {
	cmdStr := "cd ../../filesystem/submissions; ./" + strconv.FormatUint(rid, 10) + "< ../inputs/" + strconv.Itoa(pid)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		return &argError{4, "Runtime Error"}
	}

	f, err := os.Create("../../filesystem/temp/" + strconv.FormatUint(rid, 10))
	if err != nil {
		return &argError{9, "System Error: cannot create ouput file"}
	}
	defer f.Close()

	_, err = f.Write(output)
	if err != nil {
		return &argError{7, "Output Limit Exceeded"}
	}

	return nil
}

func compareFile(rid uint64, ftype int, pid int) error {
	cmdStr := "diff ../../filesystem/outputs/" + strconv.Itoa(pid) + " ../../filesystem/temp/" + strconv.FormatUint(rid, 10)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		return &argError{2, "Wrong Answer"}
	}

	if len(output) == 0 {
		return &argError{1, "Accept"}
	}

	return &argError{2, "Wrong Answer"}
}
