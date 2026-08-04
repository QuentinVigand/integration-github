package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/PlakarKorp/kloset/connectors"
	"github.com/PlakarKorp/kloset/connectors/importer"
	"github.com/PlakarKorp/kloset/location"
	"github.com/PlakarKorp/kloset/objects"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type github struct {
	client *githubv4.Client
	orgs   []string
}

func (g *github) Origin() string { return "https://api.github.com" }

func (g *github) Type() string { return "github" }

func (g *github) Root() string { return "/" }

func (g *github) Flags() location.Flags { return 0 }

func (g *github) Ping(ctx context.Context) error {
	var ping QueryPing
	err := g.client.Query(ctx, &ping, nil)
	if err != nil {
		return fmt.Errorf("running ping query: %w", err)
	}
	if ping.RateLimit.Remaining <= 0 {
		return fmt.Errorf("%w: 0/%d points left (resets at %s)",
			ErrRateLimitExceeded,
			ping.RateLimit.Limit,
			ping.RateLimit.ResetAt.Time.Format(time.RFC3339),
		)
	}
	return nil
}

func (g *github) importOrg(ctx context.Context, org NodeOrg, records chan<- *connectors.Record) error {
	basePath := fmt.Sprintf("/orgs/%s", org.Login)
	orgJson, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("marshalling org %s: %w", org.Login, err)
	}

	fi := objects.FileInfo{
		Lname:    "org.json",
		Lsize:    int64(len(orgJson)),
		Lmode:    0x644,
		LmodTime: time.Now(),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case records <- connectors.NewRecord(path.Join(basePath, "org.json"), "", fi, nil, func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(orgJson)), nil
	}):
	}

	var teamsCursor *githubv4.String
	teamsVars := map[string]any{
		"orgName":     org.Login,
		"teamsCursor": teamsCursor,
	}

	for {
		var qt QueryTeams
		if err := g.client.Query(ctx, &qt, teamsVars); err != nil {
			fmt.Fprintf(os.Stderr, "error: fetching teams: %v\n", err)
			break
		}

		for _, team := range qt.Organization.Teams.Nodes {
			members, err := g.getTeamMembers(ctx, org.Login, team.Slug)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			membersJson, err := json.Marshal(members)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: marshalling team members for %s: %v\n", team.Slug, err)
				continue
			}
			fullpath := path.Join(basePath, "teams", string(team.Slug), "members.json")

			fi := objects.FileInfo{
				Lname:    "members.json",
				Lsize:    int64(len(membersJson)),
				Lmode:    0x644,
				LmodTime: time.Now(),
			}

			select {
			case <-ctx.Done():
				fmt.Fprintf(os.Stderr, "error: saving team members for %s: %v\n", team.Slug, ctx.Err())
				continue
			case records <- connectors.NewRecord(fullpath, "", fi, nil, func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(membersJson)), nil
			}):
			}
		}
		if !qt.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		teamsVars["teamsCursor"] = new(qt.Organization.Teams.PageInfo.EndCursor)
	}
	return nil
}

func (g *github) getTeamMembers(ctx context.Context, org githubv4.String, team githubv4.String) ([]NodeMember, error) {
	var members []NodeMember
	var membersCursor *githubv4.String

	memberVars := map[string]any{
		"orgName":       org,
		"teamSlug":      team,
		"membersCursor": membersCursor,
	}

	for {
		var mq QueryMembers
		if err := g.client.Query(ctx, &mq, memberVars); err != nil {
			return members, fmt.Errorf("fetching members for team %s: %w", team, err)
		}

		members = append(members, mq.Organization.Team.Members.Nodes...)

		if !mq.Organization.Team.Members.PageInfo.HasNextPage {
			break
		}
		memberVars["membersCursor"] = new(mq.Organization.Team.Members.PageInfo.EndCursor)
	}
	return members, nil
}

func (g *github) getOrgs(ctx context.Context) ([]NodeOrg, error) {
	var orgs []NodeOrg
	var orgCursor *githubv4.String

	if g.orgs == nil {
		variables := map[string]any{
			"cursor": orgCursor,
		}

		for {
			var query QueryOrgs
			if err := g.client.Query(ctx, &query, variables); err != nil {
				return nil, fmt.Errorf("fetching orgs: %w", err)
			}
			orgs = append(orgs, query.Viewer.Organizations.Nodes...)

			if !query.Viewer.Organizations.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = new(query.Viewer.Organizations.PageInfo.EndCursor)
		}
	} else {
		var eg errgroup.Group
		orgs = make([]NodeOrg, len(g.orgs), len(g.orgs))
		for i, org := range g.orgs {
			eg.Go(func() error {
				variables := map[string]any{
					"orgLogin": githubv4.String(org),
				}
				var query QueryOrg
				if err := g.client.Query(ctx, &query, variables); err != nil {
					return fmt.Errorf("fetching org: %w", err)
				}
				orgs[i] = query.Organization
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return nil, err
		}
	}
	return orgs, nil
}

func (g *github) Import(ctx context.Context, records chan<- *connectors.Record, results <-chan *connectors.Result) error {
	defer close(records)

	orgs, err := g.getOrgs(ctx)
	if err != nil {
		return err
	}

	for _, org := range orgs {
		if err := g.importOrg(ctx, org, records); err != nil {
			fmt.Fprintf(os.Stderr, "error importing org %s: %v\n", org.Login, err)
		}
	}
	return nil
}
func (g *github) Close(context.Context) error { return nil }

func init() {
	importer.Register("github", 0, NewImporter)
}

func NewImporter(ctx context.Context, opts *connectors.Options, proto string, config map[string]string) (importer.Importer, error) {
	token, ok := config["token"]
	if !ok {
		return nil, fmt.Errorf("missing token in config")
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	orgsParam := config["orgs"]
	var orgs []string
	if orgsParam != "" {
		orgs = strings.Split(orgsParam, ",")
	}

	return &github{
		client: githubv4.NewClient(httpClient),
		orgs:   orgs,
	}, nil
}
