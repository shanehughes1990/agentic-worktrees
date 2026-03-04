import 'package:agentic_repositories/features/projects/screens/project_setup_edit_screen.dart';
import 'package:agentic_repositories/features/workers/screens/worker_sessions_screen.dart';
import 'package:agentic_repositories/features/workers/screens/worker_settings_screen.dart';
import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:agentic_repositories/shared/graph/typed/client.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

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
  late ProjectSetupConfig _projectSetup;
  String? _statusMessage;

  @override
  void initState() {
    super.initState();
    _api = ControlPlaneApi(buildGraphqlClient(widget.endpoint));
    _projectSetup = widget.projectSetup;
  }

  Future<void> _openEditProjectSetup() async {
    final updated = await Navigator.of(context).push<ProjectSetupConfig>(
      MaterialPageRoute<ProjectSetupConfig>(
        builder: (BuildContext context) => ProjectSetupEditScreen(
          projectSetup: _projectSetup,
          endpoint: widget.endpoint,
        ),
      ),
    );

    if (!mounted || updated == null) {
      return;
    }

    setState(() {
      _projectSetup = updated;
      _statusMessage = 'Project setup updated.';
    });
  }

  Future<void> _openWorkerSessions() async {
    await Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (BuildContext context) => Scaffold(
          appBar: AppBar(title: const Text('Worker Sessions')),
          body: WorkerSessionsScreen(
            api: _api,
            statusMessage: _statusMessage,
            onStatus: (String message) {
              if (!mounted) {
                return;
              }
              setState(() => _statusMessage = message);
            },
          ),
        ),
      ),
    );
  }

  Future<void> _openWorkerSettings() async {
    await Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (BuildContext context) => Scaffold(
          appBar: AppBar(title: const Text('Worker Settings')),
          body: WorkerSettingsScreen(
            api: _api,
            statusMessage: _statusMessage,
            onStatus: (String message) {
              if (!mounted) {
                return;
              }
              setState(() => _statusMessage = message);
            },
          ),
        ),
      ),
    );
  }

  void _goToDashboardHome() {
    Navigator.of(context).popUntil((Route<dynamic> route) => route.isFirst);
  }

  Future<void> _copyProjectID() async {
    await Clipboard.setData(ClipboardData(text: _projectSetup.projectID));
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(const SnackBar(content: Text('Project ID copied')));
  }

  @override
  Widget build(BuildContext context) {
    final repositories = _projectSetup.repositories;
    final scm = _projectSetup.scms.isNotEmpty ? _projectSetup.scms.first : null;
    final hasTracker = _projectSetup.boards.isNotEmpty;

    return Scaffold(
      drawer: Drawer(
        child: SafeArea(
          child: ListView(
            padding: EdgeInsets.zero,
            children: <Widget>[
              DrawerHeader(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    Text(
                      _projectSetup.projectName,
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    const SizedBox(height: 4),
                    Text(_projectSetup.projectID),
                  ],
                ),
              ),
              ListTile(
                leading: const Icon(Icons.dashboard_outlined),
                title: const Text('Dashboard Home'),
                onTap: _goToDashboardHome,
              ),
              ListTile(
                leading: const Icon(Icons.memory_outlined),
                title: const Text('Worker Sessions'),
                onTap: () {
                  Navigator.of(context).pop();
                  _openWorkerSessions();
                },
              ),
              ListTile(
                leading: const Icon(Icons.tune),
                title: const Text('Worker Settings'),
                onTap: () {
                  Navigator.of(context).pop();
                  _openWorkerSettings();
                },
              ),
            ],
          ),
        ),
      ),
      appBar: AppBar(
        automaticallyImplyLeading: true,
        title: Text(_projectSetup.projectName),
        actions: <Widget>[
          IconButton(
            onPressed: _copyProjectID,
            icon: const Icon(Icons.copy_outlined),
            tooltip: 'Copy Project ID',
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Expanded(
                  child: Text(
                    'Project Dashboard',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
                  ),
                ),
                FilledButton.icon(
                  onPressed: _openEditProjectSetup,
                  icon: const Icon(Icons.edit_outlined),
                  label: const Text('Edit Project'),
                ),
              ],
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
                    const SizedBox(height: 8),
                    Text('Project Name: ${_projectSetup.projectName}'),
                    const SizedBox(height: 4),
                    Text('Project ID: ${_projectSetup.projectID}'),
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
                    const SizedBox(height: 8),
                    Text('Provider: ${scm?.scmProvider ?? 'Not configured'}'),
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
                      'Repository Setup',
                      style: TextStyle(fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 8),
                    if (repositories.isEmpty)
                      const Text('No repositories configured.')
                    else
                      for (final repository in repositories) ...<Widget>[
                        Text(repository.repositoryURL),
                        const SizedBox(height: 6),
                      ],
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
                      'Tracker Setup',
                      style: TextStyle(fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 8),
                    Text(hasTracker ? 'Configured' : 'Not configured yet'),
                  ],
                ),
              ),
            ),
            if (_statusMessage != null) ...<Widget>[
              const SizedBox(height: 12),
              Text(_statusMessage!),
            ],
          ],
        ),
      ),
    );
  }
}
