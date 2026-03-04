import 'package:agentic_repositories/shared/logging/app_logger.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _endpointPreferenceKey = 'graphql_http_endpoint';
const defaultGraphqlHttpEndpoint = 'http://localhost:8080/query';

String normalizeGraphqlEndpoint(String endpoint) {
  final trimmed = endpoint.trim();
  if (trimmed.isEmpty) {
    return trimmed;
  }

  final withScheme = trimmed.contains('://') ? trimmed : 'http://$trimmed';
  final parsed = Uri.tryParse(withScheme);
  if (parsed == null) {
    return withScheme;
  }

  final hasPath = parsed.path.isNotEmpty && parsed.path != '/';
  final normalizedPath = hasPath
      ? parsed.path.replaceAll(RegExp(r'/+$'), '')
      : '/query';

  final normalized = parsed
      .replace(
        path: normalizedPath,
        query: parsed.query.isEmpty ? null : parsed.query,
      )
      .toString();
  return normalized.split('#').first;
}

class AppConfigState {
  const AppConfigState({required this.graphqlHttpEndpoint});

  final String graphqlHttpEndpoint;

  AppConfigState copyWith({String? graphqlHttpEndpoint}) {
    return AppConfigState(
      graphqlHttpEndpoint: graphqlHttpEndpoint ?? this.graphqlHttpEndpoint,
    );
  }
}

class AppConfigNotifier extends AsyncNotifier<AppConfigState> {
  @override
  Future<AppConfigState> build() async {
    final preferences = await SharedPreferences.getInstance();
    final endpoint = normalizeGraphqlEndpoint(
      preferences.getString(_endpointPreferenceKey) ??
          defaultGraphqlHttpEndpoint,
    );

    if ((preferences.getString(_endpointPreferenceKey) ?? '').trim() !=
        endpoint) {
      await preferences.setString(_endpointPreferenceKey, endpoint);
    }

    AppLogger.instance.logger.i(
      'Loaded endpoint config',
      error: {'graphql_http_endpoint': endpoint},
    );

    return AppConfigState(graphqlHttpEndpoint: endpoint);
  }

  Future<void> saveGraphqlEndpoint(String endpoint) async {
    final cleaned = normalizeGraphqlEndpoint(endpoint);
    if (cleaned.isEmpty) {
      return;
    }
    final preferences = await SharedPreferences.getInstance();
    await preferences.setString(_endpointPreferenceKey, cleaned);
    AppLogger.instance.logger.i(
      'Saved endpoint config',
      error: {'graphql_http_endpoint': cleaned},
    );
    state = AsyncData(
      (state.valueOrNull ??
              const AppConfigState(
                graphqlHttpEndpoint: defaultGraphqlHttpEndpoint,
              ))
          .copyWith(graphqlHttpEndpoint: cleaned),
    );
  }
}

final appConfigProvider =
    AsyncNotifierProvider<AppConfigNotifier, AppConfigState>(
      AppConfigNotifier.new,
    );
