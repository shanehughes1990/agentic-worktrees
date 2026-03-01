import 'package:agentic_worktrees/features/settings/logic/connection_settings_logic.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('normalizeEndpoint delegates to endpoint normalization', () {
    final normalized = ConnectionSettingsLogic.normalizeEndpoint('localhost:8080');

    expect(normalized, 'http://localhost:8080/query');
  });

  group('validateEndpoint', () {
    test('returns error for empty endpoint', () {
      expect(
        ConnectionSettingsLogic.validateEndpoint('  '),
        'Endpoint cannot be empty.',
      );
    });

    test('returns null for non-empty endpoint', () {
      expect(ConnectionSettingsLogic.validateEndpoint('http://x'), isNull);
    });
  });

  test('successMessage includes row count', () {
    expect(
      ConnectionSettingsLogic.successMessage(3),
      'Connection successful (3 session rows returned).',
    );
  });

  test('failureMessage includes endpoint and compact error', () {
    expect(
      ConnectionSettingsLogic.failureMessage(
        endpoint: 'http://localhost:8080/query',
        compactError: 'request failed',
      ),
      'Connection failed at http://localhost:8080/query: request failed',
    );
  });
}
