import 'package:agentic_worktrees/app/app.dart';
import 'package:agentic_worktrees/shared/config/app_config.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('App bootstraps dashboard shell', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: <Override>[
          appConfigProvider.overrideWith(() => _FakeAppConfigNotifier()),
        ],
        child: const AgenticWorktreesApp(),
      ),
    );

    await tester.pumpAndSettle();

    expect(
      find.text('Agentic Worktrees Desktop Control Plane'),
      findsOneWidget,
    );
    expect(find.byType(Scaffold), findsOneWidget);
  });
}

class _FakeAppConfigNotifier extends AppConfigNotifier {
  @override
  Future<AppConfigState> build() async {
    return const AppConfigState(
      graphqlHttpEndpoint: defaultGraphqlHttpEndpoint,
    );
  }
}
