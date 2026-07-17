package github

import "github.com/shurcooL/githubv4"

// GitHub's graphql max page size is 100 hence the (first: 100..) in graphql tags.

type PageInfo struct {
	EndCursor   githubv4.String
	HasNextPage githubv4.Boolean
}

type QueryPing struct {
	Viewer struct {
		Login githubv4.String
	}
}

type QueryOrgs struct {
	Viewer struct {
		Organizations struct {
			Nodes    []NodeOrg
			PageInfo PageInfo
		} `graphql:"organizations(first: 100, after: $cursor)"`
	}
}
type QueryTeams struct {
	Organization struct {
		Teams struct {
			Nodes    []NodeTeam
			PageInfo PageInfo
		} `graphql:"teams(first: 100, after: $teamsCursor)"`
	} `graphql:"organization(login: $orgName)"`
}

type QueryMembers struct {
	Organization struct {
		Team struct {
			Members struct {
				Nodes    []NodeMember
				PageInfo PageInfo
			} `graphql:"members(first: 100, after: $membersCursor)"`
		} `graphql:"team(slug: $teamSlug)"`
	} `graphql:"organization(login: $orgName)"`
}

type NodeOrg struct {
	Login       githubv4.String
	Name        githubv4.String
	Description githubv4.String
}

type NodeTeam struct {
	Slug githubv4.String
}
type NodeMember struct {
	Login githubv4.String
	Name  githubv4.String
}
