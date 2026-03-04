import 'package:agentic_repositories/features/projects/logic/project_setup_logic.dart';
import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:agentic_repositories/shared/graph/typed/client.dart';
import 'package:flutter/material.dart';

class ProjectSetupEditScreen extends StatefulWidget {
  const ProjectSetupEditScreen({
    required this.projectSetup,
    required this.endpoint,
    super.key,
  });

  final ProjectSetupConfig projectSetup;
  final String endpoint;

  @override
  State<ProjectSetupEditScreen> createState() => _ProjectSetupEditScreenState();
}

class _ProjectSetupEditScreenState extends State<ProjectSetupEditScreen> {
  late final ControlPlaneApi _api;
  late ProjectSetupConfig _currentSetup;
  late String _savedScmProvider;
  late final TextEditingController _projectIDController;
  late final TextEditingController _projectNameController;
  late final TextEditingController _scmTokenController;
  final Map<String, List<TextEditingController>>
  _repositoryControllersByProvider = <String, List<TextEditingController>>{};

  bool _isSaving = false;
  bool _isRegeneratingToken = false;
  String? _statusMessage;
  String _scmProvider = ProjectSetupLogic.defaultScmProvider;

  @override
  void initState() {
    super.initState();
    _api = ControlPlaneApi(buildGraphqlClient(widget.endpoint));
    _currentSetup = widget.projectSetup;
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
    for (final controllers in _repositoryControllersByProvider.values) {
      for (final TextEditingController controller in controllers) {
        controller.dispose();
      }
    }
    super.dispose();
  }

  List<TextEditingController> _repositoryControllersFor(String provider) {
    final cleanProvider = provider.trim();
    if (cleanProvider.isEmpty) {
      return <TextEditingController>[];
    }
    final existing = _repositoryControllersByProvider[cleanProvider];
    if (existing != null) {
      return existing;
    }
    final created = <TextEditingController>[TextEditingController()];
    _repositoryControllersByProvider[cleanProvider] = created;
    return created;
  }

  void _applySetup(ProjectSetupConfig setup) {
    _projectIDController.text = setup.projectID;
    _projectNameController.text = setup.projectName;
    _scmTokenController.clear();

    for (final controllers in _repositoryControllersByProvider.values) {
      for (final TextEditingController controller in controllers) {
        controller.dispose();
      }
    }
    _repositoryControllersByProvider.clear();

    final scmProviderByID = <String, String>{
      for (final scm in setup.scms)
        if (scm.scmID.trim().isNotEmpty && scm.scmProvider.trim().isNotEmpty)
          scm.scmID.trim(): scm.scmProvider.trim(),
    };

    for (final repository in setup.repositories) {
      final provider =
          scmProviderByID[repository.scmID.trim()] ??
          ProjectSetupLogic.defaultScmProvider;
      _repositoryControllersByProvider
          .putIfAbsent(provider, () => <TextEditingController>[])
          .add(TextEditingController(text: repository.repositoryURL));
    }

    _scmProvider = setup.scms.isNotEmpty
        ? setup.scms.first.scmProvider
        : ProjectSetupLogic.defaultScmProvider;
    _repositoryControllersFor(_scmProvider);
    _savedScmProvider = _scmProvider;
    _isRegeneratingToken = false;
  }

  Future<void> _onScmProviderChanged(String value) async {
    if (value == _scmProvider) {
      return;
    }
    final shouldContinue = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: const Text('Change SCM Provider?'),
          content: const Text(
            'Switching providers will override provider credentials on save.',
          ),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Continue'),
            ),
          ],
        );
      },
    );
    if (shouldContinue != true || !mounted) {
      return;
    }
    setState(() {
      _scmProvider = value;
      _isRegeneratingToken = true;
      _scmTokenController.clear();
      _statusMessage = null;
    });
  }

  void _startTokenRegeneration() {
    setState(() {
      _isRegeneratingToken = true;
      _scmTokenController.clear();
      _statusMessage = null;
    });
  }

  void _cancelTokenRegeneration() {
    setState(() {
      _isRegeneratingToken = false;
      _scmTokenController.clear();
      if (_scmProvider == _savedScmProvider) {
        _statusMessage = null;
      }
    });
  }

  void _addRepositoryBlock() {
    final controllers = _repositoryControllersFor(_scmProvider);
    setState(() {
      controllers.add(TextEditingController());
    });
  }

  void _removeRepositoryBlock(int index) {
    final controllers = _repositoryControllersFor(_scmProvider);
    if (controllers.length <= 1) {
      return;
    }
    setState(() {
      final removed = controllers.removeAt(index);
      removed.dispose();
    });
  }

  Future<void> _saveProjectSetup() async {
    final projectID = _projectIDController.text.trim();
    final projectName = _projectNameController.text.trim();
    final repositoryURLs = _repositoryControllersFor(_scmProvider)
        .map((TextEditingController controller) => controller.text.trim())
        .where((String value) => value.isNotEmpty)
        .toList(growable: false);
    String? validationError;
    if (projectID.isEmpty || projectName.isEmpty || repositoryURLs.isEmpty) {
      validationError =
          'Project ID, Project Name, and at least one Repository URL are required.';
    }
    final requiresNewCredentials =
        _isRegeneratingToken || _scmProvider != _savedScmProvider;
    if (validationError == null &&
        requiresNewCredentials &&
        _scmTokenController.text.trim().isEmpty) {
      validationError =
          'SCM Token is required when regenerating or changing provider credentials.';
    }
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
      scmToken: _scmTokenController.text.trim(),
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

    if (response.isSuccess && response.data != null) {
      _currentSetup = response.data!;
      _applySetup(_currentSetup);
    }
  }

  void _closeEditor() {
    Navigator.of(context).pop(_currentSetup);
  }

  @override
  Widget build(BuildContext context) {
    final hasProvider = _scmProvider.trim().isNotEmpty;
    final repositoryControllers = _repositoryControllersFor(_scmProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Edit Project Setup'),
        leading: IconButton(
          onPressed: _closeEditor,
          icon: const Icon(Icons.arrow_back),
          tooltip: 'Back',
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
                              _onScmProviderChanged(value);
                            },
                          ),
                          if (hasProvider) ...<Widget>[
                            const SizedBox(height: 12),
                            if (!_isRegeneratingToken)
                              Align(
                                alignment: Alignment.centerLeft,
                                child: OutlinedButton.icon(
                                  onPressed: _startTokenRegeneration,
                                  icon: const Icon(Icons.refresh),
                                  label: const Text('Regenerate Token'),
                                ),
                              ),
                            if (_isRegeneratingToken)
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: <Widget>[
                                  TextField(
                                    controller: _scmTokenController,
                                    obscureText: true,
                                    decoration: const InputDecoration(
                                      labelText: 'New SCM Token',
                                      border: OutlineInputBorder(),
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  TextButton.icon(
                                    onPressed: _cancelTokenRegeneration,
                                    icon: const Icon(Icons.delete_outline),
                                    label: const Text('Delete token field'),
                                  ),
                                ],
                              ),
                            const SizedBox(height: 12),
                            _RepositorySetupSection(
                              provider: _scmProvider,
                              controllers: repositoryControllers,
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
                    SizedBox(
                      width: double.infinity,
                      child: Card(
                        child: Padding(
                          padding: const EdgeInsets.all(12),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: const <Widget>[
                              Text(
                                'Tracker Setup',
                                style: TextStyle(fontWeight: FontWeight.w600),
                              ),
                              SizedBox(height: 8),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                  if (_statusMessage != null) ...<Widget>[
                    const SizedBox(height: 12),
                    Text(_statusMessage!),
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
                    onPressed: _closeEditor,
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
    required this.provider,
    required this.controllers,
    required this.onAdd,
    required this.onRemove,
  });

  final String provider;
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
        const SizedBox(height: 4),
        Text(
          'Attached SCM Provider: $provider',
          style: Theme.of(context).textTheme.bodySmall,
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
