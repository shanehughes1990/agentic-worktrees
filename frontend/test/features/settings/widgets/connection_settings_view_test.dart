import 'package:agentic_worktrees/features/settings/widgets/connection_settings_view.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late TextEditingController endpointController;

  setUp(() {
    endpointController = TextEditingController(text: 'http://localhost:8080/query');
  });

  tearDown(() {
    endpointController.dispose();
  });

  Future<void> pumpSubject(
    WidgetTester tester, {
    required VoidCallback onSave,
    required VoidCallback onTest,
    bool isSaving = false,
    bool isRunning = false,
    String? statusMessage,
  }) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ConnectionSettingsView(
            endpointController: endpointController,
            isSavingEndpoint: isSaving,
            isRunningAction: isRunning,
            onSave: onSave,
            onTest: onTest,
            statusMessage: statusMessage,
          ),
        ),
      ),
    );
  }

  testWidgets('renders heading, input, and status message', (
    WidgetTester tester,
  ) async {
    await pumpSubject(
      tester,
      onSave: () {},
      onTest: () {},
      statusMessage: 'saved',
    );

    expect(find.text('Connection Settings'), findsOneWidget);
    expect(find.text('GraphQL HTTP Endpoint'), findsOneWidget);
    expect(find.text('saved'), findsOneWidget);
  });

  testWidgets('invokes save and test callbacks', (WidgetTester tester) async {
    var saveCount = 0;
    var testCount = 0;

    await pumpSubject(
      tester,
      onSave: () => saveCount++,
      onTest: () => testCount++,
    );

    await tester.tap(find.text('Save'));
    await tester.pump();
    await tester.tap(find.text('Test'));
    await tester.pump();

    expect(saveCount, 1);
    expect(testCount, 1);
  });
}
