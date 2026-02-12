# Working Notes

## Enterprise hardening (part 2)

- Fork PR (merged): https://github.com/PaxMachinaOne/GoMikroBot/pull/2
  - Squash merge commit: `50308882e197487fd1fd5fabf3dcab7b1d0151f9`
  - Adds HTTP middleware (panic recovery, rate limiting, body size limit), `/health` + `/ready`, graceful shutdown.

- Upstream PR (pending): scalytics/GoMikroBot
  - Blocked: current GH_TOKEN cannot perform `createPullRequest` on `scalytics/GoMikroBot`.
  - Intended PR title: "Enterprise hardening part 2"
  - Intended body: "HTTP middleware (panic recovery, rate-limit), health/ready, graceful shutdown."
