// Package upnp is a port of pocketmine\network\upnp: forwarding the server port on the router
// with UPnP (pocketmine.yml network.upnp-forwarding).
package upnp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pocketmine-go/pocketmine/utils"
)

// UPnPError is a port of pocketmine\network\upnp\UPnPException.
type UPnPError struct{ Message string }

func (e *UPnPError) Error() string { return e.Message }

const maxDiscoveryAttempts = 3

var locationPattern = regexp.MustCompile(`(?i)location\s*:\s*(.+)\n`)

// upnpDevice is the part of a UPnP device description GetServiceURL walks.
type upnpDevice struct {
	DeviceType string       `xml:"deviceType"`
	Devices    []upnpDevice `xml:"deviceList>device"`
	Services   []struct {
		ServiceType string `xml:"serviceType"`
		ControlURL  string `xml:"controlURL"`
	} `xml:"serviceList>service"`
}

// findControlURL is the XPath query of UPnP::getServiceUrl:
// InternetGatewayDevice:1 / WANDevice:1 / WANConnectionDevice:1 / service WANIPConnection:1 /
// controlURL, searched at any depth for the gateway device.
func findControlURL(d upnpDevice) (string, bool) {
	if d.DeviceType == "urn:schemas-upnp-org:device:InternetGatewayDevice:1" {
		for _, wan := range d.Devices {
			if wan.DeviceType != "urn:schemas-upnp-org:device:WANDevice:1" {
				continue
			}
			for _, conn := range wan.Devices {
				if conn.DeviceType != "urn:schemas-upnp-org:device:WANConnectionDevice:1" {
					continue
				}
				for _, s := range conn.Services {
					if s.ServiceType == "urn:schemas-upnp-org:service:WANIPConnection:1" {
						return s.ControlURL, true
					}
				}
			}
		}
	}
	for _, child := range d.Devices {
		if u, ok := findControlURL(child); ok {
			return u, true
		}
	}
	return "", false
}

// GetServiceURL is a port of UPnP::getServiceUrl: discovers the router with SSDP and returns the
// control URL of its WANIPConnection service.
func GetServiceURL() (string, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return "", &UPnPError{Message: "Socket error: " + err.Error()}
	}
	defer conn.Close()

	contents := "M-SEARCH * HTTP/1.1\r\n" +
		"MX: 2\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"ST: upnp:rootdevice\r\n\r\n"
	multicast := &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900}

	location := ""
	buf := make([]byte, 1024)
discovery:
	for i := 0; i < maxDiscoveryAttempts; i++ {
		if n, err := conn.WriteTo([]byte(contents), multicast); err != nil {
			return "", &UPnPError{Message: "Socket error: " + err.Error()}
		} else if n != len(contents) {
			return "", &UPnPError{Message: "Socket error: Unable to send the entire contents."}
		}
		for {
			_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue discovery
				}
				return "", &UPnPError{Message: "Socket error: " + err.Error()}
			}
			if m := locationPattern.FindSubmatch(buf[:n]); m != nil { //this might be garbage from somewhere other than the router
				location = strings.TrimSpace(string(m[1]))
				break discovery
			}
		}
	}
	if location == "" {
		return "", &UPnPError{Message: "Unable to find the router. Ensure that network discovery is enabled in Control Panel."}
	}

	u, err := url.Parse(location)
	if err != nil {
		return "", &UPnPError{Message: "Failed to parse the router's url: " + location}
	}
	if u.Hostname() == "" {
		return "", &UPnPError{Message: "Failed to recognize the host name from the router's url: " + location}
	}
	if u.Port() == "" {
		return "", &UPnPError{Message: "Failed to recognize the port number from the router's url: " + location}
	}
	port, _ := strconv.Atoi(u.Port())

	response, err := utils.GetURL(location, 3, nil)
	if err != nil {
		return "", &UPnPError{Message: "Unable to access XML: " + err.Error()}
	}
	if response.Code != 200 {
		return "", &UPnPError{Message: "Unable to access XML: " + response.Body}
	}

	var root struct {
		Device upnpDevice `xml:"device"`
	}
	if err := xml.Unmarshal([]byte(response.Body), &root); err != nil {
		return "", &UPnPError{Message: "Broken XML."}
	}
	controlURL, ok := findControlURL(root.Device)
	if !ok {
		return "", &UPnPError{Message: "Your router does not support portforwarding"}
	}
	return fmt.Sprintf("%s:%d/%s", u.Hostname(), port, controlURL), nil
}

func soapEnvelope(body string) string {
	return `<?xml version="1.0"?>` +
		`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">` +
		`<s:Body>` + body + `</s:Body></s:Envelope>`
}

func soapURL(serviceURL string) string {
	if strings.Contains(serviceURL, "://") {
		return serviceURL
	}
	return "http://" + serviceURL
}

// PortForward is a port of UPnP::portForward.
func PortForward(serviceURL, internalIP string, internalPort, externalPort int) error {
	body := `<u:AddPortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:1">` +
		`<NewRemoteHost></NewRemoteHost>` +
		`<NewExternalPort>` + strconv.Itoa(externalPort) + `</NewExternalPort>` +
		`<NewProtocol>UDP</NewProtocol>` +
		`<NewInternalPort>` + strconv.Itoa(internalPort) + `</NewInternalPort>` +
		`<NewInternalClient>` + internalIP + `</NewInternalClient>` +
		`<NewEnabled>1</NewEnabled>` +
		`<NewPortMappingDescription>PocketMine-MP</NewPortMappingDescription>` +
		`<NewLeaseDuration>0</NewLeaseDuration>` +
		`</u:AddPortMapping>`
	headers := map[string]string{
		"Content-Type": "text/xml",
		"SOAPAction":   `"urn:schemas-upnp-org:service:WANIPConnection:1#AddPortMapping"`,
	}
	if _, err := utils.PostURL(soapURL(serviceURL), strings.NewReader(soapEnvelope(body)), 3, headers); err != nil {
		return &UPnPError{Message: "Failed to portforward using UPnP: " + err.Error()}
	}
	return nil
}

// RemovePortForward is a port of UPnP::removePortForward.
func RemovePortForward(serviceURL string, externalPort int) {
	body := `<u:DeletePortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:1">` +
		`<NewRemoteHost></NewRemoteHost>` +
		`<NewExternalPort>` + strconv.Itoa(externalPort) + `</NewExternalPort>` +
		`<NewProtocol>UDP</NewProtocol>` +
		`</u:DeletePortMapping>`
	headers := map[string]string{
		"Content-Type": "text/xml",
		"SOAPAction":   `"urn:schemas-upnp-org:service:WANIPConnection:1#DeletePortMapping"`,
	}
	_, _ = utils.PostURL(soapURL(serviceURL), strings.NewReader(soapEnvelope(body)), 3, headers)
}
