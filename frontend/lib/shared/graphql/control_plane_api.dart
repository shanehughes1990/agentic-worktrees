import 'dart:async';
import 'dart:convert';

import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:graphql/client.dart';

class ApiResult<T> {
  const ApiResult.success(this.data) : errorMessage = null, isSuccess = true;

  const ApiResult.failure(this.errorMessage) : data = null, isSuccess = false;

  final T? data;
  final String? errorMessage;
  final bool isSuccess;
}

class SessionSummary {
  const SessionSummary({
    required this.runID,
    required this.taskCount,
    required this.jobCount,
    required this.updatedAt,
  });

  final String runID;
  final int taskCount;
  final int jobCount;
  final DateTime updatedAt;

  factory SessionSummary.fromMap(Map<String, dynamic> value) {
    return SessionSummary(
      runID: value['runID'] as String? ?? '',
      taskCount: value['taskCount'] as int? ?? 0,
      jobCount: value['jobCount'] as int? ?? 0,
      updatedAt:
          DateTime.tryParse(value['updatedAt'] as String? ?? '')?.toLocal() ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}

class WorkerSummary {
  const WorkerSummary({
    required this.workerID,
    required this.capabilities,
    required this.lastHeartbeat,
  });

  final String workerID;
  final List<String> capabilities;
  final DateTime lastHeartbeat;

  factory WorkerSummary.fromMap(Map<String, dynamic> value) {
    return WorkerSummary(
      workerID: value['workerID'] as String? ?? '',
      capabilities:
          ((value['capabilities'] as List<dynamic>?) ?? const <dynamic>[])
              .map((entry) => entry.toString())
              .toList(growable: false),
      lastHeartbeat:
          DateTime.tryParse(
            value['lastHeartbeat'] as String? ?? '',
          )?.toLocal() ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}

class WorkflowJob {
  const WorkflowJob({
    required this.runID,
    required this.taskID,
    required this.jobID,
    required this.jobKind,
    required this.status,
    required this.queue,
    required this.queueTaskID,
    required this.duplicate,
    required this.updatedAt,
  });

  final String runID;
  final String taskID;
  final String jobID;
  final String jobKind;
  final String status;
  final String queue;
  final String queueTaskID;
  final bool duplicate;
  final DateTime updatedAt;

  factory WorkflowJob.fromMap(Map<String, dynamic> value) {
    return WorkflowJob(
      runID: value['runID'] as String? ?? '',
      taskID: value['taskID'] as String? ?? '',
      jobID: value['jobID'] as String? ?? '',
      jobKind: value['jobKind'] as String? ?? '',
      status: value['status'] as String? ?? '',
      queue: value['queue'] as String? ?? '',
      queueTaskID: value['queueTaskID'] as String? ?? '',
      duplicate: value['duplicate'] as bool? ?? false,
      updatedAt:
          DateTime.tryParse(value['updatedAt'] as String? ?? '')?.toLocal() ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}

class SupervisorDecision {
  const SupervisorDecision({
    required this.signalType,
    required this.action,
    required this.reason,
    required this.occurredAt,
  });

  final String signalType;
  final String action;
  final String reason;
  final DateTime occurredAt;

  factory SupervisorDecision.fromMap(Map<String, dynamic> value) {
    return SupervisorDecision(
      signalType: value['signalType'] as String? ?? '',
      action: value['action'] as String? ?? '',
      reason: value['reason'] as String? ?? '',
      occurredAt:
          DateTime.tryParse(value['occurredAt'] as String? ?? '')?.toLocal() ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}

class StreamEvent {
  const StreamEvent({
    required this.eventID,
    required this.eventType,
    required this.source,
    required this.payload,
    required this.occurredAt,
  });

  final String eventID;
  final String eventType;
  final String source;
  final String payload;
  final DateTime occurredAt;

  factory StreamEvent.fromMap(Map<String, dynamic> value) {
    return StreamEvent(
      eventID: value['eventID'] as String? ?? '',
      eventType: value['eventType'] as String? ?? '',
      source: value['source'] as String? ?? '',
      payload: value['payload'] as String? ?? '{}',
      occurredAt:
          DateTime.tryParse(value['occurredAt'] as String? ?? '')?.toLocal() ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}

class ControlPlaneApi {
  ControlPlaneApi(this._client);

  final GraphQLClient _client;

  static const String _sessionsQuery = r'''
query Sessions($limit: Int!) {
  sessions(limit: $limit) {
    __typename
    ... on SessionsSuccess {
      sessions { runID taskCount jobCount updatedAt }
    }
    ... on GraphError { code message field }
  }
}
''';

  static const String _workersQuery = r'''
query Workers($limit: Int!) {
  workers(limit: $limit) {
    __typename
    ... on WorkersSuccess {
      workers { workerID capabilities lastHeartbeat }
    }
    ... on GraphError { code message field }
  }
}
''';

  static const String _workflowJobsQuery = r'''
query WorkflowJobs($runID: String!, $taskID: String, $limit: Int!) {
  workflowJobs(runID: $runID, taskID: $taskID, limit: $limit) {
    __typename
    ... on WorkflowJobsSuccess {
      jobs { runID taskID jobID jobKind status queue queueTaskID duplicate updatedAt }
    }
    ... on GraphError { code message field }
  }
}
''';

  static const String _supervisorHistoryQuery = r'''
query SupervisorDecisionHistory($runID: String!, $taskID: String!, $jobID: String!) {
  supervisorDecisionHistory(correlation: {runID: $runID, taskID: $taskID, jobID: $jobID}) {
    __typename
    ... on SupervisorDecisionHistorySuccess {
      decisions { signalType action reason occurredAt }
    }
    ... on GraphError { code message field }
  }
}
''';

  static const String _enqueueScmWorkflowMutation = r'''
mutation EnqueueScmWorkflow($input: EnqueueSCMWorkflowInput!) {
  enqueueScmWorkflow(input: $input) {
    __typename
    ... on EnqueueSCMWorkflowSuccess { queueTaskID duplicate }
    ... on GraphError { code message field }
  }
}
''';

  static const String _enqueueIngestionMutation = r'''
mutation EnqueueIngestionWorkflow($input: EnqueueIngestionWorkflowInput!) {
  enqueueIngestionWorkflow(input: $input) {
    __typename
    ... on EnqueueIngestionWorkflowSuccess { queueTaskID duplicate }
    ... on GraphError { code message field }
  }
}
''';

  static const String _approveIssueMutation = r'''
mutation ApproveIssueIntake($input: ApproveIssueIntakeInput!) {
  approveIssueIntake(input: $input) {
    __typename
    ... on ApproveIssueIntakeSuccess { decision { action reason occurredAt } }
    ... on GraphError { code message field }
  }
}
''';

  static const String _sessionActivitySubscription = r'''
subscription SessionActivity($runID: String!, $taskID: String!, $jobID: String!, $fromOffset: Int!) {
  sessionActivityStream(correlation: {runID: $runID, taskID: $taskID, jobID: $jobID}, fromOffset: $fromOffset) {
    __typename
    ... on StreamEventSuccess { event { eventID eventType source payload occurredAt } }
    ... on GraphError { code message field }
  }
}
''';

  Future<ApiResult<List<SessionSummary>>> sessions({int limit = 50}) async {
    final result = await _client.query(
      QueryOptions(
        document: gql(_sessionsQuery),
        variables: <String, dynamic>{'limit': limit},
      ),
    );
    final error = _extractOperationError(result, field: 'sessions');
    if (error != null) {
      return ApiResult<List<SessionSummary>>.failure(error);
    }
    final node = _fieldNode(result.data, 'sessions');
    if (node == null) {
      return const ApiResult<List<SessionSummary>>.failure(
        'sessions returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<List<SessionSummary>>.failure(_graphErrorMessage(node));
    }
    final sessionsNode =
        node['sessions'] as List<dynamic>? ?? const <dynamic>[];
    final items = sessionsNode
        .whereType<Map<dynamic, dynamic>>()
        .map(
          (entry) => SessionSummary.fromMap(Map<String, dynamic>.from(entry)),
        )
        .toList(growable: false);
    return ApiResult<List<SessionSummary>>.success(items);
  }

  Future<ApiResult<List<WorkerSummary>>> workers({int limit = 50}) async {
    final result = await _client.query(
      QueryOptions(
        document: gql(_workersQuery),
        variables: <String, dynamic>{'limit': limit},
      ),
    );
    final error = _extractOperationError(result, field: 'workers');
    if (error != null) {
      return ApiResult<List<WorkerSummary>>.failure(error);
    }
    final node = _fieldNode(result.data, 'workers');
    if (node == null) {
      return const ApiResult<List<WorkerSummary>>.failure(
        'workers returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<List<WorkerSummary>>.failure(_graphErrorMessage(node));
    }
    final workersNode = node['workers'] as List<dynamic>? ?? const <dynamic>[];
    final items = workersNode
        .whereType<Map<dynamic, dynamic>>()
        .map((entry) => WorkerSummary.fromMap(Map<String, dynamic>.from(entry)))
        .toList(growable: false);
    return ApiResult<List<WorkerSummary>>.success(items);
  }

  Future<ApiResult<List<WorkflowJob>>> workflowJobs({
    required String runID,
    String? taskID,
    int limit = 100,
  }) async {
    final result = await _client.query(
      QueryOptions(
        document: gql(_workflowJobsQuery),
        variables: <String, dynamic>{
          'runID': runID,
          'taskID': taskID,
          'limit': limit,
        },
      ),
    );
    final error = _extractOperationError(result, field: 'workflowJobs');
    if (error != null) {
      return ApiResult<List<WorkflowJob>>.failure(error);
    }
    final node = _fieldNode(result.data, 'workflowJobs');
    if (node == null) {
      return const ApiResult<List<WorkflowJob>>.failure(
        'workflowJobs returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<List<WorkflowJob>>.failure(_graphErrorMessage(node));
    }
    final jobsNode = node['jobs'] as List<dynamic>? ?? const <dynamic>[];
    final items = jobsNode
        .whereType<Map<dynamic, dynamic>>()
        .map((entry) => WorkflowJob.fromMap(Map<String, dynamic>.from(entry)))
        .toList(growable: false);
    return ApiResult<List<WorkflowJob>>.success(items);
  }

  Future<ApiResult<List<SupervisorDecision>>> supervisorHistory({
    required String runID,
    required String taskID,
    required String jobID,
  }) async {
    final result = await _client.query(
      QueryOptions(
        document: gql(_supervisorHistoryQuery),
        variables: <String, dynamic>{
          'runID': runID,
          'taskID': taskID,
          'jobID': jobID,
        },
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'supervisorDecisionHistory',
    );
    if (error != null) {
      return ApiResult<List<SupervisorDecision>>.failure(error);
    }
    final node = _fieldNode(result.data, 'supervisorDecisionHistory');
    if (node == null) {
      return const ApiResult<List<SupervisorDecision>>.failure(
        'supervisorDecisionHistory returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<List<SupervisorDecision>>.failure(
        _graphErrorMessage(node),
      );
    }
    final decisionsNode =
        node['decisions'] as List<dynamic>? ?? const <dynamic>[];
    final items = decisionsNode
        .whereType<Map<dynamic, dynamic>>()
        .map(
          (entry) =>
              SupervisorDecision.fromMap(Map<String, dynamic>.from(entry)),
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
    final result = await _client.mutate(
      MutationOptions(
        document: gql(_enqueueScmWorkflowMutation),
        variables: <String, dynamic>{
          'input': <String, dynamic>{
            'operation': 'SOURCE_STATE',
            'provider': 'GITHUB',
            'owner': owner,
            'repository': repository,
            'runID': runID,
            'taskID': taskID,
            'jobID': jobID,
            'idempotencyKey': idempotencyKey,
          },
        },
      ),
    );
    final error = _extractOperationError(result, field: 'enqueueScmWorkflow');
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final node = _fieldNode(result.data, 'enqueueScmWorkflow');
    if (node == null) {
      return const ApiResult<String>.failure(
        'enqueueScmWorkflow returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<String>.failure(_graphErrorMessage(node));
    }
    return ApiResult<String>.success(
      node['queueTaskID'] as String? ?? 'queued',
    );
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
    final result = await _client.mutate(
      MutationOptions(
        document: gql(_enqueueIngestionMutation),
        variables: <String, dynamic>{
          'input': <String, dynamic>{
            'runID': runID,
            'taskID': taskID,
            'jobID': jobID,
            'idempotencyKey': idempotencyKey,
            'prompt': prompt,
            'projectID': projectID,
            'workflowID': workflowID,
            'boardSource': <String, dynamic>{
              'kind': 'GITHUB_ISSUES',
              'location': source,
            },
          },
        },
      ),
    );
    final error = _extractOperationError(
      result,
      field: 'enqueueIngestionWorkflow',
    );
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final node = _fieldNode(result.data, 'enqueueIngestionWorkflow');
    if (node == null) {
      return const ApiResult<String>.failure(
        'enqueueIngestionWorkflow returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<String>.failure(_graphErrorMessage(node));
    }
    return ApiResult<String>.success(
      node['queueTaskID'] as String? ?? 'queued',
    );
  }

  Future<ApiResult<String>> approveIssueIntake({
    required String runID,
    required String taskID,
    required String jobID,
    required String source,
    required String issueReference,
    required String approvedBy,
  }) async {
    final result = await _client.mutate(
      MutationOptions(
        document: gql(_approveIssueMutation),
        variables: <String, dynamic>{
          'input': <String, dynamic>{
            'runID': runID,
            'taskID': taskID,
            'jobID': jobID,
            'source': source,
            'issueReference': issueReference,
            'approvedBy': approvedBy,
          },
        },
      ),
    );
    final error = _extractOperationError(result, field: 'approveIssueIntake');
    if (error != null) {
      return ApiResult<String>.failure(error);
    }
    final node = _fieldNode(result.data, 'approveIssueIntake');
    if (node == null) {
      return const ApiResult<String>.failure(
        'approveIssueIntake returned no data',
      );
    }
    if (node['__typename'] == 'GraphError') {
      return ApiResult<String>.failure(_graphErrorMessage(node));
    }
    final decision =
        node['decision'] as Map<dynamic, dynamic>? ??
        const <dynamic, dynamic>{};
    return ApiResult<String>.success(
      '${decision['action'] ?? 'UNKNOWN'} (${decision['reason'] ?? 'no reason'})',
    );
  }

  Stream<ApiResult<StreamEvent>> sessionActivityStream({
    required String runID,
    int fromOffset = 0,
  }) {
    return _client
        .subscribe(
          SubscriptionOptions(
            document: gql(_sessionActivitySubscription),
            variables: <String, dynamic>{
              'runID': runID,
              'taskID': '',
              'jobID': '',
              'fromOffset': fromOffset,
            },
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
          final node = _fieldNode(result.data, 'sessionActivityStream');
          if (node == null) {
            return const ApiResult<StreamEvent>.failure(
              'sessionActivityStream returned no data',
            );
          }
          if (node['__typename'] == 'GraphError') {
            return ApiResult<StreamEvent>.failure(_graphErrorMessage(node));
          }
          final eventNode = node['event'] as Map<dynamic, dynamic>?;
          if (eventNode == null) {
            return const ApiResult<StreamEvent>.failure(
              'stream event payload missing',
            );
          }
          final event = StreamEvent.fromMap(
            Map<String, dynamic>.from(eventNode),
          );
          return ApiResult<StreamEvent>.success(event);
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

  Map<String, dynamic>? _fieldNode(Map<String, dynamic>? data, String field) {
    if (data == null) {
      return null;
    }
    final value = data[field];
    if (value is Map<String, dynamic>) {
      return value;
    }
    if (value is Map<dynamic, dynamic>) {
      return Map<String, dynamic>.from(value);
    }
    return null;
  }

  String _graphErrorMessage(Map<String, dynamic> node) {
    final code = node['code'] as String? ?? 'ERROR';
    final message = node['message'] as String? ?? 'Unknown GraphQL error';
    final field = node['field'] as String?;
    if (field == null || field.isEmpty) {
      return '$code: $message';
    }
    return '$code ($field): $message';
  }
}

GraphQLClient buildGraphqlClient(String httpEndpoint) {
  AppLogger.instance.logger.i(
    'Building GraphQL client',
    error: {'http_endpoint': httpEndpoint},
  );
  final parsed = Uri.parse(httpEndpoint);
  final wsScheme = parsed.scheme == 'https' ? 'wss' : 'ws';
  final wsUri = parsed.replace(scheme: wsScheme);
  final httpLink = HttpLink(httpEndpoint);
  final wsLink = WebSocketLink(
    wsUri.toString(),
    config: const SocketClientConfig(
      autoReconnect: true,
      inactivityTimeout: Duration(minutes: 2),
    ),
  );
  final link = Link.split(
    (request) => request.isSubscription,
    wsLink,
    httpLink,
  );

  return GraphQLClient(
    cache: GraphQLCache(store: InMemoryStore()),
    link: link,
    defaultPolicies: DefaultPolicies(
      query: Policies(fetch: FetchPolicy.networkOnly),
      mutate: Policies(fetch: FetchPolicy.networkOnly),
      subscribe: Policies(fetch: FetchPolicy.noCache),
    ),
  );
}

String prettyJson(String raw) {
  final trimmed = raw.trim();
  if (trimmed.isEmpty) {
    return '{}';
  }
  try {
    final decoded = jsonDecode(trimmed);
    const encoder = JsonEncoder.withIndent('  ');
    return encoder.convert(decoded);
  } catch (_) {
    return raw;
  }
}
