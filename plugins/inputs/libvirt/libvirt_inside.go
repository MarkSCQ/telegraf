package libvirt

import (
	"fmt"

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

func cpuInside(domains []golibvirt.Domain) error {
	fmt.Println("hello")

	for _, dom := range domains {
		fmt.Println(dom)
		fmt.Println(dom.Name)
		fmt.Println(dom.UUID)
		fmt.Println(dom.ID)

		uuidValue, _ := uuid.FromBytes(dom.UUID[:])

		// Convert UUID to string
		uuidString := uuidValue.String()
	
		fmt.Println(uuidString)
	}
	// QEMUDomainAgentCommand
	fmt.Println("world")

	return nil
}

func memoryInside(domains []golibvirt.Domain) error {
	return nil
}

func diskInside(domains []golibvirt.Domain) {
}
