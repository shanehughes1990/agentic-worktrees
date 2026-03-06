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

class ProjectRepositoryConfig {
  const ProjectRepositoryConfig({
    required this.repositoryID,
    required this.scmID,
    required this.repositoryURL,
    required this.isPrimary,
  });

  final String repositoryID;
  final String scmID;
  final String repositoryURL;
  final bool isPrimary;
}

class ProjectRepositoryBranchOption {
  const ProjectRepositoryBranchOption({
    required this.repositoryID,
    required this.repositoryURL,
    required this.defaultBranch,
    required this.branches,
  });

  final String repositoryID;
  final String repositoryURL;
  final String? defaultBranch;
  final List<String> branches;
}

class ProjectScmConfig {
  const ProjectScmConfig({required this.scmID, required this.scmProvider});

  final String scmID;
  final String scmProvider;
}

class ProjectBoardConfig {
  const ProjectBoardConfig({
    required this.boardID,
    required this.trackerProvider,
    required this.taskboardName,
    required this.appliesToAllRepositories,
    required this.repositoryIDs,
  });

  final String boardID;
  final String trackerProvider;
  final String taskboardName;
  final bool appliesToAllRepositories;
  final List<String> repositoryIDs;
}

class ProjectSetupConfig {
  const ProjectSetupConfig({
    required this.projectID,
    required this.projectName,
    required this.scms,
    required this.repositories,
    required this.boards,
    required this.createdAt,
    required this.updatedAt,
  });

  final String projectID;
  final String projectName;
  final List<ProjectScmConfig> scms;
  final List<ProjectRepositoryConfig> repositories;
  final List<ProjectBoardConfig> boards;
  final DateTime createdAt;
  final DateTime updatedAt;
}

class ProjectDocument {
  const ProjectDocument({
    required this.projectID,
    required this.documentID,
    required this.fileName,
    required this.contentType,
    required this.objectPath,
    required this.cdnURL,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
  });

  final String projectID;
  final String documentID;
  final String fileName;
  final String contentType;
  final String objectPath;
  final String cdnURL;
  final String status;
  final DateTime createdAt;
  final DateTime updatedAt;
}

class ProjectDocumentUploadTicket {
  const ProjectDocumentUploadTicket({
    required this.requestID,
    required this.projectID,
    required this.documentID,
    required this.fileName,
    required this.contentType,
    required this.objectPath,
    required this.uploadURL,
    required this.cdnURL,
    required this.expiresAt,
    required this.status,
  });

  final String requestID;
  final String projectID;
  final String documentID;
  final String fileName;
  final String contentType;
  final String objectPath;
  final String uploadURL;
  final String cdnURL;
  final DateTime expiresAt;
  final String status;
}

class IngestionRunTicket {
  const IngestionRunTicket({
    required this.runID,
    required this.taskID,
    required this.jobID,
    required this.queueTaskID,
    required this.duplicate,
  });

  final String runID;
  final String taskID;
  final String jobID;
  final String queueTaskID;
  final bool duplicate;
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
    required this.streamOffset,
    required this.eventType,
    required this.source,
    required this.payload,
    required this.occurredAt,
    this.runID,
    this.taskID,
    this.jobID,
    this.projectID,
    this.sessionID,
    this.gapDetected = false,
    this.gapReconciled = false,
    this.expectedEventSeq,
    this.observedEventSeq,
  });

  final String eventID;
  final int streamOffset;
  final String eventType;
  final String source;
  final String payload;
  final DateTime occurredAt;
  final String? runID;
  final String? taskID;
  final String? jobID;
  final String? projectID;
  final String? sessionID;
  final bool gapDetected;
  final bool gapReconciled;
  final int? expectedEventSeq;
  final int? observedEventSeq;
}

class LifecycleSessionSnapshotModel {
  const LifecycleSessionSnapshotModel({
    required this.projectID,
    required this.sessionID,
    required this.pipelineType,
    required this.currentState,
    required this.currentSeverity,
    required this.lastEventSeq,
    required this.lastProjectEventSeq,
    required this.startedAt,
    required this.updatedAt,
    this.runID,
    this.taskID,
    this.jobID,
    this.sourceRuntime,
    this.lastReasonCode,
    this.lastReasonSummary,
    this.lastLivenessAt,
    this.lastActivityAt,
    this.lastCheckpointAt,
    this.endedAt,
  });

  final String projectID;
  final String sessionID;
  final String pipelineType;
  final String currentState;
  final String currentSeverity;
  final int lastEventSeq;
  final int lastProjectEventSeq;
  final DateTime startedAt;
  final DateTime updatedAt;
  final String? runID;
  final String? taskID;
  final String? jobID;
  final String? sourceRuntime;
  final String? lastReasonCode;
  final String? lastReasonSummary;
  final DateTime? lastLivenessAt;
  final DateTime? lastActivityAt;
  final DateTime? lastCheckpointAt;
  final DateTime? endedAt;
}

class LifecycleHistoryEventModel {
  const LifecycleHistoryEventModel({
    required this.eventID,
    required this.projectID,
    required this.sessionID,
    required this.pipelineType,
    required this.sourceRuntime,
    required this.eventType,
    required this.eventSeq,
    required this.projectEventSeq,
    required this.occurredAt,
    required this.payload,
    this.runID,
    this.taskID,
    this.jobID,
  });

  final String eventID;
  final String projectID;
  final String sessionID;
  final String pipelineType;
  final String sourceRuntime;
  final String eventType;
  final int eventSeq;
  final int projectEventSeq;
  final DateTime occurredAt;
  final String payload;
  final String? runID;
  final String? taskID;
  final String? jobID;
}

class LifecycleTreeNodeModel {
  const LifecycleTreeNodeModel({
    required this.nodeID,
    required this.nodeType,
    required this.projectID,
    required this.sessionCount,
    required this.updatedAt,
    this.parentNodeID,
    this.runID,
    this.taskID,
    this.jobID,
    this.sessionID,
    this.pipelineType,
    this.sourceRuntime,
    this.currentState,
    this.currentSeverity,
  });

  final String nodeID;
  final String nodeType;
  final String projectID;
  final int sessionCount;
  final DateTime updatedAt;
  final String? parentNodeID;
  final String? runID;
  final String? taskID;
  final String? jobID;
  final String? sessionID;
  final String? pipelineType;
  final String? sourceRuntime;
  final String? currentState;
  final String? currentSeverity;
}

class InterventionMetricsModel {
  const InterventionMetricsModel({
    required this.projectID,
    required this.interventionCount,
    required this.successfulOutcomeCount,
    required this.failedOutcomeCount,
    required this.averageRecoverySeconds,
  });

  final String projectID;
  final int interventionCount;
  final int successfulOutcomeCount;
  final int failedOutcomeCount;
  final int averageRecoverySeconds;
}

class WorkerSession {
  const WorkerSession({
    required this.workerID,
    required this.epoch,
    required this.state,
    required this.lastHeartbeat,
    required this.leaseExpiresAt,
    required this.updatedAt,
  });

  final String workerID;
  final int epoch;
  final String state;
  final DateTime lastHeartbeat;
  final DateTime leaseExpiresAt;
  final DateTime updatedAt;
}

class WorkerSettings {
  const WorkerSettings({
    required this.heartbeatIntervalSeconds,
    required this.responseDeadlineSeconds,
    required this.updatedAt,
  });

  final int heartbeatIntervalSeconds;
  final int responseDeadlineSeconds;
  final DateTime updatedAt;
}

class TaskboardModel {
  const TaskboardModel({
    required this.boardID,
    required this.projectID,
    required this.name,
    required this.state,
    required this.epics,
    required this.ingestionAudits,
    required this.ingestionFilesAdded,
    required this.ingestionUserPrompt,
    required this.createdAt,
    required this.updatedAt,
  });

  final String boardID;
  final String projectID;
  final String name;
  final String state;
  final List<TaskboardEpicModel> epics;
  final List<TaskModelAuditModel> ingestionAudits;
  final List<String> ingestionFilesAdded;
  final String? ingestionUserPrompt;
  final DateTime createdAt;
  final DateTime updatedAt;
}

class TaskModelAuditModel {
  const TaskModelAuditModel({
    required this.modelProvider,
    required this.modelName,
    required this.modelVersion,
    required this.modelRunID,
    required this.agentSessionID,
    required this.agentStreamID,
    required this.promptFingerprint,
    required this.inputTokens,
    required this.outputTokens,
    required this.startedAt,
    required this.completedAt,
  });

  final String modelProvider;
  final String modelName;
  final String? modelVersion;
  final String? modelRunID;
  final String? agentSessionID;
  final String? agentStreamID;
  final String? promptFingerprint;
  final int? inputTokens;
  final int? outputTokens;
  final DateTime? startedAt;
  final DateTime? completedAt;
}

class TaskboardEpicModel {
  const TaskboardEpicModel({
    required this.id,
    required this.boardID,
    required this.title,
    required this.objective,
    required this.repositoryIDs,
    required this.deliverables,
    required this.state,
    required this.rank,
    required this.dependsOnEpicIDs,
    required this.tasks,
  });

  final String id;
  final String boardID;
  final String title;
  final String? objective;
  final List<String> repositoryIDs;
  final List<String> deliverables;
  final String state;
  final int rank;
  final List<String> dependsOnEpicIDs;
  final List<TaskboardTaskModel> tasks;
}

class TaskboardTaskModel {
  const TaskboardTaskModel({
    required this.id,
    required this.boardID,
    required this.epicID,
    required this.title,
    required this.description,
    required this.repositoryIDs,
    required this.deliverables,
    required this.taskType,
    required this.state,
    required this.rank,
    required this.dependsOnTaskIDs,
    required this.audits,
  });

  final String id;
  final String boardID;
  final String epicID;
  final String title;
  final String? description;
  final List<String> repositoryIDs;
  final List<String> deliverables;
  final String taskType;
  final String state;
  final int rank;
  final List<String> dependsOnTaskIDs;
  final List<TaskModelAuditModel> audits;
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
