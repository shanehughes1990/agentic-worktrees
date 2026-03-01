import 'dart:async';
import 'dart:io';

import 'package:agentic_worktrees/shared/config/app_config.dart';
import 'package:agentic_worktrees/graph/typed/control_plane.dart';
import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/services.dart';

enum _DashboardView { dashboard, projectSetup, settings }

class DashboardScreen extends ConsumerStatefulWidget {
  const DashboardScreen({required this.initialEndpoint, super.key});

  final String initialEndpoint;

  @override
  ConsumerState<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends ConsumerState<DashboardScreen> {
  late final TextEditingController _endpointController;
  final TextEditingController _sourceController = TextEditingController(
    text: 'octo/repo',
  );
  final TextEditingController _issueReferenceController = TextEditingController(
    text: 'octo/repo#1',
  );
  final TextEditingController _approvedByController = TextEditingController(
    text: 'operator',
  );
  final TextEditingController _projectController = TextEditingController(
    text: 'project-1',
  );
  final TextEditingController _projectNameController = TextEditingController(
    text: 'Project 1',
  );
  final TextEditingController _repositoryUrlController = TextEditingController(
    text: 'https://github.com/acme/repo',
  );
  final TextEditingController _trackerLocationController =
      TextEditingController(text: 'acme/repo');
  final TextEditingController _trackerBoardIDController =
      TextEditingController();
  final TextEditingController _workflowController = TextEditingController(
    text: 'workflow-1',
  );
  final TextEditingController _promptController = TextEditingController(
    text: 'Ingest latest issue board state',
  );
  final TextEditingController _scmOwnerController = TextEditingController(
    text: 'acme',
  );
  final TextEditingController _scmRepoController = TextEditingController(
    text: 'repo',
  );

  SessionSummary? _selectedSession;
  WorkflowJob? _selectedJob;
  String? _statusMessage;
  bool _isSavingEndpoint = false;
  bool _isSavingProjectSetup = false;
  bool _isRunningAction = false;
  int _refreshToken = 0;
  _DashboardView _activeView = _DashboardView.dashboard;
  String _setupScmProvider = 'GITHUB';
  String _setupTrackerProvider = 'GITHUB_ISSUES';
  List<ProjectSetupConfig> _projectSetups = const <ProjectSetupConfig>[];
  final List<StreamEvent> _streamEvents = <StreamEvent>[];
  StreamSubscription<ApiResult<StreamEvent>>? _streamSubscription;

  @override
  void initState() {
    super.initState();
    _endpointController = TextEditingController(text: widget.initialEndpoint);
    unawaited(_loadProjectSetups());
  }

  void _showDashboard(BuildContext context) {
    Navigator.of(context).pop();
    setState(() => _activeView = _DashboardView.dashboard);
  }

  void _showSettings(BuildContext context) {
    Navigator.of(context).pop();
    setState(() => _activeView = _DashboardView.settings);
  }

  void _startNewProjectSetup(BuildContext context) {
    Navigator.of(context).pop();
    setState(() {
      _activeView = _DashboardView.projectSetup;
      _projectController.text = '';
      _projectNameController.text = '';
      _repositoryUrlController.text = '';
      _trackerLocationController.text = '';
      _trackerBoardIDController.text = '';
      _setupScmProvider = 'GITHUB';
      _setupTrackerProvider = 'GITHUB_ISSUES';
      _statusMessage = 'Creating a new project setup';
    });
  }

  Future<void> _exitApp(BuildContext context) async {
    Navigator.of(context).pop();
    await SystemNavigator.pop();
    if (Platform.isMacOS) {
      exit(0);
    }
  }

  @override
  void dispose() {
    _endpointController.dispose();
    _sourceController.dispose();
    _issueReferenceController.dispose();
    _approvedByController.dispose();
    _projectController.dispose();
    _projectNameController.dispose();
    _repositoryUrlController.dispose();
    _trackerLocationController.dispose();
    _trackerBoardIDController.dispose();
    _workflowController.dispose();
    _promptController.dispose();
    _scmOwnerController.dispose();
    _scmRepoController.dispose();
    _streamSubscription?.cancel();
    super.dispose();
  }

  Future<void> _saveEndpoint() async {
    final endpoint = normalizeGraphqlEndpoint(_endpointController.text);
    if (endpoint.isEmpty) {
      setState(() => _statusMessage = 'Endpoint cannot be empty.');
      return;
    }
    setState(() {
      _isSavingEndpoint = true;
      _statusMessage = null;
    });
    await ref.read(appConfigProvider.notifier).saveGraphqlEndpoint(endpoint);
    setState(() {
      _endpointController.text = endpoint;
      _isSavingEndpoint = false;
      _statusMessage = 'Saved endpoint $endpoint';
      _refreshToken++;
    });
  }

  Future<void> _testConnection() async {
    final endpoint = normalizeGraphqlEndpoint(_endpointController.text);
    if (endpoint.isEmpty) {
      setState(() => _statusMessage = 'Endpoint cannot be empty.');
      return;
    }

    setState(() {
      _isRunningAction = true;
      _statusMessage = 'Testing GraphQL connection to $endpoint...';
    });

    AppLogger.instance.logger.i(
      'Testing GraphQL connection from dashboard',
      error: {'endpoint': endpoint},
    );

    await ref.read(appConfigProvider.notifier).saveGraphqlEndpoint(endpoint);
    final api = ControlPlaneApi(buildGraphqlClient(endpoint));
    final result = await api.sessions(limit: 1);

    if (!result.isSuccess) {
      AppLogger.instance.logger.e(
        'Dashboard connection test failed',
        error: {'endpoint': endpoint, 'error': result.errorMessage},
      );
    }

    setState(() {
      _endpointController.text = endpoint;
      _isRunningAction = false;
      _statusMessage = result.isSuccess
          ? 'Connection successful (${result.data?.length ?? 0} session rows returned).'
          : 'Connection failed at $endpoint: ${_compactError(result.errorMessage)}';
      _refreshToken++;
    });

    if (result.isSuccess) {
      await _loadProjectSetups();
    }
  }

  Future<void> _loadProjectSetups() async {
    final endpoint = normalizeGraphqlEndpoint(_endpointController.text);
    if (endpoint.isEmpty) {
      return;
    }
    final api = ControlPlaneApi(buildGraphqlClient(endpoint));
    final response = await api.projectSetups(limit: 50);
    if (!mounted) {
      return;
    }
    if (!response.isSuccess || response.data == null) {
      setState(() {
        _statusMessage =
            'Loading project setups failed: ${_compactError(response.errorMessage)}';
      });
      return;
    }
    setState(() {
      _projectSetups = response.data!;
      final selectedProjectID = _projectController.text.trim();
      final selected = _projectSetups
          .where((ProjectSetupConfig setup) {
            return setup.projectID == selectedProjectID;
          })
          .toList(growable: false);
      if (selected.isNotEmpty) {
        _applyProjectSetup(selected.first);
      } else if (_projectSetups.isNotEmpty) {
        _applyProjectSetup(_projectSetups.first);
      }
    });
  }

  Future<void> _saveProjectSetup() async {
    final endpoint = normalizeGraphqlEndpoint(_endpointController.text);
    if (endpoint.isEmpty) {
      setState(() => _statusMessage = 'Save endpoint settings first.');
      return;
    }
    final projectID = _projectController.text.trim();
    final projectName = _projectNameController.text.trim();
    final repositoryURL = _repositoryUrlController.text.trim();
    if (projectID.isEmpty || projectName.isEmpty || repositoryURL.isEmpty) {
      setState(
        () => _statusMessage =
            'Project ID, Project Name, and Repository URL are required.',
      );
      return;
    }
    setState(() => _isSavingProjectSetup = true);
    final api = ControlPlaneApi(buildGraphqlClient(endpoint));
    final response = await api.upsertProjectSetup(
      projectID: projectID,
      projectName: projectName,
      scmProvider: _setupScmProvider,
      repositoryURL: repositoryURL,
      trackerProvider: _setupTrackerProvider,
      trackerLocation: _trackerLocationController.text.trim(),
      trackerBoardID: _trackerBoardIDController.text.trim(),
    );
    if (!mounted) {
      return;
    }
    setState(() => _isSavingProjectSetup = false);
    if (!response.isSuccess || response.data == null) {
      setState(
        () => _statusMessage =
            'Saving project setup failed: ${_compactError(response.errorMessage)}',
      );
      return;
    }
    setState(() {
      _statusMessage = 'Saved project setup for ${response.data!.projectID}';
    });
    await _loadProjectSetups();
  }

  void _applyProjectSetup(ProjectSetupConfig setup) {
    _projectController.text = setup.projectID;
    _projectNameController.text = setup.projectName;
    _repositoryUrlController.text = setup.repositoryURL;
    _setupScmProvider = setup.scmProvider;
    _setupTrackerProvider = setup.trackerProvider;
    _trackerLocationController.text = setup.trackerLocation;
    _trackerBoardIDController.text = setup.trackerBoardID;
  }

  String _compactError(String? message) {
    final fallback = 'unknown error';
    final raw = (message ?? fallback).trim();
    if (raw.isEmpty) {
      return fallback;
    }
    final firstLine = raw.split('\n').first.trim();
    if (firstLine.length <= 180) {
      return firstLine;
    }
    return '${firstLine.substring(0, 177)}...';
  }

  Future<void> _runEnqueueIngestion(ControlPlaneApi api) async {
    if (_selectedSession == null) {
      setState(
        () => _statusMessage = 'Select a session before enqueueing ingestion.',
      );
      return;
    }
    final now = DateTime.now().millisecondsSinceEpoch;
    setState(() => _isRunningAction = true);
    final response = await api.enqueueIngestionWorkflow(
      runID: _selectedSession!.runID,
      taskID: _selectedJob?.taskID ?? 'task-ingestion',
      jobID: _selectedJob?.jobID ?? 'job-ingestion',
      idempotencyKey: 'ingest-$now',
      prompt: _promptController.text.trim(),
      projectID: _projectController.text.trim(),
      workflowID: _workflowController.text.trim(),
      source: _sourceController.text.trim(),
    );
    setState(() {
      _isRunningAction = false;
      _statusMessage = response.isSuccess
          ? 'Enqueued ingestion task ${response.data}'
          : 'Enqueue ingestion failed: ${response.errorMessage}';
      _refreshToken++;
    });
  }

  Future<void> _runApproveIssue(ControlPlaneApi api) async {
    if (_selectedSession == null || _selectedJob == null) {
      setState(
        () => _statusMessage =
            'Select a workflow job before approving issue intake.',
      );
      return;
    }
    setState(() => _isRunningAction = true);
    final response = await api.approveIssueIntake(
      runID: _selectedSession!.runID,
      taskID: _selectedJob!.taskID,
      jobID: _selectedJob!.jobID,
      source: _sourceController.text.trim(),
      issueReference: _issueReferenceController.text.trim(),
      approvedBy: _approvedByController.text.trim(),
    );
    setState(() {
      _isRunningAction = false;
      _statusMessage = response.isSuccess
          ? 'Issue approval decision: ${response.data}'
          : 'Approve issue failed: ${response.errorMessage}';
      _refreshToken++;
    });
  }

  Future<void> _runEnqueueScm(ControlPlaneApi api) async {
    if (_selectedSession == null) {
      setState(
        () => _statusMessage = 'Select a session before enqueueing SCM.',
      );
      return;
    }
    final now = DateTime.now().millisecondsSinceEpoch;
    setState(() => _isRunningAction = true);
    final response = await api.enqueueScmWorkflow(
      runID: _selectedSession!.runID,
      taskID: _selectedJob?.taskID ?? 'task-scm',
      jobID: _selectedJob?.jobID ?? 'job-scm',
      idempotencyKey: 'scm-$now',
      owner: _scmOwnerController.text.trim(),
      repository: _scmRepoController.text.trim(),
    );
    setState(() {
      _isRunningAction = false;
      _statusMessage = response.isSuccess
          ? 'Enqueued SCM task ${response.data}'
          : 'Enqueue SCM failed: ${response.errorMessage}';
      _refreshToken++;
    });
  }

  void _selectSession(ControlPlaneApi api, SessionSummary session) {
    setState(() {
      _selectedSession = session;
      _selectedJob = null;
      _streamEvents.clear();
      _statusMessage = 'Selected session ${session.runID}';
    });
    _streamSubscription?.cancel();
    _streamSubscription = api
        .sessionActivityStream(runID: session.runID)
        .listen((ApiResult<StreamEvent> eventResult) {
          if (!mounted) {
            return;
          }
          if (!eventResult.isSuccess || eventResult.data == null) {
            setState(
              () =>
                  _statusMessage = 'Stream error: ${eventResult.errorMessage}',
            );
            return;
          }
          setState(() {
            _streamEvents.insert(0, eventResult.data!);
            if (_streamEvents.length > 100) {
              _streamEvents.removeRange(100, _streamEvents.length);
            }
          });
        });
  }

  @override
  Widget build(BuildContext context) {
    final configState = ref.watch(appConfigProvider).valueOrNull;
    final endpoint = configState?.graphqlHttpEndpoint ?? widget.initialEndpoint;
    final api = ControlPlaneApi(buildGraphqlClient(endpoint));
    final isDashboard = _activeView == _DashboardView.dashboard;
    final isProjectSetup = _activeView == _DashboardView.projectSetup;
    final title = isDashboard
        ? 'Agentic Worktrees Desktop Control Plane'
        : isProjectSetup
        ? 'New Project Setup'
        : 'Settings';

    return Scaffold(
      drawer: Drawer(
        child: SafeArea(
          child: Column(
            children: <Widget>[
              const DrawerHeader(
                child: Align(
                  alignment: Alignment.bottomLeft,
                  child: Text(
                    'Agentic Worktrees',
                    style: TextStyle(fontSize: 22, fontWeight: FontWeight.w600),
                  ),
                ),
              ),
              Expanded(
                child: ListView(
                  children: <Widget>[
                    ListTile(
                      leading: const Icon(Icons.dashboard_outlined),
                      title: const Text('Dashboard'),
                      selected: isDashboard,
                      onTap: () => _showDashboard(context),
                    ),
                    ListTile(
                      leading: const Icon(Icons.add_box_outlined),
                      title: const Text('New Project Setup'),
                      selected: isProjectSetup,
                      onTap: () => _startNewProjectSetup(context),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              ListTile(
                leading: const Icon(Icons.settings_outlined),
                title: const Text('Settings'),
                selected: _activeView == _DashboardView.settings,
                onTap: () => _showSettings(context),
              ),
              ListTile(
                leading: const Icon(Icons.exit_to_app),
                title: const Text('Exit'),
                onTap: () => _exitApp(context),
              ),
            ],
          ),
        ),
      ),
      appBar: AppBar(
        title: Text(title),
        actions: <Widget>[
          if (isDashboard || isProjectSetup)
            IconButton(
              onPressed: () {
                setState(() => _refreshToken++);
                unawaited(_loadProjectSetups());
              },
              icon: const Icon(Icons.refresh),
              tooltip: 'Refresh queries',
            ),
        ],
      ),
      body: isDashboard
          ? _buildDashboardBody(api)
          : isProjectSetup
          ? _buildProjectSetupBody()
          : _buildSettingsBody(),
    );
  }

  Widget _buildDashboardBody(ControlPlaneApi api) {
    return Column(
      children: <Widget>[
        if (_statusMessage != null)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Align(
              alignment: Alignment.centerLeft,
              child: Text(
                _statusMessage!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ),
        Expanded(
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 330,
                child: Column(
                  children: <Widget>[
                    Card(
                      margin: const EdgeInsets.fromLTRB(12, 12, 12, 6),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            const ListTile(
                              dense: true,
                              title: Text('Configured Projects'),
                            ),
                            if (_projectSetups.isEmpty)
                              const Padding(
                                padding: EdgeInsets.fromLTRB(16, 0, 16, 12),
                                child: Text('No project setups configured.'),
                              )
                            else
                              ..._projectSetups.map((ProjectSetupConfig setup) {
                                final selected =
                                    _projectController.text.trim() ==
                                    setup.projectID;
                                return ListTile(
                                  dense: true,
                                  selected: selected,
                                  title: Text(setup.projectID),
                                  subtitle: Text(
                                    '${setup.projectName}\n${setup.repositoryURL}',
                                  ),
                                  onTap: () {
                                    setState(() {
                                      _applyProjectSetup(setup);
                                      _statusMessage =
                                          'Selected project ${setup.projectID}';
                                    });
                                  },
                                );
                              }),
                          ],
                        ),
                      ),
                    ),
                    Expanded(
                      child: Card(
                        margin: const EdgeInsets.fromLTRB(12, 6, 12, 12),
                        child: FutureBuilder<ApiResult<List<SessionSummary>>>(
                          future: api.sessions(limit: 50 + _refreshToken),
                          builder:
                              (
                                BuildContext context,
                                AsyncSnapshot<ApiResult<List<SessionSummary>>>
                                snapshot,
                              ) {
                                if (!snapshot.hasData) {
                                  return const Center(
                                    child: CircularProgressIndicator(),
                                  );
                                }
                                final value = snapshot.data!;
                                if (!value.isSuccess || value.data == null) {
                                  return Center(
                                    child: Text(
                                      value.errorMessage ??
                                          'Failed loading sessions',
                                    ),
                                  );
                                }
                                final sessions = value.data!;
                                if (sessions.isEmpty) {
                                  return const Center(
                                    child: Text('No sessions found.'),
                                  );
                                }
                                return ListView.builder(
                                  itemCount: sessions.length,
                                  itemBuilder: (BuildContext context, int index) {
                                    final item = sessions[index];
                                    final selected =
                                        _selectedSession?.runID == item.runID;
                                    return ListTile(
                                      selected: selected,
                                      title: Text(item.runID),
                                      subtitle: Text(
                                        'tasks: ${item.taskCount} jobs: ${item.jobCount}\nupdated: ${item.updatedAt}',
                                      ),
                                      onTap: () => _selectSession(api, item),
                                    );
                                  },
                                );
                              },
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: _selectedSession == null
                    ? const Center(
                        child: Text('Select a session to view details.'),
                      )
                    : _SessionDetailPanel(
                        api: api,
                        refreshToken: _refreshToken,
                        session: _selectedSession!,
                        selectedJob: _selectedJob,
                        streamEvents: _streamEvents,
                        sourceController: _sourceController,
                        issueReferenceController: _issueReferenceController,
                        approvedByController: _approvedByController,
                        projectController: _projectController,
                        workflowController: _workflowController,
                        promptController: _promptController,
                        scmOwnerController: _scmOwnerController,
                        scmRepoController: _scmRepoController,
                        isRunningAction: _isRunningAction,
                        onJobSelected: (WorkflowJob job) {
                          setState(() => _selectedJob = job);
                        },
                        onEnqueueIngestion: () => _runEnqueueIngestion(api),
                        onApproveIssue: () => _runApproveIssue(api),
                        onEnqueueScm: () => _runEnqueueScm(api),
                      ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildSettingsBody() {
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
                  controller: _endpointController,
                  decoration: const InputDecoration(
                    labelText: 'GraphQL HTTP Endpoint',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: _isSavingEndpoint ? null : _saveEndpoint,
                child: const Text('Save'),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: _isRunningAction ? null : _testConnection,
                child: const Text('Test'),
              ),
            ],
          ),
          if (_statusMessage != null) ...<Widget>[
            const SizedBox(height: 12),
            Text(_statusMessage!, maxLines: 2, overflow: TextOverflow.ellipsis),
          ],
        ],
      ),
    );
  }

  Widget _buildProjectSetupBody() {
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
            controller: _projectController,
            decoration: const InputDecoration(
              labelText: 'Project ID',
              border: OutlineInputBorder(),
            ),
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
          DropdownButtonFormField<String>(
            initialValue: _setupScmProvider,
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
              setState(() => _setupScmProvider = value);
            },
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _repositoryUrlController,
            decoration: const InputDecoration(
              labelText: 'Repository URL',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            initialValue: _setupTrackerProvider,
            decoration: const InputDecoration(
              labelText: 'Tracker Provider',
              border: OutlineInputBorder(),
            ),
            items: const <DropdownMenuItem<String>>[
              DropdownMenuItem<String>(
                value: 'GITHUB_ISSUES',
                child: Text('GitHub Issues'),
              ),
              DropdownMenuItem<String>(value: 'JIRA', child: Text('Jira')),
              DropdownMenuItem<String>(
                value: 'LOCAL_JSON',
                child: Text('Local JSON'),
              ),
              DropdownMenuItem<String>(value: 'LINEAR', child: Text('Linear')),
            ],
            onChanged: (String? value) {
              if (value == null) {
                return;
              }
              setState(() => _setupTrackerProvider = value);
            },
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _trackerLocationController,
            decoration: const InputDecoration(
              labelText: 'Tracker Location',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _trackerBoardIDController,
            decoration: const InputDecoration(
              labelText: 'Tracker Board ID (optional)',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              FilledButton(
                onPressed: _isSavingProjectSetup ? null : _saveProjectSetup,
                child: const Text('Save Project Setup'),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: _loadProjectSetups,
                child: const Text('Reload'),
              ),
            ],
          ),
          if (_projectSetups.isNotEmpty) ...<Widget>[
            const SizedBox(height: 12),
            const Text(
              'Configured Projects',
              style: TextStyle(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 8),
            ..._projectSetups.map((ProjectSetupConfig setup) {
              return Card(
                child: ListTile(
                  dense: true,
                  selected: _projectController.text.trim() == setup.projectID,
                  title: Text(setup.projectID),
                  subtitle: Text(
                    '${setup.projectName}\n${setup.repositoryURL}',
                  ),
                  onTap: () {
                    setState(() {
                      _applyProjectSetup(setup);
                      _statusMessage =
                          'Loaded project setup ${setup.projectID}';
                    });
                  },
                ),
              );
            }),
          ],
          if (_statusMessage != null) ...<Widget>[
            const SizedBox(height: 12),
            Text(_statusMessage!, maxLines: 2, overflow: TextOverflow.ellipsis),
          ],
        ],
      ),
    );
  }
}

class _SessionDetailPanel extends StatelessWidget {
  const _SessionDetailPanel({
    required this.api,
    required this.refreshToken,
    required this.session,
    required this.selectedJob,
    required this.streamEvents,
    required this.sourceController,
    required this.issueReferenceController,
    required this.approvedByController,
    required this.projectController,
    required this.workflowController,
    required this.promptController,
    required this.scmOwnerController,
    required this.scmRepoController,
    required this.isRunningAction,
    required this.onJobSelected,
    required this.onEnqueueIngestion,
    required this.onApproveIssue,
    required this.onEnqueueScm,
  });

  final ControlPlaneApi api;
  final int refreshToken;
  final SessionSummary session;
  final WorkflowJob? selectedJob;
  final List<StreamEvent> streamEvents;
  final TextEditingController sourceController;
  final TextEditingController issueReferenceController;
  final TextEditingController approvedByController;
  final TextEditingController projectController;
  final TextEditingController workflowController;
  final TextEditingController promptController;
  final TextEditingController scmOwnerController;
  final TextEditingController scmRepoController;
  final bool isRunningAction;
  final ValueChanged<WorkflowJob> onJobSelected;
  final VoidCallback onEnqueueIngestion;
  final VoidCallback onApproveIssue;
  final VoidCallback onEnqueueScm;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Session ${session.runID}',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          FutureBuilder<ApiResult<List<WorkerSummary>>>(
            future: api.workers(limit: 50 + refreshToken),
            builder:
                (
                  BuildContext context,
                  AsyncSnapshot<ApiResult<List<WorkerSummary>>> snapshot,
                ) {
                  if (!snapshot.hasData) {
                    return const LinearProgressIndicator();
                  }
                  final value = snapshot.data!;
                  if (!value.isSuccess || value.data == null) {
                    return Text('Workers error: ${value.errorMessage}');
                  }
                  final workers = value.data!;
                  return Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: workers
                        .map(
                          (WorkerSummary worker) => Chip(
                            label: Text(
                              '${worker.workerID} (${worker.capabilities.join(', ')})',
                            ),
                          ),
                        )
                        .toList(growable: false),
                  );
                },
          ),
          const SizedBox(height: 12),
          FutureBuilder<ApiResult<List<WorkflowJob>>>(
            future: api.workflowJobs(
              runID: session.runID,
              limit: 100 + refreshToken,
            ),
            builder:
                (
                  BuildContext context,
                  AsyncSnapshot<ApiResult<List<WorkflowJob>>> snapshot,
                ) {
                  if (!snapshot.hasData) {
                    return const LinearProgressIndicator();
                  }
                  final value = snapshot.data!;
                  if (!value.isSuccess || value.data == null) {
                    return Text('Workflow jobs error: ${value.errorMessage}');
                  }
                  final jobs = value.data!;
                  if (jobs.isEmpty) {
                    return const Text(
                      'No workflow jobs found for this session.',
                    );
                  }
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: jobs
                        .map(
                          (WorkflowJob job) => Card(
                            child: ListTile(
                              selected: selectedJob?.jobID == job.jobID,
                              title: Text(
                                '${job.taskID}/${job.jobID} • ${job.jobKind}',
                              ),
                              subtitle: Text('${job.status} • ${job.queue}'),
                              trailing: Text(job.updatedAt.toString()),
                              onTap: () => onJobSelected(job),
                            ),
                          ),
                        )
                        .toList(growable: false),
                  );
                },
          ),
          const SizedBox(height: 12),
          if (selectedJob != null)
            FutureBuilder<ApiResult<List<SupervisorDecision>>>(
              future: api.supervisorHistory(
                runID: session.runID,
                taskID: selectedJob!.taskID,
                jobID: selectedJob!.jobID,
              ),
              builder:
                  (
                    BuildContext context,
                    AsyncSnapshot<ApiResult<List<SupervisorDecision>>> snapshot,
                  ) {
                    if (!snapshot.hasData) {
                      return const LinearProgressIndicator();
                    }
                    final value = snapshot.data!;
                    if (!value.isSuccess || value.data == null) {
                      return Text(
                        'Supervisor history error: ${value.errorMessage}',
                      );
                    }
                    final history = value.data!;
                    return ExpansionTile(
                      title: Text('Supervisor History (${history.length})'),
                      children: history
                          .map(
                            (SupervisorDecision decision) => ListTile(
                              dense: true,
                              title: Text(
                                '${decision.signalType} → ${decision.action}',
                              ),
                              subtitle: Text(
                                '${decision.reason} • ${decision.occurredAt}',
                              ),
                            ),
                          )
                          .toList(growable: false),
                    );
                  },
            ),
          const SizedBox(height: 12),
          ExpansionTile(
            initiallyExpanded: true,
            title: Text('Realtime Session Events (${streamEvents.length})'),
            children: streamEvents
                .map(
                  (StreamEvent event) => ListTile(
                    dense: true,
                    title: Text('${event.eventType} • ${event.source}'),
                    subtitle: Text(
                      '${event.occurredAt}\n${prettyJson(event.payload)}',
                    ),
                  ),
                )
                .toList(growable: false),
          ),
          const SizedBox(height: 12),
          Text(
            'Control Actions',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              SizedBox(
                width: 260,
                child: TextField(
                  controller: sourceController,
                  decoration: const InputDecoration(
                    labelText: 'Source (owner/repo)',
                  ),
                ),
              ),
              SizedBox(
                width: 260,
                child: TextField(
                  controller: issueReferenceController,
                  decoration: const InputDecoration(
                    labelText: 'Issue reference',
                  ),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: approvedByController,
                  decoration: const InputDecoration(labelText: 'Approved by'),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: projectController,
                  decoration: const InputDecoration(labelText: 'Project ID'),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: workflowController,
                  decoration: const InputDecoration(labelText: 'Workflow ID'),
                ),
              ),
              SizedBox(
                width: 420,
                child: TextField(
                  controller: promptController,
                  decoration: const InputDecoration(
                    labelText: 'Ingestion prompt',
                  ),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: scmOwnerController,
                  decoration: const InputDecoration(labelText: 'SCM owner'),
                ),
              ),
              SizedBox(
                width: 220,
                child: TextField(
                  controller: scmRepoController,
                  decoration: const InputDecoration(
                    labelText: 'SCM repository',
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              FilledButton(
                onPressed: isRunningAction ? null : onEnqueueIngestion,
                child: const Text('Enqueue Ingestion'),
              ),
              FilledButton(
                onPressed: isRunningAction ? null : onApproveIssue,
                child: const Text('Approve Issue Intake'),
              ),
              FilledButton(
                onPressed: isRunningAction ? null : onEnqueueScm,
                child: const Text('Enqueue SCM Source State'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
