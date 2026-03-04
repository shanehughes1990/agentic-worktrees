import 'package:agentic_repositories/features/projects/logic/project_setup_logic.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../../support/test_data.dart';

void main() {
  group('ProjectSetupLogic.validateRequiredFields', () {
    test('returns error when required fields are missing', () {
      final error = ProjectSetupLogic.validateRequiredFields(
        projectID: '',
        projectName: 'name',
        repositoryURLs: const <String>['https://example.com/repo'],
        scmToken: 'token',
      );

      expect(
        error,
        'Project ID, Project Name, and at least one Repository URL are required.',
      );
    });

    test('returns null for valid input', () {
      final error = ProjectSetupLogic.validateRequiredFields(
        projectID: 'projectOne',
        projectName: 'Project One',
        repositoryURLs: const <String>['https://example.com/repo'],
        scmToken: 'token',
      );

      expect(error, isNull);
    });
  });

  test('applySetupToForm updates all controllers and providers', () {
    final setup = sampleProjectSetup();
    final projectController = TextEditingController();
    final projectNameController = TextEditingController();
    final repositoryController = TextEditingController();
    String scmProvider = '';

    ProjectSetupLogic.applySetupToForm(
      setup: setup,
      projectController: projectController,
      projectNameController: projectNameController,
      repositoryUrlController: repositoryController,
      onScmProviderChanged: (String value) => scmProvider = value,
    );

    expect(projectController.text, setup.projectID);
    expect(projectNameController.text, setup.projectName);
    expect(repositoryController.text, setup.repositories.first.repositoryURL);
    expect(scmProvider, setup.scms.first.scmProvider);

    projectController.dispose();
    projectNameController.dispose();
    repositoryController.dispose();
  });

  test('parseMultilineEntries filters blanks and trims', () {
    final entries = ProjectSetupLogic.parseMultilineEntries(
      '  https://github.com/acme/repo-1  \n\n https://github.com/acme/repo-2 ',
    );

    expect(entries, <String>[
      'https://github.com/acme/repo-1',
      'https://github.com/acme/repo-2',
    ]);
  });

  test('projectIDFromName converts name to lower snake case', () {
    expect(ProjectSetupLogic.projectIDFromName('Project One'), 'project_one');
    expect(
      ProjectSetupLogic.projectIDFromName('My cool_project 42'),
      'my_cool_project_42',
    );
    expect(
      ProjectSetupLogic.projectIDFromName('Project ABC 123'),
      'project_abc_123',
    );
    expect(ProjectSetupLogic.projectIDFromName('   '), '');
  });
}
