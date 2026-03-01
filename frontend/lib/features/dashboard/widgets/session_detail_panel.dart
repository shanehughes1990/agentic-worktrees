import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class SessionDetailPanel extends StatelessWidget {
  const SessionDetailPanel({
    required this.api,
    required this.refreshToken,
    required this.session,
    required this.selectedJob,
    required this.streamEvents,
    required this.sourceController,
    required this.issueReferenceController,
    required this.approvedByController,
    required this.projectController,
    required this.workflowController,
    required this.promptController,
    required this.scmOwnerController,
    required this.scmRepoController,
    required this.isRunningAction,
    required this.onJobSelected,
    required this.onEnqueueIngestion,
    required this.onApproveIssue,
    required this.onEnqueueScm,
    super.key,
  });

  final ControlPlaneApi api;
  final int refreshToken;
  final SessionSummary session;
  final WorkflowJob? selectedJob;
  final List<StreamEvent> streamEvents;
  final TextEditingController sourceController;
  final TextEditingController issueReferenceController;
  final TextEditingController approvedByController;
  final TextEditingController projectController;
  final TextEditingController workflowController;
  final TextEditingController promptController;
  final TextEditingController scmOwnerController;
  final TextEditingController scmRepoController;
  final bool isRunningAction;
  final ValueChanged<WorkflowJob> onJobSelected;
  final VoidCallback onEnqueueIngestion;
  final VoidCallback onApproveIssue;
  final VoidCallback onEnqueueScm;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Session ${session.runID}',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          FutureBuilder<ApiResult<List<WorkerSummary>>>(
            future: api.workers(limit: 50 + refreshToken),
            builder:
                (
                  BuildContext context,
                  AsyncSnapshot<ApiResult<List<WorkerSummary>>> snapshot,
                ) {
                  if (!snapshot.hasData) {
                    return const LinearProgressIndicator();
                  }
                  final value = snapshot.data!;
                  if (!value.isSuccess || value.data == null) {
                    return Text('Workers error: ${value.errorMessage}');
                  }
                  final workers = value.data!;
                  return Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: workers
                        .map(
                          (WorkerSummary worker) => Chip(
                            label: Text(
                              '${worker.workerID} (${worker.capabilities.join(', ')})',
                            ),
                          ),
                        )
                        .toList(growable: false),
                  );
                },
          ),
          const SizedBox(height: 12),
          FutureBuilder<ApiResult<List<WorkflowJob>>>(
            future: api.workflowJobs(
              runID: session.runID,
              limit: 100 + refreshToken,
            ),
            builder:
                (
                  BuildContext context,
                  AsyncSnapshot<ApiResult<List<WorkflowJob>>> snapshot,
                ) {
                  if (!snapshot.hasData) {
                    return const LinearProgressIndicator();
                  }
                  final value = snapshot.data!;
                  if (!value.isSuccess || value.data == null) {
                    return Text('Workflow jobs error: ${value.errorMessage}');
                  }
                  final jobs = value.data!;
                  if (jobs.isEmpty) {
                    return const Text(
                      'No workflow jobs found for this session.',
                    );
                  }
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: jobs
                        .map(
                          (WorkflowJob job) => Card(
                            child: ListTile(
                              selected: selectedJob?.jobID == job.jobID,
                              title: Text(
                                '${job.taskID}/${job.jobID} • ${job.jobKind}',
                              ),
                              subtitle: Text('${job.status} • ${job.queue}'),
                              trailing: Text(job.updatedAt.toString()),
                              onTap: () => onJobSelected(job),
                            ),
                          ),
                        )
                        .toList(growable: false),
                  );
                },
          ),
          const SizedBox(height: 12),
          if (selectedJob != null)
            FutureBuilder<ApiResult<List<SupervisorDecision>>>(
              future: api.supervisorHistory(
                runID: session.runID,
                taskID: selectedJob!.taskID,
                jobID: selectedJob!.jobID,
              ),
              builder:
                  (
                    BuildContext context,
                    AsyncSnapshot<ApiResult<List<SupervisorDecision>>> snapshot,
                  ) {
                    if (!snapshot.hasData) {
                      return const LinearProgressIndicator();
                    }
                    final value = snapshot.data!;
                    if (!value.isSuccess || value.data == null) {
                      return Text(
                        'Supervisor history error: ${value.errorMessage}',
                      );
                    }
                    final history = value.data!;
                    return ExpansionTile(
                      title: Text('Supervisor History (${history.length})'),
                      children: history
                          .map(
                            (SupervisorDecision decision) => ListTile(
                              dense: true,
                              title: Text(
                                '${decision.signalType} → ${decision.action}',
                              ),
                              subtitle: Text(
                                '${decision.reason} • ${decision.occurredAt}',
                              ),
                            ),
                          )
                          .toList(growable: false),
                    );
                  },
            ),
          const SizedBox(height: 12),
          ExpansionTile(
            initiallyExpanded: true,
            title: Text('Realtime Session Events (${streamEvents.length})'),
            children: streamEvents
                .map(
                  (StreamEvent event) => ListTile(
                    dense: true,
                    title: Text('${event.eventType} • ${event.source}'),
                    subtitle: Text(
                      '${event.occurredAt}\n${prettyJson(event.payload)}',
                    ),
                  ),
                )
                .toList(growable: false),
          ),
          const SizedBox(height: 12),
          Text(
            'Control Actions',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              SizedBox(
                width: 260,
                child: TextField(
                  controller: sourceController,
                  decoration: const InputDecoration(
                    labelText: 'Source (owner/repo)',
                  ),
                ),
              ),
              SizedBox(
                width: 260,
                child: TextField(
                  controller: issueReferenceController,
                  decoration: const InputDecoration(
                    labelText: 'Issue reference',
                  ),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: approvedByController,
                  decoration: const InputDecoration(labelText: 'Approved by'),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: projectController,
                  decoration: const InputDecoration(labelText: 'Project ID'),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: workflowController,
                  decoration: const InputDecoration(labelText: 'Workflow ID'),
                ),
              ),
              SizedBox(
                width: 420,
                child: TextField(
                  controller: promptController,
                  decoration: const InputDecoration(
                    labelText: 'Ingestion prompt',
                  ),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: scmOwnerController,
                  decoration: const InputDecoration(labelText: 'SCM owner'),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: scmRepoController,
                  decoration: const InputDecoration(
                    labelText: 'SCM repository',
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              FilledButton(
                onPressed: isRunningAction ? null : onEnqueueIngestion,
                child: const Text('Enqueue Ingestion'),
              ),
              FilledButton(
                onPressed: isRunningAction ? null : onApproveIssue,
                child: const Text('Approve Issue Intake'),
              ),
              FilledButton(
                onPressed: isRunningAction ? null : onEnqueueScm,
                child: const Text('Enqueue SCM Source State'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
