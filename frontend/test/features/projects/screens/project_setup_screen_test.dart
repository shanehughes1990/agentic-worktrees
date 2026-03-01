import 'package:agentic_worktrees/features/projects/screens/project_setup_screen.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../../support/test_data.dart';

void main() {
  late TextEditingController projectController;
  late TextEditingController projectNameController;
  late TextEditingController repositoryController;
  late TextEditingController trackerLocationController;
  late TextEditingController trackerBoardController;

  setUp(() {
    projectController = TextEditingController();
    projectNameController = TextEditingController();
    repositoryController = TextEditingController();
    trackerLocationController = TextEditingController();
    trackerBoardController = TextEditingController();
  });

  tearDown(() {
    projectController.dispose();
    projectNameController.dispose();
    repositoryController.dispose();
    trackerLocationController.dispose();
    trackerBoardController.dispose();
  });

  Future<void> pumpSubject(
    WidgetTester tester, {
    required VoidCallback onSave,
    required VoidCallback onReload,
    required ValueChanged<String> onScmProviderChanged,
    required ValueChanged<String> onTrackerProviderChanged,
    required ValueChanged<ProjectSetupConfig> onProjectSelected,
    bool isSaving = false,
  }) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ProjectSetupScreen(
            projectController: projectController,
            projectNameController: projectNameController,
            repositoryUrlController: repositoryController,
            trackerLocationController: trackerLocationController,
            trackerBoardIDController: trackerBoardController,
            setupScmProvider: 'GITHUB',
            setupTrackerProvider: 'GITHUB_ISSUES',
            onSetupScmProviderChanged: onScmProviderChanged,
            onSetupTrackerProviderChanged: onTrackerProviderChanged,
            isSavingProjectSetup: isSaving,
            onSaveProjectSetup: onSave,
            onReloadProjectSetups: onReload,
            projectSetups: <ProjectSetupConfig>[sampleProjectSetup()],
            selectedProjectID: 'project-1',
            onProjectSelected: onProjectSelected,
            statusMessage: 'status',
          ),
        ),
      ),
    );
  }

  testWidgets('renders form sections and status', (WidgetTester tester) async {
    await pumpSubject(
      tester,
      onSave: () {},
      onReload: () {},
      onScmProviderChanged: (_) {},
      onTrackerProviderChanged: (_) {},
      onProjectSelected: (_) {},
    );

    expect(find.text('Project Setup'), findsOneWidget);
    expect(find.text('Save Project Setup'), findsOneWidget);
    expect(find.text('Configured Projects'), findsOneWidget);
    expect(find.text('status'), findsOneWidget);
  });

  testWidgets('invokes save and reload callbacks', (WidgetTester tester) async {
    var saveCount = 0;
    var reloadCount = 0;

    await pumpSubject(
      tester,
      onSave: () => saveCount++,
      onReload: () => reloadCount++,
      onScmProviderChanged: (_) {},
      onTrackerProviderChanged: (_) {},
      onProjectSelected: (_) {},
    );

    await tester.tap(find.text('Save Project Setup'));
    await tester.pump();
    await tester.tap(find.text('Reload'));
    await tester.pump();

    expect(saveCount, 1);
    expect(reloadCount, 1);
  });

  testWidgets('disables save button while saving', (WidgetTester tester) async {
    await pumpSubject(
      tester,
      isSaving: true,
      onSave: () {},
      onReload: () {},
      onScmProviderChanged: (_) {},
      onTrackerProviderChanged: (_) {},
      onProjectSelected: (_) {},
    );

    final saveButton = tester.widget<FilledButton>(
      find.widgetWithText(FilledButton, 'Save Project Setup'),
    );
    expect(saveButton.onPressed, isNull);
  });
}
