package main

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	compileFile(8, 0, 1000)
	runFile(8, 0, 1000)
	compareFile(8, 0, 1000)
}

func compileFile(rid int, ftype int, pid int) error {
	cmd := exec.Command("")
	switch ftype {
	case 0:
		cmd = exec.Command("gcc", "-o", strconv.Itoa(rid), strconv.Itoa(rid)+".c")
	case 1:
		cmd = exec.Command("gcc", "-o", strconv.Itoa(rid), strconv.Itoa(rid)+".cpp", "-lstdc++")
	}
	cmd.Dir = "../../filesystem/submissions"

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

func runFile(rid int, ftype int, pid int) error {
	cmdStr := "cd ../../filesystem/submissions; ./" + strconv.Itoa(rid) + "< ../inputs/" + strconv.Itoa(pid)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()

	f, err := os.Create("../../filesystem/temp/" + strconv.Itoa(rid))
	if err != nil {
		return errors.New("cannot create output file")
	}
	defer f.Close()

	_, err = f.Write(output)
	if err != nil {
		return errors.New("write file error")
	}

	return nil
}

func compareFile(rid int, ftype int, pid int) error {
	return nil
}
