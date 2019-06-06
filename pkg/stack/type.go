package stack

type Stack struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Scope           string          `json:"scope"`
	WorkspaceConfig WorkspaceConfig `json:"workspaceConfig"`
	Components      []Component     `json:"components"`
	Creator         string          `json:"creator"`
	Tags            []string        `json:"tags"`
}

type WorkspaceConfig struct {
	Environments Environments `json:"environments"`
	Commands     []Command    `json:"commands"`
	Projects     []string     `json:"projects"`
	DefaultEnv   string       `json:"defaultEnv"`
	Name         string       `json:"name"`
	Links        []string     `json:"links"`
}

type Environments struct {
	Default Environment `json:"default"`
}

type Environment struct {
	Recipe   Recipe   `json:"recipe"`
	Machines Machines `json:"machines"`
}

type Recipe struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type Machines struct {
	DevMachine DevMachine `json:"dev-machine"`
}

type DevMachine struct {
	Env        map[string]string `json:"env"`
	Servers    map[string]Server `json:"servers"`
	Volumes    map[string]string `json:"volumes"`
	Installers []string          `json:"installers"`
	Attributes map[string]string `json:"attributes"`
}

type Server struct {
	Attributes map[string]string `json:"attributes"`
	Protocol   string            `json:"protocol"`
	Port       string            `json:"port"`
}

type Command struct {
	CommandLine string            `json:"commandLine"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Attributes  map[string]string `json:"attributes"`
}

type Component struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

type StackPermission struct {
	UserID     string   `json:"userId"`
	DomainID   string   `json:"domainId"`
	InstanceID string   `json:"instanceId"`
	Actions    []string `json:"actions"`
}
