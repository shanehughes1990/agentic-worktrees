import 'package:agentic_worktrees/features/projects/logic/project_setup_logic.dart';
import 'package:agentic_worktrees/features/projects/widgets/project_setups_list.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class ProjectSetupScreen extends StatefulWidget {
  const ProjectSetupScreen({
    required this.projectController,
    required this.projectNameController,
    required this.repositoryUrlController,
    required this.taskboardNameController,
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
  final TextEditingController taskboardNameController;
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
  State<ProjectSetupScreen> createState() => _ProjectSetupScreenState();
}

class _ProjectSetupScreenState extends State<ProjectSetupScreen> {
  final List<TextEditingController> _repositoryControllers =
      <TextEditingController>[];
  final TextEditingController _taskboardNameController =
      TextEditingController();

  String _lastRepositoryRaw = '';
  String _lastTaskboardNameRaw = '';

  @override
  void initState() {
    super.initState();
    _syncDraftControllersFromForm(force: true);
  }

  @override
  void didUpdateWidget(covariant ProjectSetupScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    _syncDraftControllersFromForm();
  }

  @override
  void dispose() {
    for (final TextEditingController controller in _repositoryControllers) {
      controller.dispose();
    }
    _taskboardNameController.dispose();
    super.dispose();
  }

  void _syncDraftControllersFromForm({bool force = false}) {
    final repositoryRaw = widget.repositoryUrlController.text;
    final taskboardNameRaw = widget.taskboardNameController.text;

    if (!force &&
        repositoryRaw == _lastRepositoryRaw &&
        taskboardNameRaw == _lastTaskboardNameRaw) {
      return;
    }

    _lastRepositoryRaw = repositoryRaw;
    _lastTaskboardNameRaw = taskboardNameRaw;

    for (final TextEditingController controller in _repositoryControllers) {
      controller.dispose();
    }
    _repositoryControllers.clear();

    final repositoryURLs = ProjectSetupLogic.parseMultilineEntries(
      repositoryRaw,
    );
    if (repositoryURLs.isEmpty) {
      _repositoryControllers.add(TextEditingController());
    } else {
      _repositoryControllers.addAll(
        repositoryURLs.map((String url) => TextEditingController(text: url)),
      );
    }

    final taskboardNames = ProjectSetupLogic.parseMultilineEntries(
      taskboardNameRaw,
    );
    _taskboardNameController.text = taskboardNames.isNotEmpty
        ? taskboardNames.first
        : '';

    if (mounted) {
      setState(() {});
    }
  }

  void _syncFormControllersFromDraft() {
    widget.repositoryUrlController.text = _repositoryControllers
        .map((TextEditingController controller) => controller.text.trim())
        .where((String value) => value.isNotEmpty)
        .join('\n');

    widget.taskboardNameController.text = _taskboardNameController.text.trim();

    _lastRepositoryRaw = widget.repositoryUrlController.text;
    _lastTaskboardNameRaw = widget.taskboardNameController.text;
  }

  void _onProjectNameChanged(String value) {
    final generatedID = ProjectSetupLogic.projectIDFromName(value);
    if (widget.projectController.text != generatedID) {
      widget.projectController.text = generatedID;
    }
  }

  void _addRepositoryBlock() {
    setState(() {
      _repositoryControllers.add(TextEditingController());
    });
  }

  void _removeRepositoryBlock(int index) {
    if (_repositoryControllers.length <= 1) {
      return;
    }
    setState(() {
      final removed = _repositoryControllers.removeAt(index);
      removed.dispose();
    });
  }

  @override
  Widget build(BuildContext context) {
    final hasProvider = widget.setupScmProvider.trim().isNotEmpty;

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
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  const Text(
                    'Project Setup',
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: widget.projectNameController,
                    decoration: const InputDecoration(
                      labelText: 'Project Name',
                      border: OutlineInputBorder(),
                    ),
                    onChanged: _onProjectNameChanged,
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: widget.projectController,
                    readOnly: true,
                    decoration: const InputDecoration(
                      labelText: 'Project ID',
                      border: OutlineInputBorder(),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  const Text(
                    'SCM Provider',
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 12),
                  DropdownButtonFormField<String>(
                    initialValue: widget.setupScmProvider,
                    decoration: const InputDecoration(
                      labelText: 'Provider',
                      border: OutlineInputBorder(),
                    ),
                    items: const <DropdownMenuItem<String>>[
                      DropdownMenuItem<String>(
                        value: 'GITHUB',
                        child: Text('GitHub'),
                      ),
                    ],
                    onChanged: (String? value) {
                      if (value == null) {
                        return;
                      }
                      widget.onSetupScmProviderChanged(value);
                      widget.onSetupTrackerProviderChanged(
                        ProjectSetupLogic.defaultTrackerProvider,
                      );
                    },
                  ),
                  if (hasProvider) ...<Widget>[
                    const SizedBox(height: 12),
                    _RepositorySetupSection(
                      controllers: _repositoryControllers,
                      onAdd: _addRepositoryBlock,
                      onRemove: _removeRepositoryBlock,
                    ),
                  ],
                ],
              ),
            ),
          ),
          if (hasProvider) ...<Widget>[
            const SizedBox(height: 12),
            _TrackerSetupSection(
              trackerProvider: widget.setupTrackerProvider,
              onTrackerProviderChanged: widget.onSetupTrackerProviderChanged,
              taskboardNameController: _taskboardNameController,
            ),
          ],
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              FilledButton(
                onPressed: widget.isSavingProjectSetup
                    ? null
                    : () {
                        _syncFormControllersFromDraft();
                        widget.onSaveProjectSetup();
                      },
                child: const Text('Save Project Setup'),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: widget.onReloadProjectSetups,
                child: const Text('Reload'),
              ),
            ],
          ),
          if (widget.projectSetups.isNotEmpty) ...<Widget>[
            const SizedBox(height: 12),
            const Text(
              'Configured Projects',
              style: TextStyle(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 8),
            ProjectSetupsList(
              projectSetups: widget.projectSetups,
              selectedProjectID: widget.selectedProjectID,
              onProjectSelected: widget.onProjectSelected,
            ),
          ],
          if (widget.statusMessage != null) ...<Widget>[
            const SizedBox(height: 12),
            Text(
              widget.statusMessage!,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ],
      ),
    );
  }
}

class _RepositorySetupSection extends StatelessWidget {
  const _RepositorySetupSection({
    required this.controllers,
    required this.onAdd,
    required this.onRemove,
  });

  final List<TextEditingController> controllers;
  final VoidCallback onAdd;
  final ValueChanged<int> onRemove;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          'Repository Setup',
          style: TextStyle(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        Row(
          children: <Widget>[
            const Spacer(),
            OutlinedButton.icon(
              onPressed: onAdd,
              icon: const Icon(Icons.add),
              label: const Text('Add Repository'),
            ),
          ],
        ),
        const SizedBox(height: 8),
        for (var index = 0; index < controllers.length; index++) ...<Widget>[
          Card(
            color: Theme.of(context).colorScheme.surfaceContainerHighest,
            child: Padding(
              padding: const EdgeInsets.all(10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Row(
                    children: <Widget>[
                      Text('Repository ${index + 1}'),
                      const Spacer(),
                      if (controllers.length > 1)
                        IconButton(
                          onPressed: () => onRemove(index),
                          icon: const Icon(Icons.delete_outline),
                          tooltip: 'Remove Repository',
                        ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: controllers[index],
                    decoration: const InputDecoration(
                      labelText: 'Repository URL',
                      border: OutlineInputBorder(),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),
        ],
      ],
    );
  }
}

class _TrackerSetupSection extends StatelessWidget {
  const _TrackerSetupSection({
    required this.trackerProvider,
    required this.onTrackerProviderChanged,
    required this.taskboardNameController,
  });

  final String trackerProvider;
  final ValueChanged<String> onTrackerProviderChanged;
  final TextEditingController taskboardNameController;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Text(
              'Tracker Setup',
              style: TextStyle(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 8),
            DropdownButtonFormField<String>(
              initialValue: trackerProvider,
              decoration: const InputDecoration(
                labelText: 'Tracker Provider',
                border: OutlineInputBorder(),
              ),
              items: const <DropdownMenuItem<String>>[
                DropdownMenuItem<String>(
                  value: 'GITHUB_ISSUES',
                  child: Text('GitHub Issues'),
                ),
                DropdownMenuItem<String>(
                  value: 'INTERNAL',
                  child: Text('Internal Taskboard'),
                ),
              ],
              onChanged: (String? value) {
                if (value == null) {
                  return;
                }
                onTrackerProviderChanged(value);
              },
            ),
            const SizedBox(height: 8),
            TextField(
              controller: taskboardNameController,
              decoration: const InputDecoration(
                labelText: 'Taskboard Name',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
