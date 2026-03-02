import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';

class DashboardWorkflowLogic {
  const DashboardWorkflowLogic._();

  static String compactError(String? message) {
    const fallback = 'unknown error';
    final raw = (message ?? fallback).trim();
    if (raw.isEmpty) {
      return fallback;
    }
    final firstLine = raw.split('\n').first.trim();
    if (firstLine.length <= 180) {
      return firstLine;
    }
    return '${firstLine.substring(0, 177)}...';
  }

  static String scmIdempotencyKey(DateTime now) {
    return 'scm-${now.millisecondsSinceEpoch}';
  }

  static void appendStreamEvent(
    List<StreamEvent> streamEvents,
    StreamEvent event, {
    int maxEvents = 100,
  }) {
    streamEvents.insert(0, event);
    if (streamEvents.length > maxEvents) {
      streamEvents.removeRange(maxEvents, streamEvents.length);
    }
  }
}
