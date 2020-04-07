package cmd

import (
	"fmt"
	"time"

	"github.com/mvisonneau/go-helpers/logger"
	"github.com/nlopes/slack"
	"github.com/xanzy/go-gitlab"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

type client struct {
	gitlab      *gitlab.Client
	gitlabAdmin *gitlab.Client
	slack       *slack.Client
}

// EmailMappings ...
type EmailMappings map[string]*Mapping

// Mapping ...
type Mapping struct {
	GitlabUserID int
	SlackUserID  string
}

var start time.Time
var c *client

func configure(ctx *cli.Context) error {
	start = ctx.App.Metadata["startTime"].(time.Time)

	lc := &logger.Config{
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return err
	}

	requiredFlags := []string{
		"gitlab-url",
		"gitlab-token",
	}

	if err := mandatoryStringOptions(ctx, requiredFlags); err != nil {
		return err
	}

	c = &client{}
	var err error
	opts := []gitlab.ClientOptionFunc{
		gitlab.WithBaseURL(ctx.GlobalString("gitlab-url")),
	}

	c.gitlab, err = gitlab.NewClient(ctx.GlobalString("gitlab-token"), opts...)
	if err != nil {
		return err
	}

	c = &client{}

	if ctx.String("gitlab-admin-token") != "" {
		c.gitlabAdmin, err = gitlab.NewClient(ctx.GlobalString("gitlab-admin-token"), opts...)
		if err != nil {
			return err
		}
	} else {
		c.gitlabAdmin = c.gitlab
	}

	if ctx.String("slack-token") != "" {
		c.slack = slack.New(ctx.String("slack-token"))
	}

	return nil
}

func mandatoryStringOptions(ctx *cli.Context, opts []string) (err error) {
	for _, o := range opts {
		if ctx.GlobalString(o) == "" && ctx.String(o) == "" {
			return fmt.Errorf("%s is required", o)
		}
	}
	return nil
}

func exit(err error, exitCode int) *cli.ExitError {
	defer log.Debugf("Executed in %s, exiting..", time.Since(start))
	if err != nil {
		log.Error(err.Error())
		return cli.NewExitError("", exitCode)
	}

	return cli.NewExitError("", 0)
}
