import 'dart:async';

import 'package:agentic_repositories/features/projects/widgets/taskboard_ingestion_dialog.dart';
import 'package:agentic_repositories/shared/graph/typed/api.dart';
import 'package:agentic_repositories/shared/graph/typed/models.dart';
import 'package:flutter/material.dart';

class TaskboardManagementScreen extends StatefulWidget {
  const TaskboardManagementScreen({
    required this.api,
    required this.projectID,
    required this.boardID,
    super.key,
  });

  final ControlPlaneApi api;
  final String projectID;
  final String boardID;

  @override
  State<TaskboardManagementScreen> createState() =>
      _TaskboardManagementScreenState();
}

class _TaskboardManagementScreenState extends State<TaskboardManagementScreen> {
  TaskboardModel? _board;
  final Set<String> _expandedEpicIDs = <String>{};
  bool _isLoading = true;
  bool _isMutating = false;
  String? _statusMessage;

  static const String _readOnlyMessage =
      'This taskboard is ended and read-only.';

  static const List<String> _boardStates = <String>[
    'PENDING',
    'ACTIVE',
    'COMPLETED',
    'CANCELLED',
  ];
  static const List<String> _epicStates = <String>[
    'PENDING',
    'ACTIVE',
    'COMPLETED',
    'BLOCKED',
    'CANCELLED',
  ];
  static const List<String> _taskStates = <String>[
    'TODO',
    'IN_PROGRESS',
    'BLOCKED',
    'DONE',
    'CANCELLED',
  ];

  @override
  void initState() {
    super.initState();
    unawaited(_loadBoard());
  }

  Future<void> _loadBoard() async {
    setState(() {
      _isLoading = true;
    });
    final result = await widget.api.taskboard(
      projectID: widget.projectID,
      boardID: widget.boardID,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isLoading = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed loading taskboard: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
    });
  }

  Future<void> _runTaskboardAgain() async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (!_canRunAgain(board)) {
      setState(() {
        _statusMessage =
            'Run Again is only available for taskboards that have not started yet.';
      });
      return;
    }

    final documentsResult = await widget.api.projectDocuments(
      projectID: widget.projectID,
      limit: 100,
    );
    if (!mounted) {
      return;
    }
    if (!documentsResult.isSuccess || documentsResult.data == null) {
      setState(() {
        _statusMessage =
            'Failed loading project documents: ${documentsResult.errorMessage ?? 'unknown error'}';
      });
      return;
    }

    final branchOptionsResult = await widget.api.projectRepositoryBranches(
      projectID: widget.projectID,
    );
    if (!mounted) {
      return;
    }
    if (!branchOptionsResult.isSuccess || branchOptionsResult.data == null) {
      setState(() {
        _statusMessage =
            'Failed loading repository branches: ${branchOptionsResult.errorMessage ?? 'unknown error'}';
      });
      return;
    }

    final preselectedIDs = _matchDocumentIDsFromLocations(
      documentsResult.data!,
      board.ingestionFilesAdded,
    );
    final draft = await showTaskboardIngestionDialog(
      context: context,
      api: widget.api,
      projectID: widget.projectID,
      projectDocuments: documentsResult.data!,
      repositoryBranchOptions: branchOptionsResult.data!,
      title: 'Run Taskboard Again',
      submitLabel: 'Run Again',
      initialTaskboardName: board.name,
      initialUserPrompt: board.ingestionUserPrompt,
      initialSelectedDocumentIDs: preselectedIDs,
    );

    if (!mounted || draft == null) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.runIngestionAgent(
      projectID: widget.projectID,
      taskboardName: draft.taskboardName,
      boardID: board.boardID,
      selectedDocumentIDs: draft.selectedDocumentIDs,
      userPrompt: draft.userPrompt,
      repositorySourceBranches: draft.repositorySourceBranches,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed to run taskboard again: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _statusMessage =
          'Taskboard rerun enqueued (run=${result.data!.runID}, task=${result.data!.taskID}).';
    });
  }

  bool _canRunAgain(TaskboardModel board) {
    return board.state.trim().toLowerCase() == 'pending';
  }

  Set<String> _matchDocumentIDsFromLocations(
    List<ProjectDocument> documents,
    List<String> locations,
  ) {
    final normalizedLocations = locations
        .map((String value) => value.trim().toLowerCase())
        .where((String value) => value.isNotEmpty)
        .toSet();
    final selected = <String>{};
    for (final document in documents) {
      final candidates = <String>{
        document.documentID.trim().toLowerCase(),
        document.fileName.trim().toLowerCase(),
        document.objectPath.trim().toLowerCase(),
        document.cdnURL.trim().toLowerCase(),
      };
      if (candidates.any(normalizedLocations.contains)) {
        selected.add(document.documentID);
      }
    }
    return selected;
  }

  Future<void> _editBoard() async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final nameController = TextEditingController(text: board.name);
    String selectedState = _boardStates.contains(board.state)
        ? board.state
        : _boardStates.first;
    final updated = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            return AlertDialog(
              title: const Text('Edit Taskboard'),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  TextField(
                    controller: nameController,
                    decoration: const InputDecoration(labelText: 'Name'),
                  ),
                  const SizedBox(height: 12),
                  DropdownButtonFormField<String>(
                    initialValue: selectedState,
                    items: _boardStates
                        .map(
                          (String state) => DropdownMenuItem<String>(
                            value: state,
                            child: Text(state),
                          ),
                        )
                        .toList(growable: false),
                    onChanged: (String? value) {
                      if (value == null) {
                        return;
                      }
                      setDialogState(() => selectedState = value);
                    },
                    decoration: const InputDecoration(labelText: 'State'),
                  ),
                ],
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Save'),
                ),
              ],
            );
          },
        );
      },
    );
    if (updated != true) {
      return;
    }
    setState(() => _isMutating = true);
    final result = await widget.api.updateTaskboard(
      projectID: widget.projectID,
      boardID: board.boardID,
      name: nameController.text.trim(),
      state: selectedState,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed updating taskboard: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Taskboard updated.';
    });
  }

  Future<void> _deleteBoard() async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: const Text('Delete Taskboard'),
          content: Text('Delete "${board.name}" and all epics/tasks?'),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Delete'),
            ),
          ],
        );
      },
    );
    if (confirmed != true) {
      return;
    }
    setState(() => _isMutating = true);
    final result = await widget.api.deleteTaskboard(
      projectID: widget.projectID,
      boardID: board.boardID,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess) {
        _statusMessage =
            'Failed deleting taskboard: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
    });
    if (!mounted) {
      return;
    }
    Navigator.of(context).pop(true);
  }

  Future<void> _createEpic() async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final titleController = TextEditingController();
    final objectiveController = TextEditingController();
    final rankController = TextEditingController(text: '0');
    String selectedState = _epicStates.first;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            return AlertDialog(
              title: const Text('Create Epic'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    TextField(
                      controller: titleController,
                      decoration: const InputDecoration(labelText: 'Title'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: objectiveController,
                      decoration: const InputDecoration(labelText: 'Objective'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: rankController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(labelText: 'Rank'),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<String>(
                      initialValue: selectedState,
                      items: _epicStates
                          .map(
                            (String state) => DropdownMenuItem<String>(
                              value: state,
                              child: Text(state),
                            ),
                          )
                          .toList(growable: false),
                      onChanged: (String? value) {
                        if (value == null) {
                          return;
                        }
                        setDialogState(() => selectedState = value);
                      },
                      decoration: const InputDecoration(labelText: 'State'),
                    ),
                  ],
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Create'),
                ),
              ],
            );
          },
        );
      },
    );
    if (confirmed != true) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.createTaskboardEpic(
      projectID: widget.projectID,
      boardID: board.boardID,
      title: titleController.text.trim(),
      objective: objectiveController.text.trim().isEmpty
          ? null
          : objectiveController.text.trim(),
      state: selectedState,
      rank: int.tryParse(rankController.text.trim()) ?? 0,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed creating epic: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Epic created.';
    });
  }

  Future<void> _editEpic(TaskboardEpicModel epic) async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final titleController = TextEditingController(text: epic.title);
    final objectiveController = TextEditingController(
      text: epic.objective ?? '',
    );
    final rankController = TextEditingController(text: '${epic.rank}');
    String selectedState = _epicStates.contains(epic.state)
        ? epic.state
        : _epicStates.first;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            return AlertDialog(
              title: const Text('Edit Epic'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    TextField(
                      controller: titleController,
                      decoration: const InputDecoration(labelText: 'Title'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: objectiveController,
                      decoration: const InputDecoration(labelText: 'Objective'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: rankController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(labelText: 'Rank'),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<String>(
                      initialValue: selectedState,
                      items: _epicStates
                          .map(
                            (String state) => DropdownMenuItem<String>(
                              value: state,
                              child: Text(state),
                            ),
                          )
                          .toList(growable: false),
                      onChanged: (String? value) {
                        if (value == null) {
                          return;
                        }
                        setDialogState(() => selectedState = value);
                      },
                      decoration: const InputDecoration(labelText: 'State'),
                    ),
                  ],
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Save'),
                ),
              ],
            );
          },
        );
      },
    );
    if (confirmed != true) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.updateTaskboardEpic(
      projectID: widget.projectID,
      boardID: board.boardID,
      epicID: epic.id,
      title: titleController.text.trim(),
      objective: objectiveController.text.trim().isEmpty
          ? null
          : objectiveController.text.trim(),
      state: selectedState,
      rank: int.tryParse(rankController.text.trim()) ?? 0,
      dependsOnEpicIDs: epic.dependsOnEpicIDs,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed updating epic: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Epic updated.';
    });
  }

  Future<void> _deleteEpic(TaskboardEpicModel epic) async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: const Text('Delete Epic'),
          content: Text('Delete epic "${epic.title}" and its tasks?'),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Delete'),
            ),
          ],
        );
      },
    );
    if (confirmed != true) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.deleteTaskboardEpic(
      projectID: widget.projectID,
      boardID: board.boardID,
      epicID: epic.id,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed deleting epic: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Epic deleted.';
    });
  }

  Future<void> _createTask(TaskboardEpicModel epic) async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final titleController = TextEditingController();
    final descriptionController = TextEditingController();
    final taskTypeController = TextEditingController(text: 'feature');
    final rankController = TextEditingController(text: '0');
    String selectedState = _taskStates.first;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            return AlertDialog(
              title: const Text('Create Task'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    TextField(
                      controller: titleController,
                      decoration: const InputDecoration(labelText: 'Title'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: descriptionController,
                      decoration: const InputDecoration(
                        labelText: 'Description',
                      ),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: taskTypeController,
                      decoration: const InputDecoration(labelText: 'Task Type'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: rankController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(labelText: 'Rank'),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<String>(
                      initialValue: selectedState,
                      items: _taskStates
                          .map(
                            (String state) => DropdownMenuItem<String>(
                              value: state,
                              child: Text(state),
                            ),
                          )
                          .toList(growable: false),
                      onChanged: (String? value) {
                        if (value == null) {
                          return;
                        }
                        setDialogState(() => selectedState = value);
                      },
                      decoration: const InputDecoration(labelText: 'State'),
                    ),
                  ],
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Create'),
                ),
              ],
            );
          },
        );
      },
    );
    if (confirmed != true) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.createTaskboardTask(
      projectID: widget.projectID,
      boardID: board.boardID,
      epicID: epic.id,
      title: titleController.text.trim(),
      description: descriptionController.text.trim().isEmpty
          ? null
          : descriptionController.text.trim(),
      taskType: taskTypeController.text.trim(),
      state: selectedState,
      rank: int.tryParse(rankController.text.trim()) ?? 0,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed creating task: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Task created.';
    });
  }

  Future<void> _editTask(
    TaskboardEpicModel epic,
    TaskboardTaskModel task,
  ) async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final titleController = TextEditingController(text: task.title);
    final descriptionController = TextEditingController(
      text: task.description ?? '',
    );
    final taskTypeController = TextEditingController(text: task.taskType);
    final rankController = TextEditingController(text: '${task.rank}');
    String selectedState = _taskStates.contains(task.state)
        ? task.state
        : _taskStates.first;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setDialogState) {
            return AlertDialog(
              title: const Text('Edit Task'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    TextField(
                      controller: titleController,
                      decoration: const InputDecoration(labelText: 'Title'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: descriptionController,
                      decoration: const InputDecoration(
                        labelText: 'Description',
                      ),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: taskTypeController,
                      decoration: const InputDecoration(labelText: 'Task Type'),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: rankController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(labelText: 'Rank'),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<String>(
                      initialValue: selectedState,
                      items: _taskStates
                          .map(
                            (String state) => DropdownMenuItem<String>(
                              value: state,
                              child: Text(state),
                            ),
                          )
                          .toList(growable: false),
                      onChanged: (String? value) {
                        if (value == null) {
                          return;
                        }
                        setDialogState(() => selectedState = value);
                      },
                      decoration: const InputDecoration(labelText: 'State'),
                    ),
                  ],
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  child: const Text('Save'),
                ),
              ],
            );
          },
        );
      },
    );
    if (confirmed != true) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.updateTaskboardTask(
      projectID: widget.projectID,
      boardID: board.boardID,
      epicID: epic.id,
      taskID: task.id,
      title: titleController.text.trim(),
      description: descriptionController.text.trim().isEmpty
          ? null
          : descriptionController.text.trim(),
      taskType: taskTypeController.text.trim(),
      state: selectedState,
      rank: int.tryParse(rankController.text.trim()) ?? 0,
      dependsOnTaskIDs: task.dependsOnTaskIDs,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed updating task: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Task updated.';
    });
  }

  Future<void> _deleteTask(TaskboardTaskModel task) async {
    final board = _board;
    if (board == null || _isMutating) {
      return;
    }
    if (_isBoardEnded(board)) {
      _setReadOnlyStatus();
      return;
    }
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: const Text('Delete Task'),
          content: Text('Delete task "${task.title}"?'),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Delete'),
            ),
          ],
        );
      },
    );
    if (confirmed != true) {
      return;
    }

    setState(() => _isMutating = true);
    final result = await widget.api.deleteTaskboardTask(
      projectID: widget.projectID,
      boardID: board.boardID,
      taskID: task.id,
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _isMutating = false;
      if (!result.isSuccess || result.data == null) {
        _statusMessage =
            'Failed deleting task: ${result.errorMessage ?? 'unknown error'}';
        return;
      }
      _board = result.data;
      _statusMessage = 'Task deleted.';
    });
  }

  bool _isBoardEnded(TaskboardModel board) {
    final state = board.state.trim().toLowerCase();
    return state == 'completed' || state == 'failed';
  }

  int _taskCount(TaskboardModel board) {
    var count = 0;
    for (final epic in board.epics) {
      count += epic.tasks.length;
    }
    return count;
  }

  Color _stateAccentColor(String value) {
    final normalized = value.trim().toLowerCase();
    if (normalized == 'failed' ||
        normalized == 'blocked' ||
        normalized == 'cancelled') {
      return Theme.of(context).colorScheme.error;
    }
    if (normalized == 'completed' || normalized == 'done') {
      return Colors.green.shade600;
    }
    if (normalized == 'active' || normalized == 'in_progress') {
      return Colors.amber.shade700;
    }
    return Theme.of(context).colorScheme.primary;
  }

  Widget _buildStatusPill({
    required String label,
    required String value,
    Color? accent,
  }) {
    final color = accent ?? Theme.of(context).colorScheme.primary;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: color,
          fontWeight: FontWeight.w600,
          fontSize: 12,
        ),
      ),
    );
  }

  Widget _buildPanel({
    required Widget child,
    Color? accent,
    EdgeInsetsGeometry padding = const EdgeInsets.all(12),
  }) {
    final borderColor = (accent ?? Theme.of(context).colorScheme.outline)
        .withValues(alpha: 0.35);
    return Container(
      padding: padding,
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: borderColor),
      ),
      child: child,
    );
  }

  Widget _buildEpicCard(
    TaskboardEpicModel epic,
    bool mutationControlsDisabled,
  ) {
    final accent = _stateAccentColor(epic.state);
    final isExpanded = _expandedEpicIDs.contains(epic.id);
    return _buildPanel(
      accent: accent,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  epic.title,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
              ),
              _buildStatusPill(
                label: 'STATE',
                value: epic.state,
                accent: accent,
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            epic.objective?.trim().isNotEmpty == true
                ? epic.objective!
                : 'No objective set.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              _buildStatusPill(
                label: 'RANK',
                value: '${epic.rank}',
                accent: accent,
              ),
              _buildStatusPill(
                label: 'TASKS',
                value: '${epic.tasks.length}',
                accent: Theme.of(context).colorScheme.primary,
              ),
            ],
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              if (epic.tasks.isNotEmpty)
                OutlinedButton.icon(
                  onPressed: () {
                    setState(() {
                      if (isExpanded) {
                        _expandedEpicIDs.remove(epic.id);
                      } else {
                        _expandedEpicIDs.add(epic.id);
                      }
                    });
                  },
                  icon: Icon(
                    isExpanded
                        ? Icons.keyboard_arrow_up
                        : Icons.keyboard_arrow_down,
                  ),
                  label: Text(isExpanded ? 'Hide Tasks' : 'Show Tasks'),
                ),
              OutlinedButton(
                onPressed: mutationControlsDisabled
                    ? null
                    : () => _editEpic(epic),
                child: const Text('Edit Epic'),
              ),
              OutlinedButton(
                onPressed: mutationControlsDisabled
                    ? null
                    : () => _deleteEpic(epic),
                child: const Text('Delete Epic'),
              ),
              OutlinedButton(
                onPressed: mutationControlsDisabled
                    ? null
                    : () => _createTask(epic),
                child: const Text('Add Task'),
              ),
            ],
          ),
          if (isExpanded) ...<Widget>[
            const SizedBox(height: 10),
            ...epic.tasks.map((TaskboardTaskModel task) {
              final taskAccent = _stateAccentColor(task.state);
              return Container(
                margin: const EdgeInsets.only(top: 8),
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: taskAccent.withValues(alpha: 0.35)),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Text(
                            task.title,
                            style: const TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Type: ${task.taskType} • State: ${task.state} • Rank: ${task.rank}',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Wrap(
                      spacing: 2,
                      children: <Widget>[
                        IconButton(
                          onPressed: mutationControlsDisabled
                              ? null
                              : () => _editTask(epic, task),
                          icon: const Icon(Icons.edit_outlined),
                          tooltip: 'Edit Task',
                        ),
                        IconButton(
                          onPressed: mutationControlsDisabled
                              ? null
                              : () => _deleteTask(task),
                          icon: const Icon(Icons.delete_outline),
                          tooltip: 'Delete Task',
                        ),
                      ],
                    ),
                  ],
                ),
              );
            }),
          ],
        ],
      ),
    );
  }

  void _setReadOnlyStatus() {
    if (!mounted) {
      return;
    }
    setState(() {
      _statusMessage = _readOnlyMessage;
    });
  }

  void _expandAllEpics(TaskboardModel board) {
    setState(() {
      _expandedEpicIDs
        ..clear()
        ..addAll(
          board.epics
              .where((TaskboardEpicModel epic) => epic.tasks.isNotEmpty)
              .map((TaskboardEpicModel epic) => epic.id),
        );
    });
  }

  void _collapseAllEpics() {
    setState(() {
      _expandedEpicIDs.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    final board = _board;
    final isReadOnly = board != null && _isBoardEnded(board);
    final mutationControlsDisabled = _isMutating || board == null || isReadOnly;
    final runAgainDisabled =
        _isMutating || board == null || !_canRunAgain(board);
    final hasAnyTasks =
        board?.epics.any((TaskboardEpicModel epic) => epic.tasks.isNotEmpty) ??
        false;
    final hasExpandedEpics = _expandedEpicIDs.isNotEmpty;
    final expandCollapseTooltip = hasExpandedEpics
        ? 'Collapse All Tasks'
        : 'Expand All Tasks';

    final appBarStateColor = board == null
        ? Theme.of(context).colorScheme.primary
        : _stateAccentColor(board.state);

    return Scaffold(
      appBar: AppBar(
        leadingWidth: 230,
        leading: Row(
          children: <Widget>[
            const BackButton(),
            if (board != null)
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: <Widget>[
                      _buildStatusPill(
                        label: 'STATE',
                        value: board.state,
                        accent: appBarStateColor,
                      ),
                      const SizedBox(width: 8),
                      _buildStatusPill(
                        label: 'EPICS',
                        value: '${board.epics.length}',
                        accent: Theme.of(context).colorScheme.primary,
                      ),
                    ],
                  ),
                ),
              ),
          ],
        ),
        centerTitle: true,
        title: const Text('Taskboard Details'),
        actions: <Widget>[
          if (_isMutating)
            const Padding(
              padding: EdgeInsets.only(right: 6),
              child: Center(
                child: SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ),
            ),
          IconButton(
            onPressed: !hasAnyTasks
                ? null
                : () {
                    final currentBoard = _board;
                    if (currentBoard == null) {
                      return;
                    }
                    if (hasExpandedEpics) {
                      _collapseAllEpics();
                      return;
                    }
                    _expandAllEpics(currentBoard);
                  },
            icon: Icon(
              hasExpandedEpics
                  ? Icons.unfold_less_outlined
                  : Icons.unfold_more_outlined,
            ),
            tooltip: expandCollapseTooltip,
          ),
          IconButton(
            onPressed: _loadBoard,
            icon: const Icon(Icons.refresh),
            tooltip: 'Reload',
          ),
          IconButton(
            onPressed: runAgainDisabled ? null : _runTaskboardAgain,
            icon: const Icon(Icons.play_circle_outline),
            tooltip: 'Run Again',
          ),
          IconButton(
            onPressed: mutationControlsDisabled ? null : _editBoard,
            icon: const Icon(Icons.edit_outlined),
            tooltip: 'Edit Taskboard',
          ),
          IconButton(
            onPressed: mutationControlsDisabled ? null : _deleteBoard,
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Delete Taskboard',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : board == null
          ? Center(child: Text(_statusMessage ?? 'Taskboard unavailable'))
          : SafeArea(
              child: RefreshIndicator(
                onRefresh: _loadBoard,
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      _buildPanel(
                        accent: appBarStateColor,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            Row(
                              children: <Widget>[
                                Expanded(
                                  child: Text(
                                    board.name,
                                    style: const TextStyle(
                                      fontSize: 18,
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                ),
                                FilledButton.icon(
                                  onPressed: mutationControlsDisabled
                                      ? null
                                      : _createEpic,
                                  icon: const Icon(Icons.add),
                                  label: const Text('Add Epic'),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Text(
                              'Board ID: ${board.boardID}',
                              style: Theme.of(context).textTheme.labelMedium,
                            ),
                            const SizedBox(height: 8),
                            Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              children: <Widget>[
                                _buildStatusPill(
                                  label: 'STATE',
                                  value: board.state,
                                  accent: appBarStateColor,
                                ),
                                _buildStatusPill(
                                  label: 'EPICS',
                                  value: '${board.epics.length}',
                                  accent: Theme.of(context).colorScheme.primary,
                                ),
                                _buildStatusPill(
                                  label: 'TASKS',
                                  value: '${_taskCount(board)}',
                                  accent: Theme.of(context).colorScheme.primary,
                                ),
                                _buildStatusPill(
                                  label: 'UPDATED',
                                  value: board.updatedAt.toIso8601String(),
                                  accent: Theme.of(
                                    context,
                                  ).colorScheme.secondary,
                                ),
                              ],
                            ),
                            if (isReadOnly) ...<Widget>[
                              const SizedBox(height: 10),
                              Text(
                                _readOnlyMessage,
                                style: TextStyle(
                                  color: Theme.of(context).colorScheme.error,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                      const SizedBox(height: 10),
                      if (board.epics.isEmpty)
                        _buildPanel(
                          child: const Text(
                            'No epics yet. Create one to begin managing tasks.',
                          ),
                        )
                      else
                        ...board.epics.map((TaskboardEpicModel epic) {
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 10),
                            child: _buildEpicCard(
                              epic,
                              mutationControlsDisabled,
                            ),
                          );
                        }),
                      if (_statusMessage != null) ...<Widget>[
                        _buildPanel(
                          accent: Theme.of(context).colorScheme.primary,
                          child: Text(_statusMessage!),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
    );
  }
}
