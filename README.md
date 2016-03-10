# gitlab2gogs

Migrate your Gitlab repositories to Gogs.

Usage:

```
./gitlab2gogs -gitlab-host https://<yourgitlabhost> \
    -gitlab-api-path /api/v3
    -gitlab-token <your gitlab token> \
    -gitlab-user <gitlab admin user> \
    -gitlab-password <password of gitlab-user> \
    -gogs-url https://<yourgogshost> \
    -gogs-token <your gogs token> \
    -gogs-user <gogs admin username>
```

Organizations are created if they do not yet exists.
Existing repositories (in Gogs) are not overwritten.

To convert newly created Repository and Organization
names to lowercase, append `--lc-names` to command line.

For migration of Repositories in a single Organization: `-namespace <organization name>`

And to migrate a single Repository within that Organization: `-project <repository name>`

To make migrated Repositories as mirror (backup) Repository: `--mirror`
