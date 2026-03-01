import 'package:agentic_worktrees/features/projects/logic/project_setup_logic.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../../support/test_data.dart';

void main() {
  group('ProjectSetupLogic.validateRequiredFields', () {
    test('returns error when required fields are missing', () {
      final error = ProjectSetupLogic.validateRequiredFields(
        projectID: '',
        projectName: 'name',
        repositoryURL: 'https://example.com/repo',
      );

      expect(error, 'Project ID, Project Name, and Repository URL are required.');
    });

    test('returns null for valid input', () {
      final error = ProjectSetupLogic.validateRequiredFields(
        projectID: 'project-1',
        projectName: 'Project',
        repositoryURL: 'https://example.com/repo',
      );

      expect(error, isNull);
    });
  });

  test('applySetupToForm updates all controllers and providers', () {
    final setup = sampleProjectSetup();
    final projectController = TextEditingController();
    final projectNameController = TextEditingController();
    final repositoryController = TextEditingController();
    final trackerLocationController = TextEditingController();
    final trackerBoardController = TextEditingController();
    String scmProvider = '';
    String trackerProvider = '';

    ProjectSetupLogic.applySetupToForm(
      setup: setup,
      projectController: projectController,
      projectNameController: projectNameController,
      repositoryUrlController: repositoryController,
      trackerLocationController: trackerLocationController,
      trackerBoardIDController: trackerBoardController,
      onScmProviderChanged: (String value) => scmProvider = value,
      onTrackerProviderChanged: (String value) => trackerProvider = value,
    );

    expect(projectController.text, setup.projectID);
    expect(projectNameController.text, setup.projectName);
    expect(repositoryController.text, setup.repositoryURL);
    expect(trackerLocationController.text, setup.trackerLocation);
    expect(trackerBoardController.text, setup.trackerBoardID);
    expect(scmProvider, setup.scmProvider);
    expect(trackerProvider, setup.trackerProvider);

    projectController.dispose();
    projectNameController.dispose();
    repositoryController.dispose();
    trackerLocationController.dispose();
    trackerBoardController.dispose();
  });

  test('tracker provider options include supported values', () {
    expect(
      ProjectSetupLogic.trackerProviderOptions,
      containsAll(<String>['GITHUB_ISSUES', 'JIRA', 'LOCAL_JSON', 'LINEAR']),
    );
  });
}
