import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:graphql/client.dart';

GraphQLClient buildGraphqlClient(String httpEndpoint) {
  AppLogger.instance.logger.i(
    'Building GraphQL client',
    error: {'http_endpoint': httpEndpoint},
  );
  final sanitizedEndpoint = httpEndpoint.trim().split('#').first;
  final parsed = Uri.parse(sanitizedEndpoint);
  final httpUri = parsed;
  final wsScheme = httpUri.scheme == 'https' ? 'wss' : 'ws';
  final wsUri = httpUri.replace(scheme: wsScheme);
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
