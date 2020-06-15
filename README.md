# Cherry Bot

![](https://user-images.githubusercontent.com/9587680/60788142-95abc100-a18e-11e9-9a42-fbf21a023449.jpg)

Cherry Bot is a bot which helps you automate some work on github.

## Usage

* database setup
  
  Cherry Picker support mysql prorocol, using TiDB or MariaDB. Import `bot.sql` into your database.

  Edit `cherry_picker.slack_users` table, this table is used send direct message to Slack from a GitHub account. Insert GitHub and Slack account email relations in this table.

* edit `config.toml`, see [configuration](#configuration).

* build and run, see [flag](#flag).

## Configuration

```toml
[database]
address = "127.0.0.1"
port = 3306
username = "root"
password = ""
dbname = "cherry_picker"

[slack]
heartbeat = "admin@pingcap.com"
token = ""
mute = false
hello = true

[github]
token = ""
bot = "cherry-picker"

[[repos]]
owner = "owner"
repo = "repo"
interval = 300000
fullupdate = 86400000
webhookSecret = "secret"
rule = "needs-cherry-pick-([0-9.]+)"
release = "release-[version]"
typeLabel = "type/[version] cherry-pick"
ignoreLabel = ".*LGT.*"
dryrun = true
cherryPick = false
cherryPickChannel = "cherry-picker-test"

labelCheck = true
labelCheckChannel = "label-notice-test"
defaultChecker = "admin@pingcap.com"

prLimit = true
maxPrOpened = 3
prLimitMode = "blocklist"

merge = true
canMergeLabel = "can merge"
```

### database

| parameter  | description |
| - | - |
| address | database address |
| port | database port |
| username | database username |
| password | database password |
| dbname | database name |

### Slack

| parameter  | description |
| - | - |
| heartbeat | email address, heartbeat will send to this Slack user every 1hour |
| token | Slack APP OAuth access token |
| mute | if mute Slack notice, turn on will disable all Slack message |
| hello | if send hello message to Slack |

### Github

| parameter  | description |
| - | - |
| token | GitHub access token |
| bot | GitHub username |

### repos `Array`

| parameter  | description |
| - | - |
| owner | repository owner, user or organization name |
| repo | repository name |
| interval | latest pull requests polling interval |
| fullupdate | 30 days pull requests polling interval |
| webhookSecret | GitHub webhook secret, for verifying webhook request |
| rule | test which PR labels that needs cherry pick, using regex to match target version |
| release | cherry pick PR merge to which branch, `[version]` will be replaced by specific version number |
| typeLabel | name template of label which will be converted and before attatched to submit cherry pick PR |
| ignoreLabel | name template of label which will be ignored when copying labels to cherry pick PR |
| dryrun  | test flag for cherry pick, turn on it will disable create PR in GitHub |
| cherryPick  | cherry pick flag, turn off to disable all cherry pick |
| cherryPickChannel  | cherry notice Slack channel, make sure channel exist |
| runTestCommand  | if test command is setting, bot will create test command after creating cherry pick PR |
| labelCheck  | label check flag, turn off to disable all label checking |
| labelCheckChannel  | label check Slack channel, make sure channel exist |
| defaultChecker | PR submitted by contributors which don't have Slack account or not in `cherry_picker.slack_users` will be sent to default checker, using `,` to seperate several checkers. |
| prLimit | pr limit flag, turn off will not have opened PR limit for one user |
| maxPrOpened | max opened PR amount of one user |
| prLimitMode | can be `allowlist` or `blocklist`. `allowlist` stands for that user out of this list will have PR count limit, `blocklist` stands for the opposite. Other values will apply PR count limit above all users |
| prLimitLabel | bot will add a label for PRs bot closed |
| contributorLabel | bot will specify if a PR is submit by contributor and add a label for it |
| prLimitOrgs | this is for bot to specify whether a user is organization member, use "," to seperate orgs |
| merge | auto merge flag, turn off to disable auto merge |
| canMergeLabel | trigger label of auto merge job |

## Flag

```
Usage of ./cherry-picker:
  -addr string
        listen address (default "0.0.0.0")
  -c string
        config path (default "./config.toml")
  -port int
        listen port (default 8080)
```

## API

* `[GET]` Check pull requests in last month

This API will fetch pull requests last month and send report to Slack channel, Send with webhook secret(for safety and GitHub API limit).

```sh
curl http://localhost:8080/history/owner/repo?secret=secret
```

* `[GET]` get allowlist

```sh
curl http://localhost:8080/prlimit/allowlist/you06/cherry-pick-playground?secret=secret
```

* `[POST]` add a user to allowlist

```sh
curl -X POST http://localhost:8080/webhook/prlimit/allowlist/you06/cherry-pick-playground/you06?secret=secret
```

* `[POST]` remove a user from allowlist

```sh
curl -X DELETE http://localhost:8080/webhook/prlimit/allowlist/you06/cherry-pick-playground/you06?secret=secret
```

