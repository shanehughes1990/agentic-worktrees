import 'package:agentic_worktrees/features/settings/widgets/connection_settings_view.dart';
import 'package:flutter/material.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({
    required this.endpointController,
    required this.isSavingEndpoint,
    required this.isRunningAction,
    required this.onSaveEndpoint,
    required this.onTestConnection,
    required this.statusMessage,
    super.key,
  });

  final TextEditingController endpointController;
  final bool isSavingEndpoint;
  final bool isRunningAction;
  final VoidCallback onSaveEndpoint;
  final VoidCallback onTestConnection;
  final String? statusMessage;

  @override
  Widget build(BuildContext context) {
    return ConnectionSettingsView(
      endpointController: endpointController,
      isSavingEndpoint: isSavingEndpoint,
      isRunningAction: isRunningAction,
      onSave: onSaveEndpoint,
      onTest: onTestConnection,
      statusMessage: statusMessage,
    );
  }
}
