import 'package:agentic_worktrees/features/dashboard/widgets/dashboard_home_view.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mockito/mockito.dart';

import '../../../support/mocks.mocks.dart';
import '../../../support/test_data.dart';

void main() {
  late MockControlPlaneApi api;
  late TextEditingController sourceController;
  late TextEditingController issueController;
  late TextEditingController approvedByController;
  late TextEditingController projectController;
  late TextEditingController workflowController;
  late TextEditingController promptController;
  late TextEditingController scmOwnerController;
  late TextEditingController scmRepoController;

  setUp(() {
    api = MockControlPlaneApi();
    sourceController = TextEditingController(text: 'acme/repo');
    issueController = TextEditingController(text: 'acme/repo#1');
    approvedByController = TextEditingController(text: 'operator');
    projectController = TextEditingController(text: 'project-1');
    workflowController = TextEditingController(text: 'workflow-1');
    promptController = TextEditingController(text: 'prompt');
    scmOwnerController = TextEditingController(text: 'acme');
    scmRepoController = TextEditingController(text: 'repo');

    when(api.sessions(limit: anyNamed('limit'))).thenAnswer(
      (_) async => ApiResult<List<SessionSummary>>.success(
        <SessionSummary>[sampleSession()],
      ),
    );
    when(api.workers(limit: anyNamed('limit'))).thenAnswer(
      (_) async => ApiResult<List<WorkerSummary>>.success(
        <WorkerSummary>[sampleWorkerSummary()],
      ),
    );
    when(
      api.workflowJobs(
        runID: anyNamed('runID'),
        taskID: anyNamed('taskID'),
        limit: anyNamed('limit'),
      ),
    ).thenAnswer(
      (_) async => ApiResult<List<WorkflowJob>>.success(
        <WorkflowJob>[sampleWorkflowJob()],
      ),
    );
    when(
      api.supervisorHistory(
        runID: anyNamed('runID'),
        taskID: anyNamed('taskID'),
        jobID: anyNamed('jobID'),
      ),
    ).thenAnswer(
      (_) async => ApiResult<List<SupervisorDecision>>.success(
        <SupervisorDecision>[sampleSupervisorDecision()],
      ),
    );
  });

  tearDown(() {
    sourceController.dispose();
    issueController.dispose();
    approvedByController.dispose();
    projectController.dispose();
    workflowController.dispose();
    promptController.dispose();
    scmOwnerController.dispose();
    scmRepoController.dispose();
  });

  Future<void> pumpSubject(
    WidgetTester tester, {
    SessionSummary? selectedSession,
    required ValueChanged<SessionSummary> onSessionSelected,
    required ValueChanged<WorkflowJob> onJobSelected,
    required VoidCallback onEnqueueIngestion,
    required VoidCallback onApproveIssue,
    required VoidCallback onEnqueueScm,
  }) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: DashboardHomeView(
            api: api,
            refreshToken: 0,
            statusMessage: 'ready',
            projectSetups: <ProjectSetupConfig>[sampleProjectSetup()],
            selectedProjectID: 'project-1',
            onProjectSelected: (_) {},
            selectedSession: selectedSession,
            onSessionSelected: onSessionSelected,
            selectedJob: sampleWorkflowJob(),
            streamEvents: <StreamEvent>[sampleStreamEvent()],
            sourceController: sourceController,
            issueReferenceController: issueController,
            approvedByController: approvedByController,
            projectController: projectController,
            workflowController: workflowController,
            promptController: promptController,
            scmOwnerController: scmOwnerController,
            scmRepoController: scmRepoController,
            isRunningAction: false,
            onJobSelected: onJobSelected,
            onEnqueueIngestion: onEnqueueIngestion,
            onApproveIssue: onApproveIssue,
            onEnqueueScm: onEnqueueScm,
          ),
        ),
      ),
    );
  }

  testWidgets('shows prompt when no selected session', (WidgetTester tester) async {
    await pumpSubject(
      tester,
      selectedSession: null,
      onSessionSelected: (_) {},
      onJobSelected: (_) {},
      onEnqueueIngestion: () {},
      onApproveIssue: () {},
      onEnqueueScm: () {},
    );

    await tester.pumpAndSettle();

    expect(find.text('Configured Projects'), findsOneWidget);
    expect(find.text('Select a session to view details.'), findsOneWidget);
  });

  testWidgets('calls onSessionSelected when a session row is tapped', (
    WidgetTester tester,
  ) async {
    SessionSummary? selected;

    await pumpSubject(
      tester,
      selectedSession: sampleSession(),
      onSessionSelected: (SessionSummary s) => selected = s,
      onJobSelected: (_) {},
      onEnqueueIngestion: () {},
      onApproveIssue: () {},
      onEnqueueScm: () {},
    );

    await tester.pumpAndSettle();
    await tester.tap(find.text('run-1').first);
    await tester.pump();

    expect(selected?.runID, 'run-1');
  });
}
