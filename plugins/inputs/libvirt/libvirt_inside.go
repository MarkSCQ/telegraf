package libvirt

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	golibvirt "github.com/digitalocean/go-libvirt"
	uuid "github.com/satori/go.uuid"
)

// osID, err := common.AgentGuestInfo(domain)

// func OsCheck(domainMap []golibvirt.Libvirt) error {
// 	domains := make([]string, len(domainMap))
// 	for _, dm := range domainMap {
// 		for k,_ := range dm {
// 			domains = append(domains, k)
// 		}
// 	}
// 	// golibvirt.QEMUDomainAgentCommandArgs()
// 	return nil
// }

type funcMem interface {
	QemuCommandMem(str string)
}

type funcCpu interface {
	QemuCommandCpu(str string)
}

type funcDisk interface {
	QemuCommandDisk(str string)
	QemuCommandDiskIO(str string)
}

type funcNet interface {
	QemuCommandNetState(str string)
	QemuCommandNetSpeed(str string)
	QemuCommandNetUsage(str string)
}

type DomainGather struct {
	funcMem
	funcCpu
	funcDisk
	funcNet
	DomainName string
	DomainUUID string
}

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

type OsInfoRet struct {
	Return OsInfo `json:"return"`
}

func DomainGatherAllLinux(domain golibvirt.Domain, wg *sync.WaitGroup) error {
	defer wg.Done()

	domain_name := domain.Name
	domain_uuid, err := uuid.FromBytes(domain.UUID[:])
	if err != nil {
		log.Printf("Parsing UUID error:  %s, error Domain Name: %s, error Domain UUID: %s", err.Error(), domain_name, domain_uuid)
		return err
	}
	dg := DomainGather{DomainName: domain_name, DomainUUID: domain_uuid.String()}
	fmt.Println("=============================")
	fmt.Println(dg.DomainName)
	fmt.Println(dg.DomainUUID)
	return nil
}

func DomainGatherAllMsWin(domain golibvirt.Domain, wg *sync.WaitGroup) error {
	defer wg.Done()
	return nil
}

func (l *utilsImpl) CheckOSType(domain golibvirt.Domain) (int, error) {
	osInfoCmd := fmt.Sprintf(`{"execute": "guest-get-osinfo"}`)

	res, err := l.libvirt.QEMUDomainAgentCommand(domain, osInfoCmd, 5, 0)
	if err != nil {
		// fmt.Println("err")

		log.Println(err.Error())
		return 0, err
	}

	var dat OsInfoRet

	err = json.Unmarshal([]byte(res[0]), &dat)

	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	vm_id := dat.Return.VmID
	if vm_id == "mswindows" {
		return 2, nil
	} else {
		return 1, nil
	}
}

func (l *utilsImpl) QemuCommandMetrics(domains []golibvirt.Domain) error {
	fmt.Println("hello")
	var wg sync.WaitGroup

	for _, domain := range domains {
		wg.Add(1)
		vm_id, err := l.CheckOSType(domain)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		if vm_id == 1 {
			// Linxu Gather All
			fmt.Println("Get Linux")
			go DomainGatherAllLinux(domain, &wg)
		} else if vm_id == 2 {
			// Windows Gather All

			fmt.Println("Get windows")
			go DomainGatherAllMsWin(domain, &wg)
		} else {
			fmt.Println("Cannot Tell")
			// Pass
		}
	}

	fmt.Println("world")
	wg.Wait()
	return nil
}

func memoryInside(domain golibvirt.Domain) error {
	return nil
}

func diskInside(domain golibvirt.Domain) error {
	return nil
}

// addQemuCommand
// func (l *Libvirt) addQemuCommand(domains []golibvirt.Domain, acc telegraf.Accumulator) {
// 	for _, dom := range domains {
// 		fmt.Println(dom)
// 		fmt.Println(dom.Name)
// 		fmt.Println(dom.UUID)
// 		fmt.Println(dom.ID)

// 		uuidValue, _ := uuid.FromBytes(dom.UUID[:])

// 		// Convert UUID to string
// 		uuidString := uuidValue.String()

// 		fmt.Println(uuidString)
// 	}
// 	// os check and split
// }
