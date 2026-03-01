import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';

DateTime get fixedTime => DateTime.parse('2026-03-01T12:00:00Z');

ProjectSetupConfig sampleProjectSetup({
  String projectID = 'project-1',
  String projectName = 'Project One',
}) {
  return ProjectSetupConfig(
    projectID: projectID,
    projectName: projectName,
    scmProvider: 'GITHUB',
    repositoryURL: 'https://github.com/acme/repo',
    trackerProvider: 'GITHUB_ISSUES',
    trackerLocation: 'acme/repo',
    trackerBoardID: 'board-1',
    createdAt: fixedTime,
    updatedAt: fixedTime,
  );
}

SessionSummary sampleSession({String runID = 'run-1'}) {
  return SessionSummary(
    runID: runID,
    taskCount: 2,
    jobCount: 3,
    updatedAt: fixedTime,
  );
}

WorkflowJob sampleWorkflowJob({
  String taskID = 'task-1',
  String jobID = 'job-1',
}) {
  return WorkflowJob(
    runID: 'run-1',
    taskID: taskID,
    jobID: jobID,
    jobKind: 'INGESTION',
    status: 'queued',
    queue: 'worker.default',
    queueTaskID: 'queue-1',
    duplicate: false,
    updatedAt: fixedTime,
  );
}

WorkerSummary sampleWorkerSummary({String workerID = 'worker-1'}) {
  return WorkerSummary(
    workerID: workerID,
    capabilities: const <String>['INGESTION', 'SCM'],
    lastHeartbeat: fixedTime,
  );
}

SupervisorDecision sampleSupervisorDecision() {
  return SupervisorDecision(
    signalType: 'TASK_COMPLETED',
    action: 'NONE',
    reason: 'no-op',
    occurredAt: fixedTime,
  );
}

StreamEvent sampleStreamEvent() {
  return StreamEvent(
    eventID: 'event-1',
    eventType: 'TASK_ENQUEUED',
    source: 'supervisor',
    payload: '{"ok":true}',
    occurredAt: fixedTime,
  );
}
