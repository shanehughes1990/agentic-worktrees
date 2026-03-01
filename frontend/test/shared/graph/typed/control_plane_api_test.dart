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
                    '__typename': 'ProjectSetupConfig',
                    'projectID': 'project-1',
                    'projectName': 'Project One',
                    'scmProvider': 'GITHUB',
                    'repositoryURL': 'https://github.com/acme/repo',
                    'trackerProvider': 'GITHUB_ISSUES',
                    'trackerLocation': 'acme/repo',
                    'trackerBoardID': 'board-1',
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
        expect(result.data!.single.scmProvider, 'GITHUB');
      },
    );
  });

  group('enqueueIngestionWorkflow', () {
    test('maps mutation success to queue task id', () async {
      when(
        client.mutate<gql_ops.Mutation$EnqueueIngestionWorkflow>(any),
      ).thenAnswer((Invocation invocation) async {
        final options =
            invocation.positionalArguments.first
                as graphql.MutationOptions<
                  gql_ops.Mutation$EnqueueIngestionWorkflow
                >;
        return graphql.QueryResult<gql_ops.Mutation$EnqueueIngestionWorkflow>(
          options: options,
          source: graphql.QueryResultSource.network,
          data: <String, dynamic>{
            '__typename': 'Mutation',
            'enqueueIngestionWorkflow': <String, dynamic>{
              '__typename': 'EnqueueIngestionWorkflowSuccess',
              'queueTaskID': 'queue-task-1',
              'duplicate': false,
            },
          },
        );
      });

      final result = await api.enqueueIngestionWorkflow(
        runID: 'run-1',
        taskID: 'task-1',
        jobID: 'job-1',
        idempotencyKey: 'idempotency-1',
        prompt: 'prompt',
        projectID: 'project-1',
        workflowID: 'workflow-1',
        source: 'acme/repo',
      );

      expect(result.isSuccess, isTrue);
      expect(result.data, 'queue-task-1');
      verify(
        client.mutate<gql_ops.Mutation$EnqueueIngestionWorkflow>(any),
      ).called(1);
    });

    test('maps mutation graph error to failure', () async {
      when(
        client.mutate<gql_ops.Mutation$EnqueueIngestionWorkflow>(any),
      ).thenAnswer((Invocation invocation) async {
        final options =
            invocation.positionalArguments.first
                as graphql.MutationOptions<
                  gql_ops.Mutation$EnqueueIngestionWorkflow
                >;
        return graphql.QueryResult<gql_ops.Mutation$EnqueueIngestionWorkflow>(
          options: options,
          source: graphql.QueryResultSource.network,
          data: <String, dynamic>{
            '__typename': 'Mutation',
            'enqueueIngestionWorkflow': <String, dynamic>{
              '__typename': 'GraphError',
              'code': 'INTERNAL',
              'message': 'enqueue failed',
              'field': 'enqueueIngestionWorkflow',
            },
          },
        );
      });

      final result = await api.enqueueIngestionWorkflow(
        runID: 'run-1',
        taskID: 'task-1',
        jobID: 'job-1',
        idempotencyKey: 'idempotency-1',
        prompt: 'prompt',
        projectID: 'project-1',
        workflowID: 'workflow-1',
        source: 'acme/repo',
      );

      expect(result.isSuccess, isFalse);
      expect(result.errorMessage, contains('enqueue failed'));
    });
  });
}
