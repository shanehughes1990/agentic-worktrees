import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:graphql/client.dart' as graphql;
import 'package:mockito/annotations.dart';

@GenerateNiceMocks(<MockSpec<dynamic>>[
  MockSpec<ControlPlaneApi>(),
  MockSpec<graphql.GraphQLClient>(),
])
void main() {}
