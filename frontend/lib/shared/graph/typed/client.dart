import 'package:agentic_worktrees/shared/logging/app_logger.dart';
import 'package:graphql/client.dart';

GraphQLClient buildGraphqlClient(String httpEndpoint) {
  AppLogger.instance.logger.i(
    'Building GraphQL client',
    error: {'http_endpoint': httpEndpoint},
  );
  final parsed = Uri.parse(httpEndpoint);
  final wsScheme = parsed.scheme == 'https' ? 'wss' : 'ws';
  final wsUri = parsed.replace(scheme: wsScheme);
  final httpLink = HttpLink(httpEndpoint);
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
