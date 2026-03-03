import 'package:agentic_worktrees/features/projects/screens/project_setup_edit_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../../support/test_data.dart';

void main() {
  testWidgets('delete token field hides token input without throwing', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: ProjectSetupEditScreen(
          projectSetup: sampleProjectSetup(),
          endpoint: 'http://localhost:8080/graphql',
        ),
      ),
    );

    expect(find.text('Regenerate Token'), findsOneWidget);
    expect(find.text('Delete token field'), findsNothing);
    expect(find.widgetWithText(TextField, 'New SCM Token'), findsNothing);

    await tester.tap(find.text('Regenerate Token'));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('Delete token field'), findsOneWidget);
    expect(find.widgetWithText(TextField, 'New SCM Token'), findsOneWidget);

    await tester.tap(find.text('Delete token field'));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('Delete token field'), findsNothing);
    expect(find.widgetWithText(TextField, 'New SCM Token'), findsNothing);
    expect(find.text('Regenerate Token'), findsOneWidget);
  });
}
