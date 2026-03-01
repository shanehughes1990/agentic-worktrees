import 'package:agentic_worktrees/features/dashboard/widgets/session_detail_panel.dart';
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

    when(api.workers(limit: anyNamed('limit'))).thenAnswer(
      (_) async => ApiResult<List<WorkerSummary>>.success(<WorkerSummary>[
        sampleWorkerSummary(),
      ]),
    );
    when(
      api.workflowJobs(
        runID: anyNamed('runID'),
        taskID: anyNamed('taskID'),
        limit: anyNamed('limit'),
      ),
    ).thenAnswer(
      (_) async => ApiResult<List<WorkflowJob>>.success(<WorkflowJob>[
        sampleWorkflowJob(),
      ]),
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
    required VoidCallback onEnqueueIngestion,
    required VoidCallback onApproveIssue,
    required VoidCallback onEnqueueScm,
    required ValueChanged<WorkflowJob> onJobSelected,
  }) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: SessionDetailPanel(
            api: api,
            refreshToken: 0,
            session: sampleSession(),
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

  testWidgets('renders sections and action buttons', (
    WidgetTester tester,
  ) async {
    await pumpSubject(
      tester,
      onEnqueueIngestion: () {},
      onApproveIssue: () {},
      onEnqueueScm: () {},
      onJobSelected: (_) {},
    );

    await tester.pumpAndSettle();

    expect(find.textContaining('Session run-1'), findsOneWidget);
    expect(find.text('Control Actions'), findsOneWidget);
    expect(find.text('Enqueue Ingestion'), findsOneWidget);
    expect(find.text('Approve Issue Intake'), findsOneWidget);
    expect(find.text('Enqueue SCM Source State'), findsOneWidget);
  });

  testWidgets('invokes action callbacks when buttons pressed', (
    WidgetTester tester,
  ) async {
    var enqueueIngestion = 0;
    var approveIssue = 0;
    var enqueueScm = 0;

    await pumpSubject(
      tester,
      onEnqueueIngestion: () => enqueueIngestion++,
      onApproveIssue: () => approveIssue++,
      onEnqueueScm: () => enqueueScm++,
      onJobSelected: (_) {},
    );

    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Enqueue Ingestion'));
    await tester.tap(find.text('Enqueue Ingestion'));
    await tester.pump();
    await tester.ensureVisible(find.text('Approve Issue Intake'));
    await tester.tap(find.text('Approve Issue Intake'));
    await tester.pump();
    await tester.ensureVisible(find.text('Enqueue SCM Source State'));
    await tester.tap(find.text('Enqueue SCM Source State'));
    await tester.pump();

    expect(enqueueIngestion, 1);
    expect(approveIssue, 1);
    expect(enqueueScm, 1);
  });
}
