package upnp

import (
	"errors"
	"fmt"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/utils"
)

// UPnPNetworkInterface is a port of pocketmine\network\upnp\UPnPNetworkInterface: forwards the
// port on the router when registered, and removes the forward on shutdown.
type UPnPNetworkInterface struct {
	logger     log.Logger
	ip         string
	port       int
	serviceURL string
}

// NewUPnPNetworkInterface fails if the server is offline (Internet::$online).
func NewUPnPNetworkInterface(logger log.Logger, ip string, port int) (*UPnPNetworkInterface, error) {
	if !utils.OnlineMode {
		return nil, errors.New("Server is offline")
	}
	return &UPnPNetworkInterface{logger: log.NewPrefixedLogger(logger, "UPnP Port Forwarder"), ip: ip, port: port}, nil
}

func (u *UPnPNetworkInterface) Start() error {
	u.logger.Info("Attempting to portforward...")
	serviceURL, err := GetServiceURL()
	if err == nil {
		u.serviceURL = serviceURL
		var internalIP string
		if internalIP, err = utils.GetInternalIP(); err == nil {
			err = PortForward(serviceURL, internalIP, u.port, u.port)
		}
	}
	if err != nil {
		u.logger.Error("UPnP portforward failed: " + err.Error())
		return nil
	}
	u.logger.Info(fmt.Sprintf("Forwarded %s:%d to external port %d", u.ip, u.port, u.port))
	return nil
}

func (u *UPnPNetworkInterface) SetName(name string) {}

func (u *UPnPNetworkInterface) Tick() {}

func (u *UPnPNetworkInterface) Shutdown() {
	if u.serviceURL == "" {
		return
	}
	RemovePortForward(u.serviceURL, u.port)
}
