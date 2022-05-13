// Copyright 2020-2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config contains the primary configuration options that are used when doing operations
// in the tanzu package.
package config

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ClusterConfigFile = "ClusterConfigFile"
	Tty               = "Tty"
	yamlIndent        = 2

	// UnmanagedClusterConfig option keys
	ClusterName               = "ClusterName"
	KubeconfigPath            = "KubeconfigPath"
	ExistingClusterKubeconfig = "ExistingClusterKubeconfig"
	NodeImage                 = "NodeImage"
	Provider                  = "Provider"
	Cni                       = "Cni"
	PodCIDR                   = "PodCidr"
	ServiceCIDR               = "ServiceCidr"
	TKRLocation               = "TkrLocation"
	AdditionalPackageRepos    = "AdditionalPackageRepos"
	PortsToForward            = "PortsToForward"
	SkipPreflightChecks       = "SkipPreflightChecks"
	ControlPlaneNodeCount     = "ControlPlaneNodeCount"
	WorkerNodeCount           = "WorkerNodeCount"
	InstallPackages           = "InstallPackages"
	LogFile                   = "LogFile"

	configDir          = ".config"
	tanzuConfigDir     = "tanzu"
	tkgConfigDir       = "tkg"
	unmanagedConfigDir = "unmanaged"
	defaultName        = "default-name"

	ProtocolTCP  = "tcp"
	ProtocolUDP  = "udp"
	ProtocolSCTP = "sctp"

	ProviderKind     = "kind"
	ProviderMinikube = "minikube"
	ProviderNone     = "none"
)

var defaultConfigValues = map[string]interface{}{
	Provider:              "kind",
	Cni:                   "calico",
	PodCIDR:               "10.244.0.0/16",
	ServiceCIDR:           "10.96.0.0/16",
	Tty:                   "true",
	ControlPlaneNodeCount: "1",
	WorkerNodeCount:       "0",
}

// Used to generate the empty, default config
var emptyConfig = map[string]interface{}{
	ClusterConfigFile:      "",
	ClusterName:            "",
	Tty:                    "",
	TKRLocation:            "",
	Provider:               "",
	Cni:                    "",
	PodCIDR:                "",
	ServiceCIDR:            "",
	ControlPlaneNodeCount:  "",
	WorkerNodeCount:        "",
	AdditionalPackageRepos: []string{},
}

// PortMap is the mapping between a host port and a container port.
type PortMap struct {
	// ListenAddress is the listening address to attach on the host machine
	ListenAddress string `yaml:"ListenAddress,omitempty"`
	// HostPort is the port on the host machine.
	HostPort int `yaml:"HostPort,omitempty"`
	// ContainerPort is the port on the container to map to.
	ContainerPort int `yaml:"ContainerPort"`
	// Protocol is the IP protocol (TCP, UDP, SCTP).
	Protocol string `yaml:"Protocol,omitempty"`
}

type InstallPackage struct {
	Name      string `yaml:"name,omitempty"`
	Config    string `yaml:"config,omitempty"`
	Version   string `yaml:"version,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

// UnmanagedClusterConfig contains all the configuration settings for creating a
// unmanaged Tanzu cluster.
type UnmanagedClusterConfig struct {
	// ClusterName is the name of the cluster.
	ClusterName string `yaml:"ClusterName"`
	// KubeconfigPath is the location where the Kubeconfig will be persisted
	// after the cluster is created.
	KubeconfigPath string `yaml:"KubeconfigPath"`
	// ExistingClusterKubeconfig is the serialized path to the kubeconfig to use of an existing cluster.
	ExistingClusterKubeconfig string `yaml:"ExistingClusterKubeconfig"`
	// NodeImage is the host OS image to use for Kubernetes nodes.
	// It is typically resolved, automatically, in the Taznu Kubernetes Release (TKR) BOM,
	// but also can be overridden in configuration.
	NodeImage string `yaml:"NodeImage"`
	// Provider is the unmanaged infrastructure provider to use (e.g. kind).
	Provider string `yaml:"Provider"`
	// ProviderConfiguration offers optional provider-specific configuration.
	// The exact keys and values accepted are determined by the provider.
	ProviderConfiguration map[string]interface{} `yaml:"ProviderConfiguration"`
	// CNI is the networking CNI to use in the cluster. Default is calico.
	Cni string `yaml:"Cni"`
	// CNIConfiguration offers optional cni-plugin specific configuration.
	// The exact keys and values accepted are determined by the CNI choice.
	CNIConfiguration map[string]interface{} `yaml:"CniConfiguration"`
	// PodCidr is the Pod CIDR range to assign pod IP addresses.
	PodCidr string `yaml:"PodCidr"`
	// ServiceCidr is the Service CIDR range to assign service IP addresses.
	ServiceCidr string `yaml:"ServiceCidr"`
	// TkrLocation is the path to the Tanzu Kubernetes Release (TKR) data.
	TkrLocation string `yaml:"TkrLocation"`
	// AdditionalPackageRepos are the extra package repositories to install during bootstrapping
	AdditionalPackageRepos []string `yaml:"AdditionalPackageRepos"`
	// PortsToForward contains a mapping of host to container ports that should
	// be exposed.
	PortsToForward []PortMap `yaml:"PortsToForward"`
	// SkipPreflightChecks determines whether preflight checks are performed prior
	// to attempting to deploy the cluster.
	SkipPreflightChecks bool `yaml:"SkipPreflightChecks"`
	// ControlPlaneNodeCount is the number of control plane nodes to deploy for the cluster.
	// Default is 1
	ControlPlaneNodeCount string `yaml:"ControlPlaneNodeCount"`
	// WorkerNodeCount is the number of worker nodes to deploy for the cluster.
	// Default is 0
	WorkerNodeCount string `yaml:"WorkerNodeCount"`
	// InstallPackages is a set of packages to install, including the package name, (optional) version, (optional) config
	InstallPackages []InstallPackage `yaml:"InstallPackages"`
	// LogFile is the log file to send provider bootstrapping logs to
	// should be a fully qualified path
	LogFile string `yaml:"LogFile"`
}

// KubeConfigPath gets the full path to the KubeConfig for this unmanaged cluster.
func (scc *UnmanagedClusterConfig) KubeConfigPath() (string, error) {
	path, err := GetTanzuConfigPath()
	if err != nil {
		return "", fmt.Errorf("")
	}

	return filepath.Join(path, scc.ClusterName+".yaml"), nil
}

// GetTanzuConfigPath returns the filepath to the config directory.
// For example, on linux, "~/.config/tanzu/"
// Returns an error if the user home directory path cannot be resolved
func GetTanzuConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve tanzu config path. Error: %s", err.Error())
	}

	return filepath.Join(home, configDir, tanzuConfigDir), nil
}

// GetTanzuTkgConfigPath returns the filepath to the tanzu tkg config directory.
// For example, on linux, "~/.config/tanzu/tkg"
// Returns an error if the path cannot be resolved
func GetTanzuTkgConfigPath() (string, error) {
	path, err := GetTanzuConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to resolve tanzu TKG config path. Error: %s", err.Error())
	}

	return filepath.Join(path, tkgConfigDir), nil
}

// GetUnmanagedConfigPath returns the filepath to the unmanaged config directory.
// For example, on linux, "~/.config/tanzu/tkg/unmanaged"
// Returns an error if the path cannot be resolved
func GetUnmanagedConfigPath() (string, error) {
	path, err := GetTanzuTkgConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to resolve unmanaged-cluster config path. Error: %s", err.Error())
	}

	return filepath.Join(path, unmanagedConfigDir), nil
}

// InitializeConfiguration determines the configuration to use for cluster creation.
//
// There are three places where configuration comes from:
// - default settings
// - configuration file
// - environment variables
// - command line arguments
//
// The effective configuration is determined by combining these sources, in ascending
// order of preference listed. So env variables override values in the config file,
// and explicit CLI arguments override config file and env variable values.
func InitializeConfiguration(commandArgs map[string]interface{}) (*UnmanagedClusterConfig, error) {
	config := &UnmanagedClusterConfig{}

	// First, populate values based on a supplied config file
	// Check if config file was passed in and can be cast as string
	if configFile, ok := commandArgs[ClusterConfigFile].(string); ok && configFile != "" {
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(configData, config)
		if err != nil {
			return nil, err
		}
	}

	// Loop through and look up each field
	element := reflect.ValueOf(config).Elem()
	for i := 0; i < element.NumField(); i++ {
		fStructField := element.Type().Field(i)
		f := element.Field(i)
		fInt := f.Interface()

		switch fInt.(type) {
		case bool:
			setBoolValue(commandArgs, &element, &fStructField)
		case string:
			setStringValue(commandArgs, &element, &fStructField)
		case []string:
			setStringSliceValue(commandArgs, &element, &fStructField)
		case []InstallPackage:
			setInstallPackageSliceValue(commandArgs, &element, &fStructField)
		case []PortMap:
			setPortMapSliceValue(commandArgs, &element, &fStructField)
		default:
		}
		// skip fields that are not supported
	}

	// Make sure cluster name was either set on the command line or in the config
	// file.
	if config.ClusterName == "" {
		return nil, fmt.Errorf("cluster name must be provided")
	}

	// Sanatize the filepath for the provided kubeconfig
	config.ExistingClusterKubeconfig = sanatizeKubeconfigPath(config.ExistingClusterKubeconfig)

	return config, nil
}

func GenerateDefaultConfig() *UnmanagedClusterConfig {
	config := &UnmanagedClusterConfig{}

	// Loop through and look up each field
	// Because emptyConfig is used, should generate the default values
	element := reflect.ValueOf(config).Elem()
	for i := 0; i < element.NumField(); i++ {
		fStructField := element.Type().Field(i)
		f := element.Field(i)
		fInt := f.Interface()

		switch fInt.(type) {
		case bool:
			setBoolValue(emptyConfig, &element, &fStructField)
		case string:
			setStringValue(emptyConfig, &element, &fStructField)
		case []string:
			setStringSliceValue(emptyConfig, &element, &fStructField)
		case []InstallPackage:
			setInstallPackageSliceValue(emptyConfig, &element, &fStructField)
		case []PortMap:
			setPortMapSliceValue(emptyConfig, &element, &fStructField)
		default:
		}
		// skip fields that are not supported
	}

	config.ClusterName = defaultName

	return config
}

// setBoolValue takes an arbitrary map of string / interfaces, a reflect.Value, and the struct field to be filled.
// Always assumes the value being passed in is a bool.
// The bool value then gets set into the struct field
func setBoolValue(commandArgs map[string]interface{}, element *reflect.Value, field *reflect.StructField) {
	// Use the yaml name if provided so it matches what we serialize to file
	fieldName := field.Tag.Get("yaml")
	if fieldName == "" {
		fieldName = field.Name
	}

	// Check if an explicit value was passed in
	if value, ok := commandArgs[fieldName]; ok {
		element.FieldByName(field.Name).SetBool(value.(bool))
	} else if value := os.Getenv(fieldNameToEnvName(fieldName)); value != "" {
		// See if there is an environment variable set for this field
		// if there is an error with parsing bool, will always set to false
		b, _ := strconv.ParseBool(value)
		element.FieldByName(field.Name).SetBool(b)
	} else {
		// Only set to the default value if it hasn't been set already
		if value, ok := defaultConfigValues[fieldName]; ok {
			element.FieldByName(field.Name).SetBool(value.(bool))
		}
	}
}

// setStringValue takes an arbitrary map of string / interfaces, a reflect.Value, and the struct field to be filled.
// Always assumes the value being passed in is a string.
// The string value then gets set into the struct field
func setStringValue(commandArgs map[string]interface{}, element *reflect.Value, field *reflect.StructField) {
	// Use the yaml name if provided so it matches what we serialize to file
	fieldName := field.Tag.Get("yaml")
	if fieldName == "" {
		fieldName = field.Name
	}

	// Check if an explicit value was passed in
	if value, ok := commandArgs[fieldName]; ok && value != "" {
		element.FieldByName(field.Name).SetString(value.(string))
	} else if value := os.Getenv(fieldNameToEnvName(fieldName)); value != "" {
		// See if there is an environment variable set for this field
		element.FieldByName(field.Name).SetString(value)
	}

	// Only set to the default value if it hasn't been set already
	if element.FieldByName(field.Name).String() == "" {
		if value, ok := defaultConfigValues[fieldName]; ok {
			element.FieldByName(field.Name).SetString(value.(string))
		}
	}
}

// setStringSliceValue takes an arbitrary map of string / interfaces, a reflect.Value, and the struct field to be filled.
// Always assumes the value being passed in is a string slice.
// A new slice is created and the struct field is set to the slice.
func setStringSliceValue(commandArgs map[string]interface{}, element *reflect.Value, field *reflect.StructField) {
	// Use the yaml name if provided so it matches what we serialize to file
	fieldName := field.Tag.Get("yaml")
	if fieldName == "" {
		fieldName = field.Name
	}

	// Check if an explicit value was passed in
	if slice, ok := commandArgs[fieldName]; ok && len(slice.([]string)) != 0 {
		for _, val := range slice.([]string) {
			oldSlice := element.FieldByName(field.Name)
			newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
			element.FieldByName(field.Name).Set(newSlice)
		}
	} else if value := os.Getenv(fieldNameToEnvName(fieldName)); value != "" {
		// Split the env var on `,` for setting multiple values
		values := strings.Split(value, ",")
		for _, val := range values {
			oldSlice := element.FieldByName(field.Name)
			newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
			element.FieldByName(field.Name).Set(newSlice)
		}
	}

	// Only set to the default value if it hasn't been set already
	if element.FieldByName(field.Name).Len() == 0 {
		if slice, ok := defaultConfigValues[fieldName]; ok {
			for _, val := range slice.([]string) {
				oldSlice := element.FieldByName(field.Name)
				newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
				element.FieldByName(field.Name).Set(newSlice)
			}
		}
	}
}

// setPortMapSliceValue takes an arbitrary map of string / interfaces, a reflect.Value, and the struct field to be filled.
// Always assumes the value being passed in is a config.PortMap slice.
// A new slice is created and the struct field is set to the slice.
func setPortMapSliceValue(commandArgs map[string]interface{}, element *reflect.Value, field *reflect.StructField) { //nolint:dupl
	// Use the yaml name if provided so it matches what we serialize to file
	fieldName := field.Tag.Get("yaml")
	if fieldName == "" {
		fieldName = field.Name
	}

	// Check if an explicit value was passed in
	if slice, ok := commandArgs[fieldName]; ok && len(slice.([]PortMap)) != 0 {
		for _, val := range slice.([]PortMap) {
			oldSlice := element.FieldByName(field.Name)
			newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
			element.FieldByName(field.Name).Set(newSlice)
		}
	} else if value := os.Getenv(fieldNameToEnvName(fieldName)); value != "" {
		portMappings, _ := ParsePortMappings([]string{value})
		oldSlice := element.FieldByName(field.Name)
		newSlice := reflect.Append(oldSlice, reflect.ValueOf(portMappings))
		element.FieldByName(field.Name).Set(newSlice)
	}

	// Only set to the default value if it hasn't been set already
	if element.FieldByName(field.Name).Len() == 0 {
		if slice, ok := defaultConfigValues[fieldName]; ok {
			for _, val := range slice.([]PortMap) {
				oldSlice := element.FieldByName(field.Name)
				newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
				element.FieldByName(field.Name).Set(newSlice)
			}
		}
	}
}

// setInstallPackageSliceValue takes an arbitrary map of string / interfaces, a reflect.Value, and the struct field to be filled.
// Always assumes the value being passed in is a config.InstallPackage slice.
// A new slice is created and the struct field is set to the slice.
func setInstallPackageSliceValue(commandArgs map[string]interface{}, element *reflect.Value, field *reflect.StructField) { //nolint:dupl
	// Use the yaml name if provided so it matches what we serialize to file
	fieldName := field.Tag.Get("yaml")
	if fieldName == "" {
		fieldName = field.Name
	}

	// Check if an explicit value was passed in
	if slice, ok := commandArgs[fieldName]; ok && len(slice.([]InstallPackage)) != 0 {
		for _, val := range slice.([]InstallPackage) {
			oldSlice := element.FieldByName(field.Name)
			newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
			element.FieldByName(field.Name).Set(newSlice)
		}
	} else if value := os.Getenv(fieldNameToEnvName(fieldName)); value != "" {
		installPackages, _ := ParseInstallPackageMappings([]string{value})
		oldSlice := element.FieldByName(field.Name)
		newSlice := reflect.Append(oldSlice, reflect.ValueOf(installPackages))
		element.FieldByName(field.Name).Set(newSlice)
	}

	// Only set to the default value if it hasn't been set already
	if element.FieldByName(field.Name).Len() == 0 {
		if slice, ok := defaultConfigValues[fieldName]; ok {
			for _, val := range slice.([]InstallPackage) {
				oldSlice := element.FieldByName(field.Name)
				newSlice := reflect.Append(oldSlice, reflect.ValueOf(val))
				element.FieldByName(field.Name).Set(newSlice)
			}
		}
	}
}

// fieldNameToEnvName converts the config values yaml name to its expected env
// variable name.
func fieldNameToEnvName(field string) string {
	namedArray := []string{"TANZU"}
	re := regexp.MustCompile(`[A-Z][^A-Z]*`)
	allWords := re.FindAllString(field, -1)
	for _, word := range allWords {
		namedArray = append(namedArray, strings.ToUpper(word))
	}
	return strings.Join(namedArray, "_")
}

func sanatizeKubeconfigPath(path string) string {
	var builder string

	// handle tildas at the beginning of the path
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		builder = filepath.Join(builder, usr.HomeDir)
		path = path[2:]
	}

	builder = filepath.Join(builder, path)

	return builder
}

// RenderConfigToFile take a file path and serializes the configuration data to that path. It expects the path
// to not exist, if it does, an error is returned.
func RenderConfigToFile(filePath string, config interface{}) error {
	// check if file exists
	// determine if directory pre-exists
	_, err := os.ReadDir(filePath)

	// if it does not exist, which is expected, create it
	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to create config file at %q, does it already exist", filePath)
	}

	var rawConfig bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&rawConfig)
	yamlEncoder.SetIndent(yamlIndent)

	err = yamlEncoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to render configuration file. Error: %s", err.Error())
	}
	err = os.WriteFile(filePath, rawConfig.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write rawConfig file. Error: %s", err.Error())
	}
	// if it does, return an error
	// otherwise, write config to file
	return nil
}

// RenderFileToConfig reads in configuration from a file and returns the
// UnmanagedClusterConfig structure based on it. If the file does not exist or there
// is a problem reading the configuration from it an error is returned.
func RenderFileToConfig(filePath string) (*UnmanagedClusterConfig, error) {
	d, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading config file. Error: %s", err.Error())
	}
	scc := &UnmanagedClusterConfig{}
	err = yaml.Unmarshal(d, scc)
	if err != nil {
		return nil, fmt.Errorf("configuration at %s was invalid. Error: %s", filePath, err.Error())
	}

	return scc, nil
}

// ParsePortMappings parses a slice of the command line string format into a slice of PortMap structs.
// Supported formats are just container port ("80"), container port to host port
// ("80:80"), container port to host port with protocol ("80:80/tcp"),
// or adding listen address prefixed to the above options ("127.0.0.1:80:80/tcp")
func ParsePortMappings(portMappings []string) ([]PortMap, error) {
	mappings := []PortMap{}

	for _, pm := range portMappings {
		result := PortMap{}

		// See if protocol is provided
		parts := strings.Split(pm, "/")
		if len(parts) == 2 { //nolint:gomnd
			p := strings.ToLower(parts[1])
			if p != ProtocolTCP && p != ProtocolUDP && p != ProtocolSCTP {
				return nil, fmt.Errorf("failed to parse protocol %q, must be tcp, udp, or sctp", p)
			}
			result.Protocol = p
		}

		// Now see if we have just container, or container:host
		parts = strings.Split(parts[0], ":")

		switch len(parts) {
		case 3:
			result.ListenAddress = parts[0]

			containerPort, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse port mapping, detected format listenAddress:port:port, invalid container port provided: %q", parts[1])
			}
			result.ContainerPort = containerPort

			hostPort, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, fmt.Errorf("failed to parse port mapping, detected format listenAddress:port:port, invalid host port provided: %q", parts[2])
			}
			result.HostPort = hostPort
		case 2:
			containerPort, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse port mapping, detected format port:port, invalid container port provided: %q", parts[0])
			}
			result.ContainerPort = containerPort

			hostPort, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse port mapping, detected format port:port, invalid host port provided: %q", parts[1])
			}
			result.HostPort = hostPort
		case 1:
			p, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse port mapping, detected format port:port, invalid container port provided: %q", parts[0])
			}
			result.ContainerPort = p
		default:
			return nil, fmt.Errorf("failed to parse port mapping, invalid port mapping provided, expected format port, port:port, or listenAddress:port:port. Actual mapping provided: %v", parts)
		}

		mappings = append(mappings, result)
	}

	return mappings, nil
}

// ParseInstallPackageMappings creates a slice of InstallPackages
// that maps a package name, version, and config file path
// based on user provided mapping.
//
// installPackageMaps is a slice of InstallPackage mappings.
// Since users can provide multiple install-package flags,
// and cobra supports this workflow, we need to be able to parse multiple strings
// that are each individually a set of installPackage maps
//
// A string in installPackageMaps is expected to be of the following format
// where each mapping is delimitted by a `,`:
//
//     mapping-0,mapping-1, ... ,mapping-N
//
// Each mapping is expected to be in the following format:
// where each field is delimited by a `:`.
// If more than 3 `:` are found, an error is returned:
//
//     package-name:package-version:package-config:package-namespace
//
// Both version and config are optional.
// It is possible to only provide a package name
// This function will create a InstallPackage that has an empty version and config with the name given in the installPackageMaps string.
//
// See tests for further examples.
func ParseInstallPackageMappings(installPackageMaps []string) ([]InstallPackage, error) {
	result := []InstallPackage{}

	for _, installPackageMap := range installPackageMaps {
		// Users can provide mappings delimited by `,`
		installPackages := strings.Split(installPackageMap, ",")
		for _, installPackage := range installPackages {
			ip := InstallPackage{}

			// Users can provide package name, version, and config file path delimited by `:`
			parts := strings.Split(installPackage, ":")

			switch len(parts) {
			case 0:
				return nil, fmt.Errorf("could not parse install-package mapping %s - no parts found after splitting on `:` ", installPackage)
			case 1:
				// Assume only a package name was provided: "my-package.example.com"
				ip.Name = parts[0]

			case 2:
				// Assume a package name and version were provided: "my-package.example.com:1.2.3"
				ip.Name = parts[0]
				ip.Version = parts[1]

			case 3:
				// Assume a full package name, version, and config were provided: "my-package.example.com:1.2.3:values.yaml"
				ip.Name = parts[0]
				ip.Version = parts[1]
				ip.Config = parts[2]

			case 4:
				// Assume a full package name, version, and config were provided: "my-package.example.com:1.2.3:values.yaml"
				ip.Name = parts[0]
				ip.Version = parts[1]
				ip.Config = parts[2]
				ip.Namespace = parts[3]

			default:
				return nil, fmt.Errorf("could not parse install-package mapping %s - should have max 2 `:` delimiting `package-name:version:config-path`", installPackage)
			}

			result = append(result, ip)
		}
	}

	return result, nil
}
