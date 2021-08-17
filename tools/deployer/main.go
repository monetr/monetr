package main

import (
	"context"
	"flag"
	"github.com/google/go-github/v38/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strings"
)

var (
	githubToken = flag.String("token", "", "GITHUB_TOKEN secret")
	repoPath    = flag.String("repo", os.Getenv("GITHUB_REPOSITORY"), "{owner}/{repository} path, defaults to $GITHUB_REPOSITORY")
	refStr      = flag.String("ref", os.Getenv("GITHUB_REF"), "ref for deployment, defaults to $GITHUB_REF")
	taskStr     = flag.String("task", "deploy", "task for new deployment, defaults to 'deploy'")
	environment = flag.String("environment", "", "environment for the deployment")
)

func main() {
	flag.Parse()

	if environment == nil || len(*environment) == 0 {
		log.Fatal("environment is required")
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: *githubToken,
		},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	pieces := strings.SplitN(*repoPath, "/", 2)
	if len(pieces) != 2 {
		panic("invalid repo path: " + *repoPath)
	}

	autoMerge := false
	_, _, err := client.Repositories.CreateDeployment(context.Background(), pieces[0], pieces[1], &github.DeploymentRequest{
		Ref:                   refStr,
		Task:                  taskStr,
		AutoMerge:             &autoMerge,
		RequiredContexts:      nil,
		Payload:               nil,
		Environment:           environment,
		Description:           nil,
		TransientEnvironment:  nil,
		ProductionEnvironment: nil,
	})
	if err != nil {
		log.Fatalf("error creating deployment: %+v", errors.Wrap(err, "failed to create deployment"))
		return
	}

}
