import 'dart:async';

import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class WorkerSessionsScreen extends StatefulWidget {
  const WorkerSessionsScreen({
    required this.api,
    required this.statusMessage,
    required this.onStatus,
    super.key,
  });

  final ControlPlaneApi api;
  final String? statusMessage;
  final void Function(String message) onStatus;

  @override
  State<WorkerSessionsScreen> createState() => _WorkerSessionsScreenState();
}

class _WorkerSessionsScreenState extends State<WorkerSessionsScreen> {
  List<WorkerSession> _sessions = const <WorkerSession>[];
  bool _loading = false;
  StreamSubscription<ApiResult<StreamEvent>>? _subscription;

  @override
  void initState() {
    super.initState();
    unawaited(_reload());
    _subscription = widget.api.workerSessionStream().listen((
      ApiResult<StreamEvent> event,
    ) {
      if (event.isSuccess) {
        unawaited(_reload());
      }
    });
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }

  Future<void> _reload() async {
    setState(() => _loading = true);
    final result = await widget.api.workerSessions(limit: 100);
    if (!mounted) {
      return;
    }
    if (!result.isSuccess || result.data == null) {
      widget.onStatus(
        'Loading worker sessions failed: ${result.errorMessage ?? 'unknown error'}',
      );
      setState(() => _loading = false);
      return;
    }
    setState(() {
      _sessions = result.data!;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        if (widget.statusMessage != null && widget.statusMessage!.isNotEmpty)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Text(widget.statusMessage!),
          ),
        Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: <Widget>[
              ElevatedButton.icon(
                onPressed: _loading ? null : _reload,
                icon: const Icon(Icons.refresh),
                label: const Text('Reload Workers'),
              ),
              const SizedBox(width: 12),
              Text('Workers: ${_sessions.length}'),
            ],
          ),
        ),
        Expanded(
          child: _loading
              ? const Center(child: CircularProgressIndicator())
              : ListView.separated(
                  itemCount: _sessions.length,
                  separatorBuilder: (_, __) => const Divider(height: 1),
                  itemBuilder: (BuildContext context, int index) {
                    final session = _sessions[index];
                    return ListTile(
                      title: Text(
                        '${session.workerID} (epoch ${session.epoch})',
                      ),
                      subtitle: Text(
                        'state=${session.state}, lastHeartbeat=${session.lastHeartbeat.toIso8601String()}, leaseExpiresAt=${session.leaseExpiresAt.toIso8601String()}',
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}
