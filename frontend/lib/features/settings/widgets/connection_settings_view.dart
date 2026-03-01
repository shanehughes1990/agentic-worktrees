import 'package:flutter/material.dart';

class ConnectionSettingsView extends StatelessWidget {
  const ConnectionSettingsView({
    required this.endpointController,
    required this.isSavingEndpoint,
    required this.isRunningAction,
    required this.onSave,
    required this.onTest,
    required this.statusMessage,
    super.key,
  });

  final TextEditingController endpointController;
  final bool isSavingEndpoint;
  final bool isRunningAction;
  final VoidCallback onSave;
  final VoidCallback onTest;
  final String? statusMessage;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            'Connection Settings',
            style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              Expanded(
                child: TextField(
                  controller: endpointController,
                  decoration: const InputDecoration(
                    labelText: 'GraphQL HTTP Endpoint',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: isSavingEndpoint ? null : onSave,
                child: const Text('Save'),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: isRunningAction ? null : onTest,
                child: const Text('Test'),
              ),
            ],
          ),
          if (statusMessage != null) ...<Widget>[
            const SizedBox(height: 12),
            Text(statusMessage!, maxLines: 2, overflow: TextOverflow.ellipsis),
          ],
        ],
      ),
    );
  }
}
