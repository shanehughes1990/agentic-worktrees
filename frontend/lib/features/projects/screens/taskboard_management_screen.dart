import 'dart:async';

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
  bool _isLoading = true;
  bool _isMutating = false;
  String? _statusMessage;

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

  Future<void> _editBoard() async {
    final board = _board;
    if (board == null || _isMutating) {
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

  @override
  Widget build(BuildContext context) {
    final board = _board;
    return Scaffold(
      appBar: AppBar(
        title: Text(board?.name ?? widget.boardID),
        actions: <Widget>[
          IconButton(
            onPressed: _isMutating || board == null ? null : _editBoard,
            icon: const Icon(Icons.edit_outlined),
            tooltip: 'Edit Taskboard',
          ),
          IconButton(
            onPressed: _isMutating || board == null ? null : _deleteBoard,
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Delete Taskboard',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : board == null
          ? Center(child: Text(_statusMessage ?? 'Taskboard unavailable'))
          : RefreshIndicator(
              onRefresh: _loadBoard,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: <Widget>[
                  if (_isMutating)
                    const Padding(
                      padding: EdgeInsets.only(bottom: 12),
                      child: LinearProgressIndicator(),
                    ),
                  Row(
                    children: <Widget>[
                      Expanded(
                        child: Text(
                          'Taskboard: ${board.name}',
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                      FilledButton.icon(
                        onPressed: _isMutating ? null : _createEpic,
                        icon: const Icon(Icons.add),
                        label: const Text('Add Epic'),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text('State: ${board.state}'),
                  Text('Board ID: ${board.boardID}'),
                  const SizedBox(height: 12),
                  if (board.epics.isEmpty)
                    const Text(
                      'No epics yet. Create one to begin managing tasks.',
                    )
                  else
                    ...board.epics.map((TaskboardEpicModel epic) {
                      return Card(
                        margin: const EdgeInsets.only(bottom: 10),
                        child: ExpansionTile(
                          tilePadding: const EdgeInsets.symmetric(
                            horizontal: 12,
                          ),
                          title: Text(epic.title),
                          subtitle: Text(
                            'State: ${epic.state} • Rank: ${epic.rank}',
                          ),
                          childrenPadding: const EdgeInsets.fromLTRB(
                            12,
                            0,
                            12,
                            12,
                          ),
                          children: <Widget>[
                            Row(
                              children: <Widget>[
                                Expanded(
                                  child: Text(
                                    epic.objective?.trim().isNotEmpty == true
                                        ? epic.objective!
                                        : 'No objective',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                                ),
                                TextButton(
                                  onPressed: _isMutating
                                      ? null
                                      : () => _editEpic(epic),
                                  child: const Text('Edit'),
                                ),
                                TextButton(
                                  onPressed: _isMutating
                                      ? null
                                      : () => _deleteEpic(epic),
                                  child: const Text('Delete'),
                                ),
                                TextButton(
                                  onPressed: _isMutating
                                      ? null
                                      : () => _createTask(epic),
                                  child: const Text('Add Task'),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            if (epic.tasks.isEmpty)
                              const Align(
                                alignment: Alignment.centerLeft,
                                child: Text('No tasks in this epic.'),
                              )
                            else
                              ...epic.tasks.map((TaskboardTaskModel task) {
                                return ListTile(
                                  dense: true,
                                  contentPadding: EdgeInsets.zero,
                                  title: Text(task.title),
                                  subtitle: Text(
                                    'Type: ${task.taskType} • State: ${task.state} • Rank: ${task.rank}',
                                  ),
                                  trailing: Wrap(
                                    spacing: 4,
                                    children: <Widget>[
                                      IconButton(
                                        onPressed: _isMutating
                                            ? null
                                            : () => _editTask(epic, task),
                                        icon: const Icon(Icons.edit_outlined),
                                        tooltip: 'Edit Task',
                                      ),
                                      IconButton(
                                        onPressed: _isMutating
                                            ? null
                                            : () => _deleteTask(task),
                                        icon: const Icon(Icons.delete_outline),
                                        tooltip: 'Delete Task',
                                      ),
                                    ],
                                  ),
                                );
                              }),
                          ],
                        ),
                      );
                    }),
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
