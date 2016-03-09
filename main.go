package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/ewoutp/go-gitlab-client"
	"github.com/gogits/go-gogs-client"
)

var (
	gitlabHost     string
	gitlabApiPath  string
	gitlabUser     string
	gitlabPassword string
	gitlabToken    string
	gogsUrl        string
	gogsToken      string
	gogsUser       string
	lcNames        bool
	gitlabOrg      string
	gitlabRepo     string
)

func init() {
	flag.StringVar(&gitlabHost, "gitlab-host", "", "")
	flag.StringVar(&gitlabApiPath, "gitlab-api-path", "", "")
	flag.StringVar(&gitlabUser, "gitlab-user", "", "")
	flag.StringVar(&gitlabPassword, "gitlab-password", "", "")
	flag.StringVar(&gitlabToken, "gitlab-token", "", "")
	flag.StringVar(&gogsUrl, "gogs-url", "", "")
	flag.StringVar(&gogsToken, "gogs-token", "", "")
	flag.StringVar(&gogsUser, "gogs-user", "", "")
	flag.BoolVar(&lcNames, "lc-names", false, "")
	flag.StringVar(&gitlabOrg, "gitlab-org", "", "")
	flag.StringVar(&gitlabRepo, "gitlab-repo", "", "")
}

func main() {
	flag.Parse()

	gc := gogs.NewClient(gogsUrl, gogsToken)
	orgMap := make(map[string]*gogs.Organization)

	getOrg := func(o *gogitlab.Namespace) *gogs.Organization {
		name := fixName(o.Name)
		org, ok := orgMap[name]
		if ok {
			return org
		}
		org, err := gc.GetOrg(name)
		if err == nil {
			orgMap[name] = org
			return org
		}
		createOpt := gogs.CreateOrgOption{
			UserName: name,
			FullName: o.Name,
			Description: o.Description,
		}
		fmt.Printf("Creating organization '%s' as '%s'...\n", o.Name, name)
		org, err = gc.AdminCreateOrg(gogsUser, createOpt)
		if err != nil {
			exitf("Failed to create organization '%s': %v\n", name, err)
		}
		orgMap[name] = org
		return org
	}

	migrate := func(p *gogitlab.Project) {
		name := fixName(p.Name)
		ns := fixName(p.Namespace.Name)
		_, err := gc.GetRepo(ns, name)
		if err == nil {
			fmt.Printf("Repository '%s/%s' already exists.\n", ns, name)
		} else {
			org := getOrg(p.Namespace)
			fmt.Printf("Migrating '%s/%s' as '%s/%s'...\n", p.Namespace.Name, p.Name, ns, name)
			opts := gogs.MigrateRepoOption{
				CloneAddr:    p.HttpRepoUrl,
				AuthUsername: gitlabUser,
				AuthPassword: gitlabPassword,
				UID:          int(org.ID),
				RepoName:     name,
				Private:      !p.Public,
				Description:  p.Description,
			}
			_, err := gc.MigrateRepo(opts)
			if err != nil {
				exitf("Failed to migrate '%s/%s': %v\n", p.Namespace.Name, p.Name, err)
			}
		}
	}

	gitlab := gogitlab.NewGitlab(gitlabHost, gitlabApiPath, gitlabToken)
	projects, err := gitlab.AllProjects()
	if err != nil {
		exitf("Cannot get gitlab projects: %v\n", err)
	}
	if gitlabOrg != "" {
		gitlabOrg = strings.ToLower(gitlabOrg)
	}
	if gitlabRepo != "" {
		gitlabRepo = strings.ToLower(gitlabRepo)
	}
	for _, p := range projects {
		if p.Archived {
			continue
		}
		if gitlabOrg != "" {
			if gitlabOrg == strings.ToLower(p.Namespace.Name) {
				if gitlabRepo != "" && gitlabRepo != strings.ToLower(p.Name) {
					continue
				}
			} else {
				continue
			}
		}
		migrate(p)
	}
}

func fixName(name string) string {
	switch name {
	case "api": // reserved
		return "theapi"
	default:
		if lcNames {
			return strings.ToLower(name)
		} else {
			return name
		}
	}
}

func exitf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
