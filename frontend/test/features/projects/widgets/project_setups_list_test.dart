import 'package:agentic_worktrees/features/projects/widgets/project_setups_list.dart';
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../../support/test_data.dart';

void main() {
  Future<void> pumpSubject(
    WidgetTester tester, {
    required List<ProjectSetupConfig> projectSetups,
    required String selectedProjectID,
    required ValueChanged<ProjectSetupConfig> onProjectSelected,
    bool dense = false,
  }) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ProjectSetupsList(
            projectSetups: projectSetups,
            selectedProjectID: selectedProjectID,
            onProjectSelected: onProjectSelected,
            dense: dense,
          ),
        ),
      ),
    );
  }

  testWidgets('renders empty state when no setups provided', (
    WidgetTester tester,
  ) async {
    await pumpSubject(
      tester,
      projectSetups: const <ProjectSetupConfig>[],
      selectedProjectID: '',
      onProjectSelected: (_) {},
    );

    expect(find.text('No project setups configured.'), findsOneWidget);
  });

  testWidgets('renders setups and handles selection', (WidgetTester tester) async {
    final setupA = sampleProjectSetup(projectID: 'project-a');
    final setupB = sampleProjectSetup(projectID: 'project-b');
    ProjectSetupConfig? selected;

    await pumpSubject(
      tester,
      projectSetups: <ProjectSetupConfig>[setupA, setupB],
      selectedProjectID: 'project-b',
      onProjectSelected: (ProjectSetupConfig setup) => selected = setup,
      dense: true,
    );

    expect(find.text('project-a'), findsOneWidget);
    expect(find.text('project-b'), findsOneWidget);

    await tester.tap(find.text('project-a'));
    await tester.pump();

    expect(selected, setupA);
  });
}
