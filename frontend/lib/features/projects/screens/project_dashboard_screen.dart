import 'package:agentic_worktrees/features/projects/logic/project_setup_logic.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:agentic_worktrees/shared/graph/typed/client.dart';
import 'package:flutter/material.dart';

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
  late final TextEditingController _projectIDController;
  late final TextEditingController _projectNameController;
  late final TextEditingController _scmTokenController;
  final List<TextEditingController> _repositoryControllers =
      <TextEditingController>[];

  bool _isSaving = false;
  String? _statusMessage;
  String _scmProvider = ProjectSetupLogic.defaultScmProvider;

  @override
  void initState() {
    super.initState();
    _api = ControlPlaneApi(buildGraphqlClient(widget.endpoint));
    _projectIDController = TextEditingController();
    _projectNameController = TextEditingController();
    _scmTokenController = TextEditingController();
    _applySetup(widget.projectSetup);
  }

  @override
  void dispose() {
    _projectIDController.dispose();
    _projectNameController.dispose();
    _scmTokenController.dispose();
    for (final TextEditingController controller in _repositoryControllers) {
      controller.dispose();
    }
    super.dispose();
  }

  void _applySetup(ProjectSetupConfig setup) {
    _projectIDController.text = setup.projectID;
    _projectNameController.text = setup.projectName;

    for (final TextEditingController controller in _repositoryControllers) {
      controller.dispose();
    }
    _repositoryControllers.clear();
    if (setup.repositories.isEmpty) {
      _repositoryControllers.add(TextEditingController());
    } else {
      _repositoryControllers.addAll(
        setup.repositories.map(
          (ProjectRepositoryConfig repository) =>
              TextEditingController(text: repository.repositoryURL),
        ),
      );
    }

    _scmProvider = setup.scms.isNotEmpty
        ? setup.scms.first.scmProvider
        : ProjectSetupLogic.defaultScmProvider;
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

  Future<void> _saveProjectSetup() async {
    final projectID = _projectIDController.text.trim();
    final projectName = _projectNameController.text.trim();
    final repositoryURLs = _repositoryControllers
        .map((TextEditingController controller) => controller.text.trim())
        .where((String value) => value.isNotEmpty)
        .toList(growable: false);
    final validationError = ProjectSetupLogic.validateRequiredFields(
      projectID: projectID,
      projectName: projectName,
      repositoryURLs: repositoryURLs,
      scmToken: _scmTokenController.text,
    );
    if (validationError != null) {
      setState(() => _statusMessage = validationError);
      return;
    }

    setState(() {
      _isSaving = true;
      _statusMessage = null;
    });

    final response = await _api.upsertProjectSetup(
      projectID: projectID,
      projectName: projectName,
      scmProvider: _scmProvider,
      repositoryURLs: repositoryURLs,
      scmToken: _scmTokenController.text,
    );

    if (!mounted) {
      return;
    }

    setState(() {
      _isSaving = false;
      _statusMessage = response.isSuccess
          ? 'Project setup saved.'
          : 'Save failed: ${response.errorMessage ?? 'unknown error'}';
    });
  }

  @override
  Widget build(BuildContext context) {
    final hasProvider = _scmProvider.trim().isNotEmpty;

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Text(widget.projectSetup.projectName),
            Text(
              _projectIDController.text,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
      ),
      body: Column(
        children: <Widget>[
          Expanded(
            child: SingleChildScrollView(
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
                            controller: _projectNameController,
                            decoration: const InputDecoration(
                              labelText: 'Project Name',
                              border: OutlineInputBorder(),
                            ),
                          ),
                          const SizedBox(height: 12),
                          TextField(
                            controller: _projectIDController,
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
                            initialValue: _scmProvider,
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
                              setState(() => _scmProvider = value);
                            },
                          ),
                          if (hasProvider) ...<Widget>[
                            const SizedBox(height: 12),
                            TextField(
                              controller: _scmTokenController,
                              obscureText: true,
                              decoration: const InputDecoration(
                                labelText: 'SCM Token',
                                border: OutlineInputBorder(),
                              ),
                            ),
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
                    Card(
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
                          ],
                        ),
                      ),
                    ),
                  ],
                  if (_statusMessage != null) ...<Widget>[
                    const SizedBox(height: 12),
                    Text(
                      _statusMessage!,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
          ),
          SafeArea(
            top: false,
            child: Container(
              padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                border: Border(
                  top: BorderSide(color: Theme.of(context).dividerColor),
                ),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  OutlinedButton(
                    onPressed: () => Navigator.of(context).maybePop(),
                    child: const Text('Back'),
                  ),
                  const SizedBox(width: 8),
                  FilledButton(
                    onPressed: _isSaving ? null : _saveProjectSetup,
                    child: const Text('Save Project Setup'),
                  ),
                ],
              ),
            ),
          ),
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
