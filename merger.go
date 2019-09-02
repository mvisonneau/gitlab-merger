package main

import (
	// "crypto/tls"
	// "io/ioutil"
	// "net/http"
	// "os"
	// "regexp"
	// "time"
	"fmt"

	"github.com/mvisonneau/go-gitlab"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type client struct {
	*gitlab.Client
}

type args struct {
	project   string
	sourceRef string
	targetRef string
	mrPrefix  string
}

func mandatoryStringOptions(ctx *cli.Context, opts []string) (err error) {
	for _, o := range opts {
		if ctx.GlobalString(o) == "" {
			return fmt.Errorf("%s is required", o)
		}
	}
	return nil
}

func run(ctx *cli.Context) error {
	configureLogging(ctx.GlobalString("log-level"), ctx.GlobalString("log-format"))

	opts := []string{
		"gitlab-url",
		"gitlab-token",
		"project",
		"source-ref",
		"target-ref",
	}

	if err := mandatoryStringOptions(ctx, opts); err != nil {
		log.Errorf("%v", err)
		return nil
	}

	c := &client{
		gitlab.NewClient(nil, ctx.GlobalString("gitlab-token")),
	}
	c.SetBaseURL(ctx.GlobalString("gitlab-url"))

	a := &args{
		project:   ctx.GlobalString("project"),
		sourceRef: ctx.GlobalString("source-ref"),
		targetRef: ctx.GlobalString("target-ref"),
		mrPrefix:  ctx.GlobalString("mr-prefix"),
	}

	log.Infof("Checking existing merge requests..")
	mr, err := c.findExistingMR(a)
	if err != nil {
		log.Errorf("%v", err)
		return nil
	}

	if mr != nil {
		log.Infof("An opened MR already exists : %s", mr.WebURL)
		return nil
	}

	log.Infof("No open MR found! continuing..")

	log.Infof("Fetching Project ID of '%s'", a.project)
	projectID, err := c.getProjectIDByName(a.project)
	if err != nil {
		log.Errorf("%v", err)
		return nil
	}

	log.Infof("Comparing refs..")
	cmp, err := c.compareRefs(a)
	if err != nil {
		log.Errorf("%v", err)
		return nil
	}

	if len(cmp.Commits) > 0 {
		log.Infof("Found %d commit(s)", len(cmp.Commits))

		commiters, notFoundEmails, err := c.getCommiters(&projectID, cmp)
		if err != nil {
			log.Errorf("%v", err)
			return nil
		}

		log.Infof("Matched %d commiter(s) in GitLab", len(commiters))

		for _, commiter := range commiters {
			log.Infof("-> User ID : %d | Email : %s", commiter.ID, commiter.Email)
		}

		if len(notFoundEmails) > 0 {
			log.Infof("%d commiter email(s) could not be matched", len(notFoundEmails))
			fmt.Println(notFoundEmails)
		}

		mr, err := c.createMR(a, commiters, notFoundEmails)
		if err != nil {
			log.Errorf("%v", err)
			return nil
		}

		log.Infof("MR created : %s", mr.WebURL)
	} else {
		log.Infof("No commit found in the refs diff, exiting..")
	}

	return nil
}

func (c *client) compareRefs(a *args) (*gitlab.Compare, error) {
	opts := &gitlab.CompareOptions{
		From: &a.targetRef,
		To:   &a.sourceRef,
	}

	cmp, _, err := c.Repositories.Compare(a.project, opts)
	return cmp, err
}

func (c *client) findExistingMR(a *args) (*gitlab.MergeRequest, error) {
	opened := "opened"
	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:        &opened,
		SourceBranch: &a.sourceRef,
		TargetBranch: &a.targetRef,
	}

	mrs, _, err := c.MergeRequests.ListProjectMergeRequests(a.project, opts)
	if len(mrs) > 0 {
		return mrs[0], err
	}

	return nil, err
}

func (c *client) getCommiters(projectID *int, cmp *gitlab.Compare) ([]*gitlab.User, []*string, error) {
	emailSearched := map[string]int{}
	commitersByID := map[int]*gitlab.User{}
	commiters := []*gitlab.User{}
	notFoundEmails := []*string{}

	// Only used starting from one user not found through it's primary email
	projectMembers := []*gitlab.ProjectMember{}

	for _, commit := range cmp.Commits {
		// Check if we haven't already searched it
		if _, searched := emailSearched[commit.CommitterEmail]; searched {
			continue
		}

		// Set as searched
		emailSearched[commit.CommitterEmail] = 0

		// Attempt to find the user by its primary email
		user, err := c.findUserByPrimaryEmail(&commit.CommitterEmail)
		if err != nil {
			return nil, nil, err
		}

		if user != nil {
			commitersByID[user.ID] = user
			continue
		}

		// List project members if not already done
		if len(projectMembers) == 0 {
			projectMembers, err = c.listProjectMembers(projectID)
			if err != nil {
				return nil, nil, err
			}
		}

		// Attempt to find the email in users secondary
		user, err = c.findUserBySecondaryEmail(&commit.CommitterEmail, projectMembers)
		if err != nil {
			return nil, nil, err
		}

		if user != nil {
			commitersByID[user.ID] = user
			continue
		}

		notFoundEmails = append(notFoundEmails, &commit.CommitterEmail)
	}

	for _, commiter := range commitersByID {
		commiters = append(commiters, commiter)
	}

	return commiters, notFoundEmails, nil
}

func (c *client) createMR(a *args, commiters []*gitlab.User, notFoundEmails []*string) (*gitlab.MergeRequest, error) {
	title := fmt.Sprintf("%s merging '%s' into '%s' 🚀", a.mrPrefix, a.sourceRef, a.targetRef)
	description := "This is an automated MR to get review/approval from the following commiters :\n"

	committerIDs := []*int{}
	for _, commiter := range commiters {
		committerIDs = append(committerIDs, &commiter.ID)
		description += fmt.Sprintf("- @%s\n", commiter.Username)
	}

	if len(notFoundEmails) > 0 {
		description += "\n\nFollowing commiters could not be found in GitLab, you need to ping them manually:\n"
		for _, email := range notFoundEmails {
			description += fmt.Sprintf("- %s", *email)
		}
	}

	// Create the MR
	mrOpt := &gitlab.CreateMergeRequestOptions{
		Title:        &title,
		Description:  &description,
		AssigneeIDs:  committerIDs,
		SourceBranch: &a.sourceRef,
		TargetBranch: &a.targetRef,
	}

	mr, _, err := c.MergeRequests.CreateMergeRequest(a.project, mrOpt)
	if err != nil {
		return nil, err
	}

	// Update the amount of approvers (EE only)
	approvalsCount := len(commiters) + len(notFoundEmails)
	cmraOpt := &gitlab.ChangeMergeRequestApprovalConfigurationOptions{
		ApprovalsRequired: &approvalsCount,
	}

	_, _, err = c.MergeRequests.ChangeMergeRequestApprovalConfiguration(a.project, mr.IID, cmraOpt)
	if err != nil {
		return mr, err
	}

	// Update the approvers list
	cmraaOpt := &gitlab.ChangeMergeRequestAllowedApproversOptions{
		ApproverIDs:      committerIDs,
		ApproverGroupIDs: []*int{},
	}

	_, _, err = c.MergeRequests.ChangeMergeRequestAllowedApprovers(a.project, mr.IID, cmraaOpt)
	if err != nil {
		return mr, err
	}

	return mr, err
}

func (c *client) getProjectIDByName(project string) (int, error) {
	p, _, err := c.Projects.GetProject(project, &gitlab.GetProjectOptions{})
	if err != nil {
		return 0, err
	}

	return p.ID, nil
}

func (c *client) getUserByID(userID int) (*gitlab.User, error) {
	user, _, err := c.Users.GetUser(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *client) findUserByPrimaryEmail(email *string) (*gitlab.User, error) {
	opt := &gitlab.ListUsersOptions{
		Search: email,
	}

	users, _, err := c.Users.ListUsers(opt)
	if len(users) > 0 {
		return users[0], err
	}

	return nil, err
}

func (c *client) listProjectMembers(projectID *int) (members []*gitlab.ProjectMember, err error) {
	var foundMembers []*gitlab.ProjectMember
	var resp *gitlab.Response

	opt := &gitlab.ListProjectMembersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		foundMembers, resp, err = c.ProjectMembers.ListAllProjectMembers(*projectID, opt)
		if err != nil {
			return
		}

		for _, member := range foundMembers {
			members = append(members, member)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		opt.ListOptions.Page = resp.NextPage
	}

	return
}

func (c *client) findUserBySecondaryEmail(email *string, members []*gitlab.ProjectMember) (*gitlab.User, error) {
	// List all secondary emails for users members of the project
	for _, member := range members {
		emails, _, err := c.Users.ListEmailsForUser(member.ID, &gitlab.ListEmailsForUserOptions{})
		if err != nil {
			return nil, err
		}

		for _, e := range emails {
			if e.Email == *email {
				user, err := c.getUserByID(member.ID)
				if err != nil {
					return nil, err
				}

				return user, nil
			}
		}
	}

	return nil, nil
}
