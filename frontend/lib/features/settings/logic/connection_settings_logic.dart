import 'package:agentic_worktrees/shared/config/app_config.dart';

class ConnectionSettingsLogic {
  const ConnectionSettingsLogic._();

  static String normalizeEndpoint(String rawEndpoint) {
    return normalizeGraphqlEndpoint(rawEndpoint);
  }

  static String? validateEndpoint(String endpoint) {
    if (endpoint.isEmpty) {
      return 'Endpoint cannot be empty.';
    }
    return null;
  }

  static String successMessage(int sessionRows) {
    return 'Connection successful ($sessionRows session rows returned).';
  }

  static String failureMessage({
    required String endpoint,
    required String compactError,
  }) {
    return 'Connection failed at $endpoint: $compactError';
  }
}
