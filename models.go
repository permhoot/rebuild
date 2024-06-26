// Copyright © 2024 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"time"
)

type event struct {
	Description  string     `json:"description"`
	MasterBranch string     `json:"master_branch"`
	PusherType   string     `json:"pusher_type"`
	Ref          string     `json:"ref"`
	RefType      string     `json:"ref_type"`
	Repository   repository `json:"repository"`
	Sender       sender     `json:"sender"`
}

type owner struct {
	AvatarURL         string `json:"avatar_url"`
	EventsURL         string `json:"events_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	GravatarID        string `json:"gravatar_id"`
	HTMLURL           string `json:"html_url"`
	ID                int    `json:"id"`
	Login             string `json:"login"`
	NodeID            string `json:"node_id"`
	OrganizationsURL  string `json:"organizations_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	ReposURL          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	Type              string `json:"type"`
	URL               string `json:"url"`
}

type repository struct {
	AllowForking             bool      `json:"allow_forking"`
	ArchiveURL               string    `json:"archive_url"`
	Archived                 bool      `json:"archived"`
	AssigneesURL             string    `json:"assignees_url"`
	BlobsURL                 string    `json:"blobs_url"`
	BranchesURL              string    `json:"branches_url"`
	CloneURL                 string    `json:"clone_url"`
	CollaboratorsURL         string    `json:"collaborators_url"`
	CommentsURL              string    `json:"comments_url"`
	CommitsURL               string    `json:"commits_url"`
	CompareURL               string    `json:"compare_url"`
	ContentsURL              string    `json:"contents_url"`
	ContributorsURL          string    `json:"contributors_url"`
	CreatedAt                time.Time `json:"created_at"`
	DefaultBranch            string    `json:"default_branch"`
	DeploymentsURL           string    `json:"deployments_url"`
	Description              any       `json:"description"`
	Disabled                 bool      `json:"disabled"`
	DownloadsURL             string    `json:"downloads_url"`
	EventsURL                string    `json:"events_url"`
	Fork                     bool      `json:"fork"`
	Forks                    int       `json:"forks"`
	ForksCount               int       `json:"forks_count"`
	ForksURL                 string    `json:"forks_url"`
	FullName                 string    `json:"full_name"`
	GitCommitsURL            string    `json:"git_commits_url"`
	GitRefsURL               string    `json:"git_refs_url"`
	GitTagsURL               string    `json:"git_tags_url"`
	GitURL                   string    `json:"git_url"`
	HasDiscussions           bool      `json:"has_discussions"`
	HasDownloads             bool      `json:"has_downloads"`
	HasIssues                bool      `json:"has_issues"`
	HasPages                 bool      `json:"has_pages"`
	HasProjects              bool      `json:"has_projects"`
	HasWiki                  bool      `json:"has_wiki"`
	Homepage                 any       `json:"homepage"`
	HooksURL                 string    `json:"hooks_url"`
	HTMLURL                  string    `json:"html_url"`
	ID                       int       `json:"id"`
	IsTemplate               bool      `json:"is_template"`
	IssueCommentURL          string    `json:"issue_comment_url"`
	IssueEventsURL           string    `json:"issue_events_url"`
	IssuesURL                string    `json:"issues_url"`
	KeysURL                  string    `json:"keys_url"`
	LabelsURL                string    `json:"labels_url"`
	Language                 any       `json:"language"`
	LanguagesURL             string    `json:"languages_url"`
	License                  any       `json:"license"`
	MergesURL                string    `json:"merges_url"`
	MilestonesURL            string    `json:"milestones_url"`
	MirrorURL                any       `json:"mirror_url"`
	Name                     string    `json:"name"`
	NodeID                   string    `json:"node_id"`
	NotificationsURL         string    `json:"notifications_url"`
	OpenIssues               int       `json:"open_issues"`
	OpenIssuesCount          int       `json:"open_issues_count"`
	Owner                    owner     `json:"owner"`
	Private                  bool      `json:"private"`
	PullsURL                 string    `json:"pulls_url"`
	PushedAt                 time.Time `json:"pushed_at"`
	ReleasesURL              string    `json:"releases_url"`
	Size                     int       `json:"size"`
	SSHURL                   string    `json:"ssh_url"`
	StargazersCount          int       `json:"stargazers_count"`
	StargazersURL            string    `json:"stargazers_url"`
	StatusesURL              string    `json:"statuses_url"`
	SubscribersURL           string    `json:"subscribers_url"`
	SubscriptionURL          string    `json:"subscription_url"`
	SvnURL                   string    `json:"svn_url"`
	TagsURL                  string    `json:"tags_url"`
	TeamsURL                 string    `json:"teams_url"`
	Topics                   []any     `json:"topics"`
	TreesURL                 string    `json:"trees_url"`
	UpdatedAt                time.Time `json:"updated_at"`
	URL                      string    `json:"url"`
	Visibility               string    `json:"visibility"`
	Watchers                 int       `json:"watchers"`
	WatchersCount            int       `json:"watchers_count"`
	WebCommitSignoffRequired bool      `json:"web_commit_signoff_required"`
}

type sender struct {
	AvatarURL         string `json:"avatar_url"`
	EventsURL         string `json:"events_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	GravatarID        string `json:"gravatar_id"`
	HTMLURL           string `json:"html_url"`
	ID                int    `json:"id"`
	Login             string `json:"login"`
	NodeID            string `json:"node_id"`
	OrganizationsURL  string `json:"organizations_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	ReposURL          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	Type              string `json:"type"`
	URL               string `json:"url"`
}
