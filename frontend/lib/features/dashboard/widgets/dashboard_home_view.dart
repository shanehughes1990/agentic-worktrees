import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
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
    required this.isRunningAction,
    required this.onJobSelected,
    required this.onApproveIssue,
    required this.onShowWorkerSessions,
    required this.onCreateProject,
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
  final bool isRunningAction;
  final ValueChanged<WorkflowJob> onJobSelected;
  final VoidCallback onApproveIssue;
  final VoidCallback onShowWorkerSessions;
  final VoidCallback onCreateProject;

  Future<_DashboardStatsData> _loadStats() async {
    final sessionsResult = await api.sessions(limit: 50 + refreshToken);
    if (!sessionsResult.isSuccess || sessionsResult.data == null) {
      return _DashboardStatsData.error(
        sessionsResult.errorMessage ?? 'Failed loading sessions',
      );
    }

    final workersResult = await api.workerSessions(limit: 100);
    var healthyWorkerCount = 0;
    if (workersResult.isSuccess && workersResult.data != null) {
      final now = DateTime.now().toUtc();
      healthyWorkerCount = workersResult.data!
          .where(
            (WorkerSession worker) =>
                worker.state.toLowerCase() == 'healthy' &&
                worker.leaseExpiresAt.toUtc().isAfter(now),
          )
          .length;
    }

    return _DashboardStatsData.success(
      sessions: sessionsResult.data!,
      healthyWorkerCount: healthyWorkerCount,
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
            final healthyWorkerCount = stats.healthyWorkerCount;
            final totalJobs = sessions.fold<int>(
              0,
              (int sum, SessionSummary item) => sum + item.jobCount,
            );

            return Padding(
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
                        _StatCard(
                          icon: Icons.memory_outlined,
                          label: 'Workers',
                          value: healthyWorkerCount.toString(),
                          onTap: onShowWorkerSessions,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    children: <Widget>[
                      Expanded(
                        child: Text(
                          'Select a Project',
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                      ),
                      FilledButton.icon(
                        onPressed: onCreateProject,
                        icon: const Icon(Icons.add),
                        label: const Text('Create New Project'),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Expanded(
                    child: Card(
                      child: projectSetups.isEmpty
                          ? Center(
                              child: Text(
                                'No project setups configured.',
                                style: Theme.of(context).textTheme.bodyLarge,
                                textAlign: TextAlign.center,
                              ),
                            )
                          : ListView.builder(
                              padding: const EdgeInsets.all(8),
                              itemCount: projectSetups.length,
                              itemBuilder: (BuildContext context, int index) {
                                final setup = projectSetups[index];
                                final selected =
                                    selectedProjectID == setup.projectID;
                                final repositoryURL =
                                    setup.repositories.isNotEmpty
                                    ? setup.repositories.first.repositoryURL
                                    : 'No repository configured';
                                return Card(
                                  elevation: selected ? 3 : 1,
                                  margin: const EdgeInsets.only(bottom: 8),
                                  child: ListTile(
                                    selected: selected,
                                    title: Text(setup.projectID),
                                    subtitle: Text(
                                      '${setup.projectName}\n$repositoryURL',
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                    trailing: const Icon(Icons.chevron_right),
                                    onTap: () => onProjectSelected(setup),
                                  ),
                                );
                              },
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
    required this.healthyWorkerCount,
  }) : errorMessage = null;

  const _DashboardStatsData.error(this.errorMessage)
    : sessions = const <SessionSummary>[],
      healthyWorkerCount = 0;

  final List<SessionSummary> sessions;
  final int healthyWorkerCount;
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
      height: 84,
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
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 2),
                      Text(label, maxLines: 1, overflow: TextOverflow.ellipsis),
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
