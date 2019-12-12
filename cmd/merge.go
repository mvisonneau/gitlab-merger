package cmd

import (
	"fmt"
	"sort"

	"github.com/mvisonneau/go-gitlab"
	"github.com/nlopes/slack"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type mergeArgs struct {
	project   string
	sourceRef string
	targetRef string
	mrPrefix  string
}

// Merge refs
func Merge(ctx *cli.Context) error {
	if err := configure(ctx); err != nil {
		return cli.NewExitError(err, 1)
	}

	requiredFlags := []string{
		"project",
		"source-ref",
		"target-ref",
		"mr-prefix",
	}

	if err := mandatoryStringOptions(ctx, requiredFlags); err != nil {
		return exit(err, 1)
	}

	a := &mergeArgs{
		project:   ctx.String("project"),
		sourceRef: ctx.String("source-ref"),
		targetRef: ctx.String("target-ref"),
		mrPrefix:  ctx.String("mr-prefix"),
	}

	log.Infof("Checking existing merge requests..")
	mr, err := c.findExistingMR(a)
	if err != nil {
		return exit(err, 1)
	}

	if mr != nil {
		log.Infof("An opened MR already exists : %s", mr.WebURL)
		return exit(nil, 0)
	}

	log.Infof("No open MR found! continuing..")

	log.Infof("Fetching Project ID of '%s'", a.project)
	projectID, err := c.getProjectIDByName(a.project)
	if err != nil {
		return exit(err, 1)
	}

	log.Infof("Comparing refs..")
	cmp, err := c.compareRefs(a)
	if err != nil {
		return exit(err, 1)
	}

	if len(cmp.Commits) > 0 {
		log.Infof("Found %d commit(s)", len(cmp.Commits))

		em, notFoundEmails, err := c.getCommiters(&projectID, cmp)
		if err != nil {
			return exit(err, 1)
		}

		log.Infof("Matched %d commiter(s) in GitLab", len(*em))

		for email, mapping := range *em {
			log.Infof("-> User ID : %d | Email : %s", mapping.GitlabUserID, email)
		}

		if len(notFoundEmails) > 0 {
			log.Infof("%d commiter email(s) could not be matched", len(notFoundEmails))
			for _, email := range notFoundEmails {
				log.Infof("--> %s", email)
			}
		}

		mr, err := c.createMR(a, em, notFoundEmails)
		if err != nil {
			return exit(err, 1)
		}

		log.Infof("MR created : %s", mr.WebURL)

		slackChannel := ctx.String("slack-channel")
		if slackChannel != "" && c.slack != nil {
			log.Infof("Notifiying slack channel '%s' about the new MR", slackChannel)
			if err = c.notifySlackChannel(slackChannel, a, mr, em); err != nil {
				return exit(err, 1)
			}
		}
	} else {
		log.Infof("No commit found in the refs diff, exiting..")
	}

	return exit(nil, 0)
}

func (c *client) compareRefs(a *mergeArgs) (*gitlab.Compare, error) {
	opts := &gitlab.CompareOptions{
		From: &a.targetRef,
		To:   &a.sourceRef,
	}

	cmp, _, err := c.gitlab.Repositories.Compare(a.project, opts)
	return cmp, err
}

func (c *client) findExistingMR(a *mergeArgs) (*gitlab.MergeRequest, error) {
	opened := "opened"
	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:        &opened,
		SourceBranch: &a.sourceRef,
		TargetBranch: &a.targetRef,
	}

	mrs, _, err := c.gitlab.MergeRequests.ListProjectMergeRequests(a.project, opts)
	if len(mrs) > 0 {
		return mrs[0], err
	}

	return nil, err
}

func (c *client) getCommiters(projectID *int, cmp *gitlab.Compare) (*EmailMappings, []string, error) {
	knownMappings, err := c.getEmailMappings()
	if err != nil {
		return nil, nil, err
	}

	em := &EmailMappings{}
	notFoundEmails := []string{}

	for _, commit := range cmp.Commits {
		// Check if we haven't already found it
		if _, found := (*em)[commit.CommitterEmail]; found {
			continue
		}

		// Attempt to find the user in the known list
		emailMapping := knownMappings.getMapping(commit.CommitterEmail)
		if emailMapping != nil {
			(*em)[commit.CommitterEmail] = emailMapping
			continue
		}

		if !stringSliceContains(notFoundEmails, commit.CommitterEmail) {
			notFoundEmails = append(notFoundEmails, commit.CommitterEmail)
		}
	}

	return em, notFoundEmails, nil
}

func (c *client) getProjectIDByName(project string) (int, error) {
	p, _, err := c.gitlab.Projects.GetProject(project, &gitlab.GetProjectOptions{})
	if err != nil {
		return 0, err
	}

	return p.ID, nil
}

func (c *client) createMR(a *mergeArgs, em *EmailMappings, notFoundEmails []string) (*gitlab.MergeRequest, error) {
	title := fmt.Sprintf("%s Merge '%s' into '%s' 🚀", a.mrPrefix, a.sourceRef, a.targetRef)
	description := "This is an automated MR to get review/approval from the following commiters:\n"

	committerIDs := []int{}
	for _, mapping := range *em {
		gitlabUser, err := c.getGitlabUserByID(mapping.GitlabUserID)
		if err != nil {
			return nil, err
		}

		if !intSliceContains(committerIDs, gitlabUser.ID) {
			committerIDs = append(committerIDs, gitlabUser.ID)
		}

		description += fmt.Sprintf("- @%s\n", gitlabUser.Username)
	}

	if len(notFoundEmails) > 0 {
		description += "\n\nFollowing commiters could not be found in GitLab, you need to ping them manually:\n"
		for _, email := range notFoundEmails {
			description += fmt.Sprintf("- %s", email)
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

	mr, _, err := c.gitlab.MergeRequests.CreateMergeRequest(a.project, mrOpt)
	if err != nil {
		return nil, err
	}

	// Update the amount of approvers (EE only)
	approvalsCount := len(*em) + len(notFoundEmails)
	cmraOpt := &gitlab.ChangeMergeRequestApprovalConfigurationOptions{
		ApprovalsRequired: &approvalsCount,
	}

	_, _, err = c.gitlab.MergeRequestApprovals.ChangeApprovalConfiguration(a.project, mr.IID, cmraOpt)
	if err != nil {
		return mr, err
	}

	// Update the approvers list
	cmraaOpt := &gitlab.ChangeMergeRequestAllowedApproversOptions{
		ApproverIDs:      committerIDs,
		ApproverGroupIDs: []int{},
	}

	_, _, err = c.gitlab.MergeRequestApprovals.ChangeAllowedApprovers(a.project, mr.IID, cmraaOpt)
	if err != nil {
		return mr, err
	}

	return mr, err
}

func (c *client) getGitlabUserByID(userID int) (user *gitlab.User, err error) {
	user, _, err = c.gitlab.Users.GetUser(userID)
	return
}

func (c *client) getMergeRequestCommits(mr *gitlab.MergeRequest) (commits []*gitlab.Commit, err error) {
	commits, _, err = c.gitlab.MergeRequests.GetMergeRequestCommits(mr.ProjectID, mr.IID, &gitlab.GetMergeRequestCommitsOptions{})
	return
}

func (c *client) notifySlackChannel(channel string, a *mergeArgs, mr *gitlab.MergeRequest, em *EmailMappings) (err error) {
	attachment := slack.Attachment{
		Title: fmt.Sprintf("Click for MR"),
		TitleLink: fmt.Sprintf("%s", mr.WebURL),
		Pretext: fmt.Sprintf("🚀 merging `%s` to `%s` in *%s*", a.sourceRef, a.targetRef, a.project),
		Text: "",
	}

	commits, err := c.getMergeRequestCommits(mr)
	if err != nil {
		return
	}

	for _, commit := range commits {
		if len(attachment.Text) > 0 {
			attachment.Text += "\n"
		}

		commitURL := fmt.Sprintf("%s/diffs?commit_id=%s", mr.WebURL, commit.ID)
		user := commit.AuthorEmail
		mapping := em.getMapping(commit.AuthorEmail)
		if mapping != nil {
			user = fmt.Sprintf("<@%s>", mapping.SlackUserID)
		}

		attachment.Text += fmt.Sprintf("• <%s|%s> - %s | %s", commitURL, commit.ShortID, user, commit.Title)
	}

	_, _, err = c.slack.PostMessage(
		channel,
		slack.MsgOptionAttachments(attachment),
	)

	return err
}

func stringSliceContains(s []string, search string) bool {
	i := sort.SearchStrings(s, search)
	return i < len(s) && s[i] == search
}

func intSliceContains(s []int, search int) bool {
	i := sort.SearchInts(s, search)
	return i < len(s) && s[i] == search
}
