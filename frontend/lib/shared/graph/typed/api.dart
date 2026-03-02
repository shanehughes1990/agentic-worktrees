import 'dart:async';

import 'package:agentic_worktrees/shared/graph/generated/operations/control_plane.graphql.dart'
    as gql_ops;
import 'package:agentic_worktrees/shared/graph/generated/schema/control_plane.graphql.dart'
    as gql_cp;
import 'package:agentic_worktrees/shared/graph/generated/schema/scm.graphql.dart'
    as gql_scm;
import 'package:agentic_worktrees/shared/graph/typed/models.dart';
import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:graphql/client.dart';

class ControlPlaneApi {
  ControlPlaneApi(this._client);

  final GraphQLClient _client;

  Future<ApiResult<List<SessionSummary>>> sessions({int limit = 50}) async {
    final result = await _client.query$Sessions(
      gql_ops.Options$Query$Sessions(
        variables: gql_ops.Variables$Query$Sessions(limit: limit),
      ),
    );
    final error = _extractOperationError(result, field: 'sessions');
    if (error != null) {
      return ApiResult<List<SessionSummary>>.failure(error);
    }
    final payload = result.parsedData?.sessions;
    if (payload == null) {
      return const ApiResult<List<SessionSummary>>.failure(
        'sessions returned no data',
      );
    }
    if (payload is gql_ops.Query$Sessions$sessions$$GraphError) {
      return ApiResult<List<SessionSummary>>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload is! gql_ops.Query$Sessions$sessions$$SessionsSuccess) {
      return const ApiResult<List<SessionSummary>>.failure(
        'sessions returned unexpected payload',
      );
    }
    final items = payload.sessions
        .map(
          (entry) => SessionSummary(
            runID: entry.runID,
            taskCount: entry.taskCount,
            jobCount: entry.jobCount,
            updatedAt: entry.updatedAt.toLocal(),
          ),
        )
        .toList(growable: false);
    return ApiResult<List<SessionSummary>>.success(items);
  }

  Future<ApiResult<List<WorkflowJob>>> workflowJobs({
    required String runID,
    String? taskID,
    int limit = 100,
  }) async {
    final result = await _client.query$WorkflowJobs(
      gql_ops.Options$Query$WorkflowJobs(
        variables: gql_ops.Variables$Query$WorkflowJobs(
          runID: runID,
          taskID: taskID,
          limit: limit,
        ),
      ),
    );
    final error = _extractOperationError(result, field: 'workflowJobs');
    if (error != null) {
      return ApiResult<List<WorkflowJob>>.failure(error);
    }
    final payload = result.parsedData?.workflowJobs;
    if (payload == null) {
      return const ApiResult<List<WorkflowJob>>.failure(
        'workflowJobs returned no data',
      );
    }
    if (payload is gql_ops.Query$WorkflowJobs$workflowJobs$$GraphError) {
      return ApiResult<List<WorkflowJob>>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess) {
      return const ApiResult<List<WorkflowJob>>.failure(
        'workflowJobs returned unexpected payload',
      );
    }
    final items = payload.jobs
        .map(
          (entry) => WorkflowJob(
            runID: entry.runID,
            taskID: entry.taskID,
            jobID: entry.jobID,
            jobKind: entry.jobKind.toJson(),
            status: entry.status,
            queue: entry.queue,
            queueTaskID: entry.queueTaskID,
            duplicate: entry.duplicate,
            updatedAt: entry.updatedAt.toLocal(),
          ),
        )
        .toList(growable: false);
    return ApiResult<List<WorkflowJob>>.success(items);
  }

  Future<ApiResult<List<SupervisorDecision>>> supervisorHistory({
    required String runID,
    required String taskID,
    required String jobID,
  }) async {
    final result = await _client.query$SupervisorDecisionHistory(
      gql_ops.Options$Query$SupervisorDecisionHistory(
        variables: gql_ops.Variables$Query$SupervisorDecisionHistory(
          runID: runID,
          taskID: taskID,
          jobID: jobID,
        ),
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'supervisorDecisionHistory',
    );
    if (error != null) {
      return ApiResult<List<SupervisorDecision>>.failure(error);
    }
    final payload = result.parsedData?.supervisorDecisionHistory;
    if (payload == null) {
      return const ApiResult<List<SupervisorDecision>>.failure(
        'supervisorDecisionHistory returned no data',
      );
    }
    if (payload
        is gql_ops.Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError) {
      return ApiResult<List<SupervisorDecision>>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess) {
      return const ApiResult<List<SupervisorDecision>>.failure(
        'supervisorDecisionHistory returned unexpected payload',
      );
    }
    final items = payload.decisions
        .map(
          (entry) => SupervisorDecision(
            signalType: entry.signalType.toJson(),
            action: entry.action.toJson(),
            reason: entry.reason.toJson(),
            occurredAt: entry.occurredAt.toLocal(),
          ),
        )
        .toList(growable: false);
    return ApiResult<List<SupervisorDecision>>.success(items);
  }

  Future<ApiResult<String>> enqueueScmWorkflow({
    required String runID,
    required String taskID,
    required String jobID,
    required String idempotencyKey,
    required String owner,
    required String repository,
  }) async {
    final result = await _client.mutate$EnqueueScmWorkflow(
      gql_ops.Options$Mutation$EnqueueScmWorkflow(
        variables: gql_ops.Variables$Mutation$EnqueueScmWorkflow(
          input: gql_scm.Input$EnqueueSCMWorkflowInput(
            operation: gql_scm.Enum$SCMOperation.SOURCE_STATE,
            provider: gql_scm.Enum$SCMProvider.GITHUB,
            owner: owner,
            repository: repository,
            runID: runID,
            taskID: taskID,
            jobID: jobID,
            idempotencyKey: idempotencyKey,
          ),
        ),
      ),
    );
    final error = _extractOperationError(result, field: 'enqueueScmWorkflow');
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final payload = result.parsedData?.enqueueScmWorkflow;
    if (payload == null) {
      return const ApiResult<String>.failure(
        'enqueueScmWorkflow returned no data',
      );
    }
    if (payload
        is gql_ops.Mutation$EnqueueScmWorkflow$enqueueScmWorkflow$$GraphError) {
      return ApiResult<String>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Mutation$EnqueueScmWorkflow$enqueueScmWorkflow$$EnqueueSCMWorkflowSuccess) {
      return const ApiResult<String>.failure(
        'enqueueScmWorkflow returned unexpected payload',
      );
    }
    return ApiResult<String>.success(payload.queueTaskID);
  }

  Future<ApiResult<String>> enqueueIngestionWorkflow({
    required String runID,
    required String taskID,
    required String jobID,
    required String idempotencyKey,
    required String prompt,
    required String projectID,
    required String workflowID,
    required String source,
  }) async {
    final boardID = '${projectID.isEmpty ? 'project' : projectID}-board';
    final result = await _client.mutate$EnqueueIngestionWorkflow(
      gql_ops.Options$Mutation$EnqueueIngestionWorkflow(
        variables: gql_ops.Variables$Mutation$EnqueueIngestionWorkflow(
          input: gql_cp.Input$EnqueueIngestionWorkflowInput(
            runID: runID,
            taskID: taskID,
            jobID: jobID,
            idempotencyKey: idempotencyKey,
            prompt: prompt,
            projectID: projectID,
            workflowID: workflowID,
            boardSources: <gql_cp.Input$IngestionBoardSourceInput>[
              gql_cp.Input$IngestionBoardSourceInput(
                boardID: boardID,
                kind: gql_cp.Enum$TrackerSourceKind.GITHUB_ISSUES,
                location: source,
                appliesToAllRepositories: true,
              ),
            ],
          ),
        ),
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'enqueueIngestionWorkflow',
    );
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final payload = result.parsedData?.enqueueIngestionWorkflow;
    if (payload == null) {
      return const ApiResult<String>.failure(
        'enqueueIngestionWorkflow returned no data',
      );
    }
    if (payload
        is gql_ops.Mutation$EnqueueIngestionWorkflow$enqueueIngestionWorkflow$$GraphError) {
      return ApiResult<String>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Mutation$EnqueueIngestionWorkflow$enqueueIngestionWorkflow$$EnqueueIngestionWorkflowSuccess) {
      return const ApiResult<String>.failure(
        'enqueueIngestionWorkflow returned unexpected payload',
      );
    }
    return ApiResult<String>.success(payload.queueTaskID);
  }

  Future<ApiResult<String>> approveIssueIntake({
    required String runID,
    required String taskID,
    required String jobID,
    required String projectID,
    required String source,
    required String issueReference,
    required String approvedBy,
  }) async {
    final result = await _client.mutate$ApproveIssueIntake(
      gql_ops.Options$Mutation$ApproveIssueIntake(
        variables: gql_ops.Variables$Mutation$ApproveIssueIntake(
          input: gql_cp.Input$ApproveIssueIntakeInput(
            runID: runID,
            taskID: taskID,
            jobID: jobID,
            projectID: projectID,
            source: source,
            issueReference: issueReference,
            approvedBy: approvedBy,
          ),
        ),
      ),
    );
    final error = _extractOperationError(result, field: 'approveIssueIntake');
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final payload = result.parsedData?.approveIssueIntake;
    if (payload == null) {
      return const ApiResult<String>.failure(
        'approveIssueIntake returned no data',
      );
    }
    if (payload
        is gql_ops.Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError) {
      return ApiResult<String>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess) {
      return const ApiResult<String>.failure(
        'approveIssueIntake returned unexpected payload',
      );
    }
    final decision = payload.decision;
    return ApiResult<String>.success(
      '${decision.action.toJson()} (${decision.reason.toJson()})',
    );
  }

  Future<ApiResult<List<ProjectSetupConfig>>> projectSetups({
    int limit = 50,
  }) async {
    final result = await _client.query$ProjectSetups(
      gql_ops.Options$Query$ProjectSetups(
        variables: gql_ops.Variables$Query$ProjectSetups(limit: limit),
      ),
    );
    final error = _extractOperationError(result, field: 'projectSetups');
    if (error != null) {
      return ApiResult<List<ProjectSetupConfig>>.failure(error);
    }
    final payload = result.parsedData?.projectSetups;
    if (payload == null) {
      return const ApiResult<List<ProjectSetupConfig>>.failure(
        'projectSetups returned no data',
      );
    }
    if (payload is gql_ops.Query$ProjectSetups$projectSetups$$GraphError) {
      return ApiResult<List<ProjectSetupConfig>>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess) {
      return const ApiResult<List<ProjectSetupConfig>>.failure(
        'projectSetups returned unexpected payload',
      );
    }
    final projects = payload.projects
        .map(
          (entry) => ProjectSetupConfig(
            projectID: entry.projectID,
            projectName: entry.projectName,
            repositories: entry.repositories
                .map(
                  (repository) => ProjectRepositoryConfig(
                    repositoryID: repository.repositoryID,
                    scmProvider: repository.scmProvider.toJson(),
                    repositoryURL: repository.repositoryURL,
                    isPrimary: repository.isPrimary,
                  ),
                )
                .toList(growable: false),
            boards: entry.boards
                .map(
                  (board) => ProjectBoardConfig(
                    boardID: board.boardID,
                    trackerProvider: board.trackerProvider.toJson(),
                    taskboardName: board.taskboardName ?? '',
                    appliesToAllRepositories: board.appliesToAllRepositories,
                    repositoryIDs: board.repositoryIDs.toList(growable: false),
                  ),
                )
                .toList(growable: false),
            createdAt: entry.createdAt.toLocal(),
            updatedAt: entry.updatedAt.toLocal(),
          ),
        )
        .toList(growable: false);
    return ApiResult<List<ProjectSetupConfig>>.success(projects);
  }

  Future<ApiResult<ProjectSetupConfig>> upsertProjectSetup({
    required String projectID,
    required String projectName,
    required String scmProvider,
    required List<String> repositoryURLs,
    required String trackerProvider,
    required String taskboardName,
  }) async {
    final repositories = repositoryURLs
        .map((String repositoryURL) => repositoryURL.trim())
        .where((String repositoryURL) => repositoryURL.isNotEmpty)
        .toList(growable: false);
    final normalizedTaskboardName = taskboardName.trim();
    final projectScmProvider = _toProjectScmProvider(scmProvider);
    if (projectScmProvider == null) {
      return ApiResult<ProjectSetupConfig>.failure(
        _graphErrorMessageTyped(
          code: 'VALIDATION',
          message: 'unsupported scm provider',
          field: 'scmProvider',
        ),
      );
    }
    final projectTrackerProvider = _toTrackerSourceKind(trackerProvider);
    if (projectTrackerProvider == null) {
      return ApiResult<ProjectSetupConfig>.failure(
        _graphErrorMessageTyped(
          code: 'VALIDATION',
          message: 'unsupported tracker provider',
          field: 'trackerProvider',
        ),
      );
    }
    if (normalizedTaskboardName.isEmpty) {
      return ApiResult<ProjectSetupConfig>.failure(
        _graphErrorMessageTyped(
          code: 'VALIDATION',
          message: 'taskboard name is required',
          field: 'taskboardName',
        ),
      );
    }
    final result = await _client.mutate$UpsertProjectSetup(
      gql_ops.Options$Mutation$UpsertProjectSetup(
        variables: gql_ops.Variables$Mutation$UpsertProjectSetup(
          input: gql_cp.Input$UpsertProjectSetupInput(
            projectID: projectID,
            projectName: projectName,
            repositories: repositories
                .asMap()
                .entries
                .map(
                  (entry) => gql_cp.Input$ProjectRepositoryInput(
                    repositoryID:
                        '${projectID.isEmpty ? 'project' : projectID}-repo-${entry.key + 1}',
                    scmProvider: projectScmProvider,
                    repositoryURL: entry.value,
                    isPrimary: entry.key == 0,
                  ),
                )
                .toList(growable: false),
            boards: <gql_cp.Input$ProjectBoardInput>[
              gql_cp.Input$ProjectBoardInput(
                trackerProvider: projectTrackerProvider,
                taskboardName: normalizedTaskboardName,
                appliesToAllRepositories: true,
                repositoryIDs: const <String>[],
              ),
            ],
          ),
        ),
      ),
    );
    final error = _extractOperationError(result, field: 'upsertProjectSetup');
    if (error != null) {
      return ApiResult<ProjectSetupConfig>.failure(error);
    }
    final payload = result.parsedData?.upsertProjectSetup;
    if (payload == null) {
      return const ApiResult<ProjectSetupConfig>.failure(
        'upsertProjectSetup returned no data',
      );
    }
    if (payload
        is gql_ops.Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError) {
      return ApiResult<ProjectSetupConfig>.failure(
        _graphErrorMessageTyped(
          code: payload.code.toJson(),
          message: payload.message,
          field: payload.field,
        ),
      );
    }
    if (payload
        is! gql_ops.Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess) {
      return const ApiResult<ProjectSetupConfig>.failure(
        'upsertProjectSetup project payload missing',
      );
    }
    final project = payload.project;
    return ApiResult<ProjectSetupConfig>.success(
      ProjectSetupConfig(
        projectID: project.projectID,
        projectName: project.projectName,
        repositories: project.repositories
            .map(
              (repository) => ProjectRepositoryConfig(
                repositoryID: repository.repositoryID,
                scmProvider: repository.scmProvider.toJson(),
                repositoryURL: repository.repositoryURL,
                isPrimary: repository.isPrimary,
              ),
            )
            .toList(growable: false),
        boards: project.boards
            .map(
              (board) => ProjectBoardConfig(
                boardID: board.boardID,
                trackerProvider: board.trackerProvider.toJson(),
                taskboardName: board.taskboardName ?? '',
                appliesToAllRepositories: board.appliesToAllRepositories,
                repositoryIDs: board.repositoryIDs.toList(growable: false),
              ),
            )
            .toList(growable: false),
        createdAt: project.createdAt.toLocal(),
        updatedAt: project.updatedAt.toLocal(),
      ),
    );
  }

  Stream<ApiResult<StreamEvent>> sessionActivityStream({
    required String runID,
    int fromOffset = 0,
  }) {
    return _client
        .subscribe$SessionActivity(
          gql_ops.Options$Subscription$SessionActivity(
            variables: gql_ops.Variables$Subscription$SessionActivity(
              runID: runID,
              taskID: '',
              jobID: '',
              fromOffset: fromOffset,
            ),
          ),
        )
        .map((result) {
          final error = _extractOperationError(
            result,
            field: 'sessionActivityStream',
          );
          if (error != null) {
            return ApiResult<StreamEvent>.failure(error);
          }
          final payload = result.parsedData?.sessionActivityStream;
          if (payload == null) {
            return const ApiResult<StreamEvent>.failure(
              'sessionActivityStream returned no data',
            );
          }
          if (payload
              is gql_ops.Subscription$SessionActivity$sessionActivityStream$$GraphError) {
            return ApiResult<StreamEvent>.failure(
              _graphErrorMessageTyped(
                code: payload.code.toJson(),
                message: payload.message,
                field: payload.field,
              ),
            );
          }
          if (payload
              is! gql_ops.Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess) {
            return const ApiResult<StreamEvent>.failure(
              'stream event payload missing',
            );
          }
          final eventData = payload.event;
          final event = StreamEvent(
            eventID: eventData.eventID,
            eventType: eventData.eventType,
            source: eventData.source.toJson(),
            payload: eventData.payload,
            occurredAt: eventData.occurredAt.toLocal(),
          );
          return ApiResult<StreamEvent>.success(event);
        })
        .asBroadcastStream();
  }

  Future<ApiResult<List<WorkerSession>>> workerSessions({
    int limit = 100,
  }) async {
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query WorkerSessions(
            \$limit: Int!
          ) {
            workerSessions(limit: \$limit) {
              __typename
              ... on WorkerSessionsSuccess {
                sessions {
                  workerID
                  epoch
                  state
                  desiredState
                  lastHeartbeat
                  leaseExpiresAt
                  rogueReason
                  updatedAt
                }
              }
              ... on GraphError {
                code
                message
                field
              }
            }
          }
        '''),
        variables: <String, dynamic>{'limit': limit},
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'workerSessions');
    if (error != null) {
      return ApiResult<List<WorkerSession>>.failure(error);
    }
    final payload = result.data?['workerSessions'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<WorkerSession>>.failure(
        'workerSessions returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<WorkerSession>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final sessions =
        (payload['sessions'] as List<dynamic>? ?? const <dynamic>[])
            .whereType<Map<String, dynamic>>()
            .map(
              (Map<String, dynamic> item) => WorkerSession(
                workerID: item['workerID'] as String,
                epoch: item['epoch'] as int,
                state: item['state'] as String,
                desiredState: item['desiredState'] as String,
                lastHeartbeat: DateTime.parse(
                  item['lastHeartbeat'] as String,
                ).toLocal(),
                leaseExpiresAt: DateTime.parse(
                  item['leaseExpiresAt'] as String,
                ).toLocal(),
                rogueReason: item['rogueReason'] as String?,
                updatedAt: DateTime.parse(
                  item['updatedAt'] as String,
                ).toLocal(),
              ),
            )
            .toList(growable: false);
    return ApiResult<List<WorkerSession>>.success(sessions);
  }

  Future<ApiResult<WorkerSettings>> workerSettings() async {
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query WorkerSettings {
            workerSettings {
              __typename
              ... on WorkerSettingsSuccess {
                settings {
                  heartbeatIntervalSeconds
                  responseDeadlineSeconds
                  staleAfterSeconds
                  drainTimeoutSeconds
                  terminateTimeoutSeconds
                  rogueThreshold
                  updatedAt
                }
              }
              ... on GraphError {
                code
                message
                field
              }
            }
          }
        '''),
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'workerSettings');
    if (error != null) {
      return ApiResult<WorkerSettings>.failure(error);
    }
    final payload = result.data?['workerSettings'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<WorkerSettings>.failure(
        'workerSettings returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<WorkerSettings>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final settings = payload['settings'] as Map<String, dynamic>?;
    if (settings == null) {
      return const ApiResult<WorkerSettings>.failure(
        'workerSettings payload missing',
      );
    }
    return ApiResult<WorkerSettings>.success(
      WorkerSettings(
        heartbeatIntervalSeconds: settings['heartbeatIntervalSeconds'] as int,
        responseDeadlineSeconds: settings['responseDeadlineSeconds'] as int,
        staleAfterSeconds: settings['staleAfterSeconds'] as int,
        drainTimeoutSeconds: settings['drainTimeoutSeconds'] as int,
        terminateTimeoutSeconds: settings['terminateTimeoutSeconds'] as int,
        rogueThreshold: settings['rogueThreshold'] as int,
        updatedAt: DateTime.parse(settings['updatedAt'] as String).toLocal(),
      ),
    );
  }

  Future<ApiResult<WorkerSettings>> updateWorkerSettings({
    required int heartbeatIntervalSeconds,
    required int responseDeadlineSeconds,
    required int staleAfterSeconds,
    required int drainTimeoutSeconds,
    required int terminateTimeoutSeconds,
    required int rogueThreshold,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
          mutation UpdateWorkerSettings(
            \$input: UpdateWorkerSettingsInput!
          ) {
            updateWorkerSettings(input: \$input) {
              __typename
              ... on WorkerSettingsSuccess {
                settings {
                  heartbeatIntervalSeconds
                  responseDeadlineSeconds
                  staleAfterSeconds
                  drainTimeoutSeconds
                  terminateTimeoutSeconds
                  rogueThreshold
                  updatedAt
                }
              }
              ... on GraphError {
                code
                message
                field
              }
            }
          }
        '''),
        variables: <String, dynamic>{
          'input': <String, dynamic>{
            'heartbeatIntervalSeconds': heartbeatIntervalSeconds,
            'responseDeadlineSeconds': responseDeadlineSeconds,
            'staleAfterSeconds': staleAfterSeconds,
            'drainTimeoutSeconds': drainTimeoutSeconds,
            'terminateTimeoutSeconds': terminateTimeoutSeconds,
            'rogueThreshold': rogueThreshold,
          },
        },
      ),
    );
    final error = _extractOperationError(result, field: 'updateWorkerSettings');
    if (error != null) {
      return ApiResult<WorkerSettings>.failure(error);
    }
    final payload =
        result.data?['updateWorkerSettings'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<WorkerSettings>.failure(
        'updateWorkerSettings returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<WorkerSettings>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final settings = payload['settings'] as Map<String, dynamic>?;
    if (settings == null) {
      return const ApiResult<WorkerSettings>.failure(
        'updateWorkerSettings payload missing',
      );
    }
    return ApiResult<WorkerSettings>.success(
      WorkerSettings(
        heartbeatIntervalSeconds: settings['heartbeatIntervalSeconds'] as int,
        responseDeadlineSeconds: settings['responseDeadlineSeconds'] as int,
        staleAfterSeconds: settings['staleAfterSeconds'] as int,
        drainTimeoutSeconds: settings['drainTimeoutSeconds'] as int,
        terminateTimeoutSeconds: settings['terminateTimeoutSeconds'] as int,
        rogueThreshold: settings['rogueThreshold'] as int,
        updatedAt: DateTime.parse(settings['updatedAt'] as String).toLocal(),
      ),
    );
  }

  Stream<ApiResult<StreamEvent>> workerSessionStream({int fromOffset = 0}) {
    return _client
        .subscribe(
          SubscriptionOptions(
            document: gql('''
              subscription WorkerSessionStream(
                \$runID: String!
                \$taskID: String!
                \$jobID: String!
                \$fromOffset: Int!
              ) {
                workerSessionStream(
                  correlation: {runID: \$runID, taskID: \$taskID, jobID: \$jobID}
                  fromOffset: \$fromOffset
                ) {
                  __typename
                  ... on StreamEventSuccess {
                    event {
                      eventID
                      eventType
                      source
                      payload
                      occurredAt
                    }
                  }
                  ... on GraphError {
                    code
                    message
                    field
                  }
                }
              }
            '''),
            variables: <String, dynamic>{
              'runID': '',
              'taskID': '',
              'jobID': '',
              'fromOffset': fromOffset,
            },
          ),
        )
        .map((QueryResult result) {
          final error = _extractOperationError(
            result,
            field: 'workerSessionStream',
          );
          if (error != null) {
            return ApiResult<StreamEvent>.failure(error);
          }
          final payload =
              result.data?['workerSessionStream'] as Map<String, dynamic>?;
          if (payload == null) {
            return const ApiResult<StreamEvent>.failure(
              'workerSessionStream returned no data',
            );
          }
          if (payload['__typename'] == 'GraphError') {
            return ApiResult<StreamEvent>.failure(
              _graphErrorMessageTyped(
                code: payload['code'] as String? ?? 'INTERNAL',
                message: payload['message'] as String? ?? 'unknown error',
                field: payload['field'] as String?,
              ),
            );
          }
          final eventData = payload['event'] as Map<String, dynamic>?;
          if (eventData == null) {
            return const ApiResult<StreamEvent>.failure(
              'workerSessionStream event payload missing',
            );
          }
          return ApiResult<StreamEvent>.success(
            StreamEvent(
              eventID: eventData['eventID'] as String,
              eventType: eventData['eventType'] as String,
              source: eventData['source'] as String,
              payload: eventData['payload'] as String,
              occurredAt: DateTime.parse(
                eventData['occurredAt'] as String,
              ).toLocal(),
            ),
          );
        })
        .asBroadcastStream();
  }

  String? _extractOperationError(QueryResult result, {required String field}) {
    if (result.hasException) {
      AppLogger.instance.logger.e(
        'GraphQL operation exception',
        error: {'field': field, 'exception': result.exception.toString()},
      );
      return result.exception.toString();
    }
    if (result.data == null) {
      AppLogger.instance.logger.w(
        'GraphQL operation returned no payload',
        error: {'field': field},
      );
      return '$field returned no response payload';
    }
    return null;
  }

  gql_scm.Enum$SCMProvider? _toProjectScmProvider(String value) {
    switch (value.toUpperCase()) {
      case 'GITHUB':
        return gql_scm.Enum$SCMProvider.GITHUB;
      default:
        return null;
    }
  }

  gql_cp.Enum$TrackerSourceKind? _toTrackerSourceKind(String value) {
    switch (value.toUpperCase()) {
      case 'INTERNAL':
        return gql_cp.Enum$TrackerSourceKind.INTERNAL;
      case 'GITHUB_ISSUES':
        return gql_cp.Enum$TrackerSourceKind.GITHUB_ISSUES;
      default:
        return null;
    }
  }

  String _graphErrorMessageTyped({
    required String code,
    required String message,
    String? field,
  }) {
    if (field == null || field.isEmpty) {
      return '$code: $message';
    }
    return '$code ($field): $message';
  }
}
