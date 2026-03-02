import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:graphql/client.dart';

Uri canonicalGraphqlHttpUri(String endpoint) {
  final trimmed = endpoint.trim().split('#').first;
  if (trimmed.isEmpty) {
    return Uri.parse('http://localhost:8080/query');
  }
  final withScheme = trimmed.contains('://') ? trimmed : 'http://$trimmed';
  final parsed = Uri.tryParse(withScheme);
  if (parsed == null) {
    return Uri.parse('http://localhost:8080/query');
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
      .toString()
      .split('#')
      .first;
  return Uri.parse(normalized);
}

GraphQLClient buildGraphqlClient(String httpEndpoint) {
  final httpUri = canonicalGraphqlHttpUri(httpEndpoint);
  final wsScheme = httpUri.scheme == 'https' ? 'wss' : 'ws';
  final wsUri = httpUri.replace(scheme: wsScheme);

  AppLogger.instance.logger.i(
    'Building GraphQL client',
    error: {
      'http_endpoint': httpUri.toString(),
      'ws_endpoint': wsUri.toString(),
    },
  );

  final httpLink = HttpLink(httpUri.toString());
  final wsLink = WebSocketLink(
    wsUri.toString(),
    config: const SocketClientConfig(
      autoReconnect: true,
      inactivityTimeout: Duration(minutes: 2),
    ),
  );
  final link = Link.split(
    (request) => request.isSubscription,
    wsLink,
    httpLink,
  );

  return GraphQLClient(
    cache: GraphQLCache(store: InMemoryStore()),
    link: link,
    defaultPolicies: DefaultPolicies(
      query: Policies(fetch: FetchPolicy.networkOnly),
      mutate: Policies(fetch: FetchPolicy.networkOnly),
      subscribe: Policies(fetch: FetchPolicy.noCache),
    ),
  );
}
