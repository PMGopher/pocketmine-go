package upnp

import (
	"encoding/xml"
	"testing"
)

const routerDescription = `<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
 <device>
  <deviceType>urn:schemas-upnp-org:device:InternetGatewayDevice:1</deviceType>
  <deviceList><device>
   <deviceType>urn:schemas-upnp-org:device:WANDevice:1</deviceType>
   <deviceList><device>
    <deviceType>urn:schemas-upnp-org:device:WANConnectionDevice:1</deviceType>
    <serviceList><service>
     <serviceType>urn:schemas-upnp-org:service:WANIPConnection:1</serviceType>
     <controlURL>/ctl/IPConn</controlURL>
    </service></serviceList>
   </device></deviceList>
  </device></deviceList>
 </device>
</root>`

func TestFindControlURL(t *testing.T) {
	var root struct {
		Device upnpDevice `xml:"device"`
	}
	if err := xml.Unmarshal([]byte(routerDescription), &root); err != nil {
		t.Fatal(err)
	}
	if u, ok := findControlURL(root.Device); !ok || u != "/ctl/IPConn" {
		t.Fatalf("got %q %v", u, ok)
	}
}
