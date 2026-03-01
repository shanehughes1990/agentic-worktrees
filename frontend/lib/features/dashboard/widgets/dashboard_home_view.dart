import 'package:agentic_worktrees/features/projects/widgets/project_setups_list.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class DashboardHomeView extends StatelessWidget {
  const DashboardHomeView({
    required this.api,
    required this.refreshToken,
    required this.statusMessage,
    required this.projectSetups,
    required this.selectedProjectID,
    required this.onProjectSelected,
    required this.selectedSession,
    required this.onSessionSelected,
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
  final String? statusMessage;
  final List<ProjectSetupConfig> projectSetups;
  final String selectedProjectID;
  final ValueChanged<ProjectSetupConfig> onProjectSelected;
  final SessionSummary? selectedSession;
  final ValueChanged<SessionSummary> onSessionSelected;
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

  Future<_DashboardStatsData> _loadStats() async {
    final sessionsResult = await api.sessions(limit: 50 + refreshToken);
    final workersResult = await api.workers(limit: 50 + refreshToken);

    if (!sessionsResult.isSuccess || sessionsResult.data == null) {
      return _DashboardStatsData.error(
        sessionsResult.errorMessage ?? 'Failed loading sessions',
      );
    }
    if (!workersResult.isSuccess || workersResult.data == null) {
      return _DashboardStatsData.error(
        workersResult.errorMessage ?? 'Failed loading workers',
      );
    }

    return _DashboardStatsData.success(
      sessions: sessionsResult.data!,
      workers: workersResult.data!,
    );
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<_DashboardStatsData>(
      future: _loadStats(),
      builder:
          (BuildContext context, AsyncSnapshot<_DashboardStatsData> snapshot) {
            if (!snapshot.hasData) {
              return const Center(child: CircularProgressIndicator());
            }

            final stats = snapshot.data!;
            if (stats.errorMessage != null) {
              return Center(child: Text(stats.errorMessage!));
            }

            final sessions = stats.sessions;
            final workers = stats.workers;
            final totalJobs = sessions.fold<int>(
              0,
              (int sum, SessionSummary item) => sum + item.jobCount,
            );

            return SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  if (statusMessage != null) ...<Widget>[
                    Text(
                      statusMessage!,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 12),
                  ],
                  Text(
                    'Summary',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Center(
                    child: Wrap(
                      spacing: 12,
                      runSpacing: 12,
                      alignment: WrapAlignment.center,
                      children: <Widget>[
                        _StatCard(
                          icon: Icons.timeline,
                          label: 'Sessions',
                          value: sessions.length.toString(),
                          onTap: () => _showSessionsSheet(context, sessions),
                        ),
                        _StatCard(
                          icon: Icons.memory_outlined,
                          label: 'Workers',
                          value: workers.length.toString(),
                          onTap: () => _showWorkersSheet(context, workers),
                        ),
                        _StatCard(
                          icon: Icons.task_alt_outlined,
                          label: 'Jobs',
                          value: totalJobs.toString(),
                          onTap: () => _showJobsSheet(context, sessions),
                        ),
                        _StatCard(
                          icon: Icons.bolt_outlined,
                          label: 'Activity',
                          value: streamEvents.length.toString(),
                          onTap: () => _showActivitySheet(context),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Configured Projects',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      child: ProjectSetupsList(
                        dense: true,
                        projectSetups: projectSetups,
                        selectedProjectID: selectedProjectID,
                        onProjectSelected: onProjectSelected,
                      ),
                    ),
                  ),
                ],
              ),
            );
          },
    );
  }

  void _showSessionsSheet(BuildContext context, List<SessionSummary> sessions) {
    showModalBottomSheet<void>(
      context: context,
      builder: (BuildContext context) {
        return SafeArea(
          child: ListView(
            children: sessions
                .map((SessionSummary item) {
                  final selected = selectedSession?.runID == item.runID;
                  return ListTile(
                    selected: selected,
                    leading: const Icon(Icons.timeline),
                    title: Text(item.runID),
                    subtitle: Text(
                      'tasks: ${item.taskCount} • jobs: ${item.jobCount}',
                    ),
                    onTap: () {
                      Navigator.of(context).pop();
                      onSessionSelected(item);
                    },
                  );
                })
                .toList(growable: false),
          ),
        );
      },
    );
  }

  void _showWorkersSheet(BuildContext context, List<WorkerSummary> workers) {
    showModalBottomSheet<void>(
      context: context,
      builder: (BuildContext context) {
        return SafeArea(
          child: ListView(
            children: workers
                .map((WorkerSummary item) {
                  return ListTile(
                    leading: const Icon(Icons.memory_outlined),
                    title: Text(item.workerID),
                    subtitle: Text(item.capabilities.join(', ')),
                  );
                })
                .toList(growable: false),
          ),
        );
      },
    );
  }

  void _showJobsSheet(BuildContext context, List<SessionSummary> sessions) {
    showModalBottomSheet<void>(
      context: context,
      builder: (BuildContext context) {
        return SafeArea(
          child: ListView(
            children: sessions
                .map((SessionSummary item) {
                  return ListTile(
                    leading: const Icon(Icons.task_alt_outlined),
                    title: Text(item.runID),
                    trailing: Text('${item.jobCount}'),
                  );
                })
                .toList(growable: false),
          ),
        );
      },
    );
  }

  void _showActivitySheet(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      builder: (BuildContext context) {
        if (streamEvents.isEmpty) {
          return const SafeArea(
            child: Center(child: Text('No realtime activity yet.')),
          );
        }
        return SafeArea(
          child: ListView(
            children: streamEvents
                .map((StreamEvent event) {
                  return ListTile(
                    leading: const Icon(Icons.bolt_outlined),
                    title: Text(event.eventType),
                    subtitle: Text(event.source),
                  );
                })
                .toList(growable: false),
          ),
        );
      },
    );
  }
}

class _DashboardStatsData {
  const _DashboardStatsData.success({
    required this.sessions,
    required this.workers,
  }) : errorMessage = null;

  const _DashboardStatsData.error(this.errorMessage)
    : sessions = const <SessionSummary>[],
      workers = const <WorkerSummary>[];

  final List<SessionSummary> sessions;
  final List<WorkerSummary> workers;
  final String? errorMessage;
}

class _StatCard extends StatelessWidget {
  const _StatCard({
    required this.icon,
    required this.label,
    required this.value,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final String value;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 180,
      child: Card(
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: <Widget>[
                Icon(icon),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        value,
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                      const SizedBox(height: 2),
                      Text(label),
                    ],
                  ),
                ),
                const Icon(Icons.chevron_right),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
