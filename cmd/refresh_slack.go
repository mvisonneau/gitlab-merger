package cmd

import (
	"github.com/slack-go/slack"
	"github.com/urfave/cli"
)

// RefreshSlackUsers list
func RefreshSlackUsers(ctx *cli.Context) error {
	if err := configure(ctx); err != nil {
		return cli.NewExitError(err, 1)
	}

	requiredFlags := []string{
		"slack-token",
	}

	if err := mandatoryStringOptions(ctx, requiredFlags); err != nil {
		return exit(err, 1)
	}

	users, err := c.listSlackUsers()
	if err != nil {
		return exit(err, 1)
	}

	em, err := c.getEmailMappings()
	if err != nil {
		return exit(err, 1)
	}

	for _, user := range users {
		if user.Profile.Email != "" {
			em.setSlackUserID(user.Profile.Email, user.ID)
		}
	}

	em.setSlackUserForAllGitlabUserEmails()

	err = c.saveEmailMappings(em)
	if err != nil {
		return exit(err, 1)
	}

	return exit(nil, 0)
}

func (em *EmailMappings) setSlackUserID(email string, SlackUserID string) {
	if _, ok := (*em)[email]; ok {
		(*em)[email].SlackUserID = SlackUserID
	}
}

func (c *client) listSlackUsers() (users []slack.User, err error) {
	users, err = c.slack.GetUsers()
	return
}

func (em *EmailMappings) setSlackUserForAllGitlabUserEmails() {
	users := map[int]string{}
	for _, mapping := range *em {
		if mapping.SlackUserID != "" {
			users[mapping.GitlabUserID] = mapping.SlackUserID
		}
	}

	for mapping := range *em {
		(*em)[mapping].SlackUserID = users[(*em)[mapping].GitlabUserID]
	}
}
