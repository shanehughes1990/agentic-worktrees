import 'dart:async';
import 'dart:typed_data';

import 'package:agentic_repositories/shared/graph/generated/operations/control_plane.graphql.dart'
    as gql_ops;
import 'package:agentic_repositories/shared/graph/generated/schema/control_plane.graphql.dart'
    as gql_cp;
import 'package:agentic_repositories/shared/graph/generated/schema/scm.graphql.dart'
    as gql_scm;
import 'package:agentic_repositories/shared/graph/typed/models.dart';
import 'package:agentic_repositories/shared/logging/app_logger.dart';
import 'package:graphql/client.dart';
import 'package:http/http.dart' as http;

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
            scms: entry.scms
                .map(
                  (scm) => ProjectScmConfig(
                    scmID: scm.scmID,
                    scmProvider: scm.scmProvider.toJson(),
                  ),
                )
                .toList(growable: false),
            repositories: entry.repositories
                .map(
                  (repository) => ProjectRepositoryConfig(
                    repositoryID: repository.repositoryID,
                    scmID: repository.scmID,
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

  Future<ApiResult<ProjectSetupConfig>> projectSetup({
    required String projectID,
  }) async {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return const ApiResult<ProjectSetupConfig>.failure(
        'project_id is required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query ProjectSetup(
            \$projectID: String!
          ) {
            projectSetup(projectID: \$projectID) {
              __typename
              ... on ProjectSetupSuccess {
                project {
                  projectID
                  projectName
                  scms {
                    scmID
                    scmProvider
                  }
                  repositories {
                    repositoryID
                    scmID
                    repositoryURL
                    isPrimary
                  }
                  boards {
                    boardID
                    trackerProvider
                    taskboardName
                    appliesToAllRepositories
                    repositoryIDs
                  }
                  createdAt
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
        variables: <String, dynamic>{'projectID': cleanProjectID},
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'projectSetup');
    if (error != null) {
      return ApiResult<ProjectSetupConfig>.failure(error);
    }
    final payload = result.data?['projectSetup'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<ProjectSetupConfig>.failure(
        'projectSetup returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<ProjectSetupConfig>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final project = payload['project'] as Map<String, dynamic>?;
    if (project == null) {
      return const ApiResult<ProjectSetupConfig>.failure(
        'projectSetup project payload missing',
      );
    }
    return ApiResult<ProjectSetupConfig>.success(
      ProjectSetupConfig(
        projectID: project['projectID'] as String,
        projectName: project['projectName'] as String,
        scms: (project['scms'] as List<dynamic>? ?? const <dynamic>[])
            .whereType<Map<String, dynamic>>()
            .map(
              (Map<String, dynamic> scm) => ProjectScmConfig(
                scmID: scm['scmID'] as String,
                scmProvider: scm['scmProvider'] as String,
              ),
            )
            .toList(growable: false),
        repositories:
            (project['repositories'] as List<dynamic>? ?? const <dynamic>[])
                .whereType<Map<String, dynamic>>()
                .map(
                  (Map<String, dynamic> repository) => ProjectRepositoryConfig(
                    repositoryID: repository['repositoryID'] as String,
                    scmID: repository['scmID'] as String,
                    repositoryURL: repository['repositoryURL'] as String,
                    isPrimary: repository['isPrimary'] as bool,
                  ),
                )
                .toList(growable: false),
        boards: (project['boards'] as List<dynamic>? ?? const <dynamic>[])
            .whereType<Map<String, dynamic>>()
            .map(
              (Map<String, dynamic> board) => ProjectBoardConfig(
                boardID: board['boardID'] as String,
                trackerProvider: board['trackerProvider'] as String,
                taskboardName: (board['taskboardName'] as String?) ?? '',
                appliesToAllRepositories:
                    board['appliesToAllRepositories'] as bool,
                repositoryIDs:
                    (board['repositoryIDs'] as List<dynamic>? ??
                            const <dynamic>[])
                        .whereType<String>()
                        .toList(growable: false),
              ),
            )
            .toList(growable: false),
        createdAt: DateTime.parse(project['createdAt'] as String).toLocal(),
        updatedAt: DateTime.parse(project['updatedAt'] as String).toLocal(),
      ),
    );
  }

  Future<ApiResult<ProjectSetupConfig>> upsertProjectSetup({
    required String projectID,
    required String projectName,
    required String scmProvider,
    required List<String> repositoryURLs,
    required String scmToken,
  }) async {
    final repositories = repositoryURLs
        .map((String repositoryURL) => repositoryURL.trim())
        .where((String repositoryURL) => repositoryURL.isNotEmpty)
        .toList(growable: false);
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
    final projectScmID = '${projectID.isEmpty ? 'project' : projectID}-scm-1';
    final result = await _client.mutate$UpsertProjectSetup(
      gql_ops.Options$Mutation$UpsertProjectSetup(
        variables: gql_ops.Variables$Mutation$UpsertProjectSetup(
          input: gql_cp.Input$UpsertProjectSetupInput(
            projectID: projectID,
            projectName: projectName,
            scms: <gql_cp.Input$ProjectSCMInput>[
              gql_cp.Input$ProjectSCMInput(
                scmID: projectScmID,
                scmProvider: projectScmProvider,
                scmToken: scmToken.trim(),
              ),
            ],
            repositories: repositories
                .asMap()
                .entries
                .map(
                  (entry) => gql_cp.Input$ProjectRepositoryInput(
                    repositoryID:
                        '${projectID.isEmpty ? 'project' : projectID}-repo-${entry.key + 1}',
                    scmID: projectScmID,
                    repositoryURL: entry.value,
                    isPrimary: entry.key == 0,
                  ),
                )
                .toList(growable: false),
            boards: const <gql_cp.Input$ProjectBoardInput>[],
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
        scms: project.scms
            .map(
              (scm) => ProjectScmConfig(
                scmID: scm.scmID,
                scmProvider: scm.scmProvider.toJson(),
              ),
            )
            .toList(growable: false),
        repositories: project.repositories
            .map(
              (repository) => ProjectRepositoryConfig(
                repositoryID: repository.repositoryID,
                scmID: repository.scmID,
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

  Future<ApiResult<List<ProjectDocument>>> projectDocuments({
    required String projectID,
    int limit = 100,
  }) async {
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query ProjectDocuments(
            \$projectID: String!
            \$limit: Int!
          ) {
            projectDocuments(projectID: \$projectID, limit: \$limit) {
              __typename
              ... on ProjectDocumentsSuccess {
                documents {
                  projectID
                  documentID
                  fileName
                  contentType
                  objectPath
                  cdnURL
                  status
                  createdAt
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
        variables: <String, dynamic>{'projectID': projectID, 'limit': limit},
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'projectDocuments');
    if (error != null) {
      return ApiResult<List<ProjectDocument>>.failure(error);
    }
    final payload = result.data?['projectDocuments'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<ProjectDocument>>.failure(
        'projectDocuments returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<ProjectDocument>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final documents =
        (payload['documents'] as List<dynamic>? ?? const <dynamic>[])
            .whereType<Map<String, dynamic>>()
            .map(
              (Map<String, dynamic> item) => ProjectDocument(
                projectID: item['projectID'] as String,
                documentID: item['documentID'] as String,
                fileName: item['fileName'] as String,
                contentType: item['contentType'] as String,
                objectPath: item['objectPath'] as String,
                cdnURL: item['cdnURL'] as String,
                status: item['status'] as String,
                createdAt: DateTime.parse(
                  item['createdAt'] as String,
                ).toLocal(),
                updatedAt: DateTime.parse(
                  item['updatedAt'] as String,
                ).toLocal(),
              ),
            )
            .toList(growable: false);
    return ApiResult<List<ProjectDocument>>.success(documents);
  }

  Future<ApiResult<ProjectDocumentUploadTicket>> requestProjectDocumentUpload({
    required String projectID,
    required String fileName,
    required String contentType,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
          mutation RequestProjectDocumentUpload(
            \$input: RequestProjectDocumentUploadInput!
          ) {
            requestProjectDocumentUpload(input: \$input) {
              __typename
              ... on RequestProjectDocumentUploadSuccess {
                requestID
                projectID
                documentID
                fileName
                contentType
                objectPath
                uploadURL
                cdnURL
                expiresAt
                status
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
            'projectID': projectID,
            'fileName': fileName,
            'contentType': contentType,
          },
        },
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'requestProjectDocumentUpload',
    );
    if (error != null) {
      return ApiResult<ProjectDocumentUploadTicket>.failure(error);
    }
    final payload =
        result.data?['requestProjectDocumentUpload'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<ProjectDocumentUploadTicket>.failure(
        'requestProjectDocumentUpload returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<ProjectDocumentUploadTicket>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    return ApiResult<ProjectDocumentUploadTicket>.success(
      ProjectDocumentUploadTicket(
        requestID: payload['requestID'] as String,
        projectID: payload['projectID'] as String,
        documentID: payload['documentID'] as String,
        fileName: payload['fileName'] as String,
        contentType: payload['contentType'] as String,
        objectPath: payload['objectPath'] as String,
        uploadURL: payload['uploadURL'] as String,
        cdnURL: payload['cdnURL'] as String,
        expiresAt: DateTime.parse(payload['expiresAt'] as String).toLocal(),
        status: payload['status'] as String,
      ),
    );
  }

  Future<ApiResult<void>> uploadProjectDocumentBytes({
    required String uploadURL,
    required Uint8List bytes,
    required String contentType,
  }) async {
    try {
      final uploadUri = Uri.parse(uploadURL);
      final response = await http.put(
        uploadUri,
        headers: <String, String>{'Content-Type': contentType},
        body: bytes,
      );
      if (response.statusCode < 200 || response.statusCode >= 300) {
        final responseBody = response.body.trim();
        final bodySnippet = responseBody.isEmpty
            ? ''
            : ' body=${responseBody.length > 300 ? '${responseBody.substring(0, 300)}...' : responseBody}';
        AppLogger.instance.logger.e(
          'Project document upload failed',
          error: {
            'status': response.statusCode,
            'url_host': uploadUri.host,
            'url_path': uploadUri.path,
            'content_type': contentType,
          },
        );
        return ApiResult<void>.failure(
          'upload failed with status ${response.statusCode} (host=${uploadUri.host})$bodySnippet',
        );
      }
      return const ApiResult<void>.success(null);
    } catch (error) {
      return ApiResult<void>.failure('upload failed: $error');
    }
  }

  Future<ApiResult<void>> deleteProjectDocument({
    required String projectID,
    required String documentID,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
          mutation DeleteProjectDocument(
            \$input: DeleteProjectDocumentInput!
          ) {
            deleteProjectDocument(input: \$input) {
              __typename
              ... on DeleteProjectDocumentSuccess {
                ok
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
            'projectID': projectID,
            'documentID': documentID,
          },
        },
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'deleteProjectDocument',
    );
    if (error != null) {
      return ApiResult<void>.failure(error);
    }
    final payload =
        result.data?['deleteProjectDocument'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<void>.failure(
        'deleteProjectDocument returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<void>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    return const ApiResult<void>.success(null);
  }

  Future<ApiResult<IngestionRunTicket>> runIngestionAgent({
    required String projectID,
    required String taskboardName,
    String? boardID,
    List<String>? selectedDocumentIDs,
    List<String>? selectedDocumentLocations,
    String? userPrompt,
    Map<String, String>? repositorySourceBranches,
  }) async {
    final cleanTaskboardName = taskboardName.trim();
    if (cleanTaskboardName.isEmpty) {
      return const ApiResult<IngestionRunTicket>.failure(
        'taskboard_name is required',
      );
    }
    final input = <String, dynamic>{
      'projectID': projectID,
      'taskboardName': cleanTaskboardName,
    };
    if (boardID != null && boardID.trim().isNotEmpty) {
      input['boardID'] = boardID.trim();
    }
    if (selectedDocumentIDs != null) {
      input['selectedDocumentIDs'] = selectedDocumentIDs;
    }
    if (selectedDocumentLocations != null) {
      input['selectedDocumentLocations'] = selectedDocumentLocations;
    }
    if (userPrompt != null && userPrompt.trim().isNotEmpty) {
      input['userPrompt'] = userPrompt.trim();
    }
    if (repositorySourceBranches != null &&
        repositorySourceBranches.isNotEmpty) {
      final selections = <Map<String, dynamic>>[];
      repositorySourceBranches.forEach((String repositoryID, String branch) {
        final cleanRepositoryID = repositoryID.trim();
        final cleanBranch = branch.trim();
        if (cleanRepositoryID.isEmpty || cleanBranch.isEmpty) {
          return;
        }
        selections.add(<String, dynamic>{
          'repositoryID': cleanRepositoryID,
          'branch': cleanBranch,
        });
      });
      if (selections.isNotEmpty) {
        input['repositorySourceBranches'] = selections;
      }
    }
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
          mutation RunIngestionAgent(
            \$input: RunIngestionAgentInput!
          ) {
            runIngestionAgent(input: \$input) {
              __typename
              ... on RunIngestionAgentSuccess {
                runID
                taskID
                jobID
                queueTaskID
                duplicate
              }
              ... on GraphError {
                code
                message
                field
              }
            }
          }
        '''),
        variables: <String, dynamic>{'input': input},
      ),
    );
    final error = _extractOperationError(result, field: 'runIngestionAgent');
    if (error != null) {
      return ApiResult<IngestionRunTicket>.failure(error);
    }
    final payload = result.data?['runIngestionAgent'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<IngestionRunTicket>.failure(
        'runIngestionAgent returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<IngestionRunTicket>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    return ApiResult<IngestionRunTicket>.success(
      IngestionRunTicket(
        runID: payload['runID'] as String,
        taskID: payload['taskID'] as String,
        jobID: payload['jobID'] as String,
        queueTaskID: payload['queueTaskID'] as String,
        duplicate: payload['duplicate'] as bool,
      ),
    );
  }

  Future<ApiResult<String>> refineIngestionPrompt({
    required String projectID,
    required String taskboardName,
    String? userPrompt,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
      mutation RefineIngestionPrompt(
      \$input: RefineIngestionPromptInput!
      ) {
      refineIngestionPrompt(input: \$input) {
        __typename
        ... on RefineIngestionPromptSuccess {
        prompt
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
            'projectID': projectID,
            'taskboardName': taskboardName,
            'userPrompt': userPrompt,
          },
        },
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'refineIngestionPrompt',
    );
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final payload =
        result.data?['refineIngestionPrompt'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<String>.failure(
        'refineIngestionPrompt returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<String>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    return ApiResult<String>.success(
      (payload['prompt'] as String? ?? '').trim(),
    );
  }

  Future<ApiResult<List<ProjectRepositoryBranchOption>>>
  projectRepositoryBranches({required String projectID}) async {
    final result = await _client.query(
      QueryOptions(
        fetchPolicy: FetchPolicy.networkOnly,
        document: gql('''
          query ProjectRepositoryBranches(
            \$projectID: String!
          ) {
            projectRepositoryBranches(projectID: \$projectID) {
              __typename
              ... on ProjectRepositoryBranchesSuccess {
                repositories {
                  repositoryID
                  repositoryURL
                  defaultBranch
                  branches
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
        variables: <String, dynamic>{'projectID': projectID},
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'projectRepositoryBranches',
    );
    if (error != null) {
      return ApiResult<List<ProjectRepositoryBranchOption>>.failure(error);
    }
    final payload =
        result.data?['projectRepositoryBranches'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<ProjectRepositoryBranchOption>>.failure(
        'projectRepositoryBranches returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<ProjectRepositoryBranchOption>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final repositories =
        payload['repositories'] as List<dynamic>? ?? const <dynamic>[];
    final items = repositories
        .whereType<Map<String, dynamic>>()
        .map(
          (Map<String, dynamic> repository) => ProjectRepositoryBranchOption(
            repositoryID: repository['repositoryID'] as String,
            repositoryURL: repository['repositoryURL'] as String,
            defaultBranch: repository['defaultBranch'] as String?,
            branches:
                (repository['branches'] as List<dynamic>? ?? const <dynamic>[])
                    .whereType<String>()
                    .toList(growable: false),
          ),
        )
        .toList(growable: false);
    return ApiResult<List<ProjectRepositoryBranchOption>>.success(items);
  }

  Future<ApiResult<List<TaskboardModel>>> taskboards({
    required String projectID,
  }) async {
    final result = await _client.query(
      QueryOptions(
        fetchPolicy: FetchPolicy.networkOnly,
        document: gql('''
          query Taskboards(
            \$projectID: String!
          ) {
            taskboards(projectID: \$projectID) {
              __typename
              ... on TaskboardsSuccess {
                boards {
                  boardID
                  projectID
                  name
                  state
                  createdAt
                  updatedAt
                  ingestionAudits {
                    modelProvider
                    modelName
                    modelVersion
                    modelRunID
                    agentSessionID
                    agentStreamID
                    promptFingerprint
                    inputTokens
                    outputTokens
                    startedAt
                    completedAt
                  }
                  ingestionDetails {
                    filesAdded
                    userPrompt
                  }
                  epics {
                    id
                    boardID
                    title
                    objective
                repositoryIDs
                deliverables
                state
                    rank
                    dependsOnEpicIDs
                    tasks {
                      id
                      boardID
                      epicID
                      title
                      description
                  repositoryIDs
                  deliverables
                  taskType
                      state
                      rank
                      dependsOnTaskIDs
                    audits {
                      modelProvider
                      modelName
                      modelVersion
                      modelRunID
                      agentSessionID
                      agentStreamID
                      promptFingerprint
                      inputTokens
                      outputTokens
                      startedAt
                      completedAt
                    }
                    }
                  }
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
        variables: <String, dynamic>{'projectID': projectID},
      ),
    );
    final error = _extractOperationError(result, field: 'taskboards');
    if (error != null) {
      return ApiResult<List<TaskboardModel>>.failure(error);
    }
    final payload = result.data?['taskboards'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<TaskboardModel>>.failure(
        'taskboards returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<TaskboardModel>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final boards = payload['boards'] as List<dynamic>? ?? const <dynamic>[];
    return ApiResult<List<TaskboardModel>>.success(
      boards
          .whereType<Map<String, dynamic>>()
          .map(_parseTaskboard)
          .toList(growable: false),
    );
  }

  Future<ApiResult<TaskboardModel>> taskboard({
    required String projectID,
    required String boardID,
  }) async {
    final result = await _client.query(
      QueryOptions(
        fetchPolicy: FetchPolicy.networkOnly,
        document: gql('''
          query Taskboard(
            \$projectID: String!
            \$boardID: String!
          ) {
            taskboard(projectID: \$projectID, boardID: \$boardID) {
              __typename
              ... on TaskboardSuccess {
                board {
                  boardID
                  projectID
                  name
                  state
                  createdAt
                  updatedAt
                  ingestionAudits {
                    modelProvider
                    modelName
                    modelVersion
                    modelRunID
                    agentSessionID
                    agentStreamID
                    promptFingerprint
                    inputTokens
                    outputTokens
                    startedAt
                    completedAt
                  }
                  ingestionDetails {
                    filesAdded
                    userPrompt
                  }
                  epics {
                    id
                    boardID
                    title
                    objective
                repositoryIDs
                deliverables
                state
                    rank
                    dependsOnEpicIDs
                    tasks {
                      id
                      boardID
                      epicID
                      title
                      description
                  repositoryIDs
                  deliverables
                  taskType
                      state
                      rank
                      dependsOnTaskIDs
                    audits {
                      modelProvider
                      modelName
                      modelVersion
                      modelRunID
                      agentSessionID
                      agentStreamID
                      promptFingerprint
                      inputTokens
                      outputTokens
                      startedAt
                      completedAt
                    }
                    }
                  }
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
          'projectID': projectID,
          'boardID': boardID,
        },
      ),
    );
    final error = _extractOperationError(result, field: 'taskboard');
    if (error != null) {
      return ApiResult<TaskboardModel>.failure(error);
    }
    final payload = result.data?['taskboard'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<TaskboardModel>.failure(
        'taskboard returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<TaskboardModel>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final board = payload['board'] as Map<String, dynamic>?;
    if (board == null) {
      return const ApiResult<TaskboardModel>.failure(
        'taskboard payload missing board',
      );
    }
    return ApiResult<TaskboardModel>.success(_parseTaskboard(board));
  }

  Future<ApiResult<TaskboardModel>> createTaskboard({
    required String projectID,
    required String name,
  }) => _runTaskboardMutation(
    operationName: 'createTaskboard',
    document: '''
      mutation CreateTaskboard(
        \$input: CreateTaskboardInput!
      ) {
        createTaskboard(input: \$input) {
          __typename
          ... on TaskboardMutationSuccess {
            board {
              boardID
              projectID
              name
              state
              createdAt
              updatedAt
              ingestionAudits {
                modelProvider
                modelName
                modelVersion
                modelRunID
                agentSessionID
                agentStreamID
                promptFingerprint
                inputTokens
                outputTokens
                startedAt
                completedAt
              }
              ingestionDetails {
                filesAdded
                userPrompt
              }
              epics {
                id
                boardID
                title
                objective
                repositoryIDs
                deliverables
                state
                rank
                dependsOnEpicIDs
                tasks {
                  id
                  boardID
                  epicID
                  title
                  description
                  repositoryIDs
                  deliverables
                  taskType
                  state
                  rank
                  dependsOnTaskIDs
                audits {
                  modelProvider
                  modelName
                  modelVersion
                  modelRunID
                  agentSessionID
                  agentStreamID
                  promptFingerprint
                  inputTokens
                  outputTokens
                  startedAt
                  completedAt
                }
                }
              }
            }
          }
          ... on GraphError {
            code
            message
            field
          }
        }
      }
    ''',
    input: <String, dynamic>{'projectID': projectID, 'name': name},
  );

  Future<ApiResult<TaskboardModel>> updateTaskboard({
    required String projectID,
    required String boardID,
    required String name,
    required String state,
  }) => _runTaskboardMutation(
    operationName: 'updateTaskboard',
    document: '''
      mutation UpdateTaskboard(
        \$input: UpdateTaskboardInput!
      ) {
        updateTaskboard(input: \$input) {
          __typename
          ... on TaskboardMutationSuccess {
            board {
              boardID
              projectID
              name
              state
              createdAt
              updatedAt
              ingestionAudits {
                modelProvider
                modelName
                modelVersion
                modelRunID
                agentSessionID
                agentStreamID
                promptFingerprint
                inputTokens
                outputTokens
                startedAt
                completedAt
              }
              ingestionDetails {
                filesAdded
                userPrompt
              }
              epics {
                id
                boardID
                title
                objective
                repositoryIDs
                deliverables
                state
                rank
                dependsOnEpicIDs
                tasks {
                  id
                  boardID
                  epicID
                  title
                  description
                  repositoryIDs
                  deliverables
                  taskType
                  state
                  rank
                  dependsOnTaskIDs
                audits {
                  modelProvider
                  modelName
                  modelVersion
                  modelRunID
                  agentSessionID
                  agentStreamID
                  promptFingerprint
                  inputTokens
                  outputTokens
                  startedAt
                  completedAt
                }
                }
              }
            }
          }
          ... on GraphError {
            code
            message
            field
          }
        }
      }
    ''',
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'name': name,
      'state': state,
    },
  );

  Future<ApiResult<void>> deleteTaskboard({
    required String projectID,
    required String boardID,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
          mutation DeleteTaskboard(
            \$input: DeleteTaskboardInput!
          ) {
            deleteTaskboard(input: \$input) {
              __typename
              ... on TaskboardDeleteSuccess {
                ok
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
            'projectID': projectID,
            'boardID': boardID,
          },
        },
      ),
    );
    final error = _extractOperationError(result, field: 'deleteTaskboard');
    if (error != null) {
      return ApiResult<void>.failure(error);
    }
    final payload = result.data?['deleteTaskboard'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<void>.failure('deleteTaskboard returned no data');
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<void>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    return const ApiResult<void>.success(null);
  }

  Future<ApiResult<TaskboardModel>> createTaskboardEpic({
    required String projectID,
    required String boardID,
    required String title,
    String? objective,
    List<String> repositoryIDs = const <String>[],
    List<String> deliverables = const <String>[],
    required String state,
    int rank = 0,
  }) => _runTaskboardMutation(
    operationName: 'createTaskboardEpic',
    document: _epicMutationDocument(
      'CreateTaskboardEpic',
      'createTaskboardEpic',
      'CreateTaskboardEpicInput',
    ),
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'title': title,
      'objective': objective,
      'repositoryIDs': repositoryIDs,
      'deliverables': deliverables,
      'state': state,
      'rank': rank,
      'dependsOnEpicIDs': <String>[],
    },
  );

  Future<ApiResult<TaskboardModel>> updateTaskboardEpic({
    required String projectID,
    required String boardID,
    required String epicID,
    required String title,
    String? objective,
    List<String> repositoryIDs = const <String>[],
    List<String> deliverables = const <String>[],
    required String state,
    int rank = 0,
    List<String> dependsOnEpicIDs = const <String>[],
  }) => _runTaskboardMutation(
    operationName: 'updateTaskboardEpic',
    document: _epicMutationDocument(
      'UpdateTaskboardEpic',
      'updateTaskboardEpic',
      'UpdateTaskboardEpicInput',
    ),
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'epicID': epicID,
      'title': title,
      'objective': objective,
      'repositoryIDs': repositoryIDs,
      'deliverables': deliverables,
      'state': state,
      'rank': rank,
      'dependsOnEpicIDs': dependsOnEpicIDs,
    },
  );

  Future<ApiResult<TaskboardModel>> deleteTaskboardEpic({
    required String projectID,
    required String boardID,
    required String epicID,
  }) => _runTaskboardMutation(
    operationName: 'deleteTaskboardEpic',
    document: _epicDeleteMutationDocument(),
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'epicID': epicID,
    },
  );

  Future<ApiResult<TaskboardModel>> createTaskboardTask({
    required String projectID,
    required String boardID,
    required String epicID,
    required String title,
    String? description,
    List<String> repositoryIDs = const <String>[],
    List<String> deliverables = const <String>[],
    required String taskType,
    required String state,
    int rank = 0,
  }) => _runTaskboardMutation(
    operationName: 'createTaskboardTask',
    document: _taskMutationDocument(
      'CreateTaskboardTask',
      'createTaskboardTask',
      'CreateTaskboardTaskInput',
    ),
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'epicID': epicID,
      'title': title,
      'description': description,
      'repositoryIDs': repositoryIDs,
      'deliverables': deliverables,
      'taskType': taskType,
      'state': state,
      'rank': rank,
      'dependsOnTaskIDs': <String>[],
    },
  );

  Future<ApiResult<TaskboardModel>> updateTaskboardTask({
    required String projectID,
    required String boardID,
    required String epicID,
    required String taskID,
    required String title,
    String? description,
    List<String> repositoryIDs = const <String>[],
    List<String> deliverables = const <String>[],
    required String taskType,
    required String state,
    int rank = 0,
    List<String> dependsOnTaskIDs = const <String>[],
  }) => _runTaskboardMutation(
    operationName: 'updateTaskboardTask',
    document: _taskMutationDocument(
      'UpdateTaskboardTask',
      'updateTaskboardTask',
      'UpdateTaskboardTaskInput',
    ),
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'epicID': epicID,
      'taskID': taskID,
      'title': title,
      'description': description,
      'repositoryIDs': repositoryIDs,
      'deliverables': deliverables,
      'taskType': taskType,
      'state': state,
      'rank': rank,
      'dependsOnTaskIDs': dependsOnTaskIDs,
    },
  );

  Future<ApiResult<TaskboardModel>> deleteTaskboardTask({
    required String projectID,
    required String boardID,
    required String taskID,
  }) => _runTaskboardMutation(
    operationName: 'deleteTaskboardTask',
    document: _taskDeleteMutationDocument(),
    input: <String, dynamic>{
      'projectID': projectID,
      'boardID': boardID,
      'taskID': taskID,
    },
  );

  Future<ApiResult<TaskboardModel>> _runTaskboardMutation({
    required String operationName,
    required String document,
    required Map<String, dynamic> input,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql(document),
        variables: <String, dynamic>{'input': input},
      ),
    );
    final error = _extractOperationError(result, field: operationName);
    if (error != null) {
      return ApiResult<TaskboardModel>.failure(error);
    }
    final payload = result.data?[operationName] as Map<String, dynamic>?;
    if (payload == null) {
      return ApiResult<TaskboardModel>.failure(
        '$operationName returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<TaskboardModel>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final board = payload['board'] as Map<String, dynamic>?;
    if (board == null) {
      return ApiResult<TaskboardModel>.failure(
        '$operationName payload missing board',
      );
    }
    return ApiResult<TaskboardModel>.success(_parseTaskboard(board));
  }

  TaskboardModel _parseTaskboard(Map<String, dynamic> board) {
    final ingestionDetails = board['ingestionDetails'] as Map<String, dynamic>?;
    final epics = (board['epics'] as List<dynamic>? ?? const <dynamic>[])
        .whereType<Map<String, dynamic>>()
        .map((Map<String, dynamic> epic) {
          final tasks = (epic['tasks'] as List<dynamic>? ?? const <dynamic>[])
              .whereType<Map<String, dynamic>>()
              .map(
                (Map<String, dynamic> task) => TaskboardTaskModel(
                  id: task['id'] as String,
                  boardID: task['boardID'] as String,
                  epicID: task['epicID'] as String,
                  title: task['title'] as String,
                  description: task['description'] as String?,
                  repositoryIDs:
                      (task['repositoryIDs'] as List<dynamic>? ??
                              const <dynamic>[])
                          .whereType<String>()
                          .toList(growable: false),
                  deliverables:
                      (task['deliverables'] as List<dynamic>? ??
                              const <dynamic>[])
                          .whereType<String>()
                          .toList(growable: false),
                  taskType: task['taskType'] as String,
                  state: task['state'] as String,
                  rank: (task['rank'] as num?)?.toInt() ?? 0,
                  dependsOnTaskIDs:
                      (task['dependsOnTaskIDs'] as List<dynamic>? ??
                              const <dynamic>[])
                          .whereType<String>()
                          .toList(growable: false),
                  audits:
                      (task['audits'] as List<dynamic>? ?? const <dynamic>[])
                          .whereType<Map<String, dynamic>>()
                          .map(
                            (Map<String, dynamic> audit) => TaskModelAuditModel(
                              modelProvider: audit['modelProvider'] as String,
                              modelName: audit['modelName'] as String,
                              modelVersion: audit['modelVersion'] as String?,
                              modelRunID: audit['modelRunID'] as String?,
                              agentSessionID:
                                  audit['agentSessionID'] as String?,
                              agentStreamID: audit['agentStreamID'] as String?,
                              promptFingerprint:
                                  audit['promptFingerprint'] as String?,
                              inputTokens: (audit['inputTokens'] as num?)
                                  ?.toInt(),
                              outputTokens: (audit['outputTokens'] as num?)
                                  ?.toInt(),
                              startedAt: audit['startedAt'] == null
                                  ? null
                                  : DateTime.parse(
                                      audit['startedAt'] as String,
                                    ).toLocal(),
                              completedAt: audit['completedAt'] == null
                                  ? null
                                  : DateTime.parse(
                                      audit['completedAt'] as String,
                                    ).toLocal(),
                            ),
                          )
                          .toList(growable: false),
                ),
              )
              .toList(growable: false);
          return TaskboardEpicModel(
            id: epic['id'] as String,
            boardID: epic['boardID'] as String,
            title: epic['title'] as String,
            objective: epic['objective'] as String?,
            repositoryIDs:
                (epic['repositoryIDs'] as List<dynamic>? ?? const <dynamic>[])
                    .whereType<String>()
                    .toList(growable: false),
            deliverables:
                (epic['deliverables'] as List<dynamic>? ?? const <dynamic>[])
                    .whereType<String>()
                    .toList(growable: false),
            state: epic['state'] as String,
            rank: (epic['rank'] as num?)?.toInt() ?? 0,
            dependsOnEpicIDs:
                (epic['dependsOnEpicIDs'] as List<dynamic>? ??
                        const <dynamic>[])
                    .whereType<String>()
                    .toList(growable: false),
            tasks: tasks,
          );
        })
        .toList(growable: false);
    return TaskboardModel(
      boardID: board['boardID'] as String,
      projectID: board['projectID'] as String,
      name: board['name'] as String,
      state: board['state'] as String,
      ingestionAudits:
          (board['ingestionAudits'] as List<dynamic>? ?? const <dynamic>[])
              .whereType<Map<String, dynamic>>()
              .map(
                (Map<String, dynamic> audit) => TaskModelAuditModel(
                  modelProvider: audit['modelProvider'] as String,
                  modelName: audit['modelName'] as String,
                  modelVersion: audit['modelVersion'] as String?,
                  modelRunID: audit['modelRunID'] as String?,
                  agentSessionID: audit['agentSessionID'] as String?,
                  agentStreamID: audit['agentStreamID'] as String?,
                  promptFingerprint: audit['promptFingerprint'] as String?,
                  inputTokens: (audit['inputTokens'] as num?)?.toInt(),
                  outputTokens: (audit['outputTokens'] as num?)?.toInt(),
                  startedAt: audit['startedAt'] == null
                      ? null
                      : DateTime.parse(audit['startedAt'] as String).toLocal(),
                  completedAt: audit['completedAt'] == null
                      ? null
                      : DateTime.parse(
                          audit['completedAt'] as String,
                        ).toLocal(),
                ),
              )
              .toList(growable: false),
      ingestionFilesAdded:
          (ingestionDetails?['filesAdded'] as List<dynamic>? ??
                  const <dynamic>[])
              .whereType<String>()
              .toList(growable: false),
      ingestionUserPrompt: ingestionDetails?['userPrompt'] as String?,
      createdAt: DateTime.parse(board['createdAt'] as String).toLocal(),
      updatedAt: DateTime.parse(board['updatedAt'] as String).toLocal(),
      epics: epics,
    );
  }

  static String _epicMutationDocument(
    String operationTitle,
    String operationField,
    String inputType,
  ) =>
      '''
      mutation $operationTitle(
        \$input: $inputType!
      ) {
        $operationField(input: \$input) {
          __typename
          ... on TaskboardMutationSuccess {
            board {
              boardID
              projectID
              name
              state
              createdAt
              updatedAt
              ingestionAudits {
                modelProvider
                modelName
                modelVersion
                modelRunID
                agentSessionID
                agentStreamID
                promptFingerprint
                inputTokens
                outputTokens
                startedAt
                completedAt
              }
              ingestionDetails {
                filesAdded
                userPrompt
              }
              epics {
                id
                boardID
                title
                objective
                repositoryIDs
                deliverables
                state
                rank
                dependsOnEpicIDs
                tasks {
                  id
                  boardID
                  epicID
                  title
                  description
                  repositoryIDs
                  deliverables
                  taskType
                  state
                  rank
                  dependsOnTaskIDs
                audits {
                  modelProvider
                  modelName
                  modelVersion
                  modelRunID
                  agentSessionID
                  agentStreamID
                  promptFingerprint
                  inputTokens
                  outputTokens
                  startedAt
                  completedAt
                }
                }
              }
            }
          }
          ... on GraphError {
            code
            message
            field
          }
        }
      }
    ''';

  static String _epicDeleteMutationDocument() => '''
      mutation DeleteTaskboardEpic(
        \$input: DeleteTaskboardEpicInput!
      ) {
        deleteTaskboardEpic(input: \$input) {
          __typename
          ... on TaskboardMutationSuccess {
            board {
              boardID
              projectID
              name
              state
              createdAt
              updatedAt
              ingestionAudits {
                modelProvider
                modelName
                modelVersion
                modelRunID
                agentSessionID
                agentStreamID
                promptFingerprint
                inputTokens
                outputTokens
                startedAt
                completedAt
              }
              ingestionDetails {
                filesAdded
                userPrompt
              }
              epics {
                id
                boardID
                title
                objective
                repositoryIDs
                deliverables
                state
                rank
                dependsOnEpicIDs
                tasks {
                  id
                  boardID
                  epicID
                  title
                  description
                  repositoryIDs
                  deliverables
                  taskType
                  state
                  rank
                  dependsOnTaskIDs
                audits {
                  modelProvider
                  modelName
                  modelVersion
                  modelRunID
                  agentSessionID
                  agentStreamID
                  promptFingerprint
                  inputTokens
                  outputTokens
                  startedAt
                  completedAt
                }
                }
              }
            }
          }
          ... on GraphError {
            code
            message
            field
          }
        }
      }
    ''';

  static String _taskMutationDocument(
    String operationTitle,
    String operationField,
    String inputType,
  ) =>
      '''
      mutation $operationTitle(
        \$input: $inputType!
      ) {
        $operationField(input: \$input) {
          __typename
          ... on TaskboardMutationSuccess {
            board {
              boardID
              projectID
              name
              state
              createdAt
              updatedAt
              epics {
                id
                boardID
                title
                objective
                repositoryIDs
                deliverables
                state
                rank
                dependsOnEpicIDs
                tasks {
                  id
                  boardID
                  epicID
                  title
                  description
                  repositoryIDs
                  deliverables
                  taskType
                  state
                  rank
                  dependsOnTaskIDs
                audits {
                  modelProvider
                  modelName
                  modelVersion
                  modelRunID
                  agentSessionID
                  agentStreamID
                  promptFingerprint
                  inputTokens
                  outputTokens
                  startedAt
                  completedAt
                }
                }
              }
            }
          }
          ... on GraphError {
            code
            message
            field
          }
        }
      }
    ''';

  static String _taskDeleteMutationDocument() => '''
      mutation DeleteTaskboardTask(
        \$input: DeleteTaskboardTaskInput!
      ) {
        deleteTaskboardTask(input: \$input) {
          __typename
          ... on TaskboardMutationSuccess {
            board {
              boardID
              projectID
              name
              state
              createdAt
              updatedAt
              epics {
                id
                boardID
                title
                objective
                repositoryIDs
                deliverables
                state
                rank
                dependsOnEpicIDs
                tasks {
                  id
                  boardID
                  epicID
                  title
                  description
                  repositoryIDs
                  deliverables
                  taskType
                  state
                  rank
                  dependsOnTaskIDs
                audits {
                  modelProvider
                  modelName
                  modelVersion
                  modelRunID
                  agentSessionID
                  agentStreamID
                  promptFingerprint
                  inputTokens
                  outputTokens
                  startedAt
                  completedAt
                }
                }
              }
            }
          }
          ... on GraphError {
            code
            message
            field
          }
        }
      }
    ''';

  Stream<ApiResult<StreamEvent>> sessionActivityStream({
    required String runID,
    String? projectID,
    String? taskID,
    String? jobID,
    int fromOffset = 0,
  }) {
    final cleanRunID = runID.trim();
    if (cleanRunID.isEmpty) {
      return Stream<ApiResult<StreamEvent>>.value(
        const ApiResult<StreamEvent>.failure('run_id is required'),
      );
    }
    return _client
        .subscribe(
          SubscriptionOptions(
            document: gql('''
              subscription SessionActivityStream(
                \$correlation: CorrelationInput!
                \$fromOffset: Int!
              ) {
                sessionActivityStream(correlation: \$correlation, fromOffset: \$fromOffset) {
                  __typename
                  ... on StreamEventSuccess {
                    event {
                      eventID
                      streamOffset
                      occurredAt
                      runID
                      taskID
                      jobID
                      projectID
                      sessionID
                      correlationID
                      source
                      eventType
                      gapDetected
                      gapReconciled
                      expectedEventSeq
                      observedEventSeq
                      payload
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
              'correlation': <String, dynamic>{
                'runID': cleanRunID,
                'taskID': taskID?.trim() ?? '',
                'jobID': jobID?.trim() ?? '',
                'projectID': projectID?.trim(),
              },
              'fromOffset': fromOffset,
            },
          ),
        )
        .map((QueryResult result) {
          final error = _extractOperationError(
            result,
            field: 'sessionActivityStream',
          );
          if (error != null) {
            return ApiResult<StreamEvent>.failure(error);
          }
          final payload =
              result.data?['sessionActivityStream'] as Map<String, dynamic>?;
          if (payload == null) {
            return const ApiResult<StreamEvent>.failure(
              'sessionActivityStream returned no data',
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
              'sessionActivityStream event payload missing',
            );
          }
          return ApiResult<StreamEvent>.success(
            _mapStreamEventFromMap(eventData),
          );
        })
        .asBroadcastStream();
  }

  Stream<ApiResult<StreamEvent>> pipelineEventsStream({
    required String projectID,
    String? runID,
    String? taskID,
    String? jobID,
    int fromOffset = 0,
  }) {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return Stream<ApiResult<StreamEvent>>.value(
        const ApiResult<StreamEvent>.failure('project_id is required'),
      );
    }
    return _client
        .subscribe(
          SubscriptionOptions(
            document: gql('''
              subscription PipelineEventsStream(
                \$correlation: CorrelationInput!
                \$fromOffset: Int!
              ) {
                pipelineEventsStream(correlation: \$correlation, fromOffset: \$fromOffset) {
                  __typename
                  ... on StreamEventSuccess {
                    event {
                      eventID
                      streamOffset
                      occurredAt
                      runID
                      taskID
                      jobID
                      projectID
                      sessionID
                      correlationID
                      source
                      eventType
                      gapDetected
                      gapReconciled
                      expectedEventSeq
                      observedEventSeq
                      payload
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
              'correlation': <String, dynamic>{
                'projectID': cleanProjectID,
                'runID': runID?.trim() ?? '',
                'taskID': taskID?.trim() ?? '',
                'jobID': jobID?.trim() ?? '',
              },
              'fromOffset': fromOffset,
            },
          ),
        )
        .map((QueryResult result) {
          final error = _extractOperationError(
            result,
            field: 'pipelineEventsStream',
          );
          if (error != null) {
            return ApiResult<StreamEvent>.failure(error);
          }
          final payload =
              result.data?['pipelineEventsStream'] as Map<String, dynamic>?;
          if (payload == null) {
            return const ApiResult<StreamEvent>.failure(
              'pipelineEventsStream returned no data',
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
              'pipelineEventsStream event payload missing',
            );
          }
          return ApiResult<StreamEvent>.success(
            _mapStreamEventFromMap(eventData),
          );
        })
        .asBroadcastStream();
  }

  Stream<ApiResult<StreamEvent>> taskboardStream({
    required String projectID,
    int fromOffset = 0,
  }) {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return Stream<ApiResult<StreamEvent>>.value(
        const ApiResult<StreamEvent>.failure('project_id is required'),
      );
    }
    return _client
        .subscribe(
          SubscriptionOptions(
            document: gql('''
              subscription TaskboardStream(
                \$runID: String!
                \$taskID: String!
                \$jobID: String!
                \$projectID: String
                \$fromOffset: Int!
              ) {
                taskboardStream(
                  correlation: {
                    runID: \$runID
                    taskID: \$taskID
                    jobID: \$jobID
                    projectID: \$projectID
                  }
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
              'projectID': cleanProjectID,
              'fromOffset': fromOffset,
            },
          ),
        )
        .map((QueryResult result) {
          final error = _extractOperationError(
            result,
            field: 'taskboardStream',
          );
          if (error != null) {
            return ApiResult<StreamEvent>.failure(error);
          }
          final payload =
              result.data?['taskboardStream'] as Map<String, dynamic>?;
          if (payload == null) {
            return const ApiResult<StreamEvent>.failure(
              'taskboardStream returned no data',
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
              'taskboardStream event payload missing',
            );
          }
          return ApiResult<StreamEvent>.success(
            _mapStreamEventFromMap(eventData),
          );
        })
        .asBroadcastStream();
  }

  Future<ApiResult<List<StreamEvent>>> projectEvents({
    required String projectID,
    int fromOffset = 0,
    int limit = 200,
  }) async {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return const ApiResult<List<StreamEvent>>.failure(
        'project_id is required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query ProjectEvents(
            \$projectID: String!
            \$fromOffset: Int!
            \$limit: Int!
          ) {
            projectEvents(projectID: \$projectID, fromOffset: \$fromOffset, limit: \$limit) {
              __typename
              ... on StreamEventsSuccess {
                events {
                  eventID
                  streamOffset
                  occurredAt
                  runID
                  taskID
                  jobID
                  projectID
                  sessionID
                  correlationID
                  source
                  eventType
                  gapDetected
                  gapReconciled
                  expectedEventSeq
                  observedEventSeq
                  payload
                }
                nextFromOffset
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
          'projectID': cleanProjectID,
          'fromOffset': fromOffset,
          'limit': limit,
        },
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'projectEvents');
    if (error != null) {
      return ApiResult<List<StreamEvent>>.failure(error);
    }
    final payload = result.data?['projectEvents'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<StreamEvent>>.failure(
        'projectEvents returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<StreamEvent>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final events = (payload['events'] as List<dynamic>? ?? const <dynamic>[])
        .whereType<Map<String, dynamic>>()
        .map(_mapStreamEventFromMap)
        .toList(growable: false);
    return ApiResult<List<StreamEvent>>.success(events);
  }

  Future<ApiResult<List<StreamEvent>>> pipelineEvents({
    required String projectID,
    String? runID,
    String? taskID,
    String? jobID,
    int fromOffset = 0,
    int limit = 200,
  }) async {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return const ApiResult<List<StreamEvent>>.failure(
        'project_id is required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query PipelineEvents(
            \$projectID: String!
            \$runID: String
            \$taskID: String
            \$jobID: String
            \$fromOffset: Int!
            \$limit: Int!
          ) {
            pipelineEvents(
              correlation: {
                projectID: \$projectID
                runID: \$runID
                taskID: \$taskID
                jobID: \$jobID
              }
              fromOffset: \$fromOffset
              limit: \$limit
            ) {
              __typename
              ... on StreamEventsSuccess {
                events {
                  eventID
                  streamOffset
                  occurredAt
                  runID
                  taskID
                  jobID
                  projectID
                  sessionID
                  correlationID
                  source
                  eventType
                  gapDetected
                  gapReconciled
                  expectedEventSeq
                  observedEventSeq
                  payload
                }
                nextFromOffset
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
          'projectID': cleanProjectID,
          'runID': runID?.trim().isEmpty ?? true ? null : runID!.trim(),
          'taskID': taskID?.trim().isEmpty ?? true ? null : taskID!.trim(),
          'jobID': jobID?.trim().isEmpty ?? true ? null : jobID!.trim(),
          'fromOffset': fromOffset,
          'limit': limit,
        },
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'pipelineEvents');
    if (error != null) {
      return ApiResult<List<StreamEvent>>.failure(error);
    }
    final payload = result.data?['pipelineEvents'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<StreamEvent>>.failure(
        'pipelineEvents returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<StreamEvent>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final events = (payload['events'] as List<dynamic>? ?? const <dynamic>[])
        .whereType<Map<String, dynamic>>()
        .map(_mapStreamEventFromMap)
        .toList(growable: false);
    return ApiResult<List<StreamEvent>>.success(events);
  }

  Stream<ApiResult<StreamEvent>> projectEventsStream({
    required String projectID,
    int fromOffset = 0,
  }) {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return Stream<ApiResult<StreamEvent>>.value(
        const ApiResult<StreamEvent>.failure('project_id is required'),
      );
    }
    return _client
        .subscribe(
          SubscriptionOptions(
            document: gql('''
              subscription ProjectEventsStream(
                \$projectID: String!
                \$fromOffset: Int!
              ) {
                projectEventsStream(projectID: \$projectID, fromOffset: \$fromOffset) {
                  __typename
                  ... on StreamEventSuccess {
                    event {
                      eventID
                      streamOffset
                      occurredAt
                      runID
                      taskID
                      jobID
                      projectID
                      sessionID
                      correlationID
                      source
                      eventType
                      gapDetected
                      gapReconciled
                      expectedEventSeq
                      observedEventSeq
                      payload
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
              'projectID': cleanProjectID,
              'fromOffset': fromOffset,
            },
          ),
        )
        .map((QueryResult result) {
          final error = _extractOperationError(
            result,
            field: 'projectEventsStream',
          );
          if (error != null) {
            return ApiResult<StreamEvent>.failure(error);
          }
          final payload =
              result.data?['projectEventsStream'] as Map<String, dynamic>?;
          if (payload == null) {
            return const ApiResult<StreamEvent>.failure(
              'projectEventsStream returned no data',
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
              'projectEventsStream event payload missing',
            );
          }
          return ApiResult<StreamEvent>.success(
            _mapStreamEventFromMap(eventData),
          );
        })
        .asBroadcastStream();
  }

  Future<ApiResult<List<LifecycleSessionSnapshotModel>>>
  lifecycleSessionSnapshots({
    required String projectID,
    String? pipelineType,
    int limit = 200,
  }) async {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return const ApiResult<List<LifecycleSessionSnapshotModel>>.failure(
        'project_id is required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query LifecycleSessionSnapshots(
            \$projectID: String!
            \$pipelineType: String
            \$limit: Int!
          ) {
            lifecycleSessionSnapshots(projectID: \$projectID, pipelineType: \$pipelineType, limit: \$limit) {
              __typename
              ... on LifecycleSessionSnapshotsSuccess {
                sessions {
                  projectID
                  runID
                  taskID
                  jobID
                  sessionID
                  pipelineType
                  sourceRuntime
                  currentState
                  currentSeverity
                  lastReasonCode
                  lastReasonSummary
                  lastEventSeq
                  lastProjectEventSeq
                  lastLivenessAt
                  lastActivityAt
                  lastCheckpointAt
                  startedAt
                  endedAt
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
          'projectID': cleanProjectID,
          'pipelineType': pipelineType?.trim().isEmpty ?? true
              ? null
              : pipelineType!.trim(),
          'limit': limit,
        },
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'lifecycleSessionSnapshots',
    );
    if (error != null) {
      return ApiResult<List<LifecycleSessionSnapshotModel>>.failure(error);
    }
    final payload =
        result.data?['lifecycleSessionSnapshots'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<LifecycleSessionSnapshotModel>>.failure(
        'lifecycleSessionSnapshots returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<LifecycleSessionSnapshotModel>>.failure(
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
            .map((Map<String, dynamic> session) {
              DateTime? parseOptionalDate(String key) {
                final raw = session[key] as String?;
                if (raw == null || raw.trim().isEmpty) {
                  return null;
                }
                return DateTime.parse(raw).toLocal();
              }

              return LifecycleSessionSnapshotModel(
                projectID: session['projectID'] as String,
                runID: session['runID'] as String?,
                taskID: session['taskID'] as String?,
                jobID: session['jobID'] as String?,
                sessionID: session['sessionID'] as String,
                pipelineType: session['pipelineType'] as String,
                sourceRuntime: session['sourceRuntime'] as String?,
                currentState: session['currentState'] as String,
                currentSeverity: session['currentSeverity'] as String,
                lastReasonCode: session['lastReasonCode'] as String?,
                lastReasonSummary: session['lastReasonSummary'] as String?,
                lastEventSeq: session['lastEventSeq'] as int,
                lastProjectEventSeq: session['lastProjectEventSeq'] as int,
                lastLivenessAt: parseOptionalDate('lastLivenessAt'),
                lastActivityAt: parseOptionalDate('lastActivityAt'),
                lastCheckpointAt: parseOptionalDate('lastCheckpointAt'),
                startedAt: DateTime.parse(
                  session['startedAt'] as String,
                ).toLocal(),
                endedAt: parseOptionalDate('endedAt'),
                updatedAt: DateTime.parse(
                  session['updatedAt'] as String,
                ).toLocal(),
              );
            })
            .toList(growable: false);
    return ApiResult<List<LifecycleSessionSnapshotModel>>.success(sessions);
  }

  Future<ApiResult<List<LifecycleHistoryEventModel>>> lifecycleSessionHistory({
    required String projectID,
    required String sessionID,
    int fromEventSeq = 0,
    int limit = 500,
  }) async {
    final cleanProjectID = projectID.trim();
    final cleanSessionID = sessionID.trim();
    if (cleanProjectID.isEmpty || cleanSessionID.isEmpty) {
      return const ApiResult<List<LifecycleHistoryEventModel>>.failure(
        'project_id and session_id are required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query LifecycleSessionHistory(
            \$projectID: String!
            \$sessionID: String!
            \$fromEventSeq: Int!
            \$limit: Int!
          ) {
            lifecycleSessionHistory(projectID: \$projectID, sessionID: \$sessionID, fromEventSeq: \$fromEventSeq, limit: \$limit) {
              __typename
              ... on LifecycleHistorySuccess {
                events {
                  eventID
                  projectID
                  runID
                  taskID
                  jobID
                  sessionID
                  pipelineType
                  sourceRuntime
                  eventType
                  eventSeq
                  projectEventSeq
                  occurredAt
                  payload
                }
                nextFromEventSeq
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
          'projectID': cleanProjectID,
          'sessionID': cleanSessionID,
          'fromEventSeq': fromEventSeq,
          'limit': limit,
        },
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'lifecycleSessionHistory',
    );
    if (error != null) {
      return ApiResult<List<LifecycleHistoryEventModel>>.failure(error);
    }
    final payload =
        result.data?['lifecycleSessionHistory'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<LifecycleHistoryEventModel>>.failure(
        'lifecycleSessionHistory returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<LifecycleHistoryEventModel>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final events = (payload['events'] as List<dynamic>? ?? const <dynamic>[])
        .whereType<Map<String, dynamic>>()
        .map(
          (Map<String, dynamic> event) => LifecycleHistoryEventModel(
            eventID: event['eventID'] as String,
            projectID: event['projectID'] as String,
            runID: event['runID'] as String?,
            taskID: event['taskID'] as String?,
            jobID: event['jobID'] as String?,
            sessionID: event['sessionID'] as String,
            pipelineType: event['pipelineType'] as String,
            sourceRuntime: event['sourceRuntime'] as String,
            eventType: event['eventType'] as String,
            eventSeq: event['eventSeq'] as int,
            projectEventSeq: event['projectEventSeq'] as int,
            occurredAt: DateTime.parse(event['occurredAt'] as String).toLocal(),
            payload: event['payload'] as String,
          ),
        )
        .toList(growable: false);
    return ApiResult<List<LifecycleHistoryEventModel>>.success(events);
  }

  Future<ApiResult<List<LifecycleTreeNodeModel>>> lifecycleTreeNodes({
    required String projectID,
    String? pipelineType,
    String? runID,
    String? taskID,
    String? jobID,
    int limit = 500,
  }) async {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return const ApiResult<List<LifecycleTreeNodeModel>>.failure(
        'project_id is required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query LifecycleTreeNodes(
            \$filter: LifecycleTreeFilterInput!
            \$limit: Int!
          ) {
            lifecycleTreeNodes(filter: \$filter, limit: \$limit) {
              __typename
              ... on LifecycleTreeNodesSuccess {
                nodes {
                  nodeID
                  parentNodeID
                  nodeType
                  projectID
                  runID
                  taskID
                  jobID
                  sessionID
                  pipelineType
                  sourceRuntime
                  currentState
                  currentSeverity
                  sessionCount
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
          'filter': <String, dynamic>{
            'projectID': cleanProjectID,
            'pipelineType': pipelineType?.trim().isEmpty ?? true
                ? null
                : pipelineType!.trim(),
            'runID': runID?.trim().isEmpty ?? true ? null : runID!.trim(),
            'taskID': taskID?.trim().isEmpty ?? true ? null : taskID!.trim(),
            'jobID': jobID?.trim().isEmpty ?? true ? null : jobID!.trim(),
          },
          'limit': limit,
        },
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'lifecycleTreeNodes');
    if (error != null) {
      return ApiResult<List<LifecycleTreeNodeModel>>.failure(error);
    }
    final payload = result.data?['lifecycleTreeNodes'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<List<LifecycleTreeNodeModel>>.failure(
        'lifecycleTreeNodes returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<List<LifecycleTreeNodeModel>>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final nodes = (payload['nodes'] as List<dynamic>? ?? const <dynamic>[])
        .whereType<Map<String, dynamic>>()
        .map(
          (Map<String, dynamic> node) => LifecycleTreeNodeModel(
            nodeID: node['nodeID'] as String,
            parentNodeID: node['parentNodeID'] as String?,
            nodeType: node['nodeType'] as String,
            projectID: node['projectID'] as String,
            runID: node['runID'] as String?,
            taskID: node['taskID'] as String?,
            jobID: node['jobID'] as String?,
            sessionID: node['sessionID'] as String?,
            pipelineType: node['pipelineType'] as String?,
            sourceRuntime: node['sourceRuntime'] as String?,
            currentState: node['currentState'] as String?,
            currentSeverity: node['currentSeverity'] as String?,
            sessionCount: node['sessionCount'] as int,
            updatedAt: DateTime.parse(node['updatedAt'] as String).toLocal(),
          ),
        )
        .toList(growable: false);
    return ApiResult<List<LifecycleTreeNodeModel>>.success(nodes);
  }

  Future<ApiResult<InterventionMetricsModel>> interventionMetrics({
    required String projectID,
    int limit = 200,
  }) async {
    final cleanProjectID = projectID.trim();
    if (cleanProjectID.isEmpty) {
      return const ApiResult<InterventionMetricsModel>.failure(
        'project_id is required',
      );
    }
    final result = await _client.query(
      QueryOptions(
        document: gql('''
          query InterventionMetrics(
            \$projectID: String!
            \$limit: Int!
          ) {
            interventionMetrics(projectID: \$projectID, limit: \$limit) {
              __typename
              ... on InterventionMetricsSuccess {
                metrics {
                  projectID
                  interventionCount
                  successfulOutcomeCount
                  failedOutcomeCount
                  averageRecoverySeconds
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
          'projectID': cleanProjectID,
          'limit': limit,
        },
        fetchPolicy: FetchPolicy.networkOnly,
      ),
    );
    final error = _extractOperationError(result, field: 'interventionMetrics');
    if (error != null) {
      return ApiResult<InterventionMetricsModel>.failure(error);
    }
    final payload =
        result.data?['interventionMetrics'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<InterventionMetricsModel>.failure(
        'interventionMetrics returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<InterventionMetricsModel>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    final metrics = payload['metrics'] as Map<String, dynamic>?;
    if (metrics == null) {
      return const ApiResult<InterventionMetricsModel>.failure(
        'interventionMetrics payload missing',
      );
    }
    return ApiResult<InterventionMetricsModel>.success(
      InterventionMetricsModel(
        projectID: metrics['projectID'] as String,
        interventionCount: metrics['interventionCount'] as int,
        successfulOutcomeCount: metrics['successfulOutcomeCount'] as int,
        failedOutcomeCount: metrics['failedOutcomeCount'] as int,
        averageRecoverySeconds: metrics['averageRecoverySeconds'] as int,
      ),
    );
  }

  Future<ApiResult<String>> applyManualIntervention({
    required String projectID,
    required String sessionID,
    required String action,
    required String reason,
    required String actorID,
    bool force = false,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql('''
          mutation ApplyManualIntervention(
            \$input: ApplyManualInterventionInput!
          ) {
            applyManualIntervention(input: \$input) {
              __typename
              ... on ManualInterventionSuccess {
                ok
                eventID
                eventSeq
                projectEventSeq
                action
                resultingState
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
            'projectID': projectID.trim(),
            'sessionID': sessionID.trim(),
            'action': action.trim().toUpperCase(),
            'reason': reason.trim(),
            'actorID': actorID.trim(),
            'force': force,
          },
        },
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'applyManualIntervention',
    );
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final payload =
        result.data?['applyManualIntervention'] as Map<String, dynamic>?;
    if (payload == null) {
      return const ApiResult<String>.failure(
        'applyManualIntervention returned no data',
      );
    }
    if (payload['__typename'] == 'GraphError') {
      return ApiResult<String>.failure(
        _graphErrorMessageTyped(
          code: payload['code'] as String? ?? 'INTERNAL',
          message: payload['message'] as String? ?? 'unknown error',
          field: payload['field'] as String?,
        ),
      );
    }
    return ApiResult<String>.success(payload['eventID'] as String? ?? 'ok');
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
                  lastHeartbeat
                  leaseExpiresAt
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
                lastHeartbeat: DateTime.parse(
                  item['lastHeartbeat'] as String,
                ).toLocal(),
                leaseExpiresAt: DateTime.parse(
                  item['leaseExpiresAt'] as String,
                ).toLocal(),
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
        updatedAt: DateTime.parse(settings['updatedAt'] as String).toLocal(),
      ),
    );
  }

  Future<ApiResult<WorkerSettings>> updateWorkerSettings({
    required int heartbeatIntervalSeconds,
    required int responseDeadlineSeconds,
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
            _mapStreamEventFromMap(eventData),
          );
        })
        .asBroadcastStream();
  }

  StreamEvent _mapStreamEventFromMap(Map<String, dynamic> eventData) {
    final occurredAtRaw = eventData['occurredAt'] as String?;
    return StreamEvent(
      eventID: eventData['eventID'] as String,
      streamOffset: (eventData['streamOffset'] as int?) ?? 0,
      eventType: eventData['eventType'] as String,
      source: eventData['source'] as String,
      payload: eventData['payload'] as String,
      occurredAt: occurredAtRaw == null
          ? DateTime.now().toLocal()
          : DateTime.parse(occurredAtRaw).toLocal(),
      runID: eventData['runID'] as String?,
      taskID: eventData['taskID'] as String?,
      jobID: eventData['jobID'] as String?,
      projectID: eventData['projectID'] as String?,
      sessionID: eventData['sessionID'] as String?,
      gapDetected: (eventData['gapDetected'] as bool?) ?? false,
      gapReconciled: (eventData['gapReconciled'] as bool?) ?? false,
      expectedEventSeq: eventData['expectedEventSeq'] as int?,
      observedEventSeq: eventData['observedEventSeq'] as int?,
    );
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
