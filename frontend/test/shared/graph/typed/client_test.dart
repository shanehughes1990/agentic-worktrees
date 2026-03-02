import 'package:agentic_worktrees/shared/graph/typed/client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('canonicalGraphqlHttpUri', () {
    test('defaults missing path to /query', () {
      final uri = canonicalGraphqlHttpUri('http://localhost:8080');
      expect(uri.toString(), 'http://localhost:8080/query');
    });

    test('adds scheme when missing', () {
      final uri = canonicalGraphqlHttpUri('localhost:8080/query');
      expect(uri.toString(), 'http://localhost:8080/query');
    });

    test('strips fragment and keeps query', () {
      final uri = canonicalGraphqlHttpUri(
        'http://localhost:8080/query?x=1#anchor',
      );
      expect(uri.toString(), 'http://localhost:8080/query?x=1');
    });

    test('trims trailing path slash', () {
      final uri = canonicalGraphqlHttpUri('https://example.com/graphql/');
      expect(uri.toString(), 'https://example.com/graphql');
    });
  });
}
