import 'dart:async';

import 'package:agentic_repositories/features/projects/screens/project_dashboard_screen.dart';
import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:graphql/client.dart';

class _NoopLink extends Link {
  @override
  Stream<Response> request(Request request, [NextLink? forward]) {
    return const Stream<Response>.empty();
  }
}

class _FakeControlPlaneApi extends ControlPlaneApi {
  _FakeControlPlaneApi()
    : _projectEventsController =
          StreamController<ApiResult<StreamEvent>>.broadcast(),
      super(GraphQLClient(link: _NoopLink(), cache: GraphQLCache()));

  final StreamController<ApiResult<StreamEvent>> _projectEventsController;

  void emitProjectEvent(StreamEvent event) {
    _projectEventsController.add(ApiResult<StreamEvent>.success(event));
  }

  Future<void> close() async {
    await _projectEventsController.close();
  }

  @override
  Stream<ApiResult<StreamEvent>> projectEventsStream({
    required String projectID,
    int fromOffset = 0,
  }) {
    return _projectEventsController.stream;
  }

  @override
  Future<ApiResult<List<LifecycleSessionSnapshotModel>>>
  lifecycleSessionSnapshots({
    required String projectID,
    String? pipelineType,
    int limit = 200,
  }) async {
    return const ApiResult<List<LifecycleSessionSnapshotModel>>.success(
      <LifecycleSessionSnapshotModel>[],
    );
  }
}

StreamEvent _event({
  required String id,
  required int offset,
  required String type,
  required String runID,
  required String taskID,
  required String jobID,
  required String sessionID,
}) {
  return StreamEvent(
    eventID: id,
    streamOffset: offset,
    occurredAt: DateTime.now().toUtc(),
    runID: runID,
    taskID: taskID,
    jobID: jobID,
    projectID: 'project-1',
    sessionID: sessionID,
    source: 'WORKER',
    eventType: type,
    payload: '{"runtime_activity":true}',
  );
}

void main() {
  Future<void> pumpSubject(
    WidgetTester tester,
    _FakeControlPlaneApi api,
  ) async {
    await tester.binding.setSurfaceSize(const Size(1400, 900));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    await tester.pumpWidget(
      MaterialApp(
        home: ProjectEventsMatrixPage(
          api: api,
          projectID: 'project-1',
          projectName: 'Project 1',
        ),
      ),
    );
    await tester.pump();
  }

  testWidgets('shows centered empty state with Active 0', (
    WidgetTester tester,
  ) async {
    final api = _FakeControlPlaneApi();
    addTearDown(api.close);

    await pumpSubject(tester, api);

    final emptyText = find.text('No active worker activity right now.');
    expect(emptyText, findsOneWidget);
    expect(find.text('Active 0'), findsOneWidget);
    expect(
      find.ancestor(of: emptyText, matching: find.byType(Center)),
      findsOneWidget,
    );
  });

  testWidgets(
    'global live adds, updates, and removes active rows by event flow',
    (WidgetTester tester) async {
      final api = _FakeControlPlaneApi();
      addTearDown(api.close);

      await pumpSubject(tester, api);
      await tester.pump(const Duration(milliseconds: 50));

      api.emitProjectEvent(
        _event(
          id: 'evt-started',
          offset: 1,
          type: 'stream.session.started',
          runID: 'run-1',
          taskID: 'task-1',
          jobID: 'job-1',
          sessionID: 'session-1',
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 10));

      expect(find.text('Active 1'), findsOneWidget);
      expect(find.textContaining('stream.session.started'), findsOneWidget);

      api.emitProjectEvent(
        _event(
          id: 'evt-health',
          offset: 2,
          type: 'stream.session.health',
          runID: 'run-1',
          taskID: 'task-1',
          jobID: 'job-1',
          sessionID: 'session-1',
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 10));

      expect(find.text('Active 1'), findsOneWidget);

      api.emitProjectEvent(
        _event(
          id: 'evt-ended',
          offset: 3,
          type: 'stream.session.ended',
          runID: 'run-1',
          taskID: 'task-1',
          jobID: 'job-1',
          sessionID: 'session-1',
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 10));

      expect(find.text('Active 0'), findsOneWidget);
      expect(find.text('No active worker activity right now.'), findsOneWidget);
    },
  );
}
