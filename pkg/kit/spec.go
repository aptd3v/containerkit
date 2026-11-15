package kit

import (
	"fmt"
	"strings"
	"time"
)

// spec is the shared container/service/job configuration surface.
type spec struct {
	name        field[string]
	image       field[string]
	command     field[[]string]
	entrypoint  field[[]string]
	user        field[string]
	workdir     field[string]
	hostname    field[string]
	domainName  field[string]
	macAddress  field[string]
	stopSignal  field[string]
	platform    field[string]
	env         []envVar
	labels      map[string]string
	ports       []port
	expose      []exposePort
	health      field[health]
	privileged  field[bool]
	memory      field[int64]
	restart     field[restart]
	tty         field[bool]
	interactive field[bool]
	readOnly    field[bool]
	init        field[bool]
	stopTimeout field[int]
	capAdd      []Capability
	capDrop     []Capability

	volumes     []volume
	volumesFrom []string
	volumeDrv   field[string]

	networkMode     field[string]
	networkDisabled field[bool]
	networks        []netAttach
	dns             []string
	dnsOptions      []string
	dnsSearch       []string
	extraHosts      []extraHost

	groupAdd       []string
	logging        field[logging]
	shmSize        field[int64]
	memReserve     field[int64]
	memSwap        field[int64]
	memSwapiness   field[int64]
	cpuShares      field[int64]
	cpuPeriod      field[int64]
	cpuQuota       field[int64]
	cpus           field[float64]
	cpusetCpus     field[string]
	cpusetMems     field[string]
	cpuCount       field[int64]
	cpuPercent     field[int64]
	cpuRTPeriod    field[int64]
	cpuRTRuntime   field[int64]
	pidsLimit      field[int64]
	oomScoreAdj    field[int]
	oomKillDisable field[bool]
	blkioWeight    field[uint16]
	ulimits        []ulimit
	sysctls        map[string]string
	securityOpt    []string
	storageOpt     map[string]string
	devices        []device
	devCgroup      []string
	gpus           []gpu

	ipc              field[string]
	pid              field[string]
	uts              field[string]
	userns           field[string]
	cgroup           field[string]
	cgroupParent     field[string]
	cgroupns         field[string]
	isolation        field[string]
	runtime          field[string]
	annotations      map[string]string
	umask            field[uint32]
	links            []string
	ioMaxIOps        field[uint64]
	ioMaxBandwidth   field[uint64]
	consoleH         field[uint]
	consoleW         field[uint]
	blkioWDev        []blkioWeight
	blkioReadBps     []blkioRate
	blkioWriteBps    []blkioRate
	blkioReadIOps    []blkioRate
	blkioWriteIOps   []blkioRate
	models           []serviceModel
	pullRefreshAfter field[string]
	tmpfs            []string
	perReplica       field[bool]
}

type blkioWeight struct {
	path   string
	weight uint16
}

type blkioRate struct {
	path string
	rate uint64
}

type envVar struct {
	key   string
	value field[string]
}

type port struct {
	protocol      field[Proto]
	hostIP        field[string]
	hostPort      field[uint16]
	containerPort field[uint16]
	portName      field[string]
	appProtocol   field[string]
	portMode      field[string]
	published     field[string]
}

type exposePort struct {
	port     uint16
	protocol Proto
}

type health struct {
	test          field[[]string]
	interval      field[time.Duration]
	timeout       field[time.Duration]
	retries       field[int]
	startPeriod   field[time.Duration]
	startInterval field[time.Duration]
	disabled      field[bool]
}

type restart struct {
	policy     field[RestartPolicy]
	maxRetries field[int]
	delay      field[time.Duration]
	window     field[time.Duration]
}

type depend struct {
	name      field[string]
	condition field[DependCondition]
	restart   field[bool]
	required  field[bool]
}

type volume struct {
	kind           field[VolumeKind]
	source         field[string]
	target         field[string]
	readOnly       field[bool]
	size           field[int64]
	mode           field[uint32]
	prop           field[string]
	noCopy         field[bool]
	subPath        field[string]
	labels         map[string]string
	driver         field[string]
	DriverValue    map[string]string
	selinux        field[string]
	createHostPath field[bool]
	recursive      field[string]
}

type netAttach struct {
	name       field[string]
	aliases    []string
	ipv4       field[string]
	ipv6       field[string]
	mac        field[string]
	linkLocal  []string
	priority   field[int]
	gwPriority field[int]
	ifName     field[string]
	useAlias   field[bool]
	driverOpts map[string]string
}

type extraHost struct {
	hostname string
	ip       string
}

type logging struct {
	driver  field[string]
	options map[string]string
}

type device struct {
	host      field[string]
	container field[string]
	perms     field[string]
}

type gpu struct {
	driver       field[string]
	count        field[int]
	ids          []string
	capabilities []string
}

type ulimit struct {
	name field[string]
	soft field[int64]
	hard field[int64]
}

type build struct {
	source        field[string]
	dockerfile    field[string]
	inline        field[string]
	target        field[string]
	args          []envVar
	labels        map[string]string
	pull          field[bool]
	noCache       field[bool]
	network       field[string]
	shmSize       field[int64]
	privileged    field[bool]
	tags          []string
	platforms     []string
	extraHosts    []extraHost
	isolation     field[string]
	entitlements  []string
	provenance    field[string]
	sbom          field[string]
	ssh           []sshKey
	cacheFrom     []string
	cacheTo       []string
	noCacheFilter []string
	extraContexts map[string]string
	secrets       []secretRef
	ulimits       []ulimit
}

type sshKey struct {
	id   string
	path string
}

type develop struct {
	watch []watchTrigger
}

type watchTrigger struct {
	path        field[string]
	action      field[WatchAction]
	target      field[string]
	include     []string
	ignore      []string
	initialSync field[bool]
	exec        serviceHook
}

type serviceHook struct {
	cmd        []string
	user       field[string]
	privileged field[bool]
	workdir    field[string]
	env        []envVar
}

type envFileRef struct {
	path     field[string]
	required field[bool]
	format   field[string]
}

type includeRef struct {
	paths      []string
	projectDir field[string]
	envFiles   []string
}

type serviceModel struct {
	name        field[string]
	endpointVar field[string]
	modelVar    field[string]
}

type updateCfg struct {
	parallelism     field[uint64]
	delay           field[time.Duration]
	failureAction   field[string]
	monitor         field[time.Duration]
	maxFailureRatio field[float32]
	order           field[string]
}

type placement struct {
	constraints []string
	spreads     []string
	maxReplicas field[uint64]
}

type resource struct {
	memory  field[int64]
	cpus    field[float64]
	pids    field[int64]
	devices []gpu
	generic []discreteRes
}

type discreteRes struct {
	kind  string
	value int64
}

type secretRef struct {
	source   field[string]
	target   field[string]
	uid      field[string]
	gid      field[string]
	mode     field[uint32]
	file     field[string]
	external field[bool]
	driver   field[string]
	labels   map[string]string
}

type fileConfigRef struct {
	source   field[string]
	target   field[string]
	uid      field[string]
	gid      field[string]
	mode     field[uint32]
	file     field[string]
	external field[bool]
	labels   map[string]string
}

type deploy struct {
	mode         field[string]
	replicas     field[int]
	memory       field[int64]
	cpus         field[float64]
	labels       map[string]string
	update       field[updateCfg]
	rollback     field[updateCfg]
	restart      field[restart]
	placement    field[placement]
	endpointMode field[string]
	reserve      field[resource]
	pids         field[int64]
	devices      []gpu
	generic      []discreteRes
}

type projectNet struct {
	name       field[string]
	driver     field[string]
	internal   field[bool]
	attachable field[bool]
	external   field[bool]
	ipv4       field[bool]
	ipv6       field[bool]
	labels     map[string]string
	driverOpts map[string]string
	subnet     field[string]
	gateway    field[string]
	ipRange    field[string]
	aux        map[string]string
	ipamDriver field[string]
	ipamOpts   map[string]string
}

type projectVol struct {
	name       field[string]
	driver     field[string]
	external   field[bool]
	labels     map[string]string
	driverOpts map[string]string
}

type projectSecret struct {
	name           field[string]
	file           field[string]
	env            field[string]
	content        field[string]
	external       field[bool]
	driver         field[string]
	labels         map[string]string
	templateDriver field[string]
}

type projectModel struct {
	name         field[string]
	model        field[string]
	contextSize  field[int]
	runtimeFlags []string
}

type projectFile struct {
	name           field[string]
	file           field[string]
	env            field[string]
	content        field[string]
	external       field[bool]
	driver         field[string]
	labels         map[string]string
	templateDriver field[string]
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return cfgErr("", "name", "Name", fmt.Errorf("%w: name", errEmpty))
	}
	if strings.ContainsAny(name, " \t\n") {
		return cfgErr("", "name", "Name", fmt.Errorf("%w: name contains whitespace", errInvalid))
	}
	return nil
}

func validateImage(image string) error {
	if strings.TrimSpace(image) == "" {
		return cfgErr("", "image", "Image", fmt.Errorf("%w: image", errEmpty))
	}
	return nil
}

func (s *spec) setName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	s.name.assign(name)
	return nil
}

func (s *spec) setImage(image string) error {
	if err := validateImage(image); err != nil {
		return err
	}
	s.image.assign(image)
	return nil
}

func (s *spec) setCommand(cmd []string) error {
	if len(cmd) == 0 {
		return cfgErr("", "command", "Command", fmt.Errorf("%w: command", errEmpty))
	}
	for _, c := range cmd {
		if c == "" {
			return cfgErr("", "command", "Command", fmt.Errorf("%w: empty command argument", errInvalid))
		}
	}
	s.command.assign(append([]string(nil), cmd...))
	return nil
}

func (s *spec) setEntrypoint(cmd []string) error {
	if len(cmd) == 0 {
		return cfgErr("", "entrypoint", "Entrypoint", fmt.Errorf("%w: entrypoint", errEmpty))
	}
	s.entrypoint.assign(append([]string(nil), cmd...))
	return nil
}

func (s *spec) setUser(user string) error {
	if user == "" {
		return cfgErr("", "user", "User", fmt.Errorf("%w: user", errEmpty))
	}
	s.user.assign(user)
	return nil
}

func (s *spec) setWorkingDir(dir string) error {
	if dir == "" {
		return cfgErr("", "working_dir", "WorkingDir", fmt.Errorf("%w: working dir", errEmpty))
	}
	s.workdir.assign(dir)
	return nil
}

func (s *spec) setHostname(host string) error {
	if host == "" {
		return cfgErr("", "hostname", "Hostname", fmt.Errorf("%w: hostname", errEmpty))
	}
	s.hostname.assign(host)
	return nil
}

func (s *spec) setDomainName(d string) error {
	if d == "" {
		return cfgErr("", "domainname", "DomainName", fmt.Errorf("%w: domain name", errEmpty))
	}
	s.domainName.assign(d)
	return nil
}

func (s *spec) setMacAddress(m string) error {
	if m == "" {
		return cfgErr("", "mac_address", "MacAddress", fmt.Errorf("%w: mac address", errEmpty))
	}
	s.macAddress.assign(m)
	return nil
}

func (s *spec) setStopSignal(sig string) error {
	if sig == "" {
		return cfgErr("", "stop_signal", "StopSignal", fmt.Errorf("%w: stop signal", errEmpty))
	}
	s.stopSignal.assign(sig)
	return nil
}

func (s *spec) setPlatform(p string) error {
	if p == "" {
		return cfgErr("", "platform", "Platform", fmt.Errorf("%w: platform", errEmpty))
	}
	s.platform.assign(p)
	return nil
}

func (s *spec) addEnv(key, value string) error {
	if key == "" {
		return secretErr("", "env", "Env", fmt.Errorf("%w: env key", errEmpty))
	}
	if strings.ContainsAny(key, "=\n") {
		return secretErr("", "env", "Env", fmt.Errorf("%w: env key", errInvalid))
	}
	s.env = append(s.env, envVar{key: key, value: field[string]{set: true, val: value}})
	return nil
}

func (s *spec) addLabel(key, value string) error {
	if key == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if s.labels == nil {
		s.labels = map[string]string{}
	}
	s.labels[key] = value
	return nil
}

func (s *spec) addPublish(p port) error {
	if !p.containerPort.ok() {
		return cfgErr("", "publish.container_port", "ContainerPort", fmt.Errorf("%w: container port", errRequired))
	}
	s.ports = append(s.ports, p)
	return nil
}

func (s *spec) addExpose(p exposePort) error {
	if p.port == 0 {
		return cfgErr("", "expose.port", "Expose", fmt.Errorf("%w: expose port", errRequired))
	}
	s.expose = append(s.expose, p)
	return nil
}

func (s *spec) setHealth(h health) error {
	if h.disabled.ok() && h.disabled.val {
		s.health.assign(h)
		return nil
	}
	if !h.test.ok() || len(h.test.val) == 0 {
		return cfgErr("", "health.test", "Test", fmt.Errorf("%w: health test", errRequired))
	}
	s.health.assign(h)
	return nil
}

func (s *spec) setPrivileged(v bool) error {
	s.privileged.assign(v)
	return nil
}

func (s *spec) setMemory(n int64) error {
	if n < 0 {
		return cfgErr("", "memory", "Memory", fmt.Errorf("%w: memory %d", errInvalid, n))
	}
	s.memory.assign(n)
	return nil
}

func (s *spec) setCPUs(n float64) error {
	if n < 0 {
		return cfgErr("", "cpus", "CPUs", fmt.Errorf("%w: cpus", errInvalid))
	}
	s.cpus.assign(n)
	return nil
}

func (s *spec) setConsoleSize(h, w uint) error {
	s.consoleH.assign(h)
	s.consoleW.assign(w)
	return nil
}

func (s *spec) setRestart(r restart) error {
	if !r.policy.ok() {
		return cfgErr("", "restart.policy", "Policy", fmt.Errorf("%w: restart policy", errRequired))
	}
	s.restart.assign(r)
	return nil
}

func (s *spec) setTty(v bool) error {
	s.tty.assign(v)
	return nil
}

func (s *spec) setInteractive(v bool) error {
	s.interactive.assign(v)
	return nil
}

func (s *spec) setReadOnly(v bool) error {
	s.readOnly.assign(v)
	return nil
}

func (s *spec) setInit(v bool) error {
	s.init.assign(v)
	return nil
}

func (s *spec) setStopTimeout(seconds int) error {
	if seconds < 0 {
		return cfgErr("", "stop_timeout", "StopTimeout", fmt.Errorf("%w: stop timeout %d", errInvalid, seconds))
	}
	s.stopTimeout.assign(seconds)
	return nil
}

func (s *spec) addCap(c Capability) error {
	if c == "" {
		return cfgErr("", "cap_add", "CapAdd", fmt.Errorf("%w: capability", errEmpty))
	}
	s.capAdd = append(s.capAdd, c)
	return nil
}

func (s *spec) dropCap(c Capability) error {
	if c == "" {
		return cfgErr("", "cap_drop", "CapDrop", fmt.Errorf("%w: capability", errEmpty))
	}
	s.capDrop = append(s.capDrop, c)
	return nil
}
