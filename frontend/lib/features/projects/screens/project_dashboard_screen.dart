import 'dart:io';
import 'dart:typed_data';

import 'package:agentic_repositories/features/projects/screens/project_setup_edit_screen.dart';
import 'package:agentic_repositories/features/workers/screens/worker_sessions_screen.dart';
import 'package:agentic_repositories/features/workers/screens/worker_settings_screen.dart';
import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:agentic_repositories/shared/graph/typed/client.dart';
import 'package:agentic_repositories/shared/logging/app_logger.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mime/mime.dart';

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
  List<ProjectDocument> _projectDocuments = const <ProjectDocument>[];
  bool _isUploadingFiles = false;
  bool _isCreatingTaskboard = false;
  String? _statusMessage;

  @override
  void initState() {
    super.initState();
    _api = ControlPlaneApi(buildGraphqlClient(widget.endpoint));
    _projectSetup = widget.projectSetup;
    _loadProjectDocuments();
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
    await _loadProjectDocuments();
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

  Future<void> _loadProjectDocuments() async {
    final result = await _api.projectDocuments(
      projectID: _projectSetup.projectID,
      limit: 100,
    );
    if (!mounted) {
      return;
    }
    if (!result.isSuccess || result.data == null) {
      setState(() {
        _statusMessage =
            'Failed loading project documents: ${result.errorMessage ?? 'unknown error'}';
      });
      return;
    }
    setState(() {
      _projectDocuments = result.data!;
    });
  }

  Future<void> _uploadFiles() async {
    AppLogger.instance.logger.i(
      'Project document upload requested',
      error: {'projectID': _projectSetup.projectID},
    );
    FilePickerResult? picked;
    try {
      picked = await FilePicker.platform.pickFiles(
        allowMultiple: true,
        withData: true,
      );
    } catch (error, stackTrace) {
      AppLogger.instance.logger.e(
        'File picker failed to open',
        error: error,
        stackTrace: stackTrace,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _statusMessage = 'File picker failed: $error';
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Could not open file picker: $error')),
      );
      return;
    }
    if (!mounted || picked == null || picked.files.isEmpty) {
      AppLogger.instance.logger.i('File picker closed without file selection');
      return;
    }

    setState(() {
      _isUploadingFiles = true;
      _statusMessage = null;
    });

    final failures = <String>[];
    var uploadedCount = 0;

    for (final file in picked.files) {
      Uint8List? bytes = file.bytes;
      if (bytes == null && file.path != null) {
        final path = file.path!;
        bytes = await File(path).readAsBytes();
      }
      if (bytes == null || bytes.isEmpty) {
        failures.add('${file.name}: file bytes unavailable');
        continue;
      }

      final contentType =
          lookupMimeType(file.name, headerBytes: bytes) ??
          'application/octet-stream';
      final request = await _api.requestProjectDocumentUpload(
        projectID: _projectSetup.projectID,
        fileName: file.name,
        contentType: contentType,
      );
      if (!request.isSuccess || request.data == null) {
        failures.add(
          '${file.name}: ${request.errorMessage ?? 'failed requesting upload URL'}',
        );
        continue;
      }

      final upload = await _api.uploadProjectDocumentBytes(
        uploadURL: request.data!.uploadURL,
        bytes: bytes,
        contentType: contentType,
      );
      if (!upload.isSuccess) {
        final cleanup = await _api.deleteProjectDocument(
          projectID: request.data!.projectID,
          documentID: request.data!.documentID,
        );
        if (!cleanup.isSuccess) {
          AppLogger.instance.logger.w(
            'Failed to cleanup project document after upload failure',
            error: {
              'projectID': request.data!.projectID,
              'documentID': request.data!.documentID,
              'cleanupError': cleanup.errorMessage,
            },
          );
        }
        failures.add('${file.name}: ${upload.errorMessage ?? 'upload failed'}');
        continue;
      }
      uploadedCount++;
    }

    await _loadProjectDocuments();
    if (!mounted) {
      return;
    }
    setState(() {
      _isUploadingFiles = false;
      if (failures.isEmpty) {
        _statusMessage = 'Uploaded $uploadedCount file(s).';
      } else {
        _statusMessage =
            'Uploaded $uploadedCount file(s); ${failures.length} failed: ${failures.first}';
      }
    });
  }

  Future<void> _createNewTaskboard() async {
    if (_projectDocuments.isEmpty) {
      setState(() {
        _statusMessage =
            'Upload at least one project document before creating a taskboard.';
      });
      return;
    }
    final selectedDocumentIDs = _projectDocuments
        .map((ProjectDocument document) => document.documentID)
        .toSet();
    final promptController = TextEditingController();
    final draft = await showDialog<_NewTaskboardDraft>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            final isAllSelected =
                selectedDocumentIDs.length == _projectDocuments.length;
            return AlertDialog(
              title: const Text('New Taskboard'),
              content: SizedBox(
                width: 520,
                child: SingleChildScrollView(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      CheckboxListTile(
                        value: isAllSelected,
                        onChanged: (bool? value) {
                          setDialogState(() {
                            if (value == true) {
                              selectedDocumentIDs
                                ..clear()
                                ..addAll(
                                  _projectDocuments.map(
                                    (ProjectDocument document) =>
                                        document.documentID,
                                  ),
                                );
                            } else {
                              selectedDocumentIDs.clear();
                            }
                          });
                        },
                        title: const Text('Select all project documents'),
                        contentPadding: EdgeInsets.zero,
                      ),
                      const SizedBox(height: 8),
                      ..._projectDocuments.map((ProjectDocument document) {
                        return CheckboxListTile(
                          value: selectedDocumentIDs.contains(
                            document.documentID,
                          ),
                          onChanged: (bool? value) {
                            setDialogState(() {
                              if (value == true) {
                                selectedDocumentIDs.add(document.documentID);
                              } else {
                                selectedDocumentIDs.remove(document.documentID);
                              }
                            });
                          },
                          title: Text(document.fileName),
                          subtitle: Text('Status: ${document.status}'),
                          contentPadding: EdgeInsets.zero,
                        );
                      }),
                      const SizedBox(height: 12),
                      TextField(
                        controller: promptController,
                        minLines: 3,
                        maxLines: 6,
                        decoration: const InputDecoration(
                          labelText: 'User prompt',
                          hintText:
                              'Describe what you want in the new taskboard.',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () {
                    final selected = selectedDocumentIDs.toList(
                      growable: false,
                    );
                    final prompt = promptController.text.trim();
                    if (selected.isEmpty || prompt.isEmpty) {
                      return;
                    }
                    Navigator.of(context).pop(
                      _NewTaskboardDraft(
                        selectedDocumentIDs: selected,
                        userPrompt: prompt,
                      ),
                    );
                  },
                  child: const Text('Create'),
                ),
              ],
            );
          },
        );
      },
    );
    promptController.dispose();

    if (!mounted || draft == null) {
      return;
    }

    setState(() {
      _isCreatingTaskboard = true;
      _statusMessage = null;
    });

    final result = await _api.runIngestionAgent(
      projectID: _projectSetup.projectID,
      selectedDocumentIDs: draft.selectedDocumentIDs,
      userPrompt: draft.userPrompt,
    );

    if (!mounted) {
      return;
    }

    setState(() {
      _isCreatingTaskboard = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed creating taskboard: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _statusMessage =
          'Taskboard run enqueued (run=${result.data!.runID}, task=${result.data!.taskID}).';
    });
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
            onPressed: _isUploadingFiles ? null : _uploadFiles,
            icon: const Icon(Icons.upload_file_outlined),
            tooltip: 'Upload Files',
          ),
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
                  onPressed: _isCreatingTaskboard ? null : _createNewTaskboard,
                  icon: const Icon(Icons.add_task_outlined),
                  label: const Text('New Taskboard'),
                ),
                const SizedBox(width: 8),
                FilledButton.icon(
                  onPressed: _openEditProjectSetup,
                  icon: const Icon(Icons.edit_outlined),
                  label: const Text('Edit Project'),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Expanded(
                  child: Card(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          const Text(
                            'Details',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 8),
                          Text('Project Name: ${_projectSetup.projectName}'),
                          const SizedBox(height: 4),
                          Text('Project ID: ${_projectSetup.projectID}'),
                          Text(
                            'Provider: ${scm?.scmProvider ?? 'Not configured'}',
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Tracker: ${hasTracker ? 'Configured' : 'Not configured yet'}',
                          ),
                          const SizedBox(height: 4),
                          Text('Repositories: ${repositories.length}'),
                        ],
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Card(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          const Text(
                            'Project Documentation',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 8),
                          if (_isUploadingFiles)
                            const Padding(
                              padding: EdgeInsets.only(bottom: 8),
                              child: LinearProgressIndicator(),
                            ),
                          if (_projectDocuments.isEmpty)
                            const Text('No documentation files uploaded yet.')
                          else
                            ..._projectDocuments.map((
                              ProjectDocument document,
                            ) {
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 8),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: <Widget>[
                                    const Icon(
                                      Icons.description_outlined,
                                      size: 18,
                                    ),
                                    const SizedBox(width: 8),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: <Widget>[
                                          Text(
                                            document.fileName,
                                            style: const TextStyle(
                                              fontWeight: FontWeight.w500,
                                            ),
                                          ),
                                          Text(
                                            'Status: ${document.status}',
                                            style: Theme.of(
                                              context,
                                            ).textTheme.bodySmall,
                                          ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            }),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
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

class _NewTaskboardDraft {
  const _NewTaskboardDraft({
    required this.selectedDocumentIDs,
    required this.userPrompt,
  });

  final List<String> selectedDocumentIDs;
  final String userPrompt;
}
