package container

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/abcdlsj/cr"
	"gopkg.in/yaml.v2"
)

type PortMapping struct {
	HostPort      string `yaml:"host_port" json:"host_port"`
	ContainerPort string `yaml:"container_port" json:"container_port"`
}

type VolumeBind struct {
	HostPath      string `yaml:"host_path" json:"host_path"`
	ContainerPath string `yaml:"container_path" json:"container_path"`
}

type ServiceOptions struct {
	ImageName     string
	ContainerName string
	CPU           float64
	Memory        int64 // in bytes
	PortMappings  []PortMapping
	VolumeBinds   []VolumeBind
	NetworkMode   string
	RestartPolicy string
	AutoRemove    bool
}

type OutputWriter interface {
	Write(stage string, format string, args ...interface{})
}

type Deployer interface {
	DeployService(options ServiceOptions, writer OutputWriter) error
}

type DefaultOutputWriter struct {
	Writer io.Writer
}

func (w *DefaultOutputWriter) Write(stage string, format string, args ...interface{}) {
	if stage == "" {
		fmt.Fprintf(w.Writer, "%s\n", fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(w.Writer, "%s: %s\n", cr.PLGreenBold(stage), fmt.Sprintf(format, args...))
	}
}

type ServiceConfig struct {
	Name          string        `yaml:"name" json:"name"`
	Image         string        `yaml:"image" json:"image"`
	ContainerName string        `yaml:"container_name" json:"container_name"`
	CPU           float64       `yaml:"cpu" json:"cpu"`
	Memory        string        `yaml:"memory" json:"memory"`
	PortMappings  []PortMapping `yaml:"port_mappings" json:"port_mappings"`
	VolumeBinds   []VolumeBind  `yaml:"volume_binds" json:"volume_binds"`
	NetworkMode   string        `yaml:"network_mode" json:"network_mode"`
	RestartPolicy string        `yaml:"restart_policy" json:"restart_policy"`
	AutoRemove    bool          `yaml:"auto_remove" json:"auto_remove"`
}

func (s ServiceConfig) String() string {
	m, _ := json.Marshal(s)
	return string(m)
}

type ServicesTemplate struct {
	Services []ServiceConfig `yaml:"services"`
}

func LoadServicesTemplate(filename string) (*ServicesTemplate, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var template ServicesTemplate
	err = yaml.Unmarshal(data, &template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (sc *ServiceConfig) ToServiceOptions() (ServiceOptions, error) {
	memory, err := ParseMemory(sc.Memory)
	if err != nil {
		return ServiceOptions{}, fmt.Errorf("failed to parse memory: %v", err)
	}

	option := ServiceOptions{
		ImageName:     sc.Image,
		ContainerName: sc.ContainerName,
		CPU:           sc.CPU,
		Memory:        memory,
		PortMappings:  sc.PortMappings,
		NetworkMode:   sc.NetworkMode,
		RestartPolicy: sc.RestartPolicy,
		AutoRemove:    sc.AutoRemove,
		VolumeBinds:   sc.VolumeBinds,
	}

	return option, nil
}

// ParseMemory converts a string memory value (e.g., "512M") to int64 bytes
func ParseMemory(memoryString string) (int64, error) {
	// Implement memory parsing logic here
	// This is a simplified example, you might want to add more robust parsing
	var value int64
	var unit string
	_, err := fmt.Sscanf(memoryString, "%d%s", &value, &unit)
	if err != nil {
		return 0, err
	}
	switch unit {
	case "K", "k":
		return value * 1024, nil
	case "M", "m":
		return value * 1024 * 1024, nil
	case "G", "g":
		return value * 1024 * 1024 * 1024, nil
	default:
		return value, nil
	}
}
