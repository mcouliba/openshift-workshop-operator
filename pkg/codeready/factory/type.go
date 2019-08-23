package codeready

type Factory struct {
	V         string    `json:"v"`
	Name      string    `json:"name"`
	Workspace Workspace `json:"workspace"`
	Creator   Creator   `json:"creator"`
}

type Workspace struct {
	Environments Environments `json:"environments"`
	Commands     []Command    `json:"commands"`
	Projects     []Project    `json:"projects"`
	DefaultEnv   string       `json:"defaultEnv"`
	Name         string       `json:"name"`
	Links        []string     `json:"links"`
	ID           string       `json:"id"`
}

type Project struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source Source `json:"source"`
	Path   string `json:"path"`
}

type Source struct {
	Location   string            `json:"location"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters"`
	Path       string            `json:"path"`
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

type Creator struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
