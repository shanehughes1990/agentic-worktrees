import 'package:agentic_worktrees/features/projects/screens/project_setup_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late TextEditingController projectController;
  late TextEditingController projectNameController;
  late TextEditingController repositoryController;
  late TextEditingController scmTokenController;

  setUp(() {
    projectController = TextEditingController();
    projectNameController = TextEditingController();
    repositoryController = TextEditingController();
    scmTokenController = TextEditingController();
  });

  tearDown(() {
    projectController.dispose();
    projectNameController.dispose();
    repositoryController.dispose();
    scmTokenController.dispose();
  });

  Future<void> pumpSubject(
    WidgetTester tester, {
    required VoidCallback onSave,
    required ValueChanged<String> onScmProviderChanged,
    bool isSaving = false,
  }) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ProjectSetupScreen(
            projectController: projectController,
            projectNameController: projectNameController,
            repositoryUrlController: repositoryController,
            setupScmProvider: 'GITHUB',
            scmTokenController: scmTokenController,
            onSetupScmProviderChanged: onScmProviderChanged,
            isSavingProjectSetup: isSaving,
            onSaveProjectSetup: onSave,
            statusMessage: 'status',
          ),
        ),
      ),
    );
  }

  testWidgets('renders reshaped blocks and status', (
    WidgetTester tester,
  ) async {
    await pumpSubject(tester, onSave: () {}, onScmProviderChanged: (_) {});

    expect(find.text('Project Setup'), findsNWidgets(2));
    expect(find.text('SCM Provider'), findsOneWidget);
    expect(find.text('Repository Setup'), findsOneWidget);
    expect(find.text('Tracker Setup'), findsOneWidget);
    expect(find.text('Add Repository'), findsOneWidget);
    expect(find.text('Configured Projects'), findsNothing);
    expect(find.text('Add Tracker'), findsNothing);
    expect(find.text('status'), findsOneWidget);
  });

  testWidgets('project id auto-generates without visible field', (
    WidgetTester tester,
  ) async {
    await pumpSubject(tester, onSave: () {}, onScmProviderChanged: (_) {});

    await tester.enterText(find.byType(TextField).at(0), 'My Sample Project');
    await tester.pump();

    expect(projectController.text, 'my_sample_project');
    expect(find.widgetWithText(TextField, 'Project ID'), findsNothing);
    expect(find.text('Project ID'), findsNothing);
  });

  testWidgets('invokes save callback and shows back button', (
    WidgetTester tester,
  ) async {
    var saveCount = 0;

    await pumpSubject(
      tester,
      onSave: () => saveCount++,
      onScmProviderChanged: (_) {},
    );

    await tester.ensureVisible(find.text('Save Project Setup'));
    await tester.tap(find.text('Save Project Setup'));
    await tester.pump();
    expect(find.text('Reload'), findsNothing);
    expect(find.text('Back'), findsOneWidget);

    expect(saveCount, 1);
  });

  testWidgets('adds repository setup blocks', (WidgetTester tester) async {
    await pumpSubject(tester, onSave: () {}, onScmProviderChanged: (_) {});

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
      onScmProviderChanged: (_) {},
    );

    final saveButton = tester.widget<FilledButton>(
      find.widgetWithText(FilledButton, 'Save Project Setup'),
    );
    expect(saveButton.onPressed, isNull);
  });
}
