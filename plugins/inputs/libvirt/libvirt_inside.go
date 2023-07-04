package libvirt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/influxdata/telegraf"
	uuid "github.com/satori/go.uuid"
)

type DomainGather struct {
	// QemuCommandMem()
	// QemuCommandCpu()
	// QemuCommandDisk()
	// QemuCommandDiskIO(str string)
	// QemuCommandNetState(str string)
	// QemuCommandNetSpeed(str string)
	// QemuCommandNetUsage(str string)

	DomianOsType int
	Domain       golibvirt.Domain
}

// guest-get-users (Command)
type OsInfo struct {
	Name          string `json:"name"`
	KernelRelease string `json:"kernel-release"`
	Version       string `json:"version"`
	PrettyName    string `json:"pretty-name"`
	VersionId     string `json:"version-id"`
	KernelVersion string `json:"kernel-version"`
	Machine       string `json:"machine"`
	VmID          string `json:"id"`
}

type OsInfoReturn struct {
	Return OsInfo `json:"return"`
}

type FileOpenHandler struct {
	Handler int `json:"return"`
}

type ReadReturnData struct {
	Count  int    `json:"count"`
	BufB64 string `json:"buf-b64"`
	EOF    bool   `json:"eof"`
}

type ReadReturn struct {
	Return ReadReturnData `json:"return"`
}

func GuestFileOpen(domain golibvirt.Domain, lv *golibvirt.Libvirt, filePath string) (*FileOpenHandler, error) {
	cmdOpenFile := fmt.Sprintf(`{"execute": "guest-file-open", "arguments": { "path": "%s", "mode":"r" } }`, filePath)

	res, err := lv.QEMUDomainAgentCommand(domain, cmdOpenFile, 5, 0)
	if err != nil {
		return nil, err
	}
	fileOpenHandler := &FileOpenHandler{}
	err = json.Unmarshal([]byte(res[0]), fileOpenHandler)
	if err != nil {
		return nil, err
	}
	return fileOpenHandler, err
}

type GuestExecReturn struct { 
	Return GuestExecPid `json:"return"`
}

type GuestExecPid struct {
	Pid int `json:"pid"`
}
type GuestExecStatusReturn struct {
	Return GuestExecStatusContents `json:"return"`
}
// stdOut base64-encoded stdout of the process
// exitCode process exit code if it was normally terminated.
// exited true if process has already terminated.

type GuestExecStatusContents struct {
	Exited   bool   `json:"exited"`
	ExitCode int    `json:"exitcode"`
	OutData  string `json:"out-data"`
	// Exited true if process has already terminated.
	// ExitCode process exit code if it was normally terminated.
	// OutData base64-encoded stdout of the process
	// Signal int
	ErrDatabase64 string `json:"err-data"`
	OutTruncated bool	`json:"out-truncated"`
	ErrTurncated bool	`json:"err-truncated"`
	// signal signal number (linux) or unhandled exception code (windows) if the process was abnormally terminated.
	// err-database64-encoded stderr of the process Note: out-data and err-data are present only if ‘capture-output’ was specified for ‘guest-exec’
	// out-truncated: boolean (optional) true if stdout was not fully captured due to size limitation.
	// err-truncated: boolean (optional)true if stderr was not fully captured due to size limitation.
}

func GuestCommandExec(domain golibvirt.Domain, lv *golibvirt.Libvirt, Command string, Arguements []string) (string, error) {
	// argsStr := fmt.Sprintf("\"%s\"", Arguements)
	argsStr := ""
	for _, arg := range Arguements {
		if argsStr == "" {
			argsStr = fmt.Sprintf("\"%s\"", arg)
		} else {
			argsStr = argsStr + fmt.Sprintf(", \"%s\"", arg)
		}
	}
	cmdExec := fmt.Sprintf(`{"execute": "guest-exec", "arguments": { "path": "%s", "arg": [ %s ], "capture-output":true } }`, Command, argsStr)
	// fmt.Println(cmdExec)
	res, err := lv.QEMUDomainAgentCommand(domain, cmdExec, 5, 0)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	execRes := &GuestExecReturn{}
	err = json.Unmarshal([]byte(res[0]), execRes)
	if err != nil {
		return "", err
	}


	dataChan := make(chan *GuestExecStatusContents)
	dataAcc := make(chan string)
	defer close(dataChan)
	defer close(dataAcc)
	// the guest-exec and guest-exec-status,
	// especially the second step,
	// which would cost huge time comparing with normal sequential processing
	// sequentially processing will for each round will cost 10ms 
	// however inorder to acquire data, it has to run couple of rounds
	go func() {
		start := time.Now()
		defer func() {
			fmt.Println("time elapse:", time.Since(start))
		}()
		cmdExecStatus := fmt.Sprintf(`{"execute": "guest-exec-status", "arguments": { "pid": %d } }`, execRes.Return.Pid)
		round := 1
		for {
			res, err = lv.QEMUDomainAgentCommand(domain, cmdExecStatus, 5, 0)
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			execStatusRes := &GuestExecStatusReturn{}
			err = json.Unmarshal([]byte(res[0]), execStatusRes)

			if execStatusRes.Return.Exited {
				// stdOutBytes, err := base64.StdEncoding.DecodeString(execStatusRes.Return.OutData)
				_, err := base64.StdEncoding.DecodeString(execStatusRes.Return.OutData)
				if err != nil {
					fmt.Println(err.Error())
					break
				}
			}
			round += 1
			// test if shit happens
			// time.Sleep(6*time.Second)
			// only return when Exited is True
			// meanwhile using select to monitor whether timeout or not
			if execStatusRes.Return.Exited {
				// Exited true if process has already terminated.
				// ExitCode process exit code if it was normally terminated.
				dataChan <- &(execStatusRes.Return)
				return
			}
		}
	}()
	// data := <-dataChan
	timeOutFlag := false

	go func ()  {
		select {
		case data := <-dataChan:
			// fmt.Println("Go Routine -- Select")
			// fmt.Println(domain.Name)
			// fmt.Println(data)
			// fmt.Println(data.ExitCode)
			// fmt.Println(data.OutData)
			// fmt.Println(data.ErrDatabase64)
			// fmt.Println(data.OutTruncated)
			// fmt.Println(data.ErrTurncated)
			dataAcc <- data.OutData
			// timeout should be configed in config files
		case <-time.After(1 * time.Second):
			fmt.Println("Time Out")
			timeOutFlag = true
		}
	}()
	// datastring := data.
	fmt.Println("====================")
	dataget := <- dataAcc
	fmt.Println("dataget")
	ed, _ := base64.StdEncoding.DecodeString(dataget)
	fmt.Println(string(ed))

	if timeOutFlag {
		// cannot get data
		fmt.Println("Time out")
		return "", errors.New("time out")
	}
	fmt.Println("+++++++++++++++++++++++")

	if err != nil {
		return "", err
	}

	return "", nil
}

func GuestFileRead(domain golibvirt.Domain, lv *golibvirt.Libvirt, fileOpenHandler int) (string, error) {
	// Read from an open file in the guest. Data will be base64-encoded.
	// As this command is just for limited, ad-hoc debugging, such as log file access,
	// the number of bytes to read is limited to 48 MB.
	// arg := []string{"-al"}

	contentString := ""
	cmdReadFile := fmt.Sprintf(`{"execute": "guest-file-read", "arguments": { "handle": %d } }`, fileOpenHandler)
	res, err := lv.QEMUDomainAgentCommand(domain, cmdReadFile, 5, 0)
	if err != nil {
		return "", err
	}

	readRes := &ReadReturn{}
	err = json.Unmarshal([]byte(res[0]), readRes)
	if err != nil {
		return "", err
	}
	if readRes.Return.Count > 0 {
		readBytes, err := base64.StdEncoding.DecodeString(readRes.Return.BufB64)
		if err != nil {
			return "", err
		}
		contentString = string(readBytes)
	}

	if !readRes.Return.EOF {
		return "File Reading Not Finished", nil
	}
	return contentString, nil
}

func GuestFileClose(domain golibvirt.Domain, lv *golibvirt.Libvirt, fileOpenHandler int) (bool, error) {
	cmdCloseFile := fmt.Sprintf(`{"execute": "guest-file-close", "arguments": { "handle": %d } }`, fileOpenHandler)
	// guest-file-close  Nothing on success.
	_, err := lv.QEMUDomainAgentCommand(domain, cmdCloseFile, 5, 0)
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetGuestFileContent(domain golibvirt.Domain, lv *golibvirt.Libvirt, filePath string) (string, error) {
	// check file size
	// if <=48M then
	// 		do .....
	// else
	// 		make copy of files and
	//

	// get file size

	

	fileOpenHandler, err := GuestFileOpen(domain, lv, filePath)
	if err != nil {
		return "File Openning Fail", err
	}
	fileReadContent, err := GuestFileRead(domain, lv, fileOpenHandler.Handler)
	if err != nil {
		return "File Reading Fail", err
	}
	_, err = GuestFileClose(domain, lv, fileOpenHandler.Handler)
	if err != nil {
		return "File Closing Fail", err
	}
	return fileReadContent, nil
}

func LinuxMem(domain golibvirt.Domain, lv *golibvirt.Libvirt) error {
	readFromFile, err := GetGuestFileContent(domain, lv, "/proc/meminfo")
	if err != nil {
		log.Printf("Error! %s ,%s", readFromFile, err.Error())
		return err
	}
	fmt.Println(readFromFile)
	// fmt.Printf("%T \n", readFromFile)``
	return nil
}

func LinuxCpu(domain golibvirt.Domain, lv *golibvirt.Libvirt) error {
	cpuData, err := GetGuestFileContent(domain, lv, "/proc/stat")
	if err != nil {
		log.Printf("Error! %s ,%s", cpuData, err.Error())
		return err
	}
	fmt.Println(cpuData)
	// fmt.Printf("%T \n", cpuData)
	return nil
}

func LinuxNetwork(domain golibvirt.Domain, lv *golibvirt.Libvirt) error {
	networkData, err := GetGuestFileContent(domain, lv, "/proc/net/dev")
	if err != nil {
		log.Printf("Error! %s ,%s", networkData, err.Error())
		return err
	}
	fmt.Println(networkData)
	// fmt.Printf("%T \n", networkData)
	return nil
}

func (dg *DomainGather) QemuCommandMem(acc telegraf.Accumulator, lv *golibvirt.Libvirt) error {
	fmt.Println("QemuCommandMem()")
	if dg.DomianOsType == 1 {
		fmt.Println("LinuxMem() -- Linux")
		LinuxMem(dg.Domain, lv)

		fmt.Println("LinuxCpu() -- Linux")
		LinuxCpu(dg.Domain, lv)

		fmt.Println("LinuxNetwork() -- Linux")
		LinuxNetwork(dg.Domain, lv)

		fmt.Println("GuestCommandExec() -- Linux")
		GuestCommandExec(dg.Domain, lv, "/bin/df", []string{"-Th"})

		// permission problems when using guest-exec
		// GuestCommandExec(dg.Domain, lv, "/bin/ls", []string{"-l", "/proc"})
		// qemu-agent-command '{"execute":"guest-exec", "arguments":{"path":"/bin/ls", "arg":["-l", "/path/to/myfile.txt"]}}'

	} else {
	}

	return nil
}

func DomainGatherAll(domain golibvirt.Domain, lv *golibvirt.Libvirt, vmid int, acc telegraf.Accumulator) error {
	domain_name := domain.Name
	domain_uuid, err := uuid.FromBytes(domain.UUID[:])
	if err != nil {
		log.Printf("Parsing UUID error:  %s, error Domain Name: %s, error Domain UUID: %s", err.Error(), domain_name, domain_uuid)
		return err
	}
	dg := DomainGather{Domain: domain, DomianOsType: vmid}
	dg.QemuCommandMem(acc, lv)

	return nil
}

func CheckOSType(domain golibvirt.Domain, lv *golibvirt.Libvirt) (int, error) {
	osInfoCmd := `{"execute": "guest-get-osinfo"}`

	res, err := lv.QEMUDomainAgentCommand(domain, osInfoCmd, 5, 0)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	var data OsInfoReturn
	err = json.Unmarshal([]byte(res[0]), &data)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}


	if data.Return.VmID == "mswindows" {
		return 2, nil
	} else {
		return 1, nil
	}
}

func GatherAll(domain golibvirt.Domain, wg *sync.WaitGroup, lv *golibvirt.Libvirt, acc telegraf.Accumulator) error {
	defer wg.Done()
	vm_id, err := CheckOSType(domain, lv)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	DomainGatherAll(domain, lv, vm_id, acc)

	return nil
}

func (l *utilsImpl) QemuCommandMetrics(domains []golibvirt.Domain, acc telegraf.Accumulator) error {
	fmt.Println("hello")
	var wg sync.WaitGroup

	for _, domain := range domains {
		wg.Add(1)
		go GatherAll(domain, &wg, l.libvirt, acc)
	}

	fmt.Println("world")
	wg.Wait()
	return nil
}

// func (q *Qemuga) getMemoryInfo(acc telegraf.Accumulator, domain *libvirt.Domain, name string, uuid string) error {
// 	fileExists := true
// 	memoryData, err := common.ReadGuestFile(domain, "/proc/meminfo")
// 	if err != nil && strings.Contains(err.Error(), "No such file or directory") {
// 		fileExists = false
// 	} else if err != nil {
// 		return err
// 	}
// 	if !fileExists {
// 		return err
// 	}
// 	tags := map[string]string{
// 		"domain":  name,
// 		"vm_uuid": uuid,
// 	}
// 	totalMemory := float64(0)
// 	freeMemory := float64(0)
// 	//availableMemory := float64(0)
// 	bufferMemory := float64(0)
// 	cacheMemory := float64(0)
// 	shareMemory := float64(0)
// 	srlabMemory := float64(0)
// 	memoryInfo := []string{}
// 	memoryArray := strings.Split(memoryData, "\n")
// 	for _, memory := range memoryArray {
// 		memoryStr := strings.Replace(memory, ":", " ", -1)
// 		memoryInfo = strings.Fields(memoryStr)
// 		if len(memoryInfo) == 0 {
// 			continue
// 		}
// 		switch memoryInfo[0] {
// 		case "MemTotal":
// 			totalMemory = common.StringToFloat(memoryInfo[1])
// 		case "MemFree":
// 			freeMemory = common.StringToFloat(memoryInfo[1])
// 		//case "MemAvailable":
// 		//	availableMemory = common.StringToFloat(memoryInfo[1])
// 		case "Buffers":
// 			bufferMemory = common.StringToFloat(memoryInfo[1])
// 		case "Cached":
// 			cacheMemory = common.StringToFloat(memoryInfo[1])
// 		case "Shmem":
// 			shareMemory = common.StringToFloat(memoryInfo[1])
// 		case "SReclaimable":
// 			srlabMemory = common.StringToFloat(memoryInfo[1])
// 		}
// 	}
// 	fields := map[string]interface{}{
// 		"vm_total_memory":          totalMemory,
// 		"vm_free_memory":           freeMemory,
// 		"vm_available_memory":      freeMemory + bufferMemory + cacheMemory + srlabMemory,
// 		"vm_buffer_memory":         bufferMemory,
// 		"vm_cache_memory":          cacheMemory + srlabMemory,
// 		"vm_share_memory":          shareMemory,
// 		"vm_mem_available_percent": 100 * (freeMemory + bufferMemory + cacheMemory + srlabMemory) / totalMemory,
// 	}

// 	acc.AddFields("qga", fields, tags)
// 	return nil
// }

// cmdExecStatus := fmt.Sprintf(`{"execute": "guest-exec-status", "arguments": { "pid": %d } }`, execRes.Return.Pid)

// res, err = lv.QEMUDomainAgentCommand(domain, cmdExecStatus, 5, 0)
// if err != nil {
// 	fmt.Println(err.Error())

// 	return "", err
// }
// execStatusRes := &GuestExecStatusReturn{}
// err = json.Unmarshal([]byte(res[0]), execStatusRes)
// fmt.Println("execStatusRes")
// fmt.Println(execStatusRes)

// if execStatusRes.Return.Exited {
// 	stdOutBytes, err := base64.StdEncoding.DecodeString(execStatusRes.Return.OutData)
// 	if err != nil {
// 		return "", err
// 	}
// 	stdOut := string(stdOutBytes)
// 	exitCode := execStatusRes.Return.ExitCode
// 	exited := true
// 	fmt.Println("stdOut")
// 	fmt.Println(stdOut)
// 	fmt.Println("exitCode")
// 	fmt.Println(exitCode)
// 	fmt.Println("exited")
// 	fmt.Println(exited)

// }
