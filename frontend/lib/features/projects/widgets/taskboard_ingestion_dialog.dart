import 'package:agentic_repositories/shared/graph/typed/api.dart';
import 'package:agentic_repositories/shared/graph/typed/models.dart';
import 'package:flutter/material.dart';

class NewTaskboardDraft {
  const NewTaskboardDraft({
    required this.taskboardName,
    required this.selectedDocumentIDs,
    required this.userPrompt,
    required this.repositorySourceBranches,
  });

  final String taskboardName;
  final List<String>? selectedDocumentIDs;
  final String? userPrompt;
  final Map<String, String>? repositorySourceBranches;
}

Future<NewTaskboardDraft?> showTaskboardIngestionDialog({
  required BuildContext context,
  required ControlPlaneApi api,
  required String projectID,
  required List<ProjectDocument> projectDocuments,
  required List<ProjectRepositoryBranchOption> repositoryBranchOptions,
  String title = 'New Taskboard',
  String submitLabel = 'Create',
  String? initialTaskboardName,
  String? initialUserPrompt,
  Set<String>? initialSelectedDocumentIDs,
}) async {
  final selectedDocumentIDs = <String>{
    ...(initialSelectedDocumentIDs ??
        projectDocuments
            .map((ProjectDocument document) => document.documentID)
            .toSet()),
  };
  final selectedBranches = <String, String>{
    for (final option in repositoryBranchOptions)
      if (option.branches.isNotEmpty)
        option.repositoryID: option.branches.contains(option.defaultBranch)
            ? option.defaultBranch!
            : option.branches.first,
  };
  final taskboardNameController = TextEditingController(
    text: (initialTaskboardName ?? '').trim(),
  );
  final promptController = TextEditingController(
    text: (initialUserPrompt ?? '').trim(),
  );
  var isGeneratingPrompt = false;

  final draft = await showDialog<NewTaskboardDraft>(
    context: context,
    builder: (BuildContext context) {
      return StatefulBuilder(
        builder: (BuildContext context, StateSetter setDialogState) {
          final isAllSelected =
              projectDocuments.isNotEmpty &&
              selectedDocumentIDs.length == projectDocuments.length;
          return AlertDialog(
            title: Text(title),
            content: SizedBox(
              width: 520,
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    TextField(
                      controller: taskboardNameController,
                      decoration: const InputDecoration(
                        labelText: 'Taskboard name',
                        hintText: 'Required',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    CheckboxListTile(
                      value: isAllSelected,
                      onChanged: (bool? value) {
                        setDialogState(() {
                          if (value == true) {
                            selectedDocumentIDs
                              ..clear()
                              ..addAll(
                                projectDocuments.map(
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
                    ...projectDocuments.map((ProjectDocument document) {
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
                    const SizedBox(height: 8),
                    Align(
                      alignment: Alignment.centerLeft,
                      child: OutlinedButton.icon(
                        onPressed: isGeneratingPrompt
                            ? null
                            : () async {
                                final taskboardName = taskboardNameController
                                    .text
                                    .trim();
                                if (taskboardName.isEmpty) {
                                  setDialogState(() {
                                    isGeneratingPrompt = false;
                                  });
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    const SnackBar(
                                      content: Text(
                                        'Enter a taskboard name before generating a prompt.',
                                      ),
                                    ),
                                  );
                                  return;
                                }
                                setDialogState(() {
                                  isGeneratingPrompt = true;
                                });
                                final response = await api
                                    .refineIngestionPrompt(
                                      projectID: projectID,
                                      taskboardName: taskboardName,
                                      userPrompt: promptController.text,
                                    );
                                if (!context.mounted) {
                                  return;
                                }
                                if (response.isSuccess &&
                                    response.data != null &&
                                    response.data!.trim().isNotEmpty) {
                                  final generatedPrompt = response.data!.trim();
                                  promptController.text = generatedPrompt;
                                  promptController.selection =
                                      TextSelection.collapsed(
                                        offset: generatedPrompt.length,
                                      );
                                } else {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(
                                      content: Text(
                                        'Prompt generation failed: ${response.errorMessage ?? 'unknown error'}',
                                      ),
                                    ),
                                  );
                                }
                                setDialogState(() {
                                  isGeneratingPrompt = false;
                                });
                              },
                        icon: isGeneratingPrompt
                            ? const SizedBox(
                                height: 16,
                                width: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.auto_awesome),
                        label: const Text('AI: Generate Prompt'),
                      ),
                    ),
                    if (repositoryBranchOptions.isNotEmpty) ...<Widget>[
                      const SizedBox(height: 12),
                      const Text(
                        'Repository branches',
                        style: TextStyle(fontWeight: FontWeight.w600),
                      ),
                      const SizedBox(height: 8),
                      ...repositoryBranchOptions.map((option) {
                        final branches = option.branches;
                        final selectedBranch =
                            selectedBranches[option.repositoryID];
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 8),
                          child: DropdownButtonFormField<String>(
                            initialValue: selectedBranch,
                            onChanged: branches.isEmpty
                                ? null
                                : (String? value) {
                                    if (value == null) {
                                      return;
                                    }
                                    setDialogState(() {
                                      selectedBranches[option.repositoryID] =
                                          value;
                                    });
                                  },
                            decoration: InputDecoration(
                              labelText: option.repositoryURL,
                              border: const OutlineInputBorder(),
                            ),
                            items: branches
                                .map(
                                  (String branch) => DropdownMenuItem<String>(
                                    value: branch,
                                    child: Text(branch),
                                  ),
                                )
                                .toList(growable: false),
                          ),
                        );
                      }),
                    ],
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
                  final taskboardName = taskboardNameController.text.trim();
                  final selected = selectedDocumentIDs.toList(growable: false);
                  final prompt = promptController.text.trim();
                  if (taskboardName.isEmpty) {
                    return;
                  }
                  if (selected.isEmpty && prompt.isEmpty) {
                    return;
                  }
                  Navigator.of(context).pop(
                    NewTaskboardDraft(
                      taskboardName: taskboardName,
                      selectedDocumentIDs: selected.isEmpty ? null : selected,
                      userPrompt: prompt.isEmpty ? null : prompt,
                      repositorySourceBranches: selectedBranches.isEmpty
                          ? null
                          : Map<String, String>.from(selectedBranches),
                    ),
                  );
                },
                child: Text(submitLabel),
              ),
            ],
          );
        },
      );
    },
  );

  taskboardNameController.dispose();
  promptController.dispose();

  return draft;
}
