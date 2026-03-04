import 'package:agentic_repositories/features/settings/screens/settings_screen.dart';
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

  testWidgets('renders delegated connection settings view', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: SettingsScreen(
            endpointController: endpointController,
            isSavingEndpoint: false,
            isRunningAction: false,
            onSaveEndpoint: () {},
            onTestConnection: () {},
            statusMessage: 'ok',
          ),
        ),
      ),
    );

    expect(find.text('Connection Settings'), findsOneWidget);
    expect(find.text('Save'), findsOneWidget);
    expect(find.text('Test'), findsOneWidget);
    expect(find.text('ok'), findsOneWidget);
  });
}
