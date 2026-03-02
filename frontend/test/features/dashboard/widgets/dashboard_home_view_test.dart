import 'package:agentic_worktrees/features/dashboard/widgets/dashboard_home_view.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:graphql/client.dart';

import '../../../support/test_data.dart';

class _StubControlPlaneApi extends ControlPlaneApi {
  _StubControlPlaneApi()
    : super(
        GraphQLClient(
          link: HttpLink('http://localhost/query'),
          cache: GraphQLCache(),
        ),
      );

  ApiResult<List<SessionSummary>> sessionsResult =
      ApiResult<List<SessionSummary>>.success(<SessionSummary>[
        sampleSession(),
      ]);
  ApiResult<List<WorkerSession>> workerSessionsResult =
      ApiResult<List<WorkerSession>>.success(<WorkerSession>[]);

  @override
  Future<ApiResult<List<SessionSummary>>> sessions({int limit = 50}) async {
    return sessionsResult;
  }

  @override
  Future<ApiResult<List<WorkerSession>>> workerSessions({
    int limit = 100,
  }) async {
    return workerSessionsResult;
  }
}

void main() {
  late _StubControlPlaneApi api;
  late TextEditingController sourceController;
  late TextEditingController issueController;
  late TextEditingController approvedByController;
  late TextEditingController projectController;
  late TextEditingController workflowController;
  late TextEditingController promptController;
  late TextEditingController scmOwnerController;
  late TextEditingController scmRepoController;

  setUp(() {
    api = _StubControlPlaneApi();
    sourceController = TextEditingController(text: 'acme/repo');
    issueController = TextEditingController(text: 'acme/repo#1');
    approvedByController = TextEditingController(text: 'operator');
    projectController = TextEditingController(text: 'project-1');
    workflowController = TextEditingController(text: 'workflow-1');
    promptController = TextEditingController(text: 'prompt');
    scmOwnerController = TextEditingController(text: 'acme');
    scmRepoController = TextEditingController(text: 'repo');
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
            onJobSelected: (_) {},
            onEnqueueIngestion: () {},
            onApproveIssue: () {},
            onEnqueueScm: () {},
            onShowWorkerSessions: () {},
            onCreateProject: () {},
          ),
        ),
      ),
    );
  }

  testWidgets('renders stat cards and configured projects list', (
    WidgetTester tester,
  ) async {
    await pumpSubject(tester, selectedSession: null, onSessionSelected: (_) {});

    await tester.pumpAndSettle();

    expect(find.text('Summary'), findsOneWidget);
    expect(find.text('Sessions'), findsOneWidget);
    expect(find.text('Jobs'), findsOneWidget);
    expect(find.text('Activity'), findsOneWidget);
    expect(find.text('Select a Project'), findsOneWidget);
  });

  testWidgets('traverses into sessions card and selects a session', (
    WidgetTester tester,
  ) async {
    SessionSummary? selected;

    await pumpSubject(
      tester,
      selectedSession: sampleSession(),
      onSessionSelected: (SessionSummary s) => selected = s,
    );

    await tester.pumpAndSettle();

    await tester.tap(find.text('Sessions'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('run-1').last);
    await tester.pumpAndSettle();

    expect(selected?.runID, 'run-1');
  });
}
