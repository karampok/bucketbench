package driver

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/estesp/bucketbench/utils"
)

type GardenDriver struct {
	gaolPath string
}

func NewGardenDriver(gaolPath string) (Driver, error) {
	return &GardenDriver{gaolPath: gaolPath}, nil
}

func (g *GardenDriver) Type() Type {
	return Garden
}

func (g *GardenDriver) Info() (string, error) {
	return "Info for Garden isn't implemented yet", nil
}

func (g *GardenDriver) runGaol(gaolArgs ...string) (string, error) {
	out, err := exec.Command(g.gaolPath, gaolArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running %s '%s': %s\ngaol CLI output: %s", g.gaolPath, strings.Join(gaolArgs, " "), err, string(out))
	}
	return string(out), nil
}

func (g *GardenDriver) Create(name, image string, detached bool, trace bool) (Container, error) {
	if _, err := g.runGaol("create", "-n", name); err != nil {
		return nil, err
	}
	return &gardenContainer{name: name, detached: detached}, nil
}

func (g *GardenDriver) Clean() error {
	// gaol list | xargs
	containers, err := g.runGaol("list")
	if err != nil {
		return err
	}
	for _, container := range strings.Split(containers, "\n") {
		if container == "" {
			continue
		}
		if _, err := g.runGaol("destroy", container); err != nil {
			return err
		}
	}

	return nil
}

func (g *GardenDriver) Run(ctr Container) (string, int, error) {
	gaolArgs := "run " + ctr.Name()
	if !ctr.Detached() {
		gaolArgs = gaolArgs + " -a"
	}
	gaolArgs = gaolArgs + " -c whoami"
	return utils.ExecTimedCmd(g.gaolPath, gaolArgs)
}

func (g *GardenDriver) Stop(ctr Container) (string, int, error) {
	return "", 0, nil
}

func (g *GardenDriver) Remove(ctr Container) (string, int, error) {
	return utils.ExecTimedCmd(g.gaolPath, "destroy "+ctr.Name())
}

func (g *GardenDriver) Pause(ctr Container) (string, int, error) {
	return "", 0, nil
}

func (g *GardenDriver) Unpause(ctr Container) (string, int, error) {
	return "", 0, nil
}

type gardenContainer struct {
	name     string
	detached bool
}

func (c *gardenContainer) Name() string {
	return c.name
}

func (c *gardenContainer) Detached() bool {
	return c.detached
}

func (c *gardenContainer) Trace() bool {
	return false
}

func (c *gardenContainer) Image() string {
	return ""
}
