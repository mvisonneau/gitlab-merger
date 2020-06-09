package cmd

import (
	"encoding/json"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

func (c *client) getCurrentUser() (user *gitlab.User, err error) {
	user, _, err = c.gitlab.Users.CurrentUser()
	log.Debugf("Found current user ID: %d", user.ID)
	return
}

func (c *client) findStorageSnippetID() (snippetID int, err error) {
	user, err := c.getCurrentUser()
	if err != nil {
		return
	}

	if v, err := strconv.Atoi(user.WebsiteURL); err == nil {
		log.WithFields(log.Fields{
			"snippet-id": v,
		}).Debugf("found existing snippet")
		return v, nil
	}

	log.Debugf("no existing snippet found")
	return
}

func (c *client) createSnippet(userID int) (err error) {
	title := "email mapping storage for gitlab-merger"
	visibility := gitlab.PrivateVisibility
	fileName := "mappings"
	content := "{}"

	snippetOpts := &gitlab.CreateSnippetOptions{
		Title:      &title,
		Content:    &content,
		FileName:   &fileName,
		Visibility: &visibility,
	}

	// Create the snippet
	log.WithFields(log.Fields{
		"gitlab-user-id": userID,
	}).Debugf("creating new snippet")
	snippet, _, err := c.gitlab.Snippets.CreateSnippet(snippetOpts)
	if err != nil {
		return
	}

	log.WithFields(log.Fields{
		"gitlab-user-id": userID,
		"snippet-id":     snippet.ID,
	}).Infof("new snippet created")

	// Update the current user
	snippetID := strconv.Itoa(snippet.ID)
	userOpts := &gitlab.ModifyUserOptions{
		WebsiteURL: &snippetID,
	}

	log.WithFields(log.Fields{
		"gitlab-user-id": userID,
		"snippet-id":     snippet.ID,
	}).Debugf("updating user website field to store snippet ID")

	_, _, err = c.gitlabAdmin.Users.ModifyUser(userID, userOpts)
	return
}

func (c *client) getEmailMappings() (*EmailMappings, error) {
	snippetID, err := c.findStorageSnippetID()
	if err != nil {
		return nil, err
	}

	em := &EmailMappings{}
	if snippetID != 0 {
		log.WithFields(log.Fields{
			"snippet-id": snippetID,
		}).Debugf("loading content from snippet")

		data, err := c.getSnippetContent(snippetID)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(data, &em)
		return em, nil
	}

	user, err := c.getCurrentUser()
	if err != nil {
		return nil, err
	}

	err = c.createSnippet(user.ID)
	return em, err
}

func (em *EmailMappings) getMapping(email string) *Mapping {
	if _, exists := (*em)[email]; exists {
		return (*em)[email]
	}

	return nil
}

func (c *client) saveEmailMappings(em *EmailMappings) (err error) {
	snippetID, err := c.findStorageSnippetID()
	if err != nil {
		return
	}

	err = c.updateSnippet(snippetID, em)
	return
}

func (c *client) getSnippetContent(snippetID int) (data []byte, err error) {
	data, _, err = c.gitlab.Snippets.SnippetContent(snippetID)
	return
}

func (c *client) updateSnippet(snippetID int, em *EmailMappings) (err error) {
	log.WithFields(log.Fields{
		"snippet-id": snippetID,
	}).Debugf("updating snippet content")

	data, err := json.Marshal(*em)
	if err != nil {
		return
	}

	strData := string(data)
	snippetOpts := &gitlab.UpdateSnippetOptions{
		Content: &strData,
	}

	_, _, err = c.gitlab.Snippets.UpdateSnippet(snippetID, snippetOpts)
	return
}
