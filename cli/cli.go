package cli

import (
	"os"
	"time"

	"github.com/mvisonneau/gitlab-merger/cmd"
	"github.com/urfave/cli"
)

// Run handles the instanciation of the CLI application
func Run(version string) {
	NewApp(version, time.Now()).Run(os.Args)
}

// NewApp configures the CLI application
func NewApp(version string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "gitlab-merger"
	app.Version = version
	app.Usage = "Automate your MR creation"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "GLM_LOG_LEVEL",
			Usage:  "log `level` (debug,info,warn,fatal,panic)",
			Value:  "info",
		},
		cli.StringFlag{
			Name:   "log-format",
			EnvVar: "GLM_LOG_FORMAT",
			Usage:  "log `format` (json,text)",
			Value:  "text",
		},
		cli.StringFlag{
			Name:   "gitlab-url",
			EnvVar: "GLM_GITLAB_URL",
			Usage:  "`url`",
		},
		cli.StringFlag{
			Name:   "gitlab-token",
			EnvVar: "GLM_GITLAB_TOKEN",
			Usage:  "`token`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "merge",
			Usage:  "refs together",
			Action: cmd.Merge,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "mr-prefix",
					EnvVar: "GLM_MR_PREFIX",
					Usage:  "`prefix`",
					Value:  "[GL-MERGER] -",
				},
				cli.StringFlag{
					Name:   "source-ref",
					EnvVar: "GLM_SOURCE_REF",
					Usage:  "`ref`",
				},
				cli.StringFlag{
					Name:   "target-ref",
					EnvVar: "GLM_TARGET_REF",
					Usage:  "`ref`",
				},
				cli.StringFlag{
					Name:   "project",
					EnvVar: "GLM_PROJECT",
					Usage:  "`project`",
				},
				cli.StringFlag{
					Name:   "slack-token",
					EnvVar: "GLM_SLACK_TOKEN",
					Usage:  "`token`",
				},
				cli.StringFlag{
					Name:   "slack-channel",
					EnvVar: "GLM_SLACK_CHANNEL",
					Usage:  "`channel`",
				},
			},
		},
		{
			Name:  "refresh",
			Usage: "users list",
			Subcommands: []cli.Command{
				{
					Name:   "gitlab-users",
					Action: cmd.RefreshGitlabUsers,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "gitlab-admin-token",
							EnvVar: "GLM_GITLAB_ADMIN_TOKEN",
							Usage:  "`token`",
						},
					},
				},
				{
					Name:   "slack-users",
					Action: cmd.RefreshSlackUsers,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "slack-token",
							EnvVar: "GLM_SLACK_TOKEN",
							Usage:  "`token`",
						},
					},
				},
			},
		},
	}

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
