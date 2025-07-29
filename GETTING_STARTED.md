# Getting Started

These instructions assume you want to achieve the highest SLSA Source Level.
If not you may have to modify the steps somewhat.

## Enable Controls

First, enable continuity controls within the target GitHub repo.

1. Go to the GitHub repo
2. Click the 'Settings' option
3. Click 'Rules -> Rulesets'
4. Click 'New Ruleset -> Import a ruleset'
5. Upload [docs/rulesets/source_level_3_basic.json](docs/rulesets/source_level_3_basic.json)
6. Click 'Create'

## Enable Source PoC workflow

Now, enable a workflow that will evaluate the SLSA level, create provenance, etc...

1. Create a clean checkout of the target repo
2. Create a new file named `.github/workflows/compute_slsa_source.yml`
3. Add the following content

```yaml
name: SLSA Source
on:
  push:
    branches: [ "main" ]

jobs:
  # Whenever new source is pushed recompute the slsa source information.
  check-change:
    permissions:
      contents: write # needed for storing the vsa in the repo.
      id-token: write
    uses: slsa-framework/slsa-source-poc/.github/workflows/compute_slsa_source.yml@main

```

4. Submit the change to your main branch

## Validate Source PoC workflow

Let's verify that everything is working

1. Note the digest of **merged** commit that added the workflow above
2. Run the verification command

`go run github.com/slsa-framework/slsa-source-poc/sourcetool@latest verifycommit --commit <commit digest> --owner <YOUR REPO'S ORG> --repo <YOUR REPO'S NAME> --branch main`

3. You should see the message
`SUCCESS: commit <commit digest> verified with [SLSA_SOURCE_LEVEL_1]`

Move to the next section to get to `SLSA_SOURCE_LEVEL_3`.

## Create a policy file

Now let's create the policy file that will upgrade your SLSA level to level 3.

1. Create a fork of https://github.com/slsa-framework/slsa-source-poc
2. Within that fork create a new branch for your policy
3. Within a clean working directory run

`go run github.com/slsa-framework/slsa-source-poc/sourcetool@latest createpolicy --owner <YOUR REPO'S ORG> --repo <YOUR REPO'S NAME> --branch main`

e.g.

`go run github.com/slsa-framework/slsa-source-poc/sourcetool@latest createpolicy --owner TomHennen --repo wrangle --branch main`

4. Edit the created policy file to set the `canonical_repo` field to the canonical repo for this source

(TODO: see if we can remove this annoyance at some point)

e.g. `"canonical_repo": "https://github.com/TomHennen/wrangle",`

5. Add & commit the created policy file.
6. Send a PR with the change to https://github.com/slsa-framework/slsa-source-poc
7. Once it's approved you'll be at SLSA Source Level 3 for your next change.

## Validate Source Level 3 workflow

Let's verify that everything is working

1. Make and merge a change to the protected branch (`main`) in **your** repo.
2. Note the digest of **merged** commit
3. Run the verification command

`go run github.com/slsa-framework/slsa-source-poc/sourcetool@latest verifycommit --commit <commit digest> --owner <YOUR REPO'S ORG> --repo <YOUR REPO'S NAME> --branch main`

4. You should see the message
`SUCCESS: commit <commit digest> verified with [SLSA_SOURCE_LEVEL_3]`
