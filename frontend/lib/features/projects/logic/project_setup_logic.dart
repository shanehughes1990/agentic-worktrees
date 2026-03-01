import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/widgets.dart';

class ProjectSetupLogic {
  const ProjectSetupLogic._();

  static const String defaultScmProvider = 'GITHUB';
  static const String defaultTrackerProvider = 'GITHUB_ISSUES';

  static const List<String> trackerProviderOptions = <String>[
    'GITHUB_ISSUES',
    'JIRA',
    'LOCAL_JSON',
    'LINEAR',
  ];

  static String? validateRequiredFields({
    required String projectID,
    required String projectName,
    required String repositoryURL,
  }) {
    if (projectID.isEmpty || projectName.isEmpty || repositoryURL.isEmpty) {
      return 'Project ID, Project Name, and Repository URL are required.';
    }
    return null;
  }

  static void applySetupToForm({
    required ProjectSetupConfig setup,
    required TextEditingController projectController,
    required TextEditingController projectNameController,
    required TextEditingController repositoryUrlController,
    required TextEditingController trackerLocationController,
    required TextEditingController trackerBoardIDController,
    required ValueChanged<String> onScmProviderChanged,
    required ValueChanged<String> onTrackerProviderChanged,
  }) {
    projectController.text = setup.projectID;
    projectNameController.text = setup.projectName;
    repositoryUrlController.text = setup.repositoryURL;
    onScmProviderChanged(setup.scmProvider);
    onTrackerProviderChanged(setup.trackerProvider);
    trackerLocationController.text = setup.trackerLocation;
    trackerBoardIDController.text = setup.trackerBoardID;
  }
}
