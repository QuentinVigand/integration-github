# WIP.md: Rationale and discussion topics

## Current goal
Having something that works enough to start discussion on how we should do the integration


Hello reader, this is an attempt at a github importer, here is the reasoning behind this first implementation and some discussions topics/questions:

### Rationale

- No git specific data: this can be added with a git specific integration/later
- Graphql (api v4) vs Rest (api v3): graphql should lead to less network calls and useless bloat in responses.
  - needs login for any call unlike v3, but v3 does to backup anything useful.
- github.com/shurcooL/githubv4: handles objects for query & result and is [recommended by google](https://github.com/google/go-github). I think the struct mapping (when we could just dump the returned json to a record) won't be a perf bottleneck and can be convenient for an exporter later.
  - note: considered raw go but not worth recreating a graphql client. Did not try "regular" graphql clients.
- structure: resource/resourceName/resourceX.json seemed to make sense considering it is close to recreating the graphql schema
  - /orgs/%s/org.json
  - /orgs/%s/teams/%s/members.json
- Orgs are fetched first but subresources are in nested loops (team -> members), this is an arbitrary choice for later parallelization (per org) and allow early GC (team per team), I don't know out it will hold when adding more data.

### Questions, notes & discussions

If any part of the codes or reason above seem wrong or debatable, free to comment!

Questions:
- What are parameters that make sense? One would assume that this integration could be used a lot to backup a single org, or a maybe personnal repo with no org.
- Do you believe the approach taken for this integration can scale (hold up to breaking changes, keep up to evercoming knew features..) even though it is a lot of manual data mapping.

Notes:
- Recreating the complete github data model in code (needed for graphql queries) seems redundant, but I don't see a way around it.
- Github's graphql does sometime have breaking changes in its graphql api https://docs.github.com/en/graphql/overview/breaking-changes, something to keep in mind before adding every known github resource to the repo.
- The graphql library rewrites scalar types which I am not a fan of, I hope this will be tackled

### Later
Features to add later:
- handle github's rate limiting (less straight forward with the graphql API than the REST)
- fail early if wrong access rights on token
- all other github data (issues,...)
