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

  testWidgets('renders reshaped blocks and status', (
    WidgetTester tester,
  ) async {
    await pumpSubject(
      tester,
      onSave: () {},
      onReload: () {},
      onScmProviderChanged: (_) {},
      onTrackerProviderChanged: (_) {},
      onProjectSelected: (_) {},
    );

    expect(find.text('Project Setup'), findsNWidgets(2));
    expect(find.text('SCM Provider'), findsOneWidget);
    expect(find.text('Repository Setup'), findsOneWidget);
    expect(find.text('Tracker Setup'), findsOneWidget);
    expect(find.text('Tracker Provider'), findsOneWidget);
    expect(find.text('Add Repository'), findsOneWidget);
    expect(find.text('Add Tracker'), findsNothing);
    expect(find.text('status'), findsOneWidget);
  });

  testWidgets('project id auto-generates and is read-only', (
    WidgetTester tester,
  ) async {
    await pumpSubject(
      tester,
      onSave: () {},
      onReload: () {},
      onScmProviderChanged: (_) {},
      onTrackerProviderChanged: (_) {},
      onProjectSelected: (_) {},
    );

    await tester.enterText(find.byType(TextField).at(0), 'My Sample Project');
    await tester.pump();

    expect(projectController.text, 'my_sample_project');
    final projectIDField = tester.widget<TextField>(
      find.byType(TextField).at(1),
    );
    expect(projectIDField.readOnly, isTrue);
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

    await tester.ensureVisible(find.text('Save Project Setup'));
    await tester.tap(find.text('Save Project Setup'));
    await tester.pump();
    await tester.ensureVisible(find.text('Reload'));
    await tester.tap(find.text('Reload'));
    await tester.pump();

    expect(saveCount, 1);
    expect(reloadCount, 1);
  });

  testWidgets('adds repository setup blocks', (WidgetTester tester) async {
    await pumpSubject(
      tester,
      onSave: () {},
      onReload: () {},
      onScmProviderChanged: (_) {},
      onTrackerProviderChanged: (_) {},
      onProjectSelected: (_) {},
    );

    await tester.tap(find.text('Add Repository'));
    await tester.pump();

    expect(find.text('Repository 2'), findsOneWidget);
    expect(find.text('Tracker 2'), findsNothing);
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
