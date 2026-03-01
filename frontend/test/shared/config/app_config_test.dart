import 'package:agentic_worktrees/shared/config/app_config.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('normalizeGraphqlEndpoint', () {
    test('adds http scheme and /query path when missing', () {
      final normalized = normalizeGraphqlEndpoint('localhost:8080');

      expect(normalized, 'http://localhost:8080/query');
    });

    test('preserves explicit path and trims trailing slash', () {
      final normalized = normalizeGraphqlEndpoint('https://example.com/graphql/');

      expect(normalized, 'https://example.com/graphql');
    });

    test('preserves query and fragment', () {
      final normalized = normalizeGraphqlEndpoint(
        'https://example.com/query?x=1#anchor',
      );

      expect(normalized, 'https://example.com/query?x=1#anchor');
    });

    test('returns empty string when input is empty', () {
      final normalized = normalizeGraphqlEndpoint('   ');

      expect(normalized, isEmpty);
    });

    test('normalizes default endpoint constant', () {
      final normalized = normalizeGraphqlEndpoint(defaultGraphqlHttpEndpoint);

      expect(normalized, defaultGraphqlHttpEndpoint);
    });
  });
}
