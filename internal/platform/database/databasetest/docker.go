package databasetest

import (
	"bytes"
	"encoding/json"
	"net"
	"os/exec"
	"testing"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}

// StartContainer runs a postgres container to execute commands.
func StartContainer(t *testing.T, image string, port string, args ...string) *Container {
	t.Helper() // marks this func as a test helper function

	arg := []string{"run", "-P", "-d"}
	arg = append(arg, args...)
	arg = append(arg, image)

	cmd := exec.Command("docker", arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container: %v", err)
	}

	id := out.String()[:12]
	t.Log("DB ContainerID:", id)

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}

	ip, randPort := extractIPPort(t, doc, port)

	c := Container{
		ID:   id,
		Host: net.JoinHostPort(ip, randPort),
	}

	t.Log("DB Host:", c.Host)

	return &c
}

// StopContainer stops and removes the specified container.
func StopContainer(t *testing.T, c *Container) {
	t.Helper()

	if err := exec.Command("docker", "stop", c.ID).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", c.ID)

	if err := exec.Command("docker", "rm", c.ID, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", c.ID)
}

// DumpContainerLogs runs "docker logs" against the container and send it to t.Log
func DumpContainerLogs(t *testing.T, c *Container) {
	t.Helper()

	out, err := exec.Command("docker", "logs", c.ID).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", c.ID, out)
}

// extracts network settings based on the specified port.
func extractIPPort(t *testing.T, doc []map[string]interface{}, port string) (string, string) {
	nw, exists := doc[0]["NetworkSettings"]
	if !exists {
		t.Fatal("could not get network settings")
	}
	ports, exists := nw.(map[string]interface{})["Ports"]
	if !exists {
		t.Fatal("could not get network ports settings")
	}
	tcp, exists := ports.(map[string]interface{})[port+"/tcp"]
	if !exists {
		t.Fatal("could not get network ports/tcp settings")
	}
	list, exists := tcp.([]interface{})
	if !exists {
		t.Fatal("could not get network ports/tcp list settings")
	}
	if len(list) < 1 {
		t.Fatal("could not get network ports/tcp list settings")
	}
	data, exists := list[0].(map[string]interface{})
	if !exists {
		t.Fatal("could not get network ports/tcp list data")
	}

	return data["HostIp"].(string), data["HostPort"].(string)
}
