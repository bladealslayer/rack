package manifest

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"strings"
)

type Process struct {
	Name string

	Args []string

	service Service
}

func NewProcess(app string, s Service, m Manifest) Process {
	name := fmt.Sprintf("%s-%s", app, s.Name)

	args := []string{}

	args = append(args, "-i")
	args = append(args, "--rm")
	args = append(args, "--name", name)

	if s.Entrypoint != "" {
		args = append(args, "--entrypoint", s.Entrypoint)
	}

	for k, v := range s.Environment {
		if v == "" {
			args = append(args, "-e", fmt.Sprintf("%s", k))
		} else {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	for i, p := range s.Ports {
		// Assign random ports to "internal" balancers
		if !p.Public {
			p.Public = true
			p.Balancer = RandomPort()
			s.Ports[i] = p
		}
		args = append(args, "-p", p.String())
	}

	for _, n := range s.Networks {
		for _, in := range n {
			args = append(args, "--net", in.Name)
		}
	}

	for _, link := range s.Links {
		args = append(args, linkArgs(m.Services[link], fmt.Sprintf("%s-%s", app, link))...)
	}

	for _, volume := range s.Volumes {
		if !strings.Contains(volume, ":") {
			usr, err := user.Current()
			if err != nil {
				log.Fatal(err)
			}
			hostPath, err := filepath.Abs(fmt.Sprintf("%s/.convox/volumes/%s/%s/%s", usr.HomeDir, app, s.Name, volume))
			if err != nil {
				//this won't break
			}
			volume = fmt.Sprintf("%s:%s", hostPath, volume)
		}
		args = append(args, "-v", volume)
	}

	args = append(args, s.Tag(app))

	if s.Command.String != "" {
		args = append(args, s.Command.String)
	} else if len(s.Command.Array) > 0 {
		args = append(args, s.Command.Array...)
	}

	return Process{
		Name:    name,
		Args:    args,
		service: s,
	}
}

func (p *Process) Sync(local, remote string) (*Sync, error) {
	return NewSync(p.Name, local, remote)
}
