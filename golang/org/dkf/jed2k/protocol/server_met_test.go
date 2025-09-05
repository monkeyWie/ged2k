package protocol

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseServerMet(t *testing.T) {
	buf, err := os.ReadFile("./testdata/server.met")
	if err != nil {
		t.Fatalf("failed to read server.met: %v", err)
	}

	servers, err := ParseServerMet(buf)
	if err != nil {
		t.Fatalf("failed to parse server.met: %v", err)
	}

	jsonBuf, err := json.Marshal(servers)
	if err != nil {
		t.Fatalf("failed to marshal servers to JSON: %v", err)
	}

	expectedJson, err := os.ReadFile("./testdata/server.met.json")
	if err != nil {
		t.Fatalf("failed to read expected JSON: %v", err)
	}

	if string(jsonBuf) != string(expectedJson) {
		t.Errorf("parsed servers JSON does not match expected output.\nGot: %s\nWant: %s", string(jsonBuf), string(expectedJson))
	}
}
