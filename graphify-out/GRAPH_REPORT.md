# Graph Report - agentd-protocol  (2026-06-26)

## Corpus Check
- 78 files · ~41,545 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 698 nodes · 1085 edges · 45 communities (37 shown, 8 thin omitted)
- Extraction: 89% EXTRACTED · 11% INFERRED · 0% AMBIGUOUS · INFERRED: 119 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b206603f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]

## God Nodes (most connected - your core abstractions)
1. `T` - 37 edges
2. `T` - 37 edges
3. `T` - 32 edges
4. `assertRoundtrip()` - 20 edges
5. `T` - 15 edges
6. `NegotiateProtocolHello()` - 12 edges
7. `T` - 11 edges
8. `JSONRPCID` - 11 edges
9. `Wire types` - 11 edges
10. `T` - 9 edges

## Surprising Connections (you probably didn't know these)
- `TestControlMessageEdgeCases()` --calls--> `ControlType`  [INFERRED]
  protocol_test.go → control.go
- `TestKnownEngines()` --calls--> `KnownEngines()`  [INFERRED]
  delegation_test.go → delegation.go
- `TestStartDelegationAwaitOrDefault()` --calls--> `StartDelegationAwaitOrDefault()`  [INFERRED]
  delegation_test.go → delegation.go
- `TestV2FixturesParse()` --calls--> `ValidateCommandReceipt()`  [INFERRED]
  v2_test.go → receipts.go
- `TestProtocolHelloAdvertisesCommandReceipts()` --calls--> `NegotiateProtocolHello()`  [INFERRED]
  receipts_test.go → v2.go

## Import Cycles
- None detected.

## Communities (45 total, 8 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.06
Nodes (69): AgentDErrorData, AgentDEventEnvelope, ConnectionPhase, EventType, JSONRPCError, JSONRPCID, JSONRPCRequest, JSONRPCResponse (+61 more)

### Community 1 - "Community 1"
Cohesion: 0.07
Nodes (58): IsKnownEngine(), T, TestApprovalAttribution_OmitemptyBackwardCompat(), TestApprovalAttribution_Roundtrip(), TestDelegationCancelAckPayload_Roundtrip(), TestDelegationCancelControlFramesNoContent(), TestDelegationCancelLifecycleStateConstants(), TestDelegationCancelPayload_RequestIDRoundtrip() (+50 more)

### Community 2 - "Community 2"
Cohesion: 0.12
Nodes (39): assertRoundtrip(), RawMessage, T, mustJSON(), TestAckPayloadRoundtrip(), TestApplicationLivenessMessageConstants(), TestAuditEntryPayloadLegacyWithoutReason(), TestAuditEntryPayloadRoundtrip() (+31 more)

### Community 3 - "Community 3"
Cohesion: 0.05
Nodes (37): additionalProperties, minLength, type, minLength, type, $id, const, minimum (+29 more)

### Community 4 - "Community 4"
Cohesion: 0.05
Nodes (36): additionalProperties, type, type, additionalProperties, properties, type, additionalProperties, properties (+28 more)

### Community 5 - "Community 5"
Cohesion: 0.07
Nodes (31): NormalizedSupportBundleParams, SupportBundle, SupportBundleClient, SupportBundleContext, SupportBundleDaemon, SupportBundleParams, SupportBundleTransport, SupportClientDiagnostics (+23 more)

### Community 6 - "Community 6"
Cohesion: 0.07
Nodes (29): additionalProperties, $id, minLength, type, const, properties, id, payload (+21 more)

### Community 7 - "Community 7"
Cohesion: 0.07
Nodes (27): Approval lifecycle (feature 193), Backwards-compatibility, Changelog, Documentation, Envelope and control plane, Git status / changes viewer (feature 160 / 162), Git sync (feature 172), Git worktrees (feature 173) (+19 more)

### Community 8 - "Community 8"
Cohesion: 0.08
Nodes (25): GitBranch, GitBranchListRequest, GitBranchListResponse, GitBranchSwitchRequest, GitBranchSwitchResponse, GitFetchRequest, GitFetchResponse, GitPullRequest (+17 more)

### Community 9 - "Community 9"
Cohesion: 0.08
Nodes (24): additionalProperties, minLength, type, contentEncoding, type, $id, const, properties (+16 more)

### Community 10 - "Community 10"
Cohesion: 0.09
Nodes (23): AckPayload, AuditEntryPayload, ClientConnectedPayload, ClientCountPayload, ControlMessage, ControlType, DeactivateDeveloperPayload, EntitlementUpdatePayload (+15 more)

### Community 11 - "Community 11"
Cohesion: 0.09
Nodes (21): additionalProperties, $id, oneOf, const, pattern, type, const, type (+13 more)

### Community 12 - "Community 12"
Cohesion: 0.12
Nodes (20): GitWorktreeAddedPayload, GitWorktreeAddRequest, GitWorktreeAddResponse, GitWorktreeListRequest, GitWorktreeListResponse, GitWorktreeLockRequest, GitWorktreeLockResponse, GitWorktreeProgressPayload (+12 more)

### Community 13 - "Community 13"
Cohesion: 0.22
Nodes (19): MCPListPayload, MCPListResponse, MCPMutationPayload, MCPMutationResponse, MCPReconnectPayload, MCPRemovePayload, MCPServerConfig, MCPServersChangedPayload (+11 more)

### Community 14 - "Community 14"
Cohesion: 0.11
Nodes (16): ApprovalAttribution, DelegationCancelAckPayload, DelegationCancelPayload, DelegationForceAbortPayload, DelegationLinkPayload, DelegationPreviewByteStatus, DelegationPreviewDecisionPayload, DelegationPreviewPayload (+8 more)

### Community 15 - "Community 15"
Cohesion: 0.22
Nodes (16): T, TestAgentCapability186JSONKeys(), TestAgentCapability186RoundTrip(), TestAgentCapability195JSONKeys(), TestAgentCapability195RoundTrip(), TestAgentCapabilityAnswerQuestionDeprecationCoexistence(), TestAgentCapabilityAnswerQuestionFreeTextField(), TestAgentCapabilityFields() (+8 more)

### Community 16 - "Community 16"
Cohesion: 0.29
Nodes (11): CommandReceiptPayload, CommandReceiptReasonCode, CommandReceiptStage, IsKnownCommandReceiptReasonCode(), IsKnownCommandReceiptStage(), T, TestCommandReceiptConstants(), TestCommandReceiptJSONRoundtrip() (+3 more)

### Community 17 - "Community 17"
Cohesion: 0.24
Nodes (12): fixtureAgentMessage, fixtureApprovalPayload, RawMessage, T, TestCrossRepoWireFixturesParse(), validateApprovalResolvedFixture(), validateHistoryPageFixture(), validateHistoryPageRequestFixture() (+4 more)

### Community 18 - "Community 18"
Cohesion: 0.28
Nodes (12): T, TestMCPListPayloadRoundTrip(), TestMCPListResponseRequestIDOmitempty(), TestMCPListResponseRoundTrip(), TestMCPMutationPayloadRoundTrip(), TestMCPMutationResponseRoundTrip(), TestMCPRemoveTogglReconnectRoundTrip(), TestMCPServerConfigOmitempty() (+4 more)

### Community 19 - "Community 19"
Cohesion: 0.30
Nodes (9): SessionFeature, SessionFeatureReasonCode, SessionFeatureState, SessionFeatureStatusPayload, SessionInfo, ProviderCapabilityContract, IsKnownSessionFeatureReasonCode(), IsKnownSessionFeatureState() (+1 more)

### Community 20 - "Community 20"
Cohesion: 0.31
Nodes (10): ProviderCapabilityContract, ProviderCapabilitySource, ProviderCapabilitySupport, ProviderCommandArgumentRequirement, ProviderCommandDescriptor, ProviderCommandLifecycle, ProviderCommandStatusAfterDispatch, ProviderFeatureDescriptor (+2 more)

### Community 21 - "Community 21"
Cohesion: 0.33
Nodes (10): T, TestGitDiffRequestRoundtrip(), TestGitDiffResponseRoundtrip(), TestGitFileStatusJSONTags(), TestGitFileStatusRoundtrip(), TestGitNotAvailablePayloadRoundtrip(), TestGitStatusPayload_Feature172_ExtendedFields(), TestGitStatusPayload_Feature172_OmitEmpty() (+2 more)

### Community 22 - "Community 22"
Cohesion: 0.36
Nodes (7): CodexSandboxMode, IsKnownCodexSandboxMode(), KnownCodexSandboxModes(), T, TestCodexSandboxModeConstants(), TestIsKnownCodexSandboxMode(), TestKnownCodexSandboxModes()

### Community 23 - "Community 23"
Cohesion: 0.39
Nodes (8): T, TestSessionFeatureStatusConstants(), TestSessionFeatureStatusFixtures(), TestSessionFeatureStatusPayloadJSONRoundtrip(), TestSessionFeatureStatusPayloadRejectsUnknownStateAndReason(), TestSessionFeatureStatusPayloadRequiresFields(), TestSessionInfoFeatureStatusesJSONRoundtrip(), TestSessionInfoRecoveryJSONRoundtrip()

### Community 24 - "Community 24"
Cohesion: 0.25
Nodes (7): Adding a New Wire Type, AgentD Protocol Repo, Architecture, Build & Test, Critical Rules, Full Specification, What This Is

### Community 25 - "Community 25"
Cohesion: 0.43
Nodes (6): NewTraceID(), T, TestNewTraceID(), TestNewTraceID_Uniqueness(), TestValidTraceID(), ValidTraceID()

### Community 26 - "Community 26"
Cohesion: 0.33
Nodes (6): GitDiffRequest, GitDiffResponse, GitFileStatus, GitNotAvailablePayload, GitStatusPayload, GitStatusRequest

### Community 27 - "Community 27"
Cohesion: 0.53
Nodes (5): T, TestApprovalDecisionConstants(), TestApprovalResolvedPayload_Roundtrip(), TestMsgApprovalResolvedConstant(), TestMsgPendingApprovalStateConstant()

### Community 28 - "Community 28"
Cohesion: 0.70
Nodes (4): AgentExtensionDef, AgentExtensionInvocation, AgentExtensionKind, AgentExtensionScope

### Community 29 - "Community 29"
Cohesion: 0.40
Nodes (4): HistoryPagePayload, HistoryPageRequest, SessionHeadPayload, RawMessage

### Community 30 - "Community 30"
Cohesion: 0.70
Nodes (4): SessionRecoveryAction, SessionRecoveryInfo, SessionRecoveryReason, StatusPayload

### Community 31 - "Community 31"
Cohesion: 0.60
Nodes (4): T, TestGitSync_ErrorCodes_NonEmpty(), TestGitSync_MsgConstants_NonEmpty(), TestGitSync_RoundTrip()

### Community 32 - "Community 32"
Cohesion: 0.60
Nodes (4): T, TestProviderCapabilityContractJSONKeys(), TestProviderCapabilityContractRoundTrip(), TestSessionInfoProviderContractField()

### Community 33 - "Community 33"
Cohesion: 0.60
Nodes (4): T, TestWorktrees_ErrorCodes_NonEmpty(), TestWorktrees_MsgConstants_NonEmpty(), TestWorktrees_RoundTrip()

### Community 34 - "Community 34"
Cohesion: 0.50
Nodes (3): PolicyJSON, PolicyMatchJSON, PolicyMatchJSON

### Community 35 - "Community 35"
Cohesion: 0.67
Nodes (3): T, TestCostPayloadAdditiveFieldsRoundTrip(), TestCostPayloadBackwardCompatibleRoundTrip()

### Community 36 - "Community 36"
Cohesion: 0.67
Nodes (3): T, TestInteractivePromptResolvedPayloadRoundtrip(), TestMsgInteractivePromptResolvedConstant()

## Knowledge Gaps
- **239 isolated node(s):** `ApprovalResolvedPayload`, `AgentCapability`, `RawMessage`, `RegisterPayload`, `JoinPayload` (+234 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NormalizeSupportBundleParams()` connect `Community 5` to `Community 0`?**
  _High betweenness centrality (0.045) - this node is a cross-community bridge._
- **Why does `TestControlMessageEdgeCases()` connect `Community 2` to `Community 10`?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._
- **What connects `ApprovalResolvedPayload`, `AgentCapability`, `RawMessage` to the rest of the system?**
  _239 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.06442058496853018 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.07142857142857142 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.12307692307692308 - nodes in this community are weakly interconnected._
- **Should `Community 3` be split into smaller, more focused modules?**
  _Cohesion score 0.05263157894736842 - nodes in this community are weakly interconnected._