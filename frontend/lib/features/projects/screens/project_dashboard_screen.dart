import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:agentic_repositories/features/projects/screens/project_setup_edit_screen.dart';
import 'package:agentic_repositories/features/projects/screens/taskboard_management_screen.dart';
import 'package:agentic_repositories/features/workers/screens/worker_sessions_screen.dart';
import 'package:agentic_repositories/features/workers/screens/worker_settings_screen.dart';
import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:agentic_repositories/shared/logging/app_logger.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mime/mime.dart';

enum _EventsBoardMode { globalLive, pipelineDrilldown, sessionInspection }

enum _EventSeverity { info, warning, terminal, success }

class _RealtimeSummaryEntry {
  const _RealtimeSummaryEntry({required this.event, required this.receivedAt});

  final StreamEvent event;
  final DateTime receivedAt;
}

class ProjectDashboardScreen extends StatefulWidget {
  const ProjectDashboardScreen({
    required this.projectSetup,
    required this.endpoint,
    super.key,
  });

  final ProjectSetupConfig projectSetup;
  final String endpoint;

  @override
  State<ProjectDashboardScreen> createState() => _ProjectDashboardScreenState();
}

class _ProjectDashboardScreenState extends State<ProjectDashboardScreen> {
  late final ControlPlaneApi _api;
  late ProjectSetupConfig _projectSetup;
  StreamSubscription<ApiResult<StreamEvent>>? _taskboardSubscription;
  StreamSubscription<ApiResult<StreamEvent>>? _projectEventsSubscription;
  Timer? _projectEventsReconnectTimer;
  Timer? _dashboardActivePruneTimer;
  int _nextProjectEventsOffset = 0;
  List<ProjectDocument> _projectDocuments = const <ProjectDocument>[];
  List<TaskboardModel> _taskboards = const <TaskboardModel>[];
  List<StreamEvent> _projectEvents = const <StreamEvent>[];
  final List<_RealtimeSummaryEntry> _realtimeSummaryFeed =
      <_RealtimeSummaryEntry>[];
  int _realtimeSummaryUnread = 0;
  static const Duration _dashboardActiveEntryTtl = Duration(seconds: 45);
  final Map<String, _LiveActiveEntry> _dashboardActiveEntriesByKey =
      <String, _LiveActiveEntry>{};
  final TextEditingController _eventsFilterController = TextEditingController();
  final TextEditingController _pipelineRunIDController =
      TextEditingController();
  final TextEditingController _pipelineTaskIDController =
      TextEditingController();
  final TextEditingController _pipelineJobIDController =
      TextEditingController();
  bool _isUploadingFiles = false;
  bool _isCreatingTaskboard = false;
  bool _isRefreshingProjectSetup = false;
  final Set<String> _deletingDocumentIDs = <String>{};
  String? _statusMessage;

  @override
  void initState() {
    super.initState();
    _api = ControlPlaneApi(buildGraphqlClient(widget.endpoint));
    _projectSetup = widget.projectSetup;
    _startTaskboardSubscription();
    _startProjectEventsSubscription();
    _startDashboardActivePruneTicker();
    unawaited(_refreshProjectSetup(silent: true));
    unawaited(_loadProjectDocuments());
    unawaited(_loadTaskboards());
    unawaited(_refreshEventsBoard(silent: true));
  }

  @override
  void dispose() {
    _taskboardSubscription?.cancel();
    _projectEventsSubscription?.cancel();
    _projectEventsReconnectTimer?.cancel();
    _dashboardActivePruneTimer?.cancel();
    _eventsFilterController.dispose();
    _pipelineRunIDController.dispose();
    _pipelineTaskIDController.dispose();
    _pipelineJobIDController.dispose();
    super.dispose();
  }

  void _startTaskboardSubscription() {
    _taskboardSubscription?.cancel();
    _taskboardSubscription = _api
        .taskboardStream(projectID: _projectSetup.projectID)
        .listen((ApiResult<StreamEvent> eventResult) {
          if (!mounted) {
            return;
          }
          if (!eventResult.isSuccess || eventResult.data == null) {
            setState(() {
              _statusMessage =
                  'Taskboard stream error: ${eventResult.errorMessage ?? 'unknown error'}';
            });
            return;
          }
          if (_isRefreshingProjectSetup) {
            return;
          }
          unawaited(_refreshProjectSetup(silent: true));
          unawaited(_loadTaskboards(silent: true));
        });
  }

  void _startProjectEventsSubscription() {
    _projectEventsSubscription?.cancel();
    _projectEventsReconnectTimer?.cancel();
    _projectEventsSubscription = _api
        .projectEventsStream(
          projectID: _projectSetup.projectID,
          fromOffset: _nextProjectEventsOffset,
        )
        .listen(
          (ApiResult<StreamEvent> eventResult) {
            if (!mounted) {
              return;
            }
            if (!eventResult.isSuccess || eventResult.data == null) {
              setState(() {
                _statusMessage =
                    eventResult.errorMessage ??
                    'Project events stream degraded';
              });
              _scheduleProjectEventsReconnect();
              return;
            }
            final incoming = eventResult.data!;
            setState(() {
              if (incoming.streamOffset >= _nextProjectEventsOffset) {
                _nextProjectEventsOffset = incoming.streamOffset + 1;
              }
              final deduped = <String>{incoming.eventID};
              final merged = <StreamEvent>[incoming];
              for (final existing in _projectEvents) {
                if (deduped.contains(existing.eventID)) {
                  continue;
                }
                deduped.add(existing.eventID);
                merged.add(existing);
              }
              merged.sort((a, b) => b.streamOffset.compareTo(a.streamOffset));
              _projectEvents = merged.take(300).toList(growable: false);
              _applyDashboardActiveEvent(incoming);

              final now = DateTime.now().toUtc();
              final summaryKey = _summaryKeyForEvent(incoming);
              var existingIndex = _realtimeSummaryFeed.indexWhere(
                (entry) => _summaryKeyForEvent(entry.event) == summaryKey,
              );
              if (existingIndex < 0) {
                existingIndex = _realtimeSummaryFeed.indexWhere(
                  (entry) => _sameSummaryCorrelation(entry.event, incoming),
                );
              }
              final summaryEntry = _RealtimeSummaryEntry(
                event: incoming,
                receivedAt: now,
              );
              if (existingIndex >= 0) {
                final previousEntry = _realtimeSummaryFeed[existingIndex];
                final previousStatus = _summaryStatusForEvent(
                  previousEntry.event,
                );
                final nextStatus = _summaryStatusForEvent(incoming);
                _realtimeSummaryFeed.removeAt(existingIndex);
                _realtimeSummaryFeed.insert(0, summaryEntry);
                if (_shouldRenotifyOnStatusTransition(
                  previousStatus: previousStatus,
                  nextStatus: nextStatus,
                )) {
                  if (_realtimeSummaryUnread < 999) {
                    _realtimeSummaryUnread += 1;
                  }
                }
              } else {
                _realtimeSummaryFeed.insert(0, summaryEntry);
                if (_realtimeSummaryUnread < 999) {
                  _realtimeSummaryUnread += 1;
                }
              }
              if (_realtimeSummaryFeed.length > 80) {
                _realtimeSummaryFeed.removeRange(
                  80,
                  _realtimeSummaryFeed.length,
                );
              }
            });
          },
          onError: (Object error, StackTrace stackTrace) {
            if (!mounted) {
              return;
            }
            setState(() {
              _statusMessage = 'Project events stream disconnected: $error';
            });
            _scheduleProjectEventsReconnect();
          },
          onDone: () {
            if (!mounted) {
              return;
            }
            _scheduleProjectEventsReconnect();
          },
        );
  }

  void _startDashboardActivePruneTicker() {
    _dashboardActivePruneTimer?.cancel();
    _dashboardActivePruneTimer = Timer.periodic(const Duration(seconds: 5), (
      _,
    ) {
      if (!mounted || _dashboardActiveEntriesByKey.isEmpty) {
        return;
      }
      final now = DateTime.now();
      final staleKeys = _dashboardActiveEntriesByKey.entries
          .where(
            (entry) =>
                now.difference(entry.value.lastSeenAt) >
                _dashboardActiveEntryTtl,
          )
          .map((entry) => entry.key)
          .toList(growable: false);
      if (staleKeys.isEmpty) {
        return;
      }
      setState(() {
        for (final key in staleKeys) {
          _dashboardActiveEntriesByKey.remove(key);
        }
      });
    });
  }

  String _dashboardLiveActivityKey(StreamEvent event) {
    final sessionID = event.sessionID?.trim();
    if (sessionID != null && sessionID.isNotEmpty) {
      return 'session:$sessionID';
    }
    final projectID = event.projectID?.trim() ?? '';
    final runID = event.runID?.trim() ?? '';
    final taskID = event.taskID?.trim() ?? '';
    final jobID = event.jobID?.trim() ?? '';
    if (runID.isNotEmpty || taskID.isNotEmpty || jobID.isNotEmpty) {
      return 'corr:$projectID|$runID|$taskID|$jobID';
    }
    if (event.eventID.trim().isNotEmpty) {
      return 'event:${event.eventID.trim()}';
    }
    return 'offset:${event.streamOffset}';
  }

  void _applyDashboardActiveEvent(StreamEvent incoming) {
    final key = _dashboardLiveActivityKey(incoming);
    final type = incoming.eventType.trim().toLowerCase();
    final terminalEvent =
        type == 'stream.session.ended' ||
        type == 'stream.session.completed' ||
        type == 'stream.session.failed' ||
        type.contains('completed') ||
        type.contains('failed') ||
        type.contains('terminate') ||
        type.contains('cancel') ||
        type.contains('ended');
    if (terminalEvent) {
      _dashboardActiveEntriesByKey.remove(key);
      return;
    }
    _dashboardActiveEntriesByKey[key] = _LiveActiveEntry(
      event: incoming,
      lastSeenAt: DateTime.now(),
    );
  }

  bool _isDashboardSnapshotLive(LifecycleSessionSnapshotModel snapshot) {
    if (snapshot.endedAt != null) {
      return false;
    }
    final state = snapshot.currentState.trim().toLowerCase();
    if (state.isEmpty) {
      return true;
    }
    return !state.contains('completed') &&
        !state.contains('failed') &&
        !state.contains('exited') &&
        !state.contains('terminated');
  }

  void _syncDashboardActiveFromSnapshots(
    List<LifecycleSessionSnapshotModel> snapshots,
  ) {
    _dashboardActiveEntriesByKey.clear();
    final now = DateTime.now();
    for (final snapshot in snapshots) {
      if (!_isDashboardSnapshotLive(snapshot)) {
        continue;
      }
      final syntheticEvent = StreamEvent(
        eventID: 'snapshot:${snapshot.sessionID}',
        streamOffset: snapshot.lastProjectEventSeq,
        eventType: 'snapshot.active',
        source: 'lifecycle_snapshot',
        payload: '{}',
        occurredAt: snapshot.updatedAt,
        runID: snapshot.runID,
        taskID: snapshot.taskID,
        jobID: snapshot.jobID,
        projectID: snapshot.projectID,
        sessionID: snapshot.sessionID,
      );
      final key = _dashboardLiveActivityKey(syntheticEvent);
      _dashboardActiveEntriesByKey[key] = _LiveActiveEntry(
        event: syntheticEvent,
        lastSeenAt: now,
      );
    }
  }

  String _summaryKeyForEvent(StreamEvent event) {
    final sessionID = event.sessionID?.trim();
    if (sessionID != null && sessionID.isNotEmpty) {
      return 'session:$sessionID';
    }
    final runID = event.runID?.trim() ?? '';
    final taskID = event.taskID?.trim() ?? '';
    final jobID = event.jobID?.trim() ?? '';
    if (runID.isNotEmpty || taskID.isNotEmpty || jobID.isNotEmpty) {
      return 'corr:$runID|$taskID|$jobID';
    }
    return 'event:${event.eventID.trim()}';
  }

  void _scheduleProjectEventsReconnect() {
    _projectEventsReconnectTimer?.cancel();
    _projectEventsReconnectTimer = Timer(const Duration(seconds: 2), () {
      if (!mounted) {
        return;
      }
      _startProjectEventsSubscription();
    });
  }

  Future<void> _openRealtimeSummaryFeed() async {
    if (!mounted) {
      return;
    }
    setState(() {
      _realtimeSummaryUnread = 0;
    });
    await showGeneralDialog<void>(
      context: context,
      barrierDismissible: true,
      barrierLabel: 'Realtime summary feed',
      barrierColor: Colors.black38,
      pageBuilder: (BuildContext context, _, __) {
        return SafeArea(
          child: Align(
            alignment: Alignment.topRight,
            child: Padding(
              padding: const EdgeInsets.only(top: 70, right: 12, left: 12),
              child: Material(
                elevation: 10,
                borderRadius: BorderRadius.circular(12),
                color: Theme.of(context).colorScheme.surface,
                child: ConstrainedBox(
                  constraints: const BoxConstraints(
                    maxWidth: 460,
                    maxHeight: 560,
                    minWidth: 320,
                  ),
                  child: _buildRealtimeSummaryPanel(),
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildRealtimeSummaryPanel() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
          child: Row(
            children: <Widget>[
              const Expanded(
                child: Text(
                  'Realtime Summary',
                  style: TextStyle(fontWeight: FontWeight.w700),
                ),
              ),
              Text(
                'Since app load',
                style: Theme.of(context).textTheme.labelSmall,
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        if (_realtimeSummaryFeed.isEmpty)
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text('No realtime events received yet.'),
          )
        else
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 6),
              itemCount: _realtimeSummaryFeed.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (BuildContext context, int index) {
                final entry = _realtimeSummaryFeed[index];
                final event = entry.event;
                final status = _summaryStatusForEvent(event);
                final statusLabel = _summaryStatusLabelForEvent(event, status);
                final statusColor = _summaryStatusColor(status, context);
                final activityLabel = _summaryActivityLabel(
                  entry: entry,
                  status: status,
                );
                return ListTile(
                  minVerticalPadding: 6,
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 4,
                  ),
                  onTap: () => _openSummarySessionInspection(event.sessionID),
                  title: Text(
                    _summaryTitleForEvent(event),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  subtitle: Text(
                    'session=${event.sessionID ?? '-'} • $statusLabel',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  trailing: Column(
                    mainAxisSize: MainAxisSize.min,
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: <Widget>[
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: statusColor.withValues(alpha: 0.12),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: statusColor.withValues(alpha: 0.35),
                          ),
                        ),
                        child: Text(
                          statusLabel,
                          style: TextStyle(
                            color: statusColor,
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        activityLabel,
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
      ],
    );
  }

  String _summaryTitleForEvent(StreamEvent event) {
    final normalizedType = event.eventType.trim();
    if (normalizedType.isEmpty) {
      return 'Event update';
    }
    return normalizedType;
  }

  Map<String, dynamic>? _summaryPayloadMap(StreamEvent event) {
    final payloadText = event.payload.trim();
    if (payloadText.isEmpty) {
      return null;
    }
    try {
      final decoded = jsonDecode(payloadText);
      if (decoded is Map<String, dynamic>) {
        return decoded;
      }
    } catch (_) {
      return null;
    }
    return null;
  }

  int? _summaryPayloadInt(Map<String, dynamic> payload, String key) {
    final raw = payload[key];
    if (raw is int) {
      return raw;
    }
    if (raw is double) {
      return raw.toInt();
    }
    if (raw is String) {
      return int.tryParse(raw.trim());
    }
    return null;
  }

  _EventSeverity _summaryEventSeverity(StreamEvent event) {
    final normalizedType = event.eventType.trim().toLowerCase();
    final payload = _summaryPayloadMap(event);
    final failureClass =
        (payload?['failure_class'] as String?)?.trim().toLowerCase() ?? '';
    final runtimeEvent =
        (payload?['runtime_event'] as String?)?.trim().toLowerCase() ?? '';
    final retryCount = payload != null
        ? _summaryPayloadInt(payload, 'retry_count')
        : null;
    final maxRetry = payload != null
        ? _summaryPayloadInt(payload, 'max_retry')
        : null;

    final hasRetriesRemaining =
        retryCount != null && maxRetry != null && retryCount < maxRetry;
    final failedButRetryable =
        normalizedType.contains('failed') &&
        (failureClass == 'transient' || hasRetriesRemaining);

    final warningSignal =
        normalizedType.contains('retry_scheduled') ||
        normalizedType.contains('retry_started') ||
        normalizedType.contains('degraded') ||
        normalizedType.contains('gap') ||
        normalizedType.contains('stale') ||
        normalizedType.contains('idle_suspected') ||
        normalizedType.contains('waiting_input') ||
        normalizedType.contains('pause') ||
        normalizedType.contains('heartbeat_quorum_degraded') ||
        failedButRetryable ||
        failureClass == 'transient';
    if (warningSignal) {
      return _EventSeverity.warning;
    }

    final terminalSignal =
        normalizedType.contains('dead_letter') ||
        normalizedType.contains('terminated') ||
        normalizedType.contains('failed') ||
        normalizedType.contains('error') ||
        runtimeEvent == 'failed' ||
        runtimeEvent == 'terminated' ||
        failureClass == 'terminal';
    if (terminalSignal) {
      return _EventSeverity.terminal;
    }

    if (normalizedType.contains('completed') ||
        normalizedType.contains('healthy')) {
      return _EventSeverity.success;
    }

    return _EventSeverity.info;
  }

  String _summaryStatusForEvent(StreamEvent event) {
    final eventType = event.eventType.toLowerCase();
    if (eventType.contains('ended')) {
      final payloadLower = event.payload.toLowerCase();
      final failedExit =
          payloadLower.contains('failed') ||
          payloadLower.contains('error') ||
          payloadLower.contains('cancel') ||
          payloadLower.contains('terminate');
      return failedExit ? 'failed' : 'completed';
    }
    switch (_summaryEventSeverity(event)) {
      case _EventSeverity.terminal:
        return 'failed';
      case _EventSeverity.warning:
        return 'degraded';
      case _EventSeverity.success:
        return 'completed';
      case _EventSeverity.info:
        break;
    }
    if (eventType.contains('started') || eventType.contains('heartbeat')) {
      return 'running';
    }
    return 'updated';
  }

  String _summaryStatusLabelForEvent(StreamEvent event, String status) {
    final eventType = event.eventType.trim().toLowerCase();
    if (status == 'degraded') {
      if (eventType.contains('retry_scheduled') ||
          eventType.contains('retry_started')) {
        return 'retrying';
      }
      if (eventType.contains('heartbeat_quorum_degraded')) {
        return 'heartbeat degraded';
      }
      if (eventType.contains('waiting_input')) {
        return 'waiting input';
      }
      if (eventType.contains('idle_suspected')) {
        return 'idle suspected';
      }
      if (eventType.contains('stale')) {
        return 'stale';
      }
      if (eventType.contains('gap')) {
        return 'stream gap';
      }
      return 'degraded';
    }
    if (status == 'failed') {
      if (eventType.contains('dead_letter')) {
        return 'dead-lettered';
      }
      if (eventType.contains('terminate') || eventType.contains('cancel')) {
        return 'terminated';
      }
      return 'failed';
    }
    if (status == 'completed') {
      return 'completed';
    }
    if (status == 'running') {
      return 'running';
    }
    return 'updated';
  }

  bool _isTerminalSummaryStatus(String status) {
    return status == 'completed' || status == 'failed';
  }

  bool _sameSummaryCorrelation(StreamEvent left, StreamEvent right) {
    final leftRun = left.runID?.trim() ?? '';
    final leftTask = left.taskID?.trim() ?? '';
    final leftJob = left.jobID?.trim() ?? '';
    final rightRun = right.runID?.trim() ?? '';
    final rightTask = right.taskID?.trim() ?? '';
    final rightJob = right.jobID?.trim() ?? '';
    if (leftRun.isEmpty && leftTask.isEmpty && leftJob.isEmpty) {
      return false;
    }
    if (rightRun.isEmpty && rightTask.isEmpty && rightJob.isEmpty) {
      return false;
    }
    return leftRun == rightRun && leftTask == rightTask && leftJob == rightJob;
  }

  bool _shouldRenotifyOnStatusTransition({
    required String previousStatus,
    required String nextStatus,
  }) {
    return !_isTerminalSummaryStatus(previousStatus) &&
        _isTerminalSummaryStatus(nextStatus);
  }

  String _summaryActivityLabel({
    required _RealtimeSummaryEntry entry,
    required String status,
  }) {
    if (_isTerminalSummaryStatus(status)) {
      return 'exited ${_relativeTime(entry.event.occurredAt)}';
    }
    return 'updated ${_relativeTime(entry.receivedAt)}';
  }

  Color _summaryStatusColor(String status, BuildContext context) {
    switch (status) {
      case 'failed':
        return Theme.of(context).colorScheme.error;
      case 'completed':
        return Colors.green.shade700;
      case 'degraded':
        return Colors.amber.shade800;
      case 'running':
        return Colors.blue.shade700;
      default:
        return Theme.of(context).colorScheme.primary;
    }
  }

  String _relativeTime(DateTime timestamp) {
    final delta = DateTime.now().toUtc().difference(timestamp.toUtc());
    if (delta.inSeconds < 5) {
      return 'now';
    }
    if (delta.inMinutes < 1) {
      return '${delta.inSeconds}s';
    }
    if (delta.inHours < 1) {
      return '${delta.inMinutes}m';
    }
    return '${delta.inHours}h';
  }

  Future<void> _refreshEventsBoard({bool silent = false}) async {
    final snapshotsResult = await _api.lifecycleSessionSnapshots(
      projectID: _projectSetup.projectID,
      limit: 300,
    );
    final projectEventsResult = await _api.projectEvents(
      projectID: _projectSetup.projectID,
      fromOffset: 0,
      limit: 250,
    );

    if (!mounted) {
      return;
    }

    String? nextStatus;
    if (!snapshotsResult.isSuccess && snapshotsResult.errorMessage != null) {
      nextStatus =
          'Events board snapshots error: ${snapshotsResult.errorMessage}';
    } else if (!projectEventsResult.isSuccess &&
        projectEventsResult.errorMessage != null) {
      nextStatus = 'Project events error: ${projectEventsResult.errorMessage}';
    }

    setState(() {
      if (snapshotsResult.isSuccess && snapshotsResult.data != null) {
        _syncDashboardActiveFromSnapshots(snapshotsResult.data!);
      }
      if (projectEventsResult.isSuccess && projectEventsResult.data != null) {
        _projectEvents = projectEventsResult.data!;
      }
      if (nextStatus != null) {
        _statusMessage = nextStatus;
      }
    });
  }

  Future<void> _refreshProjectSetup({bool silent = false}) async {
    if (!silent && mounted) {
      setState(() {
        _isRefreshingProjectSetup = true;
      });
    }
    final result = await _api.projectSetup(projectID: _projectSetup.projectID);
    if (!mounted) {
      return;
    }
    setState(() {
      _isRefreshingProjectSetup = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed loading project setup: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _projectSetup = result.data!;
    });
  }

  Future<void> _openEditProjectSetup() async {
    final updated = await Navigator.of(context).push<ProjectSetupConfig>(
      MaterialPageRoute<ProjectSetupConfig>(
        builder: (BuildContext context) => ProjectSetupEditScreen(
          projectSetup: _projectSetup,
          endpoint: widget.endpoint,
        ),
      ),
    );

    if (!mounted || updated == null) {
      return;
    }

    setState(() {
      _projectSetup = updated;
      _statusMessage = 'Project setup updated.';
    });
    _startTaskboardSubscription();
    await _refreshProjectSetup();
    await _loadProjectDocuments();
  }

  Future<void> _openWorkerSessions() async {
    await Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (BuildContext context) => Scaffold(
          appBar: AppBar(title: const Text('Worker Sessions')),
          body: WorkerSessionsScreen(
            api: _api,
            statusMessage: _statusMessage,
            onStatus: (String message) {
              if (!mounted) {
                return;
              }
              setState(() => _statusMessage = message);
            },
          ),
        ),
      ),
    );
  }

  Future<void> _openWorkerSettings() async {
    await Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (BuildContext context) => Scaffold(
          appBar: AppBar(title: const Text('Worker Settings')),
          body: WorkerSettingsScreen(
            api: _api,
            statusMessage: _statusMessage,
            onStatus: (String message) {
              if (!mounted) {
                return;
              }
              setState(() => _statusMessage = message);
            },
          ),
        ),
      ),
    );
  }

  Future<void> _openEventsMatrixPage({String? sessionID}) async {
    final normalizedSessionID = sessionID?.trim();
    await Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (BuildContext context) => ProjectEventsMatrixPage(
          api: _api,
          projectID: _projectSetup.projectID,
          projectName: _projectSetup.projectName,
          initialMode:
              normalizedSessionID != null && normalizedSessionID.isNotEmpty
              ? _EventsBoardMode.sessionInspection
              : _EventsBoardMode.globalLive,
          initialSessionID: normalizedSessionID,
        ),
      ),
    );
  }

  Future<void> _openSummarySessionInspection(String? sessionID) async {
    if (!mounted) {
      return;
    }
    Navigator.of(context).pop();
    await _openEventsMatrixPage(sessionID: sessionID);
  }

  void _goToDashboardHome() {
    Navigator.of(context).popUntil((Route<dynamic> route) => route.isFirst);
  }

  Future<void> _copyProjectID() async {
    await Clipboard.setData(ClipboardData(text: _projectSetup.projectID));
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(const SnackBar(content: Text('Project ID copied')));
  }

  Future<void> _loadProjectDocuments() async {
    final result = await _api.projectDocuments(
      projectID: _projectSetup.projectID,
      limit: 100,
    );
    if (!mounted) {
      return;
    }
    if (!result.isSuccess || result.data == null) {
      setState(() {
        _statusMessage =
            'Failed loading project documents: ${result.errorMessage ?? 'unknown error'}';
      });
      return;
    }
    setState(() {
      _projectDocuments = result.data!;
    });
  }

  Future<void> _loadTaskboards({bool silent = false}) async {
    if (!silent && mounted) {
      setState(() {
        _isRefreshingProjectSetup = true;
      });
    }
    final result = await _api.taskboards(projectID: _projectSetup.projectID);
    if (!mounted) {
      return;
    }
    setState(() {
      _isRefreshingProjectSetup = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed loading taskboards: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _taskboards = result.data!;
    });
  }

  Future<void> _uploadFiles() async {
    AppLogger.instance.logger.i(
      'Project document upload requested',
      error: {'projectID': _projectSetup.projectID},
    );
    FilePickerResult? picked;
    try {
      picked = await FilePicker.platform.pickFiles(
        allowMultiple: true,
        withData: true,
      );
    } catch (error, stackTrace) {
      AppLogger.instance.logger.e(
        'File picker failed to open',
        error: error,
        stackTrace: stackTrace,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _statusMessage = 'File picker failed: $error';
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Could not open file picker: $error')),
      );
      return;
    }
    if (!mounted || picked == null || picked.files.isEmpty) {
      AppLogger.instance.logger.i('File picker closed without file selection');
      return;
    }

    setState(() {
      _isUploadingFiles = true;
      _statusMessage = null;
    });

    final failures = <String>[];
    var uploadedCount = 0;

    for (final file in picked.files) {
      Uint8List? bytes = file.bytes;
      if (bytes == null && file.path != null) {
        final path = file.path!;
        bytes = await File(path).readAsBytes();
      }
      if (bytes == null || bytes.isEmpty) {
        failures.add('${file.name}: file bytes unavailable');
        continue;
      }

      final contentType =
          lookupMimeType(file.name, headerBytes: bytes) ??
          'application/octet-stream';
      final request = await _api.requestProjectDocumentUpload(
        projectID: _projectSetup.projectID,
        fileName: file.name,
        contentType: contentType,
      );
      if (!request.isSuccess || request.data == null) {
        failures.add(
          '${file.name}: ${request.errorMessage ?? 'failed requesting upload URL'}',
        );
        continue;
      }

      final upload = await _api.uploadProjectDocumentBytes(
        uploadURL: request.data!.uploadURL,
        bytes: bytes,
        contentType: contentType,
      );
      if (!upload.isSuccess) {
        final cleanup = await _api.deleteProjectDocument(
          projectID: request.data!.projectID,
          documentID: request.data!.documentID,
        );
        if (!cleanup.isSuccess) {
          AppLogger.instance.logger.w(
            'Failed to cleanup project document after upload failure',
            error: {
              'projectID': request.data!.projectID,
              'documentID': request.data!.documentID,
            },
          );
        }
        failures.add(
          '${file.name}: ${upload.errorMessage ?? 'failed uploading document'}',
        );
        continue;
      }

      uploadedCount += 1;
    }

    if (!mounted) {
      return;
    }

    setState(() {
      _isUploadingFiles = false;
      if (failures.isEmpty) {
        _statusMessage = 'Uploaded $uploadedCount file(s).';
      } else {
        _statusMessage =
            'Uploaded $uploadedCount file(s), ${failures.length} failed.';
      }
    });

    await _loadProjectDocuments();

    if (failures.isNotEmpty && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            failures.take(3).join(' | '),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      );
    }
  }

  Future<void> _deleteProjectDocument(ProjectDocument document) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: const Text('Delete Document?'),
          content: Text(
            'Delete "${document.fileName}" from this project and remote storage? This cannot be undone.',
          ),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Delete'),
            ),
          ],
        );
      },
    );

    if (!mounted || confirmed != true) {
      return;
    }

    setState(() {
      _deletingDocumentIDs.add(document.documentID);
      _statusMessage = null;
    });

    final deleteResult = await _api.deleteProjectDocument(
      projectID: _projectSetup.projectID,
      documentID: document.documentID,
    );

    if (!mounted) {
      return;
    }

    if (!deleteResult.isSuccess) {
      setState(() {
        _deletingDocumentIDs.remove(document.documentID);
        _statusMessage =
            'Failed deleting ${document.fileName}: ${deleteResult.errorMessage ?? 'unknown error'}';
      });
      return;
    }

    final removed = await _waitForDocumentRemoval(document.documentID);
    if (!mounted) {
      return;
    }

    setState(() {
      _deletingDocumentIDs.remove(document.documentID);
      _statusMessage = removed
          ? 'Deleted ${document.fileName} from project and remote storage.'
          : 'Deletion queued for ${document.fileName}; refresh shortly to confirm completion.';
    });
  }

  Future<bool> _waitForDocumentRemoval(String documentID) async {
    final deadline = DateTime.now().add(const Duration(seconds: 10));
    while (DateTime.now().isBefore(deadline)) {
      final result = await _api.projectDocuments(
        projectID: _projectSetup.projectID,
        limit: 100,
      );
      if (!mounted) {
        return false;
      }
      if (!result.isSuccess || result.data == null) {
        await Future<void>.delayed(const Duration(milliseconds: 400));
        continue;
      }

      final documents = result.data!;
      final exists = documents.any(
        (ProjectDocument item) => item.documentID == documentID,
      );
      setState(() {
        _projectDocuments = documents;
      });
      if (!exists) {
        return true;
      }
      await Future<void>.delayed(const Duration(milliseconds: 400));
    }
    await _loadProjectDocuments();
    return false;
  }

  Future<void> _createNewTaskboard() async {
    final branchOptionsResult = await _api.projectRepositoryBranches(
      projectID: _projectSetup.projectID,
    );
    if (!mounted) {
      return;
    }
    if (!branchOptionsResult.isSuccess || branchOptionsResult.data == null) {
      setState(() {
        _statusMessage =
            'Failed loading repository branches: ${branchOptionsResult.errorMessage ?? 'unknown error'}';
      });
      return;
    }

    final selectedDocumentIDs = _projectDocuments
        .map((ProjectDocument document) => document.documentID)
        .toSet();
    final branchOptionsByRepository = <String, ProjectRepositoryBranchOption>{
      for (final option in branchOptionsResult.data!)
        option.repositoryID: option,
    };
    final selectedBranches = <String, String>{
      for (final option in branchOptionsResult.data!)
        if (option.branches.isNotEmpty)
          option.repositoryID: option.branches.contains(option.defaultBranch)
              ? option.defaultBranch!
              : option.branches.first,
    };
    final taskboardNameController = TextEditingController();
    final promptController = TextEditingController();
    var isGeneratingPrompt = false;
    final draft = await showDialog<_NewTaskboardDraft>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            final isAllSelected =
                selectedDocumentIDs.length == _projectDocuments.length;
            return AlertDialog(
              title: const Text('New Taskboard'),
              content: SizedBox(
                width: 520,
                child: SingleChildScrollView(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      TextField(
                        controller: taskboardNameController,
                        decoration: const InputDecoration(
                          labelText: 'Taskboard name',
                          hintText: 'Required',
                          border: OutlineInputBorder(),
                        ),
                      ),
                      const SizedBox(height: 12),
                      CheckboxListTile(
                        value: isAllSelected,
                        onChanged: (bool? value) {
                          setDialogState(() {
                            if (value == true) {
                              selectedDocumentIDs
                                ..clear()
                                ..addAll(
                                  _projectDocuments.map(
                                    (ProjectDocument document) =>
                                        document.documentID,
                                  ),
                                );
                            } else {
                              selectedDocumentIDs.clear();
                            }
                          });
                        },
                        title: const Text('Select all project documents'),
                        contentPadding: EdgeInsets.zero,
                      ),
                      const SizedBox(height: 8),
                      ..._projectDocuments.map((ProjectDocument document) {
                        return CheckboxListTile(
                          value: selectedDocumentIDs.contains(
                            document.documentID,
                          ),
                          onChanged: (bool? value) {
                            setDialogState(() {
                              if (value == true) {
                                selectedDocumentIDs.add(document.documentID);
                              } else {
                                selectedDocumentIDs.remove(document.documentID);
                              }
                            });
                          },
                          title: Text(document.fileName),
                          subtitle: Text('Status: ${document.status}'),
                          contentPadding: EdgeInsets.zero,
                        );
                      }),
                      const SizedBox(height: 12),
                      TextField(
                        controller: promptController,
                        minLines: 3,
                        maxLines: 6,
                        decoration: const InputDecoration(
                          labelText: 'User prompt',
                          hintText:
                              'Describe what you want in the new taskboard.',
                          border: OutlineInputBorder(),
                        ),
                      ),
                      const SizedBox(height: 8),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: OutlinedButton.icon(
                          onPressed: isGeneratingPrompt
                              ? null
                              : () async {
                                  final taskboardName = taskboardNameController
                                      .text
                                      .trim();
                                  if (taskboardName.isEmpty) {
                                    setDialogState(() {
                                      isGeneratingPrompt = false;
                                    });
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(
                                        content: Text(
                                          'Enter a taskboard name before generating a prompt.',
                                        ),
                                      ),
                                    );
                                    return;
                                  }
                                  setDialogState(() {
                                    isGeneratingPrompt = true;
                                  });
                                  final response = await _api
                                      .refineIngestionPrompt(
                                        projectID: _projectSetup.projectID,
                                        taskboardName: taskboardName,
                                        userPrompt: promptController.text,
                                      );
                                  if (!context.mounted) {
                                    return;
                                  }
                                  if (response.isSuccess &&
                                      response.data != null &&
                                      response.data!.trim().isNotEmpty) {
                                    final generatedPrompt = response.data!
                                        .trim();
                                    promptController.text = generatedPrompt;
                                    promptController.selection =
                                        TextSelection.collapsed(
                                          offset: generatedPrompt.length,
                                        );
                                  } else {
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      SnackBar(
                                        content: Text(
                                          'Prompt generation failed: ${response.errorMessage ?? 'unknown error'}',
                                        ),
                                      ),
                                    );
                                  }
                                  setDialogState(() {
                                    isGeneratingPrompt = false;
                                  });
                                },
                          icon: isGeneratingPrompt
                              ? const SizedBox(
                                  height: 16,
                                  width: 16,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Icon(Icons.auto_awesome),
                          label: const Text('AI: Generate Prompt'),
                        ),
                      ),
                      if (_projectSetup.repositories.isNotEmpty) ...<Widget>[
                        const SizedBox(height: 12),
                        const Text(
                          'Repository branches',
                          style: TextStyle(fontWeight: FontWeight.w600),
                        ),
                        const SizedBox(height: 8),
                        ..._projectSetup.repositories.map((repository) {
                          final option =
                              branchOptionsByRepository[repository
                                  .repositoryID];
                          final branches = option?.branches ?? const <String>[];
                          final selectedBranch =
                              selectedBranches[repository.repositoryID];
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 8),
                            child: DropdownButtonFormField<String>(
                              initialValue: selectedBranch,
                              onChanged: branches.isEmpty
                                  ? null
                                  : (String? value) {
                                      if (value == null) {
                                        return;
                                      }
                                      setDialogState(() {
                                        selectedBranches[repository
                                                .repositoryID] =
                                            value;
                                      });
                                    },
                              decoration: InputDecoration(
                                labelText: repository.repositoryURL,
                                border: const OutlineInputBorder(),
                              ),
                              items: branches
                                  .map(
                                    (String branch) => DropdownMenuItem<String>(
                                      value: branch,
                                      child: Text(branch),
                                    ),
                                  )
                                  .toList(growable: false),
                            ),
                          );
                        }),
                      ],
                    ],
                  ),
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () {
                    final taskboardName = taskboardNameController.text.trim();
                    final selected = selectedDocumentIDs.toList(
                      growable: false,
                    );
                    final prompt = promptController.text.trim();
                    if (taskboardName.isEmpty) {
                      return;
                    }
                    if (selected.isEmpty && prompt.isEmpty) {
                      return;
                    }
                    Navigator.of(context).pop(
                      _NewTaskboardDraft(
                        taskboardName: taskboardName,
                        selectedDocumentIDs: selected.isEmpty ? null : selected,
                        userPrompt: prompt.isEmpty ? null : prompt,
                        repositorySourceBranches: selectedBranches.isEmpty
                            ? null
                            : Map<String, String>.from(selectedBranches),
                      ),
                    );
                  },
                  child: const Text('Create'),
                ),
              ],
            );
          },
        );
      },
    );
    taskboardNameController.dispose();
    promptController.dispose();

    if (!mounted || draft == null) {
      return;
    }

    setState(() {
      _isCreatingTaskboard = true;
      _statusMessage = null;
    });

    final result = await _api.runIngestionAgent(
      projectID: _projectSetup.projectID,
      taskboardName: draft.taskboardName,
      selectedDocumentIDs: draft.selectedDocumentIDs,
      userPrompt: draft.userPrompt,
      repositorySourceBranches: draft.repositorySourceBranches,
    );

    if (!mounted) {
      return;
    }

    setState(() {
      _isCreatingTaskboard = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed creating taskboard: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _statusMessage =
          'Taskboard run enqueued (run=${result.data!.runID}, task=${result.data!.taskID}).';
    });
    await _refreshProjectSetup(silent: true);
    await _loadTaskboards(silent: true);
  }

  Future<void> _openTaskboardManagement(TaskboardModel board) async {
    await Navigator.of(context).push<bool>(
      MaterialPageRoute<bool>(
        builder: (BuildContext context) => TaskboardManagementScreen(
          api: _api,
          projectID: _projectSetup.projectID,
          boardID: board.boardID,
        ),
      ),
    );
    if (!mounted) {
      return;
    }
    await _loadTaskboards(silent: true);
  }

  Widget _buildTaskboardCard(BuildContext context, TaskboardModel board) {
    return InkWell(
      onTap: () => _openTaskboardManagement(board),
      child: Card(
        elevation: 0,
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        child: Padding(
          padding: const EdgeInsets.all(10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                board.name.trim().isEmpty ? board.boardID : board.name,
                style: const TextStyle(fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 4),
              Text('Board ID: ${board.boardID}'),
              const SizedBox(height: 2),
              Text('State: ${board.state}'),
              const SizedBox(height: 2),
              Text('Epics: ${board.epics.length}'),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final repositories = _projectSetup.repositories;
    final scm = _projectSetup.scms.isNotEmpty ? _projectSetup.scms.first : null;
    final hasTracker = _taskboards.isNotEmpty;

    return Scaffold(
      drawer: Drawer(
        child: SafeArea(
          child: ListView(
            padding: EdgeInsets.zero,
            children: <Widget>[
              DrawerHeader(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    Text(
                      _projectSetup.projectName,
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    const SizedBox(height: 4),
                    Text(_projectSetup.projectID),
                  ],
                ),
              ),
              ListTile(
                leading: const Icon(Icons.dashboard_outlined),
                title: const Text('Dashboard Home'),
                onTap: _goToDashboardHome,
              ),
              ListTile(
                leading: const Icon(Icons.memory_outlined),
                title: const Text('Worker Sessions'),
                onTap: () {
                  Navigator.of(context).pop();
                  _openWorkerSessions();
                },
              ),
              ListTile(
                leading: const Icon(Icons.tune),
                title: const Text('Worker Settings'),
                onTap: () {
                  Navigator.of(context).pop();
                  _openWorkerSettings();
                },
              ),
            ],
          ),
        ),
      ),
      appBar: AppBar(
        automaticallyImplyLeading: true,
        title: Text(_projectSetup.projectName),
        actions: <Widget>[
          Padding(
            padding: const EdgeInsets.only(right: 6),
            child: Tooltip(
              message: 'Open Events Matrix (Global Live)',
              child: Material(
                color: Colors.transparent,
                child: InkWell(
                  borderRadius: BorderRadius.circular(999),
                  onTap: () {
                    unawaited(_openEventsMatrixPage());
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: Theme.of(
                        context,
                      ).colorScheme.primary.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Text(
                      'Active ${_dashboardActiveEntriesByKey.length}',
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.primary,
                        fontWeight: FontWeight.w600,
                        fontSize: 12,
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
          IconButton(
            onPressed: _openRealtimeSummaryFeed,
            tooltip: 'Realtime Summary Feed',
            icon: Badge(
              isLabelVisible: _realtimeSummaryUnread > 0,
              label: Text(
                _realtimeSummaryUnread > 99
                    ? '99+'
                    : _realtimeSummaryUnread.toString(),
              ),
              child: const Icon(Icons.notifications_none_outlined),
            ),
          ),
          IconButton(
            onPressed: _isUploadingFiles ? null : _uploadFiles,
            icon: const Icon(Icons.upload_file_outlined),
            tooltip: 'Upload Files',
          ),
          IconButton(
            onPressed: _copyProjectID,
            icon: const Icon(Icons.copy_outlined),
            tooltip: 'Copy Project ID',
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Expanded(
                  child: Text(
                    'Project Dashboard',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
                  ),
                ),
                OutlinedButton.icon(
                  onPressed: _openEventsMatrixPage,
                  icon: const Icon(Icons.account_tree_outlined),
                  label: const Text('Events Matrix'),
                ),
                const SizedBox(width: 8),
                FilledButton.icon(
                  onPressed: _isCreatingTaskboard ? null : _createNewTaskboard,
                  icon: const Icon(Icons.add_task_outlined),
                  label: const Text('New Taskboard'),
                ),
                const SizedBox(width: 8),
                FilledButton.icon(
                  onPressed: _openEditProjectSetup,
                  icon: const Icon(Icons.edit_outlined),
                  label: const Text('Edit Project'),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Expanded(
                  child: Card(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          const Text(
                            'Details',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 8),
                          Text('Project Name: ${_projectSetup.projectName}'),
                          const SizedBox(height: 4),
                          Text('Project ID: ${_projectSetup.projectID}'),
                          Text(
                            'Provider: ${scm?.scmProvider ?? 'Not configured'}',
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Tracker: ${hasTracker ? 'Configured' : 'Not configured yet'}',
                          ),
                          const SizedBox(height: 4),
                          Text('Repositories: ${repositories.length}'),
                        ],
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Card(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          const Text(
                            'Project Documentation',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 8),
                          if (_isUploadingFiles)
                            const Padding(
                              padding: EdgeInsets.only(bottom: 8),
                              child: LinearProgressIndicator(),
                            ),
                          if (_projectDocuments.isEmpty)
                            const Text('No documentation files uploaded yet.')
                          else
                            ..._projectDocuments.map((
                              ProjectDocument document,
                            ) {
                              final isDeleting = _deletingDocumentIDs.contains(
                                document.documentID,
                              );
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 8),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: <Widget>[
                                    const Icon(
                                      Icons.description_outlined,
                                      size: 18,
                                    ),
                                    const SizedBox(width: 8),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: <Widget>[
                                          Text(
                                            document.fileName,
                                            style: const TextStyle(
                                              fontWeight: FontWeight.w500,
                                            ),
                                          ),
                                          Text(
                                            'Status: ${document.status}',
                                            style: Theme.of(
                                              context,
                                            ).textTheme.bodySmall,
                                          ),
                                        ],
                                      ),
                                    ),
                                    IconButton(
                                      onPressed: isDeleting
                                          ? null
                                          : () => _deleteProjectDocument(
                                              document,
                                            ),
                                      icon: isDeleting
                                          ? const SizedBox(
                                              width: 18,
                                              height: 18,
                                              child: CircularProgressIndicator(
                                                strokeWidth: 2,
                                              ),
                                            )
                                          : const Icon(
                                              Icons.delete_outline,
                                              size: 18,
                                            ),
                                      tooltip: isDeleting
                                          ? 'Deleting...'
                                          : 'Delete document',
                                    ),
                                  ],
                                ),
                              );
                            }),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    const Text(
                      'Taskboards',
                      style: TextStyle(fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 8),
                    if (_isRefreshingProjectSetup)
                      const Padding(
                        padding: EdgeInsets.only(bottom: 8),
                        child: LinearProgressIndicator(),
                      ),
                    if (_taskboards.isEmpty)
                      const Text('No taskboards configured for this project.')
                    else
                      LayoutBuilder(
                        builder:
                            (BuildContext context, BoxConstraints constraints) {
                              final boards = _taskboards;
                              final spacing = 8.0;
                              final preferredWidth = 320.0;
                              final visibleColumns =
                                  (constraints.maxWidth / preferredWidth)
                                      .floor()
                                      .clamp(1, boards.length);

                              if (boards.length <= visibleColumns) {
                                return Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: <Widget>[
                                    for (
                                      var index = 0;
                                      index < boards.length;
                                      index++
                                    ) ...<Widget>[
                                      Expanded(
                                        child: _buildTaskboardCard(
                                          context,
                                          boards[index],
                                        ),
                                      ),
                                      if (index < boards.length - 1)
                                        const SizedBox(width: 8),
                                    ],
                                  ],
                                );
                              }

                              final itemWidth =
                                  (constraints.maxWidth -
                                      ((visibleColumns - 1) * spacing)) /
                                  visibleColumns;

                              return SingleChildScrollView(
                                scrollDirection: Axis.horizontal,
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: <Widget>[
                                    for (
                                      var index = 0;
                                      index < boards.length;
                                      index++
                                    ) ...<Widget>[
                                      SizedBox(
                                        width: itemWidth,
                                        child: _buildTaskboardCard(
                                          context,
                                          boards[index],
                                        ),
                                      ),
                                      if (index < boards.length - 1)
                                        const SizedBox(width: 8),
                                    ],
                                  ],
                                ),
                              );
                            },
                      ),
                  ],
                ),
              ),
            ),
            if (_statusMessage != null) ...<Widget>[
              const SizedBox(height: 12),
              Text(_statusMessage!),
            ],
          ],
        ),
      ),
    );
  }
}

class _NewTaskboardDraft {
  const _NewTaskboardDraft({
    required this.taskboardName,
    required this.selectedDocumentIDs,
    required this.userPrompt,
    required this.repositorySourceBranches,
  });

  final String taskboardName;
  final List<String>? selectedDocumentIDs;
  final String? userPrompt;
  final Map<String, String>? repositorySourceBranches;
}

enum _LiveFeedSubStatus { connecting, live, degraded, disconnected }

class _LiveActiveEntry {
  const _LiveActiveEntry({required this.event, required this.lastSeenAt});

  final StreamEvent event;
  final DateTime lastSeenAt;
}

class _PipelineDrilldownNode {
  _PipelineDrilldownNode({
    required this.key,
    required this.runID,
    required this.taskID,
    required this.jobID,
  });

  final String key;
  final String runID;
  final String taskID;
  final String jobID;
  final List<StreamEvent> events = <StreamEvent>[];
  final Set<String> sessionIDs = <String>{};
  DateTime latestAt = DateTime.fromMillisecondsSinceEpoch(0, isUtc: true);
}

class ProjectEventsMatrixPage extends StatefulWidget {
  const ProjectEventsMatrixPage({
    required this.api,
    required this.projectID,
    required this.projectName,
    this.initialMode = _EventsBoardMode.globalLive,
    this.initialSessionID,
    super.key,
  });

  final ControlPlaneApi api;
  final String projectID;
  final String projectName;
  final _EventsBoardMode initialMode;
  final String? initialSessionID;

  @override
  State<ProjectEventsMatrixPage> createState() =>
      _ProjectEventsMatrixPageState();
}

class _ProjectEventsMatrixPageState extends State<ProjectEventsMatrixPage> {
  StreamSubscription<ApiResult<StreamEvent>>? _projectEventsSubscription;
  StreamSubscription<ApiResult<StreamEvent>>? _pipelineEventsSubscription;
  StreamSubscription<ApiResult<StreamEvent>>? _sessionActivitySubscription;
  Timer? _liveFeedHandshakeTimer;
  Timer? _liveFeedReconnectTimer;
  Timer? _activeLivePruneTimer;
  int _nextProjectEventsOffset = 0;
  int _nextPipelineEventsOffset = 0;
  int _nextSessionActivityOffset = 0;
  final TextEditingController _runIDController = TextEditingController();
  final TextEditingController _taskIDController = TextEditingController();
  final TextEditingController _jobIDController = TextEditingController();
  final Set<String> _expandedEventIDs = <String>{};

  _EventsBoardMode _mode = _EventsBoardMode.globalLive;
  _LiveFeedSubStatus _liveFeedStatus = _LiveFeedSubStatus.connecting;
  bool _isLoading = false;
  String? _statusMessage;
  String? _selectedSessionID;

  static const Duration _activeLiveEntryTtl = Duration(seconds: 45);
  final Map<String, _LiveActiveEntry> _activeLiveEntriesByKey =
      <String, _LiveActiveEntry>{};
  List<StreamEvent> _pipelineEvents = const <StreamEvent>[];
  StreamEvent? _latestSessionLiveEvent;
  List<LifecycleSessionSnapshotModel> _snapshots =
      const <LifecycleSessionSnapshotModel>[];
  List<StreamEvent> _sessionHistory = const <StreamEvent>[];

  @override
  void initState() {
    super.initState();
    _mode = widget.initialMode;
    _selectedSessionID = widget.initialSessionID?.trim();
    _startActiveLivePruneTicker();
    _startProjectSubscription();
    unawaited(_refresh(silent: true));
  }

  @override
  void dispose() {
    _projectEventsSubscription?.cancel();
    _pipelineEventsSubscription?.cancel();
    _sessionActivitySubscription?.cancel();
    _liveFeedHandshakeTimer?.cancel();
    _liveFeedReconnectTimer?.cancel();
    _activeLivePruneTimer?.cancel();
    _runIDController.dispose();
    _taskIDController.dispose();
    _jobIDController.dispose();
    super.dispose();
  }

  void _scheduleProjectSubscriptionReconnect([String? reason]) {
    _liveFeedReconnectTimer?.cancel();
    _liveFeedReconnectTimer = Timer(const Duration(seconds: 2), () {
      if (!mounted) {
        return;
      }
      _startProjectSubscription();
    });
    if (reason != null && reason.isNotEmpty && mounted) {
      setState(() {
        _statusMessage = '$reason Reconnecting stream...';
      });
    }
  }

  void _startActiveLivePruneTicker() {
    _activeLivePruneTimer?.cancel();
    _activeLivePruneTimer = Timer.periodic(const Duration(seconds: 5), (_) {
      if (!mounted || _activeLiveEntriesByKey.isEmpty) {
        return;
      }
      final now = DateTime.now();
      final staleKeys = _activeLiveEntriesByKey.entries
          .where(
            (entry) =>
                now.difference(entry.value.lastSeenAt) > _activeLiveEntryTtl,
          )
          .map((entry) => entry.key)
          .toList(growable: false);
      if (staleKeys.isEmpty) {
        return;
      }
      setState(() {
        for (final key in staleKeys) {
          _activeLiveEntriesByKey.remove(key);
        }
      });
    });
  }

  String _liveActivityKey(StreamEvent event) {
    final projectID = event.projectID?.trim() ?? '';
    final runID = event.runID?.trim() ?? '';
    final taskID = event.taskID?.trim() ?? '';
    final jobID = event.jobID?.trim() ?? '';
    if (runID.isNotEmpty || taskID.isNotEmpty || jobID.isNotEmpty) {
      return 'corr:$projectID|$runID|$taskID|$jobID';
    }
    final sessionID = event.sessionID?.trim();
    if (sessionID != null && sessionID.isNotEmpty) {
      return 'session:$sessionID';
    }
    if (event.eventID.trim().isNotEmpty) {
      return 'event:${event.eventID.trim()}';
    }
    return 'offset:${event.streamOffset}';
  }

  void _applyActiveLiveEvent(StreamEvent incoming) {
    final key = _liveActivityKey(incoming);
    final type = incoming.eventType.trim().toLowerCase();
    final terminalEvent =
        type == 'stream.session.ended' ||
        type == 'stream.session.completed' ||
        type == 'stream.session.failed';
    if (terminalEvent) {
      _activeLiveEntriesByKey.remove(key);
      return;
    }
    _activeLiveEntriesByKey[key] = _LiveActiveEntry(
      event: incoming,
      lastSeenAt: DateTime.now(),
    );
  }

  void _startProjectSubscription() {
    _projectEventsSubscription?.cancel();
    _liveFeedHandshakeTimer?.cancel();
    _liveFeedReconnectTimer?.cancel();
    if (mounted) {
      setState(() => _liveFeedStatus = _LiveFeedSubStatus.connecting);
    }
    // Quiet streams may not emit immediately; mark live after handshake delay.
    _liveFeedHandshakeTimer = Timer(const Duration(seconds: 3), () {
      if (!mounted) {
        return;
      }
      if (_liveFeedStatus == _LiveFeedSubStatus.connecting) {
        setState(() => _liveFeedStatus = _LiveFeedSubStatus.live);
      }
    });
    _projectEventsSubscription = widget.api
        .projectEventsStream(
          projectID: widget.projectID,
          fromOffset: _nextProjectEventsOffset,
        )
        .listen(
          (ApiResult<StreamEvent> eventResult) {
            if (!mounted) {
              return;
            }
            if (!eventResult.isSuccess || eventResult.data == null) {
              _liveFeedHandshakeTimer?.cancel();
              setState(() {
                _liveFeedStatus = _LiveFeedSubStatus.degraded;
                _statusMessage =
                    eventResult.errorMessage ??
                    'Project events stream degraded';
              });
              _scheduleProjectSubscriptionReconnect();
              return;
            }
            final incoming = eventResult.data!;
            _liveFeedHandshakeTimer?.cancel();
            setState(() {
              _liveFeedStatus = _LiveFeedSubStatus.live;
              if (incoming.streamOffset >= _nextProjectEventsOffset) {
                _nextProjectEventsOffset = incoming.streamOffset + 1;
              }
              _applyActiveLiveEvent(incoming);
            });
          },
          onError: (Object error, StackTrace stackTrace) {
            if (!mounted) {
              return;
            }
            _liveFeedHandshakeTimer?.cancel();
            setState(() {
              _liveFeedStatus = _LiveFeedSubStatus.disconnected;
              _statusMessage = 'Project events stream disconnected: $error';
            });
            _scheduleProjectSubscriptionReconnect();
          },
          onDone: () {
            if (!mounted) {
              return;
            }
            _liveFeedHandshakeTimer?.cancel();
            setState(() {
              _liveFeedStatus = _LiveFeedSubStatus.disconnected;
              _statusMessage ??= 'Project events stream closed.';
            });
            _scheduleProjectSubscriptionReconnect();
          },
        );
  }

  List<StreamEvent> _mergeEvents(
    List<StreamEvent> existing,
    StreamEvent incoming,
  ) {
    final mergedByID = <String, StreamEvent>{};
    for (final event in existing) {
      mergedByID[event.eventID] = event;
    }
    mergedByID[incoming.eventID] = incoming;
    final merged = mergedByID.values.toList(growable: false)
      ..sort((a, b) => b.occurredAt.compareTo(a.occurredAt));
    return merged.take(300).toList(growable: false);
  }

  void _startPipelineSubscription() {
    _pipelineEventsSubscription?.cancel();
    _pipelineEventsSubscription = widget.api
        .pipelineEventsStream(
          projectID: widget.projectID,
          runID: _runIDController.text,
          taskID: _taskIDController.text,
          jobID: _jobIDController.text,
          fromOffset: _nextPipelineEventsOffset,
        )
        .listen((ApiResult<StreamEvent> eventResult) {
          if (!mounted || !eventResult.isSuccess || eventResult.data == null) {
            return;
          }
          final incoming = eventResult.data!;
          setState(() {
            if (incoming.streamOffset >= _nextPipelineEventsOffset) {
              _nextPipelineEventsOffset = incoming.streamOffset + 1;
            }
            _pipelineEvents = _mergeEvents(_pipelineEvents, incoming);
          });
        });
  }

  void _startSessionActivitySubscription() {
    _sessionActivitySubscription?.cancel();
    final selectedSessionID = _selectedSessionID?.trim();
    if (selectedSessionID == null || selectedSessionID.isEmpty) {
      return;
    }
    LifecycleSessionSnapshotModel? selected;
    for (final snapshot in _snapshots) {
      if (snapshot.sessionID == selectedSessionID) {
        selected = snapshot;
        break;
      }
    }
    final runID = selected?.runID?.trim();
    if (runID == null || runID.isEmpty) {
      return;
    }
    _sessionActivitySubscription = widget.api
        .sessionActivityStream(
          projectID: widget.projectID,
          runID: runID,
          taskID: selected?.taskID,
          jobID: selected?.jobID,
          fromOffset: _nextSessionActivityOffset,
        )
        .listen((ApiResult<StreamEvent> eventResult) {
          if (!mounted || !eventResult.isSuccess || eventResult.data == null) {
            return;
          }
          final incoming = _normalizeSessionStreamEventForHistory(
            eventResult.data!,
          );
          setState(() {
            if (incoming.streamOffset >= _nextSessionActivityOffset) {
              _nextSessionActivityOffset = incoming.streamOffset + 1;
            }
            _applySessionSnapshotFromEvent(incoming);
            _latestSessionLiveEvent = incoming;
            _sessionHistory = _mergeSessionHistory(_sessionHistory, incoming);
          });
        });
  }

  void _applySessionSnapshotFromEvent(StreamEvent incoming) {
    final sessionID = incoming.sessionID?.trim();
    if (sessionID == null || sessionID.isEmpty) {
      return;
    }
    final index = _snapshots.indexWhere(
      (snapshot) => snapshot.sessionID == sessionID,
    );
    if (index < 0) {
      return;
    }
    final current = _snapshots[index];
    var nextState = current.currentState;
    var nextSeverity = current.currentSeverity;
    var nextEndedAt = current.endedAt;
    final normalizedType = incoming.eventType.trim().toLowerCase();

    if (normalizedType == 'stream.session.started') {
      nextState = 'running';
      nextSeverity = 'info';
      nextEndedAt = null;
    } else if (normalizedType == 'stream.session.completed') {
      nextState = 'completed';
      nextSeverity = 'info';
      nextEndedAt = incoming.occurredAt.toUtc();
    } else if (normalizedType == 'stream.session.failed') {
      nextState = 'failed';
      nextSeverity = 'error';
      nextEndedAt = incoming.occurredAt.toUtc();
    } else if (normalizedType == 'stream.session.ended') {
      final payloadLower = incoming.payload.toLowerCase();
      if (payloadLower.contains('failed')) {
        nextState = 'failed';
        nextSeverity = 'error';
      } else {
        nextState = 'completed';
        nextSeverity = 'info';
      }
      nextEndedAt = incoming.occurredAt.toUtc();
    } else {
      if (nextEndedAt == null) {
        nextState = 'healthy_active';
      }
      if (nextSeverity.trim().isEmpty) {
        nextSeverity = 'info';
      }
    }

    final replacement = LifecycleSessionSnapshotModel(
      projectID: current.projectID,
      sessionID: current.sessionID,
      pipelineType: current.pipelineType,
      currentState: nextState,
      currentSeverity: nextSeverity,
      lastEventSeq: incoming.streamOffset > current.lastEventSeq
          ? incoming.streamOffset
          : current.lastEventSeq,
      lastProjectEventSeq: current.lastProjectEventSeq,
      startedAt: current.startedAt,
      updatedAt: incoming.occurredAt.toUtc(),
      runID: current.runID,
      taskID: current.taskID,
      jobID: current.jobID,
      sourceRuntime: current.sourceRuntime,
      lastReasonCode: current.lastReasonCode,
      lastReasonSummary: current.lastReasonSummary,
      lastLivenessAt: normalizedType == 'stream.session.health'
          ? incoming.occurredAt.toUtc()
          : current.lastLivenessAt,
      lastActivityAt: incoming.occurredAt.toUtc(),
      lastCheckpointAt: current.lastCheckpointAt,
      endedAt: nextEndedAt,
    );

    final updated = List<LifecycleSessionSnapshotModel>.from(_snapshots);
    updated[index] = replacement;
    _snapshots = updated;
  }

  StreamEvent _normalizeSessionStreamEventForHistory(StreamEvent incoming) {
    if (incoming.streamOffset > 0) {
      return incoming;
    }
    final synthesizedOffset = _nextSessionActivityOffset > 0
        ? _nextSessionActivityOffset
        : 1;
    return StreamEvent(
      eventID: incoming.eventID,
      streamOffset: synthesizedOffset,
      occurredAt: incoming.occurredAt,
      runID: incoming.runID,
      taskID: incoming.taskID,
      jobID: incoming.jobID,
      projectID: incoming.projectID,
      sessionID: incoming.sessionID,
      source: incoming.source,
      eventType: incoming.eventType,
      payload: incoming.payload,
      gapDetected: incoming.gapDetected,
      gapReconciled: incoming.gapReconciled,
      expectedEventSeq: incoming.expectedEventSeq,
      observedEventSeq: incoming.observedEventSeq,
    );
  }

  List<StreamEvent> _mergeSessionHistory(
    List<StreamEvent> existing,
    StreamEvent incoming,
  ) {
    final mergedByID = <String, StreamEvent>{};
    for (final event in existing) {
      mergedByID[event.eventID] = event;
    }
    mergedByID[incoming.eventID] = incoming;
    final merged = mergedByID.values.toList(growable: false)
      ..sort((a, b) {
        final byOffset = b.streamOffset.compareTo(a.streamOffset);
        if (byOffset != 0) {
          return byOffset;
        }
        return b.occurredAt.compareTo(a.occurredAt);
      });
    return merged;
  }

  StreamEvent _streamEventFromLifecycleHistory(
    LifecycleHistoryEventModel entry,
  ) {
    return StreamEvent(
      eventID: entry.eventID,
      streamOffset: entry.eventSeq,
      occurredAt: entry.occurredAt,
      eventType: entry.eventType,
      payload: entry.payload,
      projectID: entry.projectID,
      sessionID: entry.sessionID,
      runID: entry.runID,
      taskID: entry.taskID,
      jobID: entry.jobID,
      source: entry.sourceRuntime,
    );
  }

  bool _isSessionTerminalEventType(String eventType) {
    final normalized = eventType.trim().toLowerCase();
    return normalized == 'stream.session.ended' ||
        normalized == 'stream.session.completed' ||
        normalized == 'stream.session.failed';
  }

  bool _isSessionSnapshotLive(LifecycleSessionSnapshotModel? snapshot) {
    if (snapshot == null) {
      return false;
    }
    if (snapshot.endedAt != null) {
      return false;
    }
    final state = snapshot.currentState.trim().toLowerCase();
    if (state.contains('completed') ||
        state.contains('failed') ||
        state.contains('exited') ||
        state.contains('terminated')) {
      return false;
    }
    return true;
  }

  bool _shouldShowSessionRealtimeBlock(
    LifecycleSessionSnapshotModel? snapshot,
  ) {
    final latest = _latestSessionLiveEvent;
    if (latest == null) {
      return false;
    }
    if (_isSessionTerminalEventType(latest.eventType)) {
      return false;
    }
    return _isSessionSnapshotLive(snapshot);
  }

  bool _canRetrySession(LifecycleSessionSnapshotModel? snapshot) {
    if (snapshot == null) {
      return false;
    }
    if (_isSessionSnapshotLive(snapshot)) {
      return false;
    }

    final state = snapshot.currentState.trim().toLowerCase();
    final severity = snapshot.currentSeverity.trim().toLowerCase();
    final reason = [
      snapshot.lastReasonCode,
      snapshot.lastReasonSummary,
    ].whereType<String>().join(' ').toLowerCase();

    final isSuccessfulTerminal =
        state.contains('completed') &&
        !state.contains('failed') &&
        !state.contains('error') &&
        !severity.contains('error') &&
        !severity.contains('fatal') &&
        !reason.contains('error') &&
        !reason.contains('failed') &&
        !reason.contains('timeout') &&
        !reason.contains('deadletter');
    if (isSuccessfulTerminal) {
      return false;
    }

    return state.contains('failed') ||
        state.contains('error') ||
        state.contains('exited') ||
        state.contains('terminated') ||
        state.contains('archived') ||
        severity.contains('error') ||
        severity.contains('fatal') ||
        reason.contains('error') ||
        reason.contains('failed') ||
        reason.contains('timeout') ||
        reason.contains('deadletter') ||
        reason.contains('archive');
  }

  String _liveFeedStatusLabel() {
    switch (_liveFeedStatus) {
      case _LiveFeedSubStatus.connecting:
        return 'CONNECTING';
      case _LiveFeedSubStatus.live:
        return 'LIVE';
      case _LiveFeedSubStatus.degraded:
        return 'DEGRADED';
      case _LiveFeedSubStatus.disconnected:
        return 'DISCONNECTED';
    }
  }

  Color _liveFeedStatusDotColor() {
    switch (_liveFeedStatus) {
      case _LiveFeedSubStatus.connecting:
        return Colors.amber.shade700;
      case _LiveFeedSubStatus.live:
        return Colors.green;
      case _LiveFeedSubStatus.degraded:
        return Colors.orange.shade700;
      case _LiveFeedSubStatus.disconnected:
        return Theme.of(context).colorScheme.error;
    }
  }

  Color _liveFeedBadgeBackgroundColor() {
    switch (_liveFeedStatus) {
      case _LiveFeedSubStatus.connecting:
        return Colors.amber.shade50;
      case _LiveFeedSubStatus.live:
        return Theme.of(context).colorScheme.primary.withValues(alpha: 0.1);
      case _LiveFeedSubStatus.degraded:
        return Colors.orange.shade50;
      case _LiveFeedSubStatus.disconnected:
        return Theme.of(context).colorScheme.error.withValues(alpha: 0.12);
    }
  }

  Future<void> _refresh({bool silent = false}) async {
    if (!silent && mounted) {
      setState(() => _isLoading = true);
    }
    final snapshotsResult = await widget.api.lifecycleSessionSnapshots(
      projectID: widget.projectID,
      limit: 300,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isLoading = false;
      if (snapshotsResult.isSuccess && snapshotsResult.data != null) {
        _snapshots = snapshotsResult.data!;
        _selectedSessionID ??= _snapshots.isNotEmpty
            ? _snapshots.first.sessionID
            : null;
      }
    });

    if (_mode == _EventsBoardMode.pipelineDrilldown) {
      await _reloadPipelineEvents();
      _startPipelineSubscription();
    } else {
      _pipelineEventsSubscription?.cancel();
    }
    if (_mode == _EventsBoardMode.sessionInspection) {
      await _reloadSessionHistory();
      _startSessionActivitySubscription();
    } else {
      _sessionActivitySubscription?.cancel();
    }
  }

  Future<void> _reloadPipelineEvents() async {
    final result = await widget.api.pipelineEvents(
      projectID: widget.projectID,
      runID: _runIDController.text,
      taskID: _taskIDController.text,
      jobID: _jobIDController.text,
      fromOffset: 0,
      limit: 250,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      if (result.isSuccess && result.data != null) {
        _pipelineEvents = result.data!;
        for (final event in _pipelineEvents) {
          if (event.streamOffset >= _nextPipelineEventsOffset) {
            _nextPipelineEventsOffset = event.streamOffset + 1;
          }
        }
      } else {
        _statusMessage = result.errorMessage ?? 'Pipeline query failed';
      }
    });
    _startPipelineSubscription();
  }

  Future<void> _reloadSessionHistory() async {
    final sessionID = _selectedSessionID?.trim();
    if (sessionID == null || sessionID.isEmpty) {
      return;
    }
    const pageSize = 500;
    var fromEventSeq = 0;
    final allHistory = <LifecycleHistoryEventModel>[];
    while (true) {
      final result = await widget.api.lifecycleSessionHistory(
        projectID: widget.projectID,
        sessionID: sessionID,
        fromEventSeq: fromEventSeq,
        limit: pageSize,
      );
      if (!result.isSuccess || result.data == null) {
        if (!mounted) {
          return;
        }
        setState(() {
          _statusMessage = result.errorMessage ?? 'Session history failed';
        });
        return;
      }
      final page = result.data!;
      if (page.isEmpty) {
        break;
      }
      allHistory.addAll(page);
      final highestEventSeq = page.last.eventSeq;
      fromEventSeq = highestEventSeq;
      if (page.length < pageSize) {
        break;
      }
    }
    if (!mounted) {
      return;
    }
    setState(() {
      _sessionHistory =
          allHistory
              .map(_streamEventFromLifecycleHistory)
              .toList(growable: false)
            ..sort((a, b) {
              final byOffset = b.streamOffset.compareTo(a.streamOffset);
              if (byOffset != 0) {
                return byOffset;
              }
              return b.occurredAt.compareTo(a.occurredAt);
            });
      // Session stream uses server bootstrap on initial connect, then realtime.
      _nextSessionActivityOffset = 0;
    });
    _startSessionActivitySubscription();
  }

  Map<String, dynamic>? _eventPayloadMap(StreamEvent event) {
    final payloadText = event.payload.trim();
    if (payloadText.isEmpty) {
      return null;
    }
    try {
      final decoded = jsonDecode(payloadText);
      if (decoded is Map<String, dynamic>) {
        return decoded;
      }
    } catch (_) {
      return null;
    }
    return null;
  }

  int? _payloadInt(Map<String, dynamic> payload, String key) {
    final raw = payload[key];
    if (raw is int) {
      return raw;
    }
    if (raw is double) {
      return raw.toInt();
    }
    if (raw is String) {
      return int.tryParse(raw.trim());
    }
    return null;
  }

  _EventSeverity _eventSeverity(StreamEvent event) {
    final normalizedType = event.eventType.trim().toLowerCase();
    final payload = _eventPayloadMap(event);
    final failureClass =
        (payload?['failure_class'] as String?)?.trim().toLowerCase() ?? '';
    final runtimeEvent =
        (payload?['runtime_event'] as String?)?.trim().toLowerCase() ?? '';
    final retryCount = payload != null
        ? _payloadInt(payload, 'retry_count')
        : null;
    final maxRetry = payload != null ? _payloadInt(payload, 'max_retry') : null;

    final hasRetriesRemaining =
        retryCount != null && maxRetry != null && retryCount < maxRetry;
    final failedButRetryable =
        normalizedType.contains('failed') &&
        (failureClass == 'transient' || hasRetriesRemaining);

    final warningSignal =
        normalizedType.contains('retry_scheduled') ||
        normalizedType.contains('retry_started') ||
        normalizedType.contains('degraded') ||
        normalizedType.contains('gap') ||
        normalizedType.contains('stale') ||
        normalizedType.contains('idle_suspected') ||
        normalizedType.contains('waiting_input') ||
        normalizedType.contains('pause') ||
        normalizedType.contains('heartbeat_quorum_degraded') ||
        failedButRetryable ||
        failureClass == 'transient';

    if (warningSignal) {
      return _EventSeverity.warning;
    }

    final terminalSignal =
        normalizedType.contains('dead_letter') ||
        normalizedType.contains('terminated') ||
        normalizedType.contains('failed') ||
        normalizedType.contains('error') ||
        runtimeEvent == 'failed' ||
        runtimeEvent == 'terminated' ||
        failureClass == 'terminal';

    if (terminalSignal) {
      return _EventSeverity.terminal;
    }

    if (normalizedType.contains('completed') ||
        normalizedType.contains('healthy')) {
      return _EventSeverity.success;
    }

    return _EventSeverity.info;
  }

  Color _eventAccentColor(BuildContext context, StreamEvent event) {
    switch (_eventSeverity(event)) {
      case _EventSeverity.terminal:
        return Theme.of(context).colorScheme.error;
      case _EventSeverity.warning:
        return Colors.amber.shade700;
      case _EventSeverity.success:
        return Colors.green.shade600;
      case _EventSeverity.info:
        return Theme.of(context).colorScheme.primary;
    }
  }

  Future<void> _runManualIntervention(String action) async {
    final sessionID = _selectedSessionID?.trim();
    if (sessionID == null || sessionID.isEmpty) {
      setState(() => _statusMessage = 'Select a session for intervention.');
      return;
    }
    final reasonController = TextEditingController();
    var force = false;
    final destructive = action == 'pause' || action == 'terminate';
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            return AlertDialog(
              title: Text('Manual action: ${action.toUpperCase()}'),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  TextField(
                    controller: reasonController,
                    maxLines: 3,
                    decoration: const InputDecoration(
                      labelText: 'Reason',
                      hintText: 'Required for audit trail',
                    ),
                  ),
                  if (destructive)
                    CheckboxListTile(
                      value: force,
                      contentPadding: EdgeInsets.zero,
                      title: const Text('Force confirm destructive action'),
                      onChanged: (bool? value) {
                        setDialogState(() => force = value == true);
                      },
                    ),
                ],
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Apply'),
                ),
              ],
            );
          },
        );
      },
    );
    final reason = reasonController.text.trim();
    reasonController.dispose();
    if (confirmed != true || reason.isEmpty) {
      return;
    }
    final actorID = destructive
        ? 'admin:project-dashboard'
        : 'operator:project-dashboard';
    final response = await widget.api.applyManualIntervention(
      projectID: widget.projectID,
      sessionID: sessionID,
      action: action,
      reason: reason,
      actorID: actorID,
      force: force,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _statusMessage = response.isSuccess
          ? 'Manual intervention applied: ${action.toUpperCase()}'
          : 'Manual intervention failed: ${response.errorMessage ?? 'unknown error'}';
    });
    await _refresh(silent: true);
  }

  String? _eventFailureReason(StreamEvent event) {
    final payloadText = event.payload.trim();
    if (payloadText.isEmpty) {
      return null;
    }
    try {
      final decoded = jsonDecode(payloadText);
      if (decoded is! Map<String, dynamic>) {
        return null;
      }
      final error = (decoded['error'] as String?)?.trim();
      if (error != null && error.isNotEmpty) {
        return error;
      }
      final reason = (decoded['reason_summary'] as String?)?.trim();
      if (reason != null && reason.isNotEmpty) {
        return reason;
      }
      final failureClass = (decoded['failure_class'] as String?)?.trim();
      if (failureClass != null && failureClass.isNotEmpty) {
        return 'failure_class=$failureClass';
      }
    } catch (_) {
      return null;
    }
    return null;
  }

  Widget _buildEventCard(
    StreamEvent event, {
    bool showOffset = true,
    bool forceErrorAccent = false,
  }) {
    final eventID = event.eventID.trim();
    final isExpanded = _expandedEventIDs.contains(eventID);
    final accent = forceErrorAccent
        ? Theme.of(context).colorScheme.error
        : _eventAccentColor(context, event);
    final reasonColor = forceErrorAccent
        ? Theme.of(context).colorScheme.error
        : accent;
    final reason = _eventFailureReason(event);
    final rawPayload = event.payload.trim();
    final prettyPayload = _prettyEventPayload(event.payload);
    final hasPayload = rawPayload.isNotEmpty;
    final payloadToDisplay = isExpanded ? prettyPayload : rawPayload;
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        border: Border.all(color: accent.withValues(alpha: 0.35)),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            InkWell(
              borderRadius: BorderRadius.circular(6),
              onTap: () {
                if (!hasPayload) {
                  return;
                }
                setState(() {
                  if (isExpanded) {
                    _expandedEventIDs.remove(eventID);
                  } else {
                    _expandedEventIDs.add(eventID);
                  }
                });
              },
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 2),
                child: Row(
                  children: <Widget>[
                    if (showOffset) ...<Widget>[
                      Text(
                        '#${event.streamOffset}',
                        style: TextStyle(
                          color: accent,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(width: 8),
                    ],
                    Expanded(
                      child: Text(
                        event.eventType,
                        style: const TextStyle(fontWeight: FontWeight.w600),
                      ),
                    ),
                    if (hasPayload) ...<Widget>[
                      Icon(
                        isExpanded ? Icons.expand_less : Icons.expand_more,
                        size: 18,
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                      const SizedBox(width: 6),
                    ],
                    Text(
                      event.occurredAt.toIso8601String(),
                      style: Theme.of(context).textTheme.labelSmall,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'session=${event.sessionID ?? '-'} run=${event.runID ?? '-'} task=${event.taskID ?? '-'} job=${event.jobID ?? '-'}',
              style: Theme.of(context).textTheme.labelSmall,
            ),
            if (reason != null) ...<Widget>[
              const SizedBox(height: 6),
              Text(
                'Reason: $reason',
                style: TextStyle(
                  color: reasonColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
            if (hasPayload) ...<Widget>[
              const SizedBox(height: 8),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFF1E1E1E),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  payloadToDisplay,
                  maxLines: isExpanded ? null : 2,
                  overflow: isExpanded
                      ? TextOverflow.visible
                      : TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: Color(0xFFE2E8F0),
                    fontFamily: 'JetBrainsMono',
                    fontSize: 11,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  String _prettyEventPayload(String rawPayload) {
    final payloadText = rawPayload.trim();
    if (payloadText.isEmpty) {
      return '';
    }
    try {
      final decoded = jsonDecode(payloadText);
      const encoder = JsonEncoder.withIndent('  ');
      return encoder.convert(decoded);
    } catch (_) {
      return payloadText;
    }
  }

  Widget _buildGlobalPanel() {
    final activeEvents =
        _activeLiveEntriesByKey.values
            .map((entry) => entry.event)
            .toList(growable: false)
          ..sort((a, b) => b.occurredAt.compareTo(a.occurredAt));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        ...activeEvents
            .take(120)
            .map((event) => _buildEventCard(event, showOffset: false)),
      ],
    );
  }

  Widget _buildPipelinePanel() {
    final nodes = _buildPipelineDrilldownNodes();
    return Column(
      children: <Widget>[
        if (nodes.isEmpty)
          const Text('No pipelines loaded yet.')
        else
          ...nodes.asMap().entries.map(
            (entry) =>
                _buildPipelineNodeCard(entry.value, order: entry.key + 1),
          ),
      ],
    );
  }

  String _formatPipelineTimestamp(DateTime value) {
    final local = value.toLocal();
    final year = local.year.toString().padLeft(4, '0');
    final month = local.month.toString().padLeft(2, '0');
    final day = local.day.toString().padLeft(2, '0');
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    final second = local.second.toString().padLeft(2, '0');
    return '$year-$month-$day $hour:$minute:$second';
  }

  String _normalizedID(String? value) {
    return value?.trim() ?? '';
  }

  bool _matchesPipelineFilters({
    required String runID,
    required String taskID,
    required String jobID,
  }) {
    final runFilter = _normalizedID(_runIDController.text);
    final taskFilter = _normalizedID(_taskIDController.text);
    final jobFilter = _normalizedID(_jobIDController.text);
    if (runFilter.isNotEmpty && runFilter != runID) {
      return false;
    }
    if (taskFilter.isNotEmpty && taskFilter != taskID) {
      return false;
    }
    if (jobFilter.isNotEmpty && jobFilter != jobID) {
      return false;
    }
    return true;
  }

  String _pipelineKey({
    required String runID,
    required String taskID,
    required String jobID,
    required String fallback,
  }) {
    if (runID.isNotEmpty || taskID.isNotEmpty || jobID.isNotEmpty) {
      return 'pipeline:$runID|$taskID|$jobID';
    }
    return 'pipeline:$fallback';
  }

  List<_PipelineDrilldownNode> _buildPipelineDrilldownNodes() {
    final nodesByKey = <String, _PipelineDrilldownNode>{};

    for (final snapshot in _snapshots) {
      final runID = _normalizedID(snapshot.runID);
      final taskID = _normalizedID(snapshot.taskID);
      final jobID = _normalizedID(snapshot.jobID);
      if (!_matchesPipelineFilters(
        runID: runID,
        taskID: taskID,
        jobID: jobID,
      )) {
        continue;
      }
      final nodeKey = _pipelineKey(
        runID: runID,
        taskID: taskID,
        jobID: jobID,
        fallback: 'session:${snapshot.sessionID}',
      );
      final node = nodesByKey.putIfAbsent(
        nodeKey,
        () => _PipelineDrilldownNode(
          key: nodeKey,
          runID: runID,
          taskID: taskID,
          jobID: jobID,
        ),
      );
      node.sessionIDs.add(snapshot.sessionID);
      if (snapshot.updatedAt.isAfter(node.latestAt)) {
        node.latestAt = snapshot.updatedAt;
      }
    }

    for (final event in _pipelineEvents) {
      final runID = _normalizedID(event.runID);
      final taskID = _normalizedID(event.taskID);
      final jobID = _normalizedID(event.jobID);
      if (!_matchesPipelineFilters(
        runID: runID,
        taskID: taskID,
        jobID: jobID,
      )) {
        continue;
      }
      final nodeKey = _pipelineKey(
        runID: runID,
        taskID: taskID,
        jobID: jobID,
        fallback: 'event:${event.eventID}',
      );
      final node = nodesByKey.putIfAbsent(
        nodeKey,
        () => _PipelineDrilldownNode(
          key: nodeKey,
          runID: runID,
          taskID: taskID,
          jobID: jobID,
        ),
      );
      node.events.add(event);
      final sessionID = _normalizedID(event.sessionID);
      if (sessionID.isNotEmpty) {
        node.sessionIDs.add(sessionID);
      }
      if (event.occurredAt.isAfter(node.latestAt)) {
        node.latestAt = event.occurredAt;
      }
    }

    final nodes = nodesByKey.values.toList(growable: false)
      ..sort((a, b) => b.latestAt.compareTo(a.latestAt));
    for (final node in nodes) {
      node.events.sort((a, b) => b.occurredAt.compareTo(a.occurredAt));
    }
    return nodes;
  }

  String _pipelineIdentityLabel(_PipelineDrilldownNode node) {
    final parts = <String>[];
    if (node.runID.isNotEmpty) {
      parts.add('run=${node.runID}');
    }
    if (node.taskID.isNotEmpty) {
      parts.add('task=${node.taskID}');
    }
    if (node.jobID.isNotEmpty) {
      parts.add('job=${node.jobID}');
    }
    if (parts.isEmpty) {
      return 'Pipeline (unkeyed correlation)';
    }
    return parts.join(' • ');
  }

  Future<void> _openSessionInspectionFromPipeline(String sessionID) async {
    _pipelineEventsSubscription?.cancel();
    setState(() {
      _mode = _EventsBoardMode.sessionInspection;
      _selectedSessionID = sessionID;
      _latestSessionLiveEvent = null;
      _sessionHistory = const <StreamEvent>[];
      _nextSessionActivityOffset = 0;
    });
    await _reloadSessionHistory();
  }

  Widget _buildPipelineNodeCard(
    _PipelineDrilldownNode node, {
    required int order,
  }) {
    final sortedSessionIDs = node.sessionIDs.toList(growable: false)..sort();
    final latestEvent = node.events.isEmpty ? null : node.events.first;
    final accent = latestEvent == null
        ? Theme.of(context).colorScheme.primary
        : _eventAccentColor(context, latestEvent);

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        border: Border.all(color: accent.withValues(alpha: 0.35)),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: accent.withValues(alpha: 0.12),
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: Text(
                    '#$order',
                    style: TextStyle(
                      color: accent,
                      fontWeight: FontWeight.w700,
                      fontSize: 11,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    _pipelineIdentityLabel(node),
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: <Widget>[
                    Text(
                      _formatPipelineTimestamp(node.latestAt),
                      style: Theme.of(context).textTheme.labelSmall,
                    ),
                    Text(
                      '${sortedSessionIDs.length} sessions • ${node.events.length} events',
                      style: Theme.of(context).textTheme.labelSmall,
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              'run=${node.runID.isEmpty ? '-' : node.runID} task=${node.taskID.isEmpty ? '-' : node.taskID} job=${node.jobID.isEmpty ? '-' : node.jobID}',
              style: Theme.of(context).textTheme.labelSmall,
            ),
            const SizedBox(height: 10),
            if (sortedSessionIDs.isEmpty)
              const Text('No correlated sessions found yet.')
            else ...<Widget>[
              Text('Sessions', style: Theme.of(context).textTheme.labelMedium),
              const SizedBox(height: 6),
              ...sortedSessionIDs.map((sessionID) {
                LifecycleSessionSnapshotModel? snapshot;
                for (final item in _snapshots) {
                  if (item.sessionID == sessionID) {
                    snapshot = item;
                    break;
                  }
                }
                final stateLabel = snapshot == null
                    ? 'unknown'
                    : '${snapshot.currentState} (${snapshot.currentSeverity})';
                return InkWell(
                  borderRadius: BorderRadius.circular(6),
                  onTap: () => _openSessionInspectionFromPipeline(sessionID),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      vertical: 6,
                      horizontal: 2,
                    ),
                    child: Row(
                      children: <Widget>[
                        Icon(
                          Icons.account_tree_outlined,
                          size: 16,
                          color: accent,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            '$sessionID • $stateLabel',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        const Icon(Icons.chevron_right, size: 16),
                      ],
                    ),
                  ),
                );
              }),
            ],
            if (node.events.isNotEmpty) ...<Widget>[
              const SizedBox(height: 8),
              Text(
                'Pipeline Events',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const SizedBox(height: 8),
              ...node.events
                  .take(30)
                  .map((event) => _buildEventCard(event, showOffset: false)),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildSessionPanel() {
    LifecycleSessionSnapshotModel? selected;
    for (final snapshot in _snapshots) {
      if (snapshot.sessionID == _selectedSessionID) {
        selected = snapshot;
        break;
      }
    }
    final sessionIsLive = _isSessionSnapshotLive(selected);
    final retryEnabled = _canRetrySession(selected);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        DropdownButtonFormField<String>(
          initialValue: _selectedSessionID,
          decoration: const InputDecoration(
            border: OutlineInputBorder(),
            labelText: 'Session',
            isDense: true,
          ),
          items: _snapshots
              .map(
                (snapshot) => DropdownMenuItem<String>(
                  value: snapshot.sessionID,
                  child: Text(
                    '${snapshot.sessionID} (${snapshot.currentState})',
                  ),
                ),
              )
              .toList(growable: false),
          onChanged: (String? value) async {
            if (value == null) {
              return;
            }
            setState(() {
              _selectedSessionID = value;
              _latestSessionLiveEvent = null;
              _sessionHistory = const <StreamEvent>[];
              _nextSessionActivityOffset = 0;
            });
            await _reloadSessionHistory();
          },
        ),
        const SizedBox(height: 8),
        if (selected != null) ...<Widget>[
          Text('State: ${selected.currentState} (${selected.currentSeverity})'),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              OutlinedButton(
                onPressed: sessionIsLive
                    ? () => _runManualIntervention('nudge')
                    : null,
                child: const Text('Nudge'),
              ),
              OutlinedButton(
                onPressed: retryEnabled
                    ? () => _runManualIntervention('retry')
                    : null,
                child: const Text('Retry'),
              ),
              OutlinedButton(
                onPressed: sessionIsLive
                    ? () => _runManualIntervention('pause')
                    : null,
                child: const Text('Pause'),
              ),
              OutlinedButton(
                onPressed: sessionIsLive
                    ? () => _runManualIntervention('terminate')
                    : null,
                child: const Text('Terminate'),
              ),
            ],
          ),
          const SizedBox(height: 8),
        ],
        if (_shouldShowSessionRealtimeBlock(selected)) ...<Widget>[
          Text('Realtime', style: Theme.of(context).textTheme.labelMedium),
          const SizedBox(height: 8),
          _buildEventCard(_latestSessionLiveEvent!),
          const SizedBox(height: 12),
        ],
        if (_sessionHistory.isEmpty)
          const Text('No session history loaded yet.')
        else ...<Widget>[
          Text('History', style: Theme.of(context).textTheme.labelMedium),
          const SizedBox(height: 8),
          ..._sessionHistory.map(_buildEventCard),
        ],
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final activeNowCount = _activeLiveEntriesByKey.length;
    final showGlobalEmptyState =
        _mode == _EventsBoardMode.globalLive && activeNowCount == 0;
    return Scaffold(
      appBar: AppBar(
        leadingWidth: 260,
        leading: Row(
          children: <Widget>[
            const BackButton(),
            Expanded(
              child: SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: _liveFeedBadgeBackgroundColor(),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: <Widget>[
                          Icon(
                            Icons.circle,
                            size: 8,
                            color: _liveFeedStatusDotColor(),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            _liveFeedStatusLabel(),
                            style: Theme.of(context).textTheme.labelSmall,
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: Theme.of(
                          context,
                        ).colorScheme.surfaceContainerHighest,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        'Active $activeNowCount',
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
        centerTitle: true,
        title: const Text('Session Matrix'),
        actions: <Widget>[
          if (_isLoading)
            const Padding(
              padding: EdgeInsets.only(right: 6),
              child: Center(
                child: SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ),
            ),
          TextButton(
            onPressed: _expandedEventIDs.isEmpty
                ? null
                : () {
                    setState(() {
                      _expandedEventIDs.clear();
                    });
                  },
            child: const Text('Collapse All'),
          ),
          IconButton(
            onPressed: _refresh,
            tooltip: 'Reload',
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: SafeArea(
        child: Column(
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Wrap(
                spacing: 8,
                children: <Widget>[
                  ChoiceChip(
                    label: const Text('Global Live'),
                    selected: _mode == _EventsBoardMode.globalLive,
                    onSelected: (_) {
                      setState(() => _mode = _EventsBoardMode.globalLive);
                      _pipelineEventsSubscription?.cancel();
                      _sessionActivitySubscription?.cancel();
                    },
                  ),
                  ChoiceChip(
                    label: const Text('Pipeline Drilldown'),
                    selected: _mode == _EventsBoardMode.pipelineDrilldown,
                    onSelected: (_) async {
                      setState(
                        () => _mode = _EventsBoardMode.pipelineDrilldown,
                      );
                      _sessionActivitySubscription?.cancel();
                      await _reloadPipelineEvents();
                    },
                  ),
                  ChoiceChip(
                    label: const Text('Session Inspection'),
                    selected: _mode == _EventsBoardMode.sessionInspection,
                    onSelected: (_) async {
                      setState(
                        () => _mode = _EventsBoardMode.sessionInspection,
                      );
                      _pipelineEventsSubscription?.cancel();
                      await _reloadSessionHistory();
                    },
                  ),
                ],
              ),
            ),
            const SizedBox(height: 8),
            Expanded(
              child: showGlobalEmptyState
                  ? Center(
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: <Widget>[
                            const Text(
                              'No active worker activity right now.',
                              textAlign: TextAlign.center,
                            ),
                            if (_statusMessage != null) ...<Widget>[
                              const SizedBox(height: 10),
                              Text(_statusMessage!),
                            ],
                          ],
                        ),
                      ),
                    )
                  : SingleChildScrollView(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        children: <Widget>[
                          if (_mode == _EventsBoardMode.globalLive)
                            _buildGlobalPanel()
                          else if (_mode == _EventsBoardMode.pipelineDrilldown)
                            _buildPipelinePanel()
                          else
                            _buildSessionPanel(),
                          if (_statusMessage != null) ...<Widget>[
                            const SizedBox(height: 10),
                            Text(_statusMessage!),
                          ],
                        ],
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
