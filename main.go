package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Printf("Try running something like: go run main.go run /bin/sh (as we are using Alpine FS which only has sh not bash...\n")
		panic("what???")
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setupNamespace(cmd)
	must(cmd.Run())
}

func child() {
	fmt.Printf("running %v with PID %v\n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setupContainerFilesystem()
	must(cmd.Run())
}

func setupNamespace(cmd *exec.Cmd) {
	// with help from Liz Rice and Shida
	//	作者：shida_csdn
	//来源：CSDN
	//原文：https://blog.csdn.net/shida_csdn/article/details/84649669
	//版权声明：本文为博主原创文章，转载请附上博文链接！

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | // think this is a privileged call... works under elevated permissions!
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER | // this forces chroot to seemingly fail - might be able to 666 it though and go again...
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      0,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      0,
				Size:        1,
			},
		},
	}
}

func setupContainerFilesystem() {
	must(syscall.Chroot("./alpine_rootfs"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
