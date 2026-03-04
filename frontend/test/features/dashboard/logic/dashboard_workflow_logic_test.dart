import 'package:agentic_repositories/features/dashboard/logic/dashboard_workflow_logic.dart';
import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('DashboardWorkflowLogic.compactError', () {
    test('returns fallback when null or empty', () {
      expect(DashboardWorkflowLogic.compactError(null), 'unknown error');
      expect(DashboardWorkflowLogic.compactError('   '), 'unknown error');
    });

    test('returns first line when multiline', () {
      final message = DashboardWorkflowLogic.compactError('line one\nline two');

      expect(message, 'line one');
    });

    test('truncates long message', () {
      final longMessage = List<String>.filled(300, 'x').join();
      final compact = DashboardWorkflowLogic.compactError(longMessage);

      expect(compact.length, lessThanOrEqualTo(180));
      expect(compact.endsWith('...'), isTrue);
    });
  });

  group('idempotency keys', () {
    test('scm key includes timestamp', () {
      final key = DashboardWorkflowLogic.scmIdempotencyKey(
        DateTime.fromMillisecondsSinceEpoch(456),
      );

      expect(key, 'scm-456');
    });
  });

  group('appendStreamEvent', () {
    test('prepends newest event and enforces max cap', () {
      final events = <StreamEvent>[
        for (int i = 0; i < 3; i++)
          StreamEvent(
            eventID: 'old-$i',
            eventType: 'TYPE',
            source: 'test',
            payload: '{}',
            occurredAt: DateTime.now(),
          ),
      ];

      final newest = StreamEvent(
        eventID: 'new',
        eventType: 'TYPE',
        source: 'test',
        payload: '{}',
        occurredAt: DateTime.now(),
      );

      DashboardWorkflowLogic.appendStreamEvent(events, newest, maxEvents: 3);

      expect(events.first.eventID, 'new');
      expect(events.length, 3);
    });
  });
}
