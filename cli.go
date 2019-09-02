package main

import (
	"github.com/urfave/cli"
)

// runCli : Generates cli configuration for the application
func runCli() (app *cli.App) {
	app = cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"
	app.Version = version
	app.Usage = "Export metrics about GitLab CI pipeliens statuses"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "GLMRGR_LOG_LEVEL",
			Usage:  "log `level` (debug,info,warn,fatal,panic)",
			Value:  "info",
		},
		cli.StringFlag{
			Name:   "log-format",
			EnvVar: "GLMRGR_LOG_FORMAT",
			Usage:  "log `format` (json,text)",
			Value:  "text",
		},
		cli.StringFlag{
			Name:   "project",
			EnvVar: "GLMRGR_PROJECT",
			Usage:  "`project`",
		},
		cli.StringFlag{
			Name:   "source-ref",
			EnvVar: "GLMRGR_SOURCE_REF",
			Usage:  "`ref`",
		},
		cli.StringFlag{
			Name:   "target-ref",
			EnvVar: "GLMRGR_TARGET_REF",
			Usage:  "`ref`",
		},
		cli.StringFlag{
			Name:   "gitlab-url",
			EnvVar: "GLMRGR_GITLAB_URL",
			Usage:  "`url`",
		},
		cli.StringFlag{
			Name:   "gitlab-token",
			EnvVar: "GLMRGR_GITLAB_TOKEN",
			Usage:  "`token`",
		},
		cli.StringFlag{
			Name:   "mr-prefix",
			EnvVar: "GLMRGR_MR_PREFIX",
			Usage:  "`prefix`",
			Value: "[GL-MERGER] -",
		},
	}

	app.Action = run

	return
}
