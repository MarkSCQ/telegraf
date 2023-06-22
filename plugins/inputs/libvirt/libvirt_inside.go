package libvirt

import (
	"fmt"
	"log"

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

func DomainGatherAll(domain golibvirt.Domain) error {
	domain_name := domain.Name
	domain_uuid, err := uuid.FromBytes(domain.UUID[:])
	if err != nil {
		log.Printf("Parsing UUID error:  %s, error Domain Name: %s, error Domain UUID: %s", err.Error(), domain_name, domain_uuid)
		return err
	}
	dg := DomainGather{DomainName: domain_name, DomainUUID: domain_uuid.String()}
	fmt.Println(dg.DomainName)
	fmt.Println(dg.DomainUUID)
	return nil
}

func (l *utilsImpl) CheckOSType(domain golibvirt.Domain) error {
	osInfoExec := fmt.Sprintf(`{"execute": "guest-get-osinfo"}`)
	// osInfoExec := fmt.Sprintf(`ls `)
	// t := golibvirt.Libvirt{}
	// t.QEMUDomainAgentCommand()
	// QEMUDomainAgentCommandArgs{}

	res, err := l.libvirt.QEMUDomainAgentCommand(domain, osInfoExec, 5, 0)
	if err != nil {
		fmt.Println("err")
		fmt.Println(err)
		return err
	}
	fmt.Println("res")
	fmt.Println(res)

	domain_uuid, err := uuid.FromBytes(domain.UUID[:])
	fmt.Println(domain_uuid)

	// QEMUDomainAgentCommand(Dom Domain, Cmd string, Timeout int32, Flags uint32) (rResult OptString, err error) {

	// osInfoExec := fmt.Sprintf(`{"fexecute": "guest-get-osinfo"}`)
	// domainName, err := domain.GetName()

	// output, err := domain.QemuAgentCommand(osInfoExec, libvirt.DOMAIN_QEMU_AGENT_COMMAND_DEFAULT, 0)
	// if err != nil {
	// 	log.Println("========================> ReadGuestFile guest-get-osinfo")
	// 	log.Println(domainName + " ===> " + err.Error())
	// 	return "", err
	// }
	// var dat osInfoStruct
	// err = json.Unmarshal([]byte(output), &dat)
	// if err != nil {
	// 	return "", err
	// }

	// return dat.Return.ID, nil
	return nil
}

func (l *utilsImpl) QemuCommandMetrics(domains []golibvirt.Domain) error {
	fmt.Println("hello")
	// var wg sync.WaitGroup

	for _, domain := range domains {
		// wg.Add(1)
		l.CheckOSType(domain)
		// fmt.Println(dom)
		fmt.Println(domain.Name)
		fmt.Println(domain.UUID)
		fmt.Println(domain.ID)

		// uuidValue, _ := uuid.FromBytes(dom.UUID[:])

		// // Convert UUID to string
		// uuidString := uuidValue.String()

		// fmt.Println(uuidString)
	}
	// QEMUDomainAgentCommand
	fmt.Println("world")
	// wg.Wait()
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
