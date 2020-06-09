package cmd

import (
	"fmt"
	"sort"

	"github.com/openlyinc/pointy"
	"github.com/slack-go/slack"
	"github.com/xanzy/go-gitlab"

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
func Merge(ctx *cli.Context) (int, error) {
	if err := configure(ctx); err != nil {
		return 1, err
	}

	requiredFlags := []string{
		"project",
		"source-ref",
		"target-ref",
		"mr-prefix",
	}

	if err := mandatoryStringOptions(ctx, requiredFlags); err != nil {
		return 1, err
	}

	a := &mergeArgs{
		project:   ctx.String("project"),
		sourceRef: ctx.String("source-ref"),
		targetRef: ctx.String("target-ref"),
		mrPrefix:  ctx.String("mr-prefix"),
	}

	log.Infof("checking existing merge requests..")
	mr, err := c.findExistingMR(a)
	if err != nil {
		return 1, err
	}

	if mr != nil {
		log.WithFields(log.Fields{
			"merge-request-iid": mr.IID,
			"merge-request-url": mr.WebURL,
		}).Warnf("an opened MR already exists, exiting")
		return 0, err
	}

	log.Infof("no open MR found, comparing refs..")
	cmp, err := c.compareRefs(a)
	if err != nil {
		return 1, err
	}

	if len(cmp.Commits) > 0 {
		log.WithFields(log.Fields{
			"commit-count": len(cmp.Commits),
		}).Infof("found differences between refs")

		em, notFoundEmails, err := c.getCommiters(cmp)
		if err != nil {
			return 1, err
		}

		log.WithFields(log.Fields{
			"committer-count": len(*em),
		}).Infof("done matching committers in GitLab")

		for email, mapping := range *em {
			log.WithFields(log.Fields{
				"gitlab-user-id": mapping.GitlabUserID,
				"email":          email,
			}).Infof("found GitLab user")
		}

		if len(notFoundEmails) > 0 {
			log.WithFields(log.Fields{
				"not-found-committer-count": len(notFoundEmails),
			}).Infof("some commiters emails were not matched in GitLab")
			for _, email := range notFoundEmails {
				log.WithFields(log.Fields{
					"email": email,
				}).Infof("email not found in GitLab users")
			}
		}

		mr, err := c.createMR(a, em, notFoundEmails)
		if err != nil {
			return 1, err
		}

		log.WithFields(log.Fields{
			"merge-request-iid": mr.IID,
			"merge-request-url": mr.WebURL,
		}).Infof("merge request created")

		slackChannel := ctx.String("slack-channel")
		if slackChannel != "" && c.slack != nil {
			log.WithFields(log.Fields{
				"slack-channel": slackChannel,
			}).Infof("notifying slack channel about the new MR")
			if err = c.notifySlackChannel(slackChannel, a, mr, em); err != nil {
				return 1, err
			}
		}
	} else {
		log.Infof("no differences found between the refs, exiting")
	}

	return 0, err
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

func (c *client) getCommiters(cmp *gitlab.Compare) (*EmailMappings, []string, error) {
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
			description += fmt.Sprintf("- @%s\n", gitlabUser.Username)
		}
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

	// Create the MR
	mr, _, err := c.gitlab.MergeRequests.CreateMergeRequest(a.project, mrOpt)
	if err != nil {
		return nil, err
	}

	// List and Delete all the existing approval rules on the MR that may be part of the project config (EE starter/bronze only)
	approvalRules, _, err := c.gitlab.MergeRequestApprovals.GetApprovalRules(a.project, mr.ID)
	if err != nil {
		return nil, err
	}

	for _, rule := range approvalRules {
		if _, err = c.gitlab.MergeRequestApprovals.DeleteApprovalRule(a.project, mr.ID, rule.ID); err != nil {
			return nil, err
		}
	}

	// Create a new rule to get the actual committers to approve the MR
	// For the missing committers, we do not take them in account to avoid getting blocked
	_, _, err = c.gitlab.MergeRequestApprovals.CreateApprovalRule(a.project, mr.ID, &gitlab.CreateMergeRequestApprovalRuleOptions{
		Name:              pointy.String("committers"),
		ApprovalsRequired: pointy.Int(len(*em)),
		UserIDs:           committerIDs,
	})

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
		Pretext: fmt.Sprintf("🚀 merging `%s` to `%s` in *%s* - <%s|*!%d*>", a.sourceRef, a.targetRef, a.project, mr.WebURL, mr.IID),
		Text:    "",
		Footer:  fmt.Sprintf("%s/diffs", mr.WebURL),
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
