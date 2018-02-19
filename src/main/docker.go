package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func operateFile(rid uint64, ftype int, pid int) (int, error) {
	compileFile(8, 0, 1000)
	execFile(8, 0, 1000)
	res, _ := compareFile(8, 0, 1000)
	if res == true {
		fmt.Println("Accept")
	} else {
		fmt.Println("Wrong Answer")
	}
	return 0, nil
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
		return errors.New(string(out))
	}

	return nil
}

func execFile(rid uint64, ftype int, pid int) error {
	cmdStr := "cd ../../filesystem/submissions; ./" + strconv.FormatUint(rid, 10) + "< ../inputs/" + strconv.Itoa(pid)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		return err
	}

	f, err := os.Create("../../filesystem/temp/" + strconv.FormatUint(rid, 10))
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

func compareFile(rid uint64, ftype int, pid int) (bool, error) {
	cmdStr := "diff ../../filesystem/outputs/" + strconv.Itoa(pid) + " ../../filesystem/temp/" + strconv.FormatUint(rid, 10)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		return false, err
	}

	if len(output) == 0 {
		return true, nil
	}

	return false, nil
}
