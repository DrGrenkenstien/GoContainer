package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

type Container struct {
    ID      string `json:"id"`
    Status  string `json:"status"`
    Command string `json:"command"`
    Pid     int    `json:"pid"`
}

var containers = make(map[string]*Container)

func Must(err error) {
	if (err != nil) {
		panic(err)
	}
}

func must(err error) {
	if (err != nil) {
		panic(err)
	}
}

func create_linux_command_files(container_id string, curr_dir string){

    scriptPath := "./sync_rootfs.sh"
    root_path := "./" + container_id

    cmd := exec.Command("bash", scriptPath, root_path)

    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Error executing script: %v\n", err)
        fmt.Printf("Output: %s\n", output)
        return
    }

    // Print the output
    fmt.Printf("Script output:\n%s\n", output)

}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "shub_cg"), 0755)

	// Set the maximum number of processes allowed in the cgroup
	Must(os.WriteFile(filepath.Join(pids, "shub_cg/pids.max"), []byte("20"), 0700))

	// Removes the new cgroup in place after the container exits
	Must(os.WriteFile(filepath.Join(pids, "shub_cg/notify_on_release"), []byte("1"), 0700))

	// Add the current process to the cgroup
	Must(os.WriteFile(filepath.Join(pids, "shub_cg/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func Create(id string, command string) {

    container_temp, err := loadContainer(id)

    if(err != nil){
        
        if os.IsNotExist(err) || errors.Is(err, os.ErrNotExist){
            fmt.Println("Creating container")
            container := &Container{ID: id, Status: "created", Command: command}
            containers[id] = container

            if err := saveContainer(container); err != nil {
                fmt.Printf("Error saving container: %v\n", err)
                return
            }

            fmt.Printf("Container %s created\n", id)
            return
        }

        fmt.Printf("Error loading container : %v", err)
        return
    }

    if container_temp != nil {
        fmt.Printf("Container with ID %s already exists. Use run container command\n", id)
        return
    }

    // container := &Container{ID: id, Status: "created", Command: command}
    // containers[id] = container

    // if err := saveContainer(container); err != nil {
    //     fmt.Printf("Error saving container: %v\n", err)
    //     return
    // }

    // fmt.Printf("Container %s created\n", id)
}

func Run() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run main.go run <container_id>")
        os.Exit(1)
    }

    containerID := os.Args[2]
    container, error := loadContainer(containerID)
    if error != nil {
        fmt.Printf("Container with ID %s does not exist\n", containerID)
        os.Exit(1)
    }

    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, container.Command, container.ID)...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
    }
    cmd.Stdin = os.Stdin   
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        fmt.Printf("Error starting container: %v\n", err)
        os.Exit(1)
    }

    container.Pid = cmd.Process.Pid
    container.Status = "running"
    if err := saveContainer(container); err != nil {
        fmt.Printf("Error saving container: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Container %s is running with PID %d\n", containerID, container.Pid)

    if err := cmd.Wait(); err != nil {
        fmt.Printf("Container process exited with error: %v\n", err)
        os.Exit(1)
    }

    container.Status = "stopped"
    if err := saveContainer(container); err != nil {
        fmt.Printf("Error saving container: %v\n", err)
        os.Exit(1)
    }
}

func Child() {

    fmt.Print("Running child with process id: ")

    // pid := os.Getpid()
    fmt.Printf(strconv.Itoa(os.Getpid()))

    // if(p_err == nil){
        
    // }
        

    if len(os.Args) < 3 {
        fmt.Println("Usage: child <command> [args...]")
        os.Exit(1)
    }

	containerID := os.Args[3:][0]

    if containerID == "" {
        fmt.Println("Error: CONTAINER_ID not set")
        os.Exit(1)
    }

	cg()

	wd, err := os.Getwd()

	if err != nil {
		fmt.Println("Error getting working directory:", err)
	} else {
		fmt.Println("Current working directory:", wd)
	}

	must(syscall.Sethostname([]byte(containerID)))
	must(syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""))

	newRoot := wd + "/" + containerID
	putOld := filepath.Join(newRoot, "oldroot")
	fmt.Printf("Old root: %s\n", putOld)

    create_linux_command_files(containerID, wd)

	must(os.MkdirAll(putOld, 0755))

	must(syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC|syscall.MS_PRIVATE, ""))
	must(syscall.Mount("none", newRoot, "", syscall.MS_REMOUNT|syscall.MS_BIND|syscall.MS_RDONLY, ""))
	must(syscall.Mount(newRoot, newRoot, "", syscall.MS_REMOUNT|syscall.MS_BIND|syscall.MS_REC, ""))

	must(syscall.PivotRoot(newRoot, putOld))

	// Change directory to the new root
	must(os.Chdir("/"))

	// Unmount the old root and remove it
	must(syscall.Unmount("/oldroot", syscall.MNT_DETACH))
	must(os.RemoveAll("/oldroot"))

    // Create the directory
    sys_err := os.MkdirAll("/sys", 0755)
    proc_err := os.MkdirAll("/proc", 0755)

    if sys_err != nil {
        fmt.Printf("Error creating sys directory: %v\n", sys_err)
        return
    }

    if  proc_err != nil {
        fmt.Printf("Error creating proc directory: %v\n",  proc_err)
        return
    }


	// Mount proc and sysfs filesystems
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
	must(syscall.Mount("sysfs", "/sys", "sysfs", 0, ""))

    if err := syscall.Exec(os.Args[2], os.Args[2:3], os.Environ()); err != nil { // For future: Make sure the commands are subscripted correctly in future
        fmt.Printf("Error executing command: %v\n", err)
        os.Exit(1)
    }

	must(syscall.Unmount("/proc", 0))
	must(syscall.Unmount("/sys", 0))
}

func saveContainer(container *Container) error {
    containerDir := "./" + container.ID
    if err := os.MkdirAll(containerDir, 0755); err != nil {
        return fmt.Errorf("error creating containers directory: %w", err)
    }

    containerFilePath := filepath.Join(containerDir, container.ID + ".json")
    file, err := os.Create(containerFilePath)
    if err != nil {
        return fmt.Errorf("error creating container file: %w", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    if err := encoder.Encode(container); err != nil {
        return fmt.Errorf("error encoding container to file: %w", err)
    }

    return nil
}

func loadContainer(id string) (*Container, error) {
    containerFilePath := filepath.Join("./" + id, id + ".json")
    file, err := os.Open(containerFilePath)
    if err != nil {
        return nil, fmt.Errorf("error opening container file: %w", err)
    }
    defer file.Close()

    container := &Container{}
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(container); err != nil {
        return nil, fmt.Errorf("error decoding container from file: %w", err)
    }

    return container, nil
}

func loadAllContainers() error {
    containerDir := "/var/containers"
    files, err := os.ReadDir(containerDir)
    if err != nil {
        return fmt.Errorf("error reading containers directory: %w", err)
    }

    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
            containerID := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]
            container, err := loadContainer(containerID)
            if err != nil {
                fmt.Printf("Error loading container %s: %v\n", containerID, err)
                continue
            }
            containers[container.ID] = container
        }
    }

    return nil
}