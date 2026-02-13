# SLSA Source 1.2 Control Labels

The following table documents the control labels used in sourcetool and the
mapping to each SLSA Source level.

| Label | L1 | L2 | L3 | L4 |
| ------- | :--: | :--: | :--: | :--: |
| SLSA_SOURCE_ORG_SCS | ✓ | ✓ | ✓ | ✓ |
| SLSA_SOURCE_ORG_ACCESS_CONTROL | | ✓ | ✓ | ✓ |
| SLSA_SOURCE_ORG_SAFE_EXPUNGE | | ✓ | ✓ | ✓ |
| SLSA_SOURCE_ORG_CONTINUOUS_CONTROLS | | | ✓ | ✓ |
| SLSA_SOURCE_SCS_REPO_ID | ✓ | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_REVISION_ID | ✓ | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_DIFF_DISPLAY | ✓ | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_VSA | ✓ | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_HISTORY | | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_CONTINUITY | | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_IDENTITY | | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_PROVENANCE | | ✓ | ✓ | ✓ |
| SLSA_SOURCE_SCS_PROTECTED_REFS | | | ✓ | ✓ |
| SLSA_SOURCE_SCS_TWO_PARTY_REVIEW | | | | ✓ |

## GitHub Implementation

| Label | Implicit | Descr | Rationale |
| ------- | :--: | :--: | :--: |
| SLSA_SOURCE_ORG_SCS | ✓ | Choose an appropriate Source Control System | Project is hosted on GitHub |
| SLSA_SOURCE_ORG_ACCESS_CONTROL | ✓ | SCS to control access and enforce history | All changes in GitHub are access controlled |
| SLSA_SOURCE_ORG_SAFE_EXPUNGE | | Safe Expunging Process | There is no way to prove safe expunging but disabling force-pushes guarantees that no content is removed. Disabling it brakes continuity in this control. |
| SLSA_SOURCE_ORG_CONTINUOUS_CONTROLS | | Evidence of continuous enforcement via technical controls | Proven with evaluation against policy |
| SLSA_SOURCE_SCS_REPO_ID | ✓ | Repositories are uniquely identifiable | All GitHub repositories have a unique URI |
| SLSA_SOURCE_SCS_REVISION_ID | ✓ | Revisions are immutable and uniquely identifiable | Git commits are Merkle Trees |
| SLSA_SOURCE_SCS_DIFF_DISPLAY | ✓ | Tooling to display Changes between one Source Revision and another in a human readable form | Both git and GitHub are capable |
| SLSA_SOURCE_SCS_VSA | | SCS MUST generate a Source VSA to indicate the SLSA Source Level of any revision | Yes we can |
| SLSA_SOURCE_SCS_HISTORY | | History | On if force pushes are blocked (see [SLSA spec](https://slsa.dev/spec/v1.2/source-requirements#history)) |
| SLSA_SOURCE_SCS_CONTINUITY | | continuity MUST be established and tracked from a specific start revision | Audit from revision in policy |
| SLSA_SOURCE_SCS_IDENTITY | ✓ | The SCS MUST provide an identity management system | GitHub provides a user identity system |
| SLSA_SOURCE_SCS_PROVENANCE | | Attestations that contain information about how a specific revision was created | Signed provenance from sourcetool |
| SLSA_SOURCE_SCS_PROTECTED_REFS | | SCS MUST enforce customized controls for Named References | Complied when branch protection is on |
| SLSA_SOURCE_SCS_TWO_PARTY_REVIEW | | Changes in protected branches MUST be agreed to by two or more trusted persons | _not supported yet_ |
