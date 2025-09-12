package protocol

import (
	"bytes"
	"net"
	"testing"
)

func TestEndpointBasic(t *testing.T) {
	ep := NewEndpointFromIPPort(0x7f000001, 8080) // 127.0.0.1:8080
	
	if ep.IP() != 0x7f000001 {
		t.Errorf("Expected IP 0x7f000001, got 0x%x", ep.IP())
	}
	
	if ep.Port() != 8080 {
		t.Errorf("Expected port 8080, got %d", ep.Port())
	}
	
	if !ep.Defined() {
		t.Error("Endpoint should be defined")
	}
}

func TestEndpointFromString(t *testing.T) {
	ep, err := NewEndpointFromString("192.168.1.1:1234")
	if err != nil {
		t.Fatalf("Failed to create endpoint from string: %v", err)
	}
	
	if ep.Port() != 1234 {
		t.Errorf("Expected port 1234, got %d", ep.Port())
	}
	
	expectedIP := net.ParseIP("192.168.1.1")
	if !ep.IPNet().Equal(expectedIP) {
		t.Errorf("Expected IP %s, got %s", expectedIP.String(), ep.IPNet().String())
	}
}

func TestEndpointFromTCPAddr(t *testing.T) {
	addr := &net.TCPAddr{
		IP:   net.ParseIP("10.0.0.1"),
		Port: 4662,
	}
	
	ep := NewEndpointFromTCPAddr(addr)
	
	if ep.Port() != 4662 {
		t.Errorf("Expected port 4662, got %d", ep.Port())
	}
	
	if !ep.IPNet().Equal(net.ParseIP("10.0.0.1")) {
		t.Errorf("Expected IP 10.0.0.1, got %s", ep.IPNet().String())
	}
}

func TestEndpointSerialize(t *testing.T) {
	ep := NewEndpointFromIPPort(0x12345678, 0x9ABC)
	buf := &bytes.Buffer{}
	
	err := ep.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put endpoint: %v", err)
	}
	
	if buf.Len() != 6 {
		t.Errorf("Expected buffer length 6, got %d", buf.Len())
	}
	
	// Read back
	result := NewEndpoint()
	err = result.Get(buf)
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}
	
	if !ep.Equals(result) {
		t.Errorf("Serialization/deserialization failed: %s != %s", ep.String(), result.String())
	}
}

func TestEndpointString(t *testing.T) {
	ep := NewEndpointFromIPPort(0x7f000001, 8080) // This represents IP bytes in network order
	
	// The IP 0x7f000001 in network order represents 127.0.0.1
	expected := "127.0.0.1:8080"
	if ep.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ep.String())
	}
}

func TestEndpointEquals(t *testing.T) {
	ep1 := NewEndpointFromIPPort(0x12345678, 8080)
	ep2 := NewEndpointFromIPPort(0x12345678, 8080)
	ep3 := NewEndpointFromIPPort(0x12345679, 8080)
	
	if !ep1.Equals(ep2) {
		t.Error("Equal endpoints should be equal")
	}
	
	if ep1.Equals(ep3) {
		t.Error("Different endpoints should not be equal")
	}
}

func TestEndpointCompareTo(t *testing.T) {
	ep1 := NewEndpointFromIPPort(0x12345678, 8080)
	ep2 := NewEndpointFromIPPort(0x12345678, 8080)
	ep3 := NewEndpointFromIPPort(0x12345679, 8080)
	
	if ep1.CompareTo(ep2) != 0 {
		t.Error("Equal endpoints should compare to 0")
	}
	
	if ep1.CompareTo(ep3) >= 0 {
		t.Error("ep1 should be less than ep3")
	}
}