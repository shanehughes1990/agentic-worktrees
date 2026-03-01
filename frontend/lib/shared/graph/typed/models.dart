import 'dart:convert';

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
}

class ProjectSetupConfig {
  const ProjectSetupConfig({
    required this.projectID,
    required this.projectName,
    required this.scmProvider,
    required this.repositoryURL,
    required this.trackerProvider,
    required this.trackerLocation,
    required this.trackerBoardID,
    required this.createdAt,
    required this.updatedAt,
  });

  final String projectID;
  final String projectName;
  final String scmProvider;
  final String repositoryURL;
  final String trackerProvider;
  final String trackerLocation;
  final String trackerBoardID;
  final DateTime createdAt;
  final DateTime updatedAt;
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
}

class WorkerSession {
  const WorkerSession({
    required this.workerID,
    required this.epoch,
    required this.state,
    required this.desiredState,
    required this.lastHeartbeat,
    required this.leaseExpiresAt,
    required this.rogueReason,
    required this.updatedAt,
  });

  final String workerID;
  final int epoch;
  final String state;
  final String desiredState;
  final DateTime lastHeartbeat;
  final DateTime leaseExpiresAt;
  final String? rogueReason;
  final DateTime updatedAt;
}

class WorkerSettings {
  const WorkerSettings({
    required this.heartbeatIntervalSeconds,
    required this.responseDeadlineSeconds,
    required this.staleAfterSeconds,
    required this.drainTimeoutSeconds,
    required this.terminateTimeoutSeconds,
    required this.rogueThreshold,
    required this.updatedAt,
  });

  final int heartbeatIntervalSeconds;
  final int responseDeadlineSeconds;
  final int staleAfterSeconds;
  final int drainTimeoutSeconds;
  final int terminateTimeoutSeconds;
  final int rogueThreshold;
  final DateTime updatedAt;
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
