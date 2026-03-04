import 'package:agentic_repositories/shared/graph/typed/control_plane.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ApiResult', () {
    test('success captures data and success flag', () {
      const result = ApiResult<int>.success(42);

      expect(result.isSuccess, isTrue);
      expect(result.data, 42);
      expect(result.errorMessage, isNull);
    });

    test('failure captures error and failure flag', () {
      const result = ApiResult<int>.failure('boom');

      expect(result.isSuccess, isFalse);
      expect(result.data, isNull);
      expect(result.errorMessage, 'boom');
    });
  });

  group('prettyJson', () {
    test('formats valid json with indentation', () {
      final pretty = prettyJson('{"a":1,"b":2}');

      expect(pretty, contains('\n'));
      expect(pretty, contains('"a": 1'));
    });

    test('returns braces for empty input', () {
      expect(prettyJson('   '), '{}');
    });

    test('returns raw string when invalid json', () {
      const raw = 'not-json';

      expect(prettyJson(raw), raw);
    });
  });
}
