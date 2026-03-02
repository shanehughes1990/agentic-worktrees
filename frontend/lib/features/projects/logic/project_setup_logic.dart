import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/widgets.dart';

class ProjectSetupLogic {
  const ProjectSetupLogic._();

  static const String defaultScmProvider = 'GITHUB';
  static const String defaultTrackerProvider = 'INTERNAL';

  static String? validateRequiredFields({
    required String projectID,
    required String projectName,
    required List<String> repositoryURLs,
    required String scmToken,
    required String taskboardName,
  }) {
    if (projectID.isEmpty || projectName.isEmpty || repositoryURLs.isEmpty) {
      return 'Project ID, Project Name, and at least one Repository URL are required.';
    }
    if (scmToken.trim().isEmpty) {
      return 'SCM Token is required.';
    }
    if (taskboardName.trim().isEmpty) {
      return 'Taskboard Name is required.';
    }
    return null;
  }

  static List<String> parseMultilineEntries(String rawValue) {
    return rawValue
        .split('\n')
        .map((String line) => line.trim())
        .where((String line) => line.isNotEmpty)
        .toList(growable: false);
  }

  static String projectIDFromName(String projectName) {
    final parts = projectName
        .trim()
        .split(RegExp(r'[^A-Za-z0-9]+'))
        .where((String part) => part.isNotEmpty)
        .map((String part) => part.toLowerCase())
        .toList(growable: false);
    if (parts.isEmpty) {
      return '';
    }
    return parts.join('_');
  }

  static void applySetupToForm({
    required ProjectSetupConfig setup,
    required TextEditingController projectController,
    required TextEditingController projectNameController,
    required TextEditingController repositoryUrlController,
    required TextEditingController taskboardNameController,
    required ValueChanged<String> onScmProviderChanged,
    required ValueChanged<String> onTrackerProviderChanged,
  }) {
    projectController.text = setup.projectID;
    projectNameController.text = setup.projectName;

    final repositoryURLs = setup.repositories
        .map((ProjectRepositoryConfig repository) => repository.repositoryURL)
        .where((String repositoryURL) => repositoryURL.trim().isNotEmpty)
        .toList(growable: false);
    repositoryUrlController.text = repositoryURLs.join('\n');
    final scm = setup.scms.isNotEmpty ? setup.scms.first : null;
    onScmProviderChanged(scm?.scmProvider ?? defaultScmProvider);

    final board = setup.boards.isNotEmpty ? setup.boards.first : null;
    taskboardNameController.text = board?.taskboardName ?? '';
    onTrackerProviderChanged(board?.trackerProvider ?? defaultTrackerProvider);
  }
}
