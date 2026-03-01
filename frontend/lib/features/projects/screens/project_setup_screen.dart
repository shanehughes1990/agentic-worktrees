import 'package:agentic_worktrees/features/projects/logic/project_setup_logic.dart';
import 'package:agentic_worktrees/features/projects/widgets/project_setups_list.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class ProjectSetupScreen extends StatelessWidget {
  const ProjectSetupScreen({
    required this.projectController,
    required this.projectNameController,
    required this.repositoryUrlController,
    required this.trackerLocationController,
    required this.trackerBoardIDController,
    required this.setupScmProvider,
    required this.setupTrackerProvider,
    required this.onSetupScmProviderChanged,
    required this.onSetupTrackerProviderChanged,
    required this.isSavingProjectSetup,
    required this.onSaveProjectSetup,
    required this.onReloadProjectSetups,
    required this.projectSetups,
    required this.selectedProjectID,
    required this.onProjectSelected,
    required this.statusMessage,
    super.key,
  });

  final TextEditingController projectController;
  final TextEditingController projectNameController;
  final TextEditingController repositoryUrlController;
  final TextEditingController trackerLocationController;
  final TextEditingController trackerBoardIDController;
  final String setupScmProvider;
  final String setupTrackerProvider;
  final ValueChanged<String> onSetupScmProviderChanged;
  final ValueChanged<String> onSetupTrackerProviderChanged;
  final bool isSavingProjectSetup;
  final VoidCallback onSaveProjectSetup;
  final VoidCallback onReloadProjectSetups;
  final List<ProjectSetupConfig> projectSetups;
  final String selectedProjectID;
  final ValueChanged<ProjectSetupConfig> onProjectSelected;
  final String? statusMessage;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            'Project Setup',
            style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: projectController,
            decoration: const InputDecoration(
              labelText: 'Project ID',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: projectNameController,
            decoration: const InputDecoration(
              labelText: 'Project Name',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            initialValue: setupScmProvider,
            decoration: const InputDecoration(
              labelText: 'SCM Provider',
              border: OutlineInputBorder(),
            ),
            items: const <DropdownMenuItem<String>>[
              DropdownMenuItem<String>(value: 'GITHUB', child: Text('GitHub')),
            ],
            onChanged: (String? value) {
              if (value == null) {
                return;
              }
              onSetupScmProviderChanged(value);
            },
          ),
          const SizedBox(height: 12),
          TextField(
            controller: repositoryUrlController,
            decoration: const InputDecoration(
              labelText: 'Repository URL',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            initialValue: setupTrackerProvider,
            decoration: const InputDecoration(
              labelText: 'Tracker Provider',
              border: OutlineInputBorder(),
            ),
            items: ProjectSetupLogic.trackerProviderOptions
                .map(
                  (String provider) => DropdownMenuItem<String>(
                    value: provider,
                    child: Text(_trackerLabel(provider)),
                  ),
                )
                .toList(growable: false),
            onChanged: (String? value) {
              if (value == null) {
                return;
              }
              onSetupTrackerProviderChanged(value);
            },
          ),
          const SizedBox(height: 12),
          TextField(
            controller: trackerLocationController,
            decoration: const InputDecoration(
              labelText: 'Tracker Location',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: trackerBoardIDController,
            decoration: const InputDecoration(
              labelText: 'Tracker Board ID (optional)',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              FilledButton(
                onPressed: isSavingProjectSetup ? null : onSaveProjectSetup,
                child: const Text('Save Project Setup'),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: onReloadProjectSetups,
                child: const Text('Reload'),
              ),
            ],
          ),
          if (projectSetups.isNotEmpty) ...<Widget>[
            const SizedBox(height: 12),
            const Text(
              'Configured Projects',
              style: TextStyle(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 8),
            ProjectSetupsList(
              projectSetups: projectSetups,
              selectedProjectID: selectedProjectID,
              onProjectSelected: onProjectSelected,
            ),
          ],
          if (statusMessage != null) ...<Widget>[
            const SizedBox(height: 12),
            Text(statusMessage!, maxLines: 2, overflow: TextOverflow.ellipsis),
          ],
        ],
      ),
    );
  }

  static String _trackerLabel(String provider) {
    switch (provider) {
      case 'GITHUB_ISSUES':
        return 'GitHub Issues';
      case 'JIRA':
        return 'Jira';
      case 'LOCAL_JSON':
        return 'Local JSON';
      case 'LINEAR':
        return 'Linear';
      default:
        return provider;
    }
  }
}
