import 'package:agentic_worktrees/shared/graph/generated/operations/control_plane.graphql.dart'
    as gql_ops;
import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:graphql/client.dart' as graphql;
import 'package:mockito/mockito.dart';

import '../../../support/mocks.mocks.dart';

void main() {
  late MockGraphQLClient client;
  late ControlPlaneApi api;

  setUp(() {
    client = MockGraphQLClient();
    api = ControlPlaneApi(client);
  });

  group('sessions', () {
    test('maps SessionsSuccess payload to SessionSummary list', () async {
      when(client.query<gql_ops.Query$Sessions>(any)).thenAnswer((
        Invocation invocation,
      ) async {
        final options =
            invocation.positionalArguments.first
                as graphql.QueryOptions<gql_ops.Query$Sessions>;
        return graphql.QueryResult<gql_ops.Query$Sessions>(
          options: options,
          source: graphql.QueryResultSource.network,
          data: <String, dynamic>{
            '__typename': 'Query',
            'sessions': <String, dynamic>{
              '__typename': 'SessionsSuccess',
              'sessions': <Map<String, dynamic>>[
                <String, dynamic>{
                  '__typename': 'SessionSummary',
                  'runID': 'run-1',
                  'taskCount': 2,
                  'jobCount': 3,
                  'updatedAt': '2026-03-01T12:00:00Z',
                },
              ],
            },
          },
        );
      });

      final result = await api.sessions(limit: 1);

      expect(result.isSuccess, isTrue);
      expect(result.data, isNotNull);
      expect(result.data!.single.runID, 'run-1');
      verify(client.query<gql_ops.Query$Sessions>(any)).called(1);
    });

    test('maps GraphError payload to failure', () async {
      when(client.query<gql_ops.Query$Sessions>(any)).thenAnswer((
        Invocation invocation,
      ) async {
        final options =
            invocation.positionalArguments.first
                as graphql.QueryOptions<gql_ops.Query$Sessions>;
        return graphql.QueryResult<gql_ops.Query$Sessions>(
          options: options,
          source: graphql.QueryResultSource.network,
          data: <String, dynamic>{
            '__typename': 'Query',
            'sessions': <String, dynamic>{
              '__typename': 'GraphError',
              'code': 'INTERNAL',
              'message': 'sessions failed',
              'field': 'sessions',
            },
          },
        );
      });

      final result = await api.sessions(limit: 1);

      expect(result.isSuccess, isFalse);
      expect(result.errorMessage, contains('sessions failed'));
    });
  });

  group('projectSetups', () {
    test(
      'maps ProjectSetupsSuccess payload to ProjectSetupConfig list',
      () async {
        when(client.query<gql_ops.Query$ProjectSetups>(any)).thenAnswer((
          Invocation invocation,
        ) async {
          final options =
              invocation.positionalArguments.first
                  as graphql.QueryOptions<gql_ops.Query$ProjectSetups>;
          return graphql.QueryResult<gql_ops.Query$ProjectSetups>(
            options: options,
            source: graphql.QueryResultSource.network,
            data: <String, dynamic>{
              '__typename': 'Query',
              'projectSetups': <String, dynamic>{
                '__typename': 'ProjectSetupsSuccess',
                'projects': <Map<String, dynamic>>[
                  <String, dynamic>{
                    '__typename': 'ProjectSetup',
                    'projectID': 'project-1',
                    'projectName': 'Project One',
                    'repositories': <Map<String, dynamic>>[
                      <String, dynamic>{
                        '__typename': 'ProjectRepository',
                        'repositoryID': 'repo-1',
                        'scmProvider': 'GITHUB',
                        'repositoryURL': 'https://github.com/acme/repo',
                        'isPrimary': true,
                      },
                    ],
                    'boards': <Map<String, dynamic>>[
                      <String, dynamic>{
                        '__typename': 'ProjectBoard',
                        'boardID': 'board-1',
                        'trackerProvider': 'GITHUB_ISSUES',
                        'taskboardName': 'Acme Repo Board',
                        'appliesToAllRepositories': true,
                        'repositoryIDs': <String>[],
                      },
                    ],
                    'createdAt': '2026-03-01T12:00:00Z',
                    'updatedAt': '2026-03-01T12:00:00Z',
                  },
                ],
              },
            },
          );
        });

        final result = await api.projectSetups(limit: 10);

        expect(result.isSuccess, isTrue);
        expect(result.data, isNotNull);
        expect(result.data!.single.projectID, 'project-1');
        expect(result.data!.single.repositories.single.scmProvider, 'GITHUB');
      },
    );
  });

  group('worker api', () {
    test('maps workerSessions success payload', () async {
      when(client.query<dynamic>(any)).thenAnswer((
        Invocation invocation,
      ) async {
        final options =
            invocation.positionalArguments.first
                as graphql.QueryOptions<dynamic>;
        return graphql.QueryResult<dynamic>(
          options: options,
          source: graphql.QueryResultSource.network,
          data: <String, dynamic>{
            '__typename': 'Query',
            'workerSessions': <String, dynamic>{
              '__typename': 'WorkerSessionsSuccess',
              'sessions': <Map<String, dynamic>>[
                <String, dynamic>{
                  'workerID': 'worker-1',
                  'epoch': 1,
                  'state': 'healthy',
                  'desiredState': 'healthy',
                  'lastHeartbeat': '2026-03-01T12:00:00Z',
                  'leaseExpiresAt': '2026-03-01T12:00:30Z',
                  'rogueReason': null,
                  'updatedAt': '2026-03-01T12:00:00Z',
                },
              ],
            },
          },
        );
      });

      final result = await api.workerSessions(limit: 10);

      expect(result.isSuccess, isTrue);
      expect(result.data, isNotNull);
      expect(result.data!.single.workerID, 'worker-1');
    });

    test('maps workerSettings success payload', () async {
      when(client.query<dynamic>(any)).thenAnswer((
        Invocation invocation,
      ) async {
        final options =
            invocation.positionalArguments.first
                as graphql.QueryOptions<dynamic>;
        return graphql.QueryResult<dynamic>(
          options: options,
          source: graphql.QueryResultSource.network,
          data: <String, dynamic>{
            '__typename': 'Query',
            'workerSettings': <String, dynamic>{
              '__typename': 'WorkerSettingsSuccess',
              'settings': <String, dynamic>{
                'heartbeatIntervalSeconds': 15,
                'responseDeadlineSeconds': 5,
                'staleAfterSeconds': 45,
                'drainTimeoutSeconds': 20,
                'terminateTimeoutSeconds': 10,
                'rogueThreshold': 3,
                'updatedAt': '2026-03-01T12:00:00Z',
              },
            },
          },
        );
      });

      final result = await api.workerSettings();

      expect(result.isSuccess, isTrue);
      expect(result.data, isNotNull);
      expect(result.data!.heartbeatIntervalSeconds, 15);
    });
  });

  group('upsertProjectSetup', () {
    test('rejects unsupported scm provider before mutation', () async {
      final result = await api.upsertProjectSetup(
        projectID: 'project_1',
        projectName: 'Project 1',
        scmProvider: 'UNSUPPORTED',
        repositoryURLs: const <String>['https://example.com/acme/repo'],
        scmToken: 'token',
        trackerProvider: 'GITHUB_ISSUES',
        taskboardName: 'Acme Repo Board',
      );

      expect(result.isSuccess, isFalse);
      expect(result.errorMessage, contains('unsupported scm provider'));
      verifyNever(client.mutate<gql_ops.Mutation$UpsertProjectSetup>(any));
    });
  });
}
