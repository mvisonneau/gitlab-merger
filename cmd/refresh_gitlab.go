package cmd

import (
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
)

// RefreshGitlabUsers list
func RefreshGitlabUsers(ctx *cli.Context) (int, error) {
	if err := configure(ctx); err != nil {
		return 1, err
	}

	users, err := c.listGitlabUsers()
	if err != nil {
		return 1, err
	}

	em, err := c.getEmailMappings()
	if err != nil {
		return 1, err
	}

	for _, user := range users {
		em.setGitlabUserID(user.Email, user.ID)
		userSecondaryEmails, err := c.listGitlabUserEmails(user.ID)
		if err != nil {
			return 1, err
		}

		for _, email := range userSecondaryEmails {
			em.setGitlabUserID(email.Email, user.ID)
		}
	}

	err = c.saveEmailMappings(em)
	if err != nil {
		return 1, err
	}

	return 0, nil
}

func (em *EmailMappings) setGitlabUserID(email string, GitlabUserID int) {
	if _, ok := (*em)[email]; ok {
		(*em)[email].GitlabUserID = GitlabUserID
	} else {
		(*em)[email] = &Mapping{GitlabUserID: GitlabUserID}
	}
}

func (c *client) listGitlabUserEmails(userID int) (emails []*gitlab.Email, err error) {
	emails, _, err = c.gitlabAdmin.Users.ListEmailsForUser(userID, &gitlab.ListEmailsForUserOptions{})
	return
}

func (c *client) listGitlabUsers() (users []*gitlab.User, err error) {
	// Initialize this variable as we have been experiencing some nil pointer deference
	// from ListUsers()
	foundUsers := []*gitlab.User{}
	_ = foundUsers // this is a workaround for the go vet, ineffassign check

	var resp *gitlab.Response

	opt := &gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	for {
		foundUsers, resp, err = c.gitlabAdmin.Users.ListUsers(opt)
		if err != nil {
			return
		}

		for _, user := range foundUsers {
			users = append(users, user)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		opt.ListOptions.Page = resp.NextPage
	}

	return
}
