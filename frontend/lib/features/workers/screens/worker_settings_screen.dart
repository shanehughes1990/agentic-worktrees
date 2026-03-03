import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class WorkerSettingsScreen extends StatefulWidget {
  const WorkerSettingsScreen({
    required this.api,
    required this.statusMessage,
    required this.onStatus,
    super.key,
  });

  final ControlPlaneApi api;
  final String? statusMessage;
  final void Function(String message) onStatus;

  @override
  State<WorkerSettingsScreen> createState() => _WorkerSettingsScreenState();
}

class _WorkerSettingsScreenState extends State<WorkerSettingsScreen> {
  final TextEditingController _heartbeatController = TextEditingController();
  final TextEditingController _deadlineController = TextEditingController();
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _heartbeatController.dispose();
    _deadlineController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final result = await widget.api.workerSettings();
    if (!mounted || !result.isSuccess || result.data == null) {
      widget.onStatus(
        'Loading worker settings failed: ${result.errorMessage ?? 'unknown error'}',
      );
      return;
    }
    final settings = result.data!;
    _heartbeatController.text = settings.heartbeatIntervalSeconds.toString();
    _deadlineController.text = settings.responseDeadlineSeconds.toString();
    widget.onStatus('Loaded worker settings.');
  }

  Future<void> _save() async {
    final heartbeat = int.tryParse(_heartbeatController.text.trim());
    final deadline = int.tryParse(_deadlineController.text.trim());
    if (heartbeat == null || deadline == null) {
      widget.onStatus('All settings fields must be valid integers.');
      return;
    }
    setState(() => _saving = true);
    final result = await widget.api.updateWorkerSettings(
      heartbeatIntervalSeconds: heartbeat,
      responseDeadlineSeconds: deadline,
    );
    if (!mounted) {
      return;
    }
    setState(() => _saving = false);
    widget.onStatus(
      result.isSuccess
          ? 'Worker settings updated successfully.'
          : 'Worker settings update failed: ${result.errorMessage}',
    );
  }

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: <Widget>[
        if (widget.statusMessage != null && widget.statusMessage!.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: Text(widget.statusMessage!),
          ),
        _numberField(
          label: 'Heartbeat Interval (seconds)',
          controller: _heartbeatController,
        ),
        _numberField(
          label: 'Response Deadline (seconds)',
          controller: _deadlineController,
        ),
        const SizedBox(height: 16),
        Row(
          children: <Widget>[
            ElevatedButton(
              onPressed: _saving ? null : _save,
              child: const Text('Save Worker Settings'),
            ),
            const SizedBox(width: 12),
            OutlinedButton(
              onPressed: _saving ? null : _load,
              child: const Text('Reload'),
            ),
          ],
        ),
      ],
    );
  }

  Widget _numberField({
    required String label,
    required TextEditingController controller,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: TextField(
        controller: controller,
        keyboardType: TextInputType.number,
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
        ),
      ),
    );
  }
}
