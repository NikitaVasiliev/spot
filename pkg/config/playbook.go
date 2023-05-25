package config

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/umputun/spot/pkg/config/deepcopy"
)

//go:generate moq -out mocks/secrets.go -pkg mocks -skip-ensure -fmt goimports . secretsProvider:SecretProvider

// PlayBook defines the top-level config object
type PlayBook struct {
	User      string            `yaml:"user" toml:"user"`           // ssh user
	SSHKey    string            `yaml:"ssh_key" toml:"ssh_key"`     // ssh key
	Inventory string            `yaml:"inventory" toml:"inventory"` // inventory file or url
	Targets   map[string]Target `yaml:"targets" toml:"targets"`     // list of targets/environments
	Tasks     []Task            `yaml:"tasks" toml:"tasks"`         // list of tasks

	inventory       *InventoryData    // loaded inventory
	overrides       *Overrides        // overrides passed from cli
	secrets         map[string]string // list of all discovered secrets
	secretsProvider SecretsProvider   // secrets provider to use
}

// SecretsProvider defines interface for secrets providers
type SecretsProvider interface {
	Get(key string) (string, error)
}

// SimplePlayBook defines simplified top-level config
// It is used for unmarshalling only, and result used to make the usual PlayBook
type SimplePlayBook struct {
	User      string   `yaml:"user" toml:"user"`           // ssh user
	SSHKey    string   `yaml:"ssh_key" toml:"ssh_key"`     // ssh key
	Inventory string   `yaml:"inventory" toml:"inventory"` // inventory file or url
	Targets   []string `yaml:"targets" toml:"targets"`     // list of names
	Target    string   `yaml:"target" toml:"target"`       // a single target to run task on
	Task      []Cmd    `yaml:"task" toml:"task"`           // single task is a list of commands
}

// Task defines multiple commands runs together
type Task struct {
	Name     string   `yaml:"name" toml:"name"` // name of task, mandatory
	User     string   `yaml:"user" toml:"user"`
	Commands []Cmd    `yaml:"commands" toml:"commands"`
	OnError  string   `yaml:"on_error" toml:"on_error"`
	Targets  []string `yaml:"targets" toml:"targets"` // optional list of targets to run task on, names or groups
}

// Target defines hosts to run commands on
type Target struct {
	Name   string        `yaml:"-" toml:"-"`           // name of target, set from the map key
	Hosts  []Destination `yaml:"hosts" toml:"hosts"`   // direct list of hosts to run commands on, no need to use inventory
	Groups []string      `yaml:"groups" toml:"groups"` // list of groups to run commands on, matches to inventory
	Names  []string      `yaml:"names" toml:"names"`   // list of host names to run commands on, matches to inventory
	Tags   []string      `yaml:"tags" toml:"tags"`     // list of tags to run commands on, matches to inventory
}

// Destination defines destination info
type Destination struct {
	Name string   `yaml:"name" toml:"name"`
	Host string   `yaml:"host" toml:"host"`
	Port int      `yaml:"port" toml:"port"`
	User string   `yaml:"user" toml:"user"`
	Tags []string `yaml:"tags" toml:"tags"`
}

// Overrides defines override for task passed from cli
type Overrides struct {
	User         string
	Inventory    string
	Environment  map[string]string
	AdHocCommand string
}

// InventoryData defines inventory data format
type InventoryData struct {
	Groups map[string][]Destination `yaml:"groups" toml:"groups"`
	Hosts  []Destination            `yaml:"hosts" toml:"hosts"`
}

const (
	allHostsGrp  = "all"
	inventoryEnv = "SPOT_INVENTORY"
)

// New creates a new PlayBook instance by loading the playbook configuration from the specified file. If the file cannot be
// found, and an ad-hoc command is specified in the overrides, a fake playbook with the ad-hoc command is created.
// The method also loads any secrets from the specified secrets provider and the inventory data from the specified
// location (if set). Returns an error if the playbook configuration cannot be loaded or parsed,
// or if the inventory data cannot be loaded.
func New(fname string, overrides *Overrides, secProvider SecretsProvider) (res *PlayBook, err error) {
	log.Printf("[DEBUG] request to load playbook %q", fname)
	res = &PlayBook{
		overrides:       overrides,
		secretsProvider: secProvider,
		inventory:       &InventoryData{Groups: make(map[string][]Destination)},
	}

	// load playbook
	data, err := os.ReadFile(fname) // nolint
	if err != nil {
		log.Printf("[DEBUG] no playbook file %s found", fname)
		if overrides != nil && overrides.AdHocCommand != "" {
			// no config file but adhoc set, just return empty config with overrides
			inventoryLoc := os.Getenv(inventoryEnv) // default inventory location from env
			if overrides.Inventory != "" {
				inventoryLoc = overrides.Inventory // inventory set in cli overrides
			}
			if inventoryLoc != "" { // load inventory if set in cli or env
				res.inventory, err = res.loadInventory(inventoryLoc)
				if err != nil {
					return nil, fmt.Errorf("can't load inventory %s: %w", overrides.Inventory, err)
				}
				log.Printf("[INFO] inventory loaded from %s with %d hosts", inventoryLoc, len(res.inventory.Groups[allHostsGrp]))
			} else {
				log.Printf("[INFO] no inventory loaded")
			}
			return res, nil
		}
		return nil, err
	}

	if err = unmarshalPlaybookFile(fname, data, overrides, res); err != nil {
		return nil, fmt.Errorf("can't unmarshal config: %w", err)
	}

	if err = res.checkConfig(); err != nil {
		return nil, fmt.Errorf("config %s is invalid: %w", fname, err)
	}

	// load secrets from secrets provider
	if secErr := res.loadSecrets(); secErr != nil {
		return nil, secErr
	}

	// log loaded config info
	log.Printf("[INFO] playbook loaded with %d tasks", len(res.Tasks))
	for _, tsk := range res.Tasks {
		for _, c := range tsk.Commands {
			log.Printf("[DEBUG] load command %q (task: %s)", c.Name, tsk.Name)
		}
	}

	// load inventory if set
	inventoryLoc := os.Getenv(inventoryEnv) // default inventory location from env
	if res.Inventory != "" {
		inventoryLoc = res.Inventory // inventory set in playbook
	}
	if overrides != nil && overrides.Inventory != "" {
		inventoryLoc = overrides.Inventory // inventory set in cli overrides
	}
	if inventoryLoc != "" { // load inventory if set. if not set, assume direct hosts in targets are used
		log.Printf("[DEBUG] inventory location %q", inventoryLoc)
		res.inventory, err = res.loadInventory(inventoryLoc)
		if err != nil {
			return nil, fmt.Errorf("can't load inventory %s: %w", inventoryLoc, err)
		}
	}
	if len(res.inventory.Groups) > 0 { // even with hosts only it will make a group "all"
		log.Printf("[INFO] inventory loaded with %d hosts", len(res.inventory.Groups[allHostsGrp]))
	} else {
		log.Printf("[INFO] no inventory loaded")
	}

	// populate target names from map keys to be able to use them from caller getting back just a target
	for k, v := range res.Targets {
		v.Name = k
		res.Targets[k] = v
	}

	return res, nil
}

// unmarshalPlaybookFile is trying to parse playbook from the data bytes.
// It will try to guess format by file extension or use yaml as toml.
// First it will try to unmarshal to a complete PlayBook struct, if it fails,
// it will try to unmarshal to a SimplePlayBook struct and convert it to a complete PlayBook struct.
func unmarshalPlaybookFile(fname string, data []byte, overrides *Overrides, res *PlayBook) (err error) {

	unmarshal := func(data []byte, v interface{}) error {
		// try to unmarshal yml first and then toml
		switch {
		case strings.HasSuffix(fname, ".yml") || strings.HasSuffix(fname, ".yaml") || !strings.Contains(fname, "."):
			if err = yaml.Unmarshal(data, v); err != nil {
				return fmt.Errorf("can't unmarshal config %s: %w", fname, err)
			}
		case strings.HasSuffix(fname, ".toml"):
			if err = toml.Unmarshal(data, v); err != nil {
				return fmt.Errorf("can't unmarshal config %s: %w", fname, err)
			}
		default:
			return fmt.Errorf("unknown config format %s", fname)
		}
		return nil
	}

	splitIPAddress := func(inp string) (string, int) {
		host, portStr, e := net.SplitHostPort(inp)
		if e != nil {
			return inp, 22
		}
		port, e := strconv.Atoi(portStr)
		if e != nil {
			return host, 22
		}
		return host, port
	}

	errs := new(multierror.Error)
	if err = unmarshal(data, res); err == nil && len(res.Tasks) > 0 {
		return nil // success, this is full PlayBook config
	}
	errs = multierror.Append(errs, err)

	simple := &SimplePlayBook{}
	if err = unmarshal(data, simple); err == nil && len(simple.Task) > 0 {
		// success, this is SimplePlayBook config, convert it to full PlayBook config
		res.Inventory = simple.Inventory
		res.Tasks = []Task{{Commands: simple.Task}} // simple playbook has just a list of commands as the task
		res.Tasks[0].Name = "default"               // we have only one task, set it as default

		hasInventory := simple.Inventory != "" || (overrides != nil && overrides.Inventory != "") || os.Getenv(inventoryEnv) != ""

		target := Target{}
		targets := append([]string{}, simple.Targets...)
		if simple.Target != "" {
			targets = append(targets, simple.Target) // append target from simple playbook
		}

		for _, t := range targets {
			if strings.Contains(t, ":") {
				ip, port := splitIPAddress(t)
				target.Hosts = append(target.Hosts, Destination{Host: ip, Port: port}) // set as hosts in case of ip:port
				log.Printf("[DEBUG] set target host %s:%d", ip, port)
			}

			if hasInventory && !strings.Contains(t, ":") {
				target.Names = append(target.Names, t) // set as names in case of just name and inventory is set
				log.Printf("[DEBUG] set target name %s", t)
			}

			if !hasInventory && !strings.Contains(t, ":") { // set as host with :22 in case of just name and no inventory
				target.Hosts = append(target.Hosts, Destination{Host: t, Port: 22}) // set as hosts in case of ip:port
				log.Printf("[DEBUG] set target host %s:22", t)
			}
		}
		res.Targets = map[string]Target{"default": target}
		return nil
	}

	return multierror.Append(errs, err).Unwrap()
}

// AllTasks returns the playbook's list of tasks.
// This method performs a deep copy of the tasks to avoid side effects of overrides on the original config.
func (p *PlayBook) AllTasks() []Task {
	cp := deepcopy.Copy(p.Tasks)
	res, ok := cp.([]Task)
	if !ok {
		// this should never happen
		return p.Tasks
	}

	return res
}

// Task returns the task with the specified name from the playbook's list of tasks. If the name is "ad-hoc" and an ad-hoc
// command is specified in the playbook's overrides, a fake task with a single command is created.
// The method performs a deep copy of the task to avoid side effects of overrides on the original config and also applies
// any overrides for the user and environment variables to the task and its commands.
// Returns an error if the task cannot be found or copied.
func (p *PlayBook) Task(name string) (*Task, error) {
	searchTask := func(tsk []Task, name string) (*Task, error) {
		if name == "ad-hoc" && p.overrides.AdHocCommand != "" {
			// special case for ad-hoc command, make a fake task with a single command from overrides.AdHocCommand
			return &Task{Name: "ad-hoc", Commands: []Cmd{{Name: "ad-hoc", Script: p.overrides.AdHocCommand}}}, nil
		}
		for _, t := range tsk {
			if strings.EqualFold(t.Name, name) {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("task %q not found", name)
	}

	t, err := searchTask(p.Tasks, name)
	if err != nil {
		return nil, err
	}

	cp := deepcopy.Copy(t) // deep copy to avoid side effects of overrides on original config
	res, ok := cp.(*Task)
	if !ok {
		return nil, fmt.Errorf("can't copy task %s", name)
	}

	res.Name = name
	if res.User == "" {
		res.User = p.User // if user not set in task, use default from playbook
	}

	// apply overrides of user
	if p.overrides != nil && p.overrides.User != "" {
		res.User = p.overrides.User
	}

	// apply overrides of environment variables, to each command
	if p.overrides != nil && p.overrides.Environment != nil {
		for envKey, envVal := range p.overrides.Environment {
			for cmdIdx := range res.Commands {
				if res.Commands[cmdIdx].Environment == nil {
					res.Commands[cmdIdx].Environment = make(map[string]string)
				}
				res.Commands[cmdIdx].Environment[envKey] = envVal
			}
		}
	}

	return res, nil
}

// TargetHosts returns target hosts for given target name.
func (p *PlayBook) TargetHosts(name string) ([]Destination, error) {

	userOverride := func(u string) string {
		// apply overrides of user
		if p.overrides != nil && p.overrides.User != "" {
			return p.overrides.User
		}
		// no overrides, use user from target if set
		if u != "" {
			return u
		}
		// no overrides, no user in target, use default from playbook
		return p.User
	}

	tgExtractor := newTargetExtractor(p.Targets, p.User, p.inventory)
	res, err := tgExtractor.Destinations(name)
	if err != nil {
		return nil, err
	}

	for i, h := range res {
		if h.Port == 0 {
			h.Port = 22 // the default port is 22 if not set
		}
		h.User = userOverride(h.User)
		res[i] = h
	}

	return res, nil
}

// AllSecretValues returns all secret values from all tasks and all commands.
// It is used to mask Secrets in logs.
func (p *PlayBook) AllSecretValues() []string {
	res := make([]string, 0, len(p.secrets))
	for _, v := range p.secrets {
		res = append(res, v)
	}
	sort.Strings(res)
	return res
}

// UpdateTasksTargets updates the targets of all tasks in the playbook with the values from the specified map of variables.
// The method is used to replace variables in the targets of tasks with their actual values and this way provide dynamic targets.
func (p *PlayBook) UpdateTasksTargets(vars map[string]string) {
	for i, task := range p.Tasks {
		targets := []string{}
		for _, tg := range task.Targets {
			if len(tg) > 1 && strings.HasPrefix(tg, "$") {
				if vars == nil {
					continue
				}
				if v, ok := vars[tg[1:]]; ok {
					log.Printf("[DEBUG] set target %s to %q", tg, v)
					targets = append(targets, v)
				}
				continue
			}
			targets = append(targets, tg)
		}
		p.Tasks[i].Targets = targets
	}
}

// loadInventory loads the inventory data from the specified location (file or URL) and returns it as an InventoryData struct.
// The inventory data is parsed as either YAML or TOML, depending on the file extension.
// The method also performs some additional processing on the inventory data:
// - It creates a group "all" that contains all hosts from all groups.
// - It sorts the hosts in the "all" group by host name for predictable order in tests and processing.
// - It sets default port and user values for all inventory groups if not already set.
// Returns an error if the inventory data cannot be loaded or parsed, or if the "all" group is reserved for all hosts.
func (p *PlayBook) loadInventory(loc string) (*InventoryData, error) {

	reader := func(loc string) (r io.ReadCloser, err error) {
		// get reader for inventory file or url
		switch {
		case strings.HasPrefix(loc, "http"): // location is a url
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(loc)
			if err != nil {
				return nil, fmt.Errorf("can't get inventory from http %s: %w", loc, err)
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("can't get inventory from http %s, status: %s", loc, resp.Status)
			}
			return resp.Body, nil
		default: // location is a file
			f, err := os.Open(loc) // nolint
			if err != nil {
				return nil, fmt.Errorf("can't open inventory file %s: %w", loc, err)
			}
			return f, nil
		}
	}

	rdr, err := reader(loc) // inventory ReadCloser, has to be closed
	if err != nil {
		return nil, err
	}
	defer rdr.Close() // nolint

	var data InventoryData
	if !strings.HasSuffix(loc, ".toml") {
		// we assume it is yaml. Can't do strict check, as we can have urls without any extension
		if err = yaml.NewDecoder(rdr).Decode(&data); err != nil {
			return nil, fmt.Errorf("can't parse inventory %s: %w", loc, err)
		}
	} else {
		if err = toml.NewDecoder(rdr).Decode(&data); err != nil {
			return nil, fmt.Errorf("can't parse inventory %s: %w", loc, err)
		}
	}

	if len(data.Groups[allHostsGrp]) > 0 {
		return nil, fmt.Errorf("group %q is reserved for all hosts", allHostsGrp)
	}

	if len(data.Groups) > 0 {
		// create group "all" with all hosts from all groups
		data.Groups[allHostsGrp] = []Destination{}
		for key, g := range data.Groups {
			if key == "all" {
				continue
			}
			data.Groups[allHostsGrp] = append(data.Groups[allHostsGrp], g...)
		}
	}
	if len(data.Hosts) > 0 {
		// add hosts to group "all"
		if data.Groups == nil {
			data.Groups = make(map[string][]Destination)
		}
		if _, ok := data.Groups[allHostsGrp]; !ok {
			data.Groups[allHostsGrp] = []Destination{}
		}
		data.Groups[allHostsGrp] = append(data.Groups[allHostsGrp], data.Hosts...)
	}
	// sort hosts in group "all" by host name, for predictable order in tests and in the processing
	sort.Slice(data.Groups[allHostsGrp], func(i, j int) bool {
		return data.Groups[allHostsGrp][i].Host < data.Groups[allHostsGrp][j].Host
	})

	// set default port and user if not set for all inventory groups
	for _, gr := range data.Groups {
		for i := range gr {
			if gr[i].Port == 0 {
				gr[i].Port = 22 // the default port is 22 if not set
			}
			if gr[i].User == "" {
				gr[i].User = p.User // default user is playbook's user or override, if not set by inventory
			}
		}
	}

	return &data, nil
}

// checkConfig validates the PlayBook configuration by ensuring that:
// - all tasks have unique names and no empty names
// - all commands have a single type set
// - the target set is not called "all"
// Returns an error if any of these conditions are not met.
func (p *PlayBook) checkConfig() error {

	// check that all tasks have unique names in the playbook and no empty names
	names := make(map[string]bool)
	for _, t := range p.Tasks {
		if t.Name == "" { // task name is required
			return fmt.Errorf("task name is required")
		}
		if names[t.Name] { // task name must be unique
			return fmt.Errorf("duplicate task name %q", t.Name)
		}
		names[t.Name] = true
	}

	// check what all commands have a single type set
	for _, t := range p.Tasks {
		if len(t.Commands) == 0 {
			return fmt.Errorf("task %q has no commands", t.Name)
		}
		for _, c := range t.Commands {
			if err := c.validate(); err != nil {
				return fmt.Errorf("task %q rejected, invalid command %q: %w", t.Name, c.Name, err)
			}
		}
	}

	// check what target set is not called "all"
	for k := range p.Targets {
		if strings.EqualFold(k, allHostsGrp) {
			return fmt.Errorf("target %q is reserved for all hosts", allHostsGrp)
		}
	}

	return nil
}

// loadSecrets loads secrets from secrets provider and stores them in secrets map
func (p *PlayBook) loadSecrets() error {
	// check if secrets are defined in playbook
	secretsCount := 0
	for _, t := range p.Tasks {
		for _, c := range t.Commands {
			if c.Options.NoAuto {
				continue // skip commands with noauto flag
			}
			secretsCount += len(c.Options.Secrets)
		}
	}

	if p.secretsProvider == nil && secretsCount == 0 {
		return nil
	}
	if p.secretsProvider == nil && secretsCount > 0 {
		return fmt.Errorf("secrets are defined in playbook (%d secrets), but provider is not set", secretsCount)
	}

	if p.secrets == nil {
		p.secrets = make(map[string]string)
	}

	// collect Secrets from all command's, retrieve them from provider and store in the secrets map
	for _, t := range p.Tasks {
		for i, c := range t.Commands {
			for _, key := range c.Options.Secrets {
				val, err := p.secretsProvider.Get(key)
				if err != nil {
					return fmt.Errorf("can't get secret %q defined in task %q, command %q: %w", key, t.Name, c.Name, err)
				}
				p.secrets[key] = val // store secret in the secrets map of playbook
				if c.Secrets == nil {
					c.Secrets = make(map[string]string)
				}
				c.Secrets[key] = val // store secret in the secrets map of command
			}
			t.Commands[i] = c
		}
	}
	return nil
}
