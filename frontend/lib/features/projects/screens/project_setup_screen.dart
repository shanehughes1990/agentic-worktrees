import 'package:agentic_worktrees/features/projects/logic/project_setup_logic.dart';
import 'package:flutter/material.dart';

class ProjectSetupScreen extends StatefulWidget {
  const ProjectSetupScreen({
    required this.projectController,
    required this.projectNameController,
    required this.repositoryUrlController,
    required this.setupScmProvider,
    required this.scmTokenController,
    required this.onSetupScmProviderChanged,
    required this.isSavingProjectSetup,
    required this.onSaveProjectSetup,
    required this.statusMessage,
    super.key,
  });

  final TextEditingController projectController;
  final TextEditingController projectNameController;
  final TextEditingController repositoryUrlController;
  final String setupScmProvider;
  final TextEditingController scmTokenController;
  final ValueChanged<String> onSetupScmProviderChanged;
  final bool isSavingProjectSetup;
  final VoidCallback onSaveProjectSetup;
  final String? statusMessage;

  @override
  State<ProjectSetupScreen> createState() => _ProjectSetupScreenState();
}

class _ProjectSetupScreenState extends State<ProjectSetupScreen> {
  final List<TextEditingController> _repositoryControllers =
      <TextEditingController>[];
  String _lastRepositoryRaw = '';

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
    super.dispose();
  }

  void _syncDraftControllersFromForm({bool force = false}) {
    final repositoryRaw = widget.repositoryUrlController.text;

    if (!force && repositoryRaw == _lastRepositoryRaw) {
      return;
    }

    _lastRepositoryRaw = repositoryRaw;

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

    if (mounted) {
      setState(() {});
    }
  }

  void _syncFormControllersFromDraft() {
    widget.repositoryUrlController.text = _repositoryControllers
        .map((TextEditingController controller) => controller.text.trim())
        .where((String value) => value.isNotEmpty)
        .join('\n');

    _lastRepositoryRaw = widget.repositoryUrlController.text;
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

    return Column(
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
                          },
                        ),
                        if (hasProvider) ...<Widget>[
                          const SizedBox(height: 12),
                          TextField(
                            controller: widget.scmTokenController,
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
                  onPressed: widget.isSavingProjectSetup
                      ? null
                      : () {
                          _syncFormControllersFromDraft();
                          widget.onSaveProjectSetup();
                        },
                  child: const Text('Save Project Setup'),
                ),
              ],
            ),
          ),
        ),
      ],
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
