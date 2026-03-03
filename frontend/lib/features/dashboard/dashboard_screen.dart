import 'dart:async';
import 'dart:io';

import 'package:agentic_worktrees/features/dashboard/logic/dashboard_workflow_logic.dart';
import 'package:agentic_worktrees/features/dashboard/widgets/dashboard_home_view.dart';
import 'package:agentic_worktrees/features/projects/logic/project_setup_logic.dart';
import 'package:agentic_worktrees/features/projects/screens/project_dashboard_screen.dart';
import 'package:agentic_worktrees/features/projects/screens/project_setup_screen.dart';
import 'package:agentic_worktrees/features/settings/logic/connection_settings_logic.dart';
import 'package:agentic_worktrees/features/settings/screens/settings_screen.dart';
import 'package:agentic_worktrees/features/workers/screens/worker_sessions_screen.dart';
import 'package:agentic_worktrees/features/workers/screens/worker_settings_screen.dart';
import 'package:agentic_worktrees/shared/config/app_config.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

enum _DashboardView { dashboard, workerSessions, workerSettings, settings }

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
  final TextEditingController _scmTokenController = TextEditingController();
  final TextEditingController _workflowController = TextEditingController(
    text: 'workflow-1',
  );
  final TextEditingController _promptController = TextEditingController(
    text: 'Ingest latest issue board state',
  );

  SessionSummary? _selectedSession;
  WorkflowJob? _selectedJob;
  String? _statusMessage;
  bool _isSavingEndpoint = false;
  bool _isSavingProjectSetup = false;
  bool _isRunningAction = false;
  int _refreshToken = 0;
  _DashboardView _activeView = _DashboardView.dashboard;
  String _setupScmProvider = ProjectSetupLogic.defaultScmProvider;
  List<ProjectSetupConfig> _projectSetups = const <ProjectSetupConfig>[];
  final List<StreamEvent> _streamEvents = <StreamEvent>[];
  StreamSubscription<ApiResult<StreamEvent>>? _streamSubscription;
  StreamSubscription<ApiResult<StreamEvent>>? _workerSessionSubscription;
  String _workerSessionSubscriptionEndpoint = '';
  String _apiEndpoint = '';
  ControlPlaneApi? _api;

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

  void _showWorkerSessions(BuildContext context) {
    Navigator.of(context).pop();
    setState(() => _activeView = _DashboardView.workerSessions);
  }

  void _showWorkerSettings(BuildContext context) {
    Navigator.of(context).pop();
    setState(() => _activeView = _DashboardView.workerSettings);
  }

  void _prepareNewProjectSetup() {
    setState(() {
      _projectController.text = '';
      _projectNameController.text = '';
      _repositoryUrlController.text = '';
      _scmTokenController.clear();
      _setupScmProvider = ProjectSetupLogic.defaultScmProvider;
      _statusMessage = null;
    });
  }

  void _openProjectSetup() {
    _prepareNewProjectSetup();
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (BuildContext context) {
          return Scaffold(
            appBar: AppBar(title: const Text('New Project Setup')),
            body: _buildProjectSetupBody(),
          );
        },
      ),
    );
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
    _scmTokenController.dispose();
    _workflowController.dispose();
    _promptController.dispose();
    _streamSubscription?.cancel();
    _workerSessionSubscription?.cancel();
    super.dispose();
  }

  void _ensureWorkerSessionSubscription(String endpoint) {
    final trimmedEndpoint = endpoint.trim();
    if (trimmedEndpoint.isEmpty ||
        trimmedEndpoint == _workerSessionSubscriptionEndpoint) {
      return;
    }
    _workerSessionSubscription?.cancel();
    _workerSessionSubscription =
        ControlPlaneApi(
          buildGraphqlClient(trimmedEndpoint),
        ).workerSessionStream().listen((ApiResult<StreamEvent> eventResult) {
          if (!mounted || !eventResult.isSuccess) {
            return;
          }
          setState(() => _refreshToken++);
        });
    _workerSessionSubscriptionEndpoint = trimmedEndpoint;
  }

  ControlPlaneApi _apiFor(String endpoint) {
    final trimmedEndpoint = endpoint.trim();
    if (_api != null && trimmedEndpoint == _apiEndpoint) {
      return _api!;
    }
    _apiEndpoint = trimmedEndpoint;
    _api = ControlPlaneApi(buildGraphqlClient(trimmedEndpoint));
    return _api!;
  }

  Future<void> _saveEndpoint() async {
    final endpoint = ConnectionSettingsLogic.normalizeEndpoint(
      _endpointController.text,
    );
    final validationError = ConnectionSettingsLogic.validateEndpoint(endpoint);
    if (validationError != null) {
      setState(() => _statusMessage = validationError);
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
    final endpoint = ConnectionSettingsLogic.normalizeEndpoint(
      _endpointController.text,
    );
    final validationError = ConnectionSettingsLogic.validateEndpoint(endpoint);
    if (validationError != null) {
      setState(() => _statusMessage = validationError);
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

    final nextStatus = result.isSuccess
        ? ConnectionSettingsLogic.successMessage(result.data?.length ?? 0)
        : ConnectionSettingsLogic.failureMessage(
            endpoint: endpoint,
            compactError: DashboardWorkflowLogic.compactError(
              result.errorMessage,
            ),
          );

    setState(() {
      _endpointController.text = endpoint;
      _isRunningAction = false;
      _statusMessage = nextStatus;
      _refreshToken++;
    });

    if (result.isSuccess) {
      await _loadProjectSetups();
    }
  }

  Future<void> _loadProjectSetups() async {
    final endpoint = ConnectionSettingsLogic.normalizeEndpoint(
      _endpointController.text,
    );
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
            'Loading project setups failed: ${DashboardWorkflowLogic.compactError(response.errorMessage)}';
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
    final endpoint = ConnectionSettingsLogic.normalizeEndpoint(
      _endpointController.text,
    );
    if (endpoint.isEmpty) {
      setState(() => _statusMessage = 'Save endpoint settings first.');
      return;
    }
    final projectID = _projectController.text.trim();
    final projectName = _projectNameController.text.trim();
    final repositoryURLs = ProjectSetupLogic.parseMultilineEntries(
      _repositoryUrlController.text,
    );
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
    setState(() => _isSavingProjectSetup = true);
    final api = ControlPlaneApi(buildGraphqlClient(endpoint));
    final response = await api.upsertProjectSetup(
      projectID: projectID,
      projectName: projectName,
      scmProvider: _setupScmProvider,
      repositoryURLs: repositoryURLs,
      scmToken: _scmTokenController.text.trim(),
    );
    if (!mounted) {
      return;
    }
    setState(() => _isSavingProjectSetup = false);
    if (!response.isSuccess || response.data == null) {
      setState(
        () => _statusMessage =
            'Saving project setup failed: ${DashboardWorkflowLogic.compactError(response.errorMessage)}',
      );
      return;
    }
    final savedSetup = response.data!;
    setState(() {
      _statusMessage = 'Saved project setup for ${savedSetup.projectID}';
      _applyProjectSetup(savedSetup);
    });
    await _loadProjectSetups();

    if (!mounted) {
      return;
    }
    Navigator.of(context).pop();
    _openProjectDashboard(savedSetup, endpoint);
  }

  void _applyProjectSetup(ProjectSetupConfig setup) {
    ProjectSetupLogic.applySetupToForm(
      setup: setup,
      projectController: _projectController,
      projectNameController: _projectNameController,
      repositoryUrlController: _repositoryUrlController,
      onScmProviderChanged: (String provider) {
        _setupScmProvider = provider;
      },
    );
  }

  void _openProjectDashboard(ProjectSetupConfig setup, String endpoint) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (BuildContext context) =>
            ProjectDashboardScreen(projectSetup: setup, endpoint: endpoint),
      ),
    );
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
      projectID: _projectController.text.trim(),
      source: _sourceController.text.trim(),
      issueReference: _issueReferenceController.text.trim(),
      approvedBy: _approvedByController.text.trim(),
    );
    setState(() {
      _isRunningAction = false;
      _statusMessage = response.isSuccess
          ? 'Issue approval decision: ${response.data}'
          : 'Approve issue failed: ${DashboardWorkflowLogic.compactError(response.errorMessage)}';
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
    _streamSubscription = api.sessionActivityStream(runID: session.runID).listen((
      ApiResult<StreamEvent> eventResult,
    ) {
      if (!mounted) {
        return;
      }
      if (!eventResult.isSuccess || eventResult.data == null) {
        setState(
          () => _statusMessage =
              'Stream error: ${DashboardWorkflowLogic.compactError(eventResult.errorMessage)}',
        );
        return;
      }
      setState(() {
        DashboardWorkflowLogic.appendStreamEvent(
          _streamEvents,
          eventResult.data!,
        );
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    final configState = ref.watch(appConfigProvider).valueOrNull;
    final endpoint = configState?.graphqlHttpEndpoint ?? widget.initialEndpoint;
    _ensureWorkerSessionSubscription(endpoint);
    final api = _apiFor(endpoint);
    final isDashboard = _activeView == _DashboardView.dashboard;
    final isWorkerSessions = _activeView == _DashboardView.workerSessions;
    final isWorkerSettings = _activeView == _DashboardView.workerSettings;
    final title = isDashboard
        ? 'Dashboard'
        : isWorkerSessions
        ? 'Worker Sessions'
        : isWorkerSettings
        ? 'Worker Settings'
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
                      leading: const Icon(Icons.memory_outlined),
                      title: const Text('Worker Sessions'),
                      selected: isWorkerSessions,
                      onTap: () => _showWorkerSessions(context),
                    ),
                    ListTile(
                      leading: const Icon(Icons.tune_outlined),
                      title: const Text('Worker Settings'),
                      selected: isWorkerSettings,
                      onTap: () => _showWorkerSettings(context),
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
          if (isDashboard)
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
          ? _buildDashboardBody(api, endpoint)
          : isWorkerSessions
          ? _buildWorkerSessionsBody(api)
          : isWorkerSettings
          ? _buildWorkerSettingsBody(api)
          : _buildSettingsBody(),
    );
  }

  Widget _buildDashboardBody(ControlPlaneApi api, String endpoint) {
    return DashboardHomeView(
      api: api,
      refreshToken: _refreshToken,
      statusMessage: _statusMessage,
      projectSetups: _projectSetups,
      selectedProjectID: _projectController.text.trim(),
      onProjectSelected: (ProjectSetupConfig setup) {
        setState(() {
          _applyProjectSetup(setup);
          _statusMessage = 'Selected project ${setup.projectID}';
        });
        _openProjectDashboard(setup, endpoint);
      },
      selectedSession: _selectedSession,
      onSessionSelected: (SessionSummary session) =>
          _selectSession(api, session),
      selectedJob: _selectedJob,
      streamEvents: _streamEvents,
      sourceController: _sourceController,
      issueReferenceController: _issueReferenceController,
      approvedByController: _approvedByController,
      projectController: _projectController,
      workflowController: _workflowController,
      promptController: _promptController,
      isRunningAction: _isRunningAction,
      onJobSelected: (WorkflowJob job) {
        setState(() => _selectedJob = job);
      },
      onApproveIssue: () => _runApproveIssue(api),
      onShowWorkerSessions: () {
        setState(() => _activeView = _DashboardView.workerSessions);
      },
      onCreateProject: _openProjectSetup,
    );
  }

  Widget _buildSettingsBody() {
    return SettingsScreen(
      endpointController: _endpointController,
      isSavingEndpoint: _isSavingEndpoint,
      isRunningAction: _isRunningAction,
      onSaveEndpoint: _saveEndpoint,
      onTestConnection: _testConnection,
      statusMessage: _statusMessage,
    );
  }

  Widget _buildProjectSetupBody() {
    return ProjectSetupScreen(
      projectController: _projectController,
      projectNameController: _projectNameController,
      repositoryUrlController: _repositoryUrlController,
      setupScmProvider: _setupScmProvider,
      scmTokenController: _scmTokenController,
      onSetupScmProviderChanged: (String value) {
        setState(() => _setupScmProvider = value);
      },
      isSavingProjectSetup: _isSavingProjectSetup,
      onSaveProjectSetup: _saveProjectSetup,
      statusMessage: _statusMessage,
    );
  }

  Widget _buildWorkerSessionsBody(ControlPlaneApi api) {
    return WorkerSessionsScreen(
      api: api,
      statusMessage: _statusMessage,
      onStatus: (String message) {
        if (!mounted) {
          return;
        }
        setState(() => _statusMessage = message);
      },
    );
  }

  Widget _buildWorkerSettingsBody(ControlPlaneApi api) {
    return WorkerSettingsScreen(
      api: api,
      statusMessage: _statusMessage,
      onStatus: (String message) {
        if (!mounted) {
          return;
        }
        setState(() => _statusMessage = message);
      },
    );
  }
}
