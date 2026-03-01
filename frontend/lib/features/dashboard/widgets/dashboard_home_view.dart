import 'package:agentic_worktrees/features/dashboard/widgets/session_detail_panel.dart';
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

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        if (statusMessage != null)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Align(
              alignment: Alignment.centerLeft,
              child: Text(
                statusMessage!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ),
        Expanded(
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 330,
                child: Column(
                  children: <Widget>[
                    Card(
                      margin: const EdgeInsets.fromLTRB(12, 12, 12, 6),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            const ListTile(
                              dense: true,
                              title: Text('Configured Projects'),
                            ),
                            ProjectSetupsList(
                              dense: true,
                              projectSetups: projectSetups,
                              selectedProjectID: selectedProjectID,
                              onProjectSelected: onProjectSelected,
                            ),
                          ],
                        ),
                      ),
                    ),
                    Expanded(
                      child: Card(
                        margin: const EdgeInsets.fromLTRB(12, 6, 12, 12),
                        child: FutureBuilder<ApiResult<List<SessionSummary>>>(
                          future: api.sessions(limit: 50 + refreshToken),
                          builder:
                              (
                                BuildContext context,
                                AsyncSnapshot<ApiResult<List<SessionSummary>>>
                                snapshot,
                              ) {
                                if (!snapshot.hasData) {
                                  return const Center(
                                    child: CircularProgressIndicator(),
                                  );
                                }
                                final value = snapshot.data!;
                                if (!value.isSuccess || value.data == null) {
                                  return Center(
                                    child: Text(
                                      value.errorMessage ??
                                          'Failed loading sessions',
                                    ),
                                  );
                                }
                                final sessions = value.data!;
                                if (sessions.isEmpty) {
                                  return const Center(
                                    child: Text('No sessions found.'),
                                  );
                                }
                                return ListView.builder(
                                  itemCount: sessions.length,
                                  itemBuilder: (BuildContext context, int index) {
                                    final item = sessions[index];
                                    final selected =
                                        selectedSession?.runID == item.runID;
                                    return ListTile(
                                      selected: selected,
                                      title: Text(item.runID),
                                      subtitle: Text(
                                        'tasks: ${item.taskCount} jobs: ${item.jobCount}\nupdated: ${item.updatedAt}',
                                      ),
                                      onTap: () => onSessionSelected(item),
                                    );
                                  },
                                );
                              },
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: selectedSession == null
                    ? const Center(
                        child: Text('Select a session to view details.'),
                      )
                    : SessionDetailPanel(
                        api: api,
                        refreshToken: refreshToken,
                        session: selectedSession!,
                        selectedJob: selectedJob,
                        streamEvents: streamEvents,
                        sourceController: sourceController,
                        issueReferenceController: issueReferenceController,
                        approvedByController: approvedByController,
                        projectController: projectController,
                        workflowController: workflowController,
                        promptController: promptController,
                        scmOwnerController: scmOwnerController,
                        scmRepoController: scmRepoController,
                        isRunningAction: isRunningAction,
                        onJobSelected: onJobSelected,
                        onEnqueueIngestion: onEnqueueIngestion,
                        onApproveIssue: onApproveIssue,
                        onEnqueueScm: onEnqueueScm,
                      ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
