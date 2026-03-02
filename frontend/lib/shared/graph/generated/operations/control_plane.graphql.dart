import '../schema/control_plane.graphql.dart';
import '../schema/schema.graphql.dart';
import '../schema/scm.graphql.dart';
import '../schema/supervisor.graphql.dart';
import 'dart:async';
import 'package:agentic_worktrees/shared/graph/scalars/date_time_scalar.dart';
import 'package:gql/ast.dart';
import 'package:graphql/client.dart' as graphql;

class Variables$Query$Sessions {
  factory Variables$Query$Sessions({required int limit}) =>
      Variables$Query$Sessions._({r'limit': limit});

  Variables$Query$Sessions._(this._$data);

  factory Variables$Query$Sessions.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$limit = data['limit'];
    result$data['limit'] = (l$limit as int);
    return Variables$Query$Sessions._(result$data);
  }

  Map<String, dynamic> _$data;

  int get limit => (_$data['limit'] as int);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$limit = limit;
    result$data['limit'] = l$limit;
    return result$data;
  }

  CopyWith$Variables$Query$Sessions<Variables$Query$Sessions> get copyWith =>
      CopyWith$Variables$Query$Sessions(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Query$Sessions ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$limit = limit;
    final lOther$limit = other.limit;
    if (l$limit != lOther$limit) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$limit = limit;
    return Object.hashAll([l$limit]);
  }
}

abstract class CopyWith$Variables$Query$Sessions<TRes> {
  factory CopyWith$Variables$Query$Sessions(
    Variables$Query$Sessions instance,
    TRes Function(Variables$Query$Sessions) then,
  ) = _CopyWithImpl$Variables$Query$Sessions;

  factory CopyWith$Variables$Query$Sessions.stub(TRes res) =
      _CopyWithStubImpl$Variables$Query$Sessions;

  TRes call({int? limit});
}

class _CopyWithImpl$Variables$Query$Sessions<TRes>
    implements CopyWith$Variables$Query$Sessions<TRes> {
  _CopyWithImpl$Variables$Query$Sessions(this._instance, this._then);

  final Variables$Query$Sessions _instance;

  final TRes Function(Variables$Query$Sessions) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? limit = _undefined}) => _then(
    Variables$Query$Sessions._({
      ..._instance._$data,
      if (limit != _undefined && limit != null) 'limit': (limit as int),
    }),
  );
}

class _CopyWithStubImpl$Variables$Query$Sessions<TRes>
    implements CopyWith$Variables$Query$Sessions<TRes> {
  _CopyWithStubImpl$Variables$Query$Sessions(this._res);

  TRes _res;

  call({int? limit}) => _res;
}

class Query$Sessions {
  Query$Sessions({required this.sessions, this.$__typename = 'Query'});

  factory Query$Sessions.fromJson(Map<String, dynamic> json) {
    final l$sessions = json['sessions'];
    final l$$__typename = json['__typename'];
    return Query$Sessions(
      sessions: Query$Sessions$sessions.fromJson(
        (l$sessions as Map<String, dynamic>),
      ),
      $__typename: (l$$__typename as String),
    );
  }

  final Query$Sessions$sessions sessions;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$sessions = sessions;
    _resultData['sessions'] = l$sessions.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$sessions = sessions;
    final l$$__typename = $__typename;
    return Object.hashAll([l$sessions, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$Sessions || runtimeType != other.runtimeType) {
      return false;
    }
    final l$sessions = sessions;
    final lOther$sessions = other.sessions;
    if (l$sessions != lOther$sessions) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$Sessions on Query$Sessions {
  CopyWith$Query$Sessions<Query$Sessions> get copyWith =>
      CopyWith$Query$Sessions(this, (i) => i);
}

abstract class CopyWith$Query$Sessions<TRes> {
  factory CopyWith$Query$Sessions(
    Query$Sessions instance,
    TRes Function(Query$Sessions) then,
  ) = _CopyWithImpl$Query$Sessions;

  factory CopyWith$Query$Sessions.stub(TRes res) =
      _CopyWithStubImpl$Query$Sessions;

  TRes call({Query$Sessions$sessions? sessions, String? $__typename});
  CopyWith$Query$Sessions$sessions<TRes> get sessions;
}

class _CopyWithImpl$Query$Sessions<TRes>
    implements CopyWith$Query$Sessions<TRes> {
  _CopyWithImpl$Query$Sessions(this._instance, this._then);

  final Query$Sessions _instance;

  final TRes Function(Query$Sessions) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? sessions = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$Sessions(
      sessions: sessions == _undefined || sessions == null
          ? _instance.sessions
          : (sessions as Query$Sessions$sessions),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Query$Sessions$sessions<TRes> get sessions {
    final local$sessions = _instance.sessions;
    return CopyWith$Query$Sessions$sessions(
      local$sessions,
      (e) => call(sessions: e),
    );
  }
}

class _CopyWithStubImpl$Query$Sessions<TRes>
    implements CopyWith$Query$Sessions<TRes> {
  _CopyWithStubImpl$Query$Sessions(this._res);

  TRes _res;

  call({Query$Sessions$sessions? sessions, String? $__typename}) => _res;

  CopyWith$Query$Sessions$sessions<TRes> get sessions =>
      CopyWith$Query$Sessions$sessions.stub(_res);
}

const documentNodeQuerySessions = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.query,
      name: NameNode(value: 'Sessions'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'limit')),
          type: NamedTypeNode(name: NameNode(value: 'Int'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'sessions'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'limit'),
                value: VariableNode(name: NameNode(value: 'limit')),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'SessionsSuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'sessions'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'runID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'taskCount'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'jobCount'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'updatedAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Query$Sessions _parserFn$Query$Sessions(Map<String, dynamic> data) =>
    Query$Sessions.fromJson(data);
typedef OnQueryComplete$Query$Sessions =
    FutureOr<void> Function(Map<String, dynamic>?, Query$Sessions?);

class Options$Query$Sessions extends graphql.QueryOptions<Query$Sessions> {
  Options$Query$Sessions({
    String? operationName,
    required Variables$Query$Sessions variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$Sessions? typedOptimisticResult,
    Duration? pollInterval,
    graphql.Context? context,
    OnQueryComplete$Query$Sessions? onComplete,
    graphql.OnQueryError? onError,
  }) : onCompleteWithParsed = onComplete,
       super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         pollInterval: pollInterval,
         context: context,
         onComplete: onComplete == null
             ? null
             : (data) => onComplete(
                 data,
                 data == null ? null : _parserFn$Query$Sessions(data),
               ),
         onError: onError,
         document: documentNodeQuerySessions,
         parserFn: _parserFn$Query$Sessions,
       );

  final OnQueryComplete$Query$Sessions? onCompleteWithParsed;

  @override
  List<Object?> get properties => [
    ...super.onComplete == null
        ? super.properties
        : super.properties.where((property) => property != onComplete),
    onCompleteWithParsed,
  ];
}

class WatchOptions$Query$Sessions
    extends graphql.WatchQueryOptions<Query$Sessions> {
  WatchOptions$Query$Sessions({
    String? operationName,
    required Variables$Query$Sessions variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$Sessions? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeQuerySessions,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Query$Sessions,
       );
}

class FetchMoreOptions$Query$Sessions extends graphql.FetchMoreOptions {
  FetchMoreOptions$Query$Sessions({
    required graphql.UpdateQuery updateQuery,
    required Variables$Query$Sessions variables,
  }) : super(
         updateQuery: updateQuery,
         variables: variables.toJson(),
         document: documentNodeQuerySessions,
       );
}

extension ClientExtension$Query$Sessions on graphql.GraphQLClient {
  Future<graphql.QueryResult<Query$Sessions>> query$Sessions(
    Options$Query$Sessions options,
  ) async => await this.query(options);

  graphql.ObservableQuery<Query$Sessions> watchQuery$Sessions(
    WatchOptions$Query$Sessions options,
  ) => this.watchQuery(options);

  void writeQuery$Sessions({
    required Query$Sessions data,
    required Variables$Query$Sessions variables,
    bool broadcast = true,
  }) => this.writeQuery(
    graphql.Request(
      operation: graphql.Operation(document: documentNodeQuerySessions),
      variables: variables.toJson(),
    ),
    data: data.toJson(),
    broadcast: broadcast,
  );

  Query$Sessions? readQuery$Sessions({
    required Variables$Query$Sessions variables,
    bool optimistic = true,
  }) {
    final result = this.readQuery(
      graphql.Request(
        operation: graphql.Operation(document: documentNodeQuerySessions),
        variables: variables.toJson(),
      ),
      optimistic: optimistic,
    );
    return result == null ? null : Query$Sessions.fromJson(result);
  }
}

class Query$Sessions$sessions {
  Query$Sessions$sessions({required this.$__typename});

  factory Query$Sessions$sessions.fromJson(Map<String, dynamic> json) {
    switch (json["__typename"] as String) {
      case "SessionsSuccess":
        return Query$Sessions$sessions$$SessionsSuccess.fromJson(json);

      case "GraphError":
        return Query$Sessions$sessions$$GraphError.fromJson(json);

      default:
        final l$$__typename = json['__typename'];
        return Query$Sessions$sessions($__typename: (l$$__typename as String));
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$Sessions$sessions || runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$Sessions$sessions on Query$Sessions$sessions {
  CopyWith$Query$Sessions$sessions<Query$Sessions$sessions> get copyWith =>
      CopyWith$Query$Sessions$sessions(this, (i) => i);

  _T when<_T>({
    required _T Function(Query$Sessions$sessions$$SessionsSuccess)
    sessionsSuccess,
    required _T Function(Query$Sessions$sessions$$GraphError) graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "SessionsSuccess":
        return sessionsSuccess(
          this as Query$Sessions$sessions$$SessionsSuccess,
        );

      case "GraphError":
        return graphError(this as Query$Sessions$sessions$$GraphError);

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(Query$Sessions$sessions$$SessionsSuccess)? sessionsSuccess,
    _T Function(Query$Sessions$sessions$$GraphError)? graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "SessionsSuccess":
        if (sessionsSuccess != null) {
          return sessionsSuccess(
            this as Query$Sessions$sessions$$SessionsSuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(this as Query$Sessions$sessions$$GraphError);
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Query$Sessions$sessions<TRes> {
  factory CopyWith$Query$Sessions$sessions(
    Query$Sessions$sessions instance,
    TRes Function(Query$Sessions$sessions) then,
  ) = _CopyWithImpl$Query$Sessions$sessions;

  factory CopyWith$Query$Sessions$sessions.stub(TRes res) =
      _CopyWithStubImpl$Query$Sessions$sessions;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Query$Sessions$sessions<TRes>
    implements CopyWith$Query$Sessions$sessions<TRes> {
  _CopyWithImpl$Query$Sessions$sessions(this._instance, this._then);

  final Query$Sessions$sessions _instance;

  final TRes Function(Query$Sessions$sessions) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Query$Sessions$sessions(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$Sessions$sessions<TRes>
    implements CopyWith$Query$Sessions$sessions<TRes> {
  _CopyWithStubImpl$Query$Sessions$sessions(this._res);

  TRes _res;

  call({String? $__typename}) => _res;
}

class Query$Sessions$sessions$$SessionsSuccess
    implements Query$Sessions$sessions {
  Query$Sessions$sessions$$SessionsSuccess({
    required this.sessions,
    this.$__typename = 'SessionsSuccess',
  });

  factory Query$Sessions$sessions$$SessionsSuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$sessions = json['sessions'];
    final l$$__typename = json['__typename'];
    return Query$Sessions$sessions$$SessionsSuccess(
      sessions: (l$sessions as List<dynamic>)
          .map(
            (e) => Query$Sessions$sessions$$SessionsSuccess$sessions.fromJson(
              (e as Map<String, dynamic>),
            ),
          )
          .toList(),
      $__typename: (l$$__typename as String),
    );
  }

  final List<Query$Sessions$sessions$$SessionsSuccess$sessions> sessions;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$sessions = sessions;
    _resultData['sessions'] = l$sessions.map((e) => e.toJson()).toList();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$sessions = sessions;
    final l$$__typename = $__typename;
    return Object.hashAll([
      Object.hashAll(l$sessions.map((v) => v)),
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$Sessions$sessions$$SessionsSuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$sessions = sessions;
    final lOther$sessions = other.sessions;
    if (l$sessions.length != lOther$sessions.length) {
      return false;
    }
    for (int i = 0; i < l$sessions.length; i++) {
      final l$sessions$entry = l$sessions[i];
      final lOther$sessions$entry = lOther$sessions[i];
      if (l$sessions$entry != lOther$sessions$entry) {
        return false;
      }
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$Sessions$sessions$$SessionsSuccess
    on Query$Sessions$sessions$$SessionsSuccess {
  CopyWith$Query$Sessions$sessions$$SessionsSuccess<
    Query$Sessions$sessions$$SessionsSuccess
  >
  get copyWith =>
      CopyWith$Query$Sessions$sessions$$SessionsSuccess(this, (i) => i);
}

abstract class CopyWith$Query$Sessions$sessions$$SessionsSuccess<TRes> {
  factory CopyWith$Query$Sessions$sessions$$SessionsSuccess(
    Query$Sessions$sessions$$SessionsSuccess instance,
    TRes Function(Query$Sessions$sessions$$SessionsSuccess) then,
  ) = _CopyWithImpl$Query$Sessions$sessions$$SessionsSuccess;

  factory CopyWith$Query$Sessions$sessions$$SessionsSuccess.stub(TRes res) =
      _CopyWithStubImpl$Query$Sessions$sessions$$SessionsSuccess;

  TRes call({
    List<Query$Sessions$sessions$$SessionsSuccess$sessions>? sessions,
    String? $__typename,
  });
  TRes sessions(
    Iterable<Query$Sessions$sessions$$SessionsSuccess$sessions> Function(
      Iterable<
        CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions<
          Query$Sessions$sessions$$SessionsSuccess$sessions
        >
      >,
    )
    _fn,
  );
}

class _CopyWithImpl$Query$Sessions$sessions$$SessionsSuccess<TRes>
    implements CopyWith$Query$Sessions$sessions$$SessionsSuccess<TRes> {
  _CopyWithImpl$Query$Sessions$sessions$$SessionsSuccess(
    this._instance,
    this._then,
  );

  final Query$Sessions$sessions$$SessionsSuccess _instance;

  final TRes Function(Query$Sessions$sessions$$SessionsSuccess) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? sessions = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$Sessions$sessions$$SessionsSuccess(
      sessions: sessions == _undefined || sessions == null
          ? _instance.sessions
          : (sessions
                as List<Query$Sessions$sessions$$SessionsSuccess$sessions>),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  TRes sessions(
    Iterable<Query$Sessions$sessions$$SessionsSuccess$sessions> Function(
      Iterable<
        CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions<
          Query$Sessions$sessions$$SessionsSuccess$sessions
        >
      >,
    )
    _fn,
  ) => call(
    sessions: _fn(
      _instance.sessions.map(
        (e) => CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions(
          e,
          (i) => i,
        ),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Query$Sessions$sessions$$SessionsSuccess<TRes>
    implements CopyWith$Query$Sessions$sessions$$SessionsSuccess<TRes> {
  _CopyWithStubImpl$Query$Sessions$sessions$$SessionsSuccess(this._res);

  TRes _res;

  call({
    List<Query$Sessions$sessions$$SessionsSuccess$sessions>? sessions,
    String? $__typename,
  }) => _res;

  sessions(_fn) => _res;
}

class Query$Sessions$sessions$$SessionsSuccess$sessions {
  Query$Sessions$sessions$$SessionsSuccess$sessions({
    required this.runID,
    required this.taskCount,
    required this.jobCount,
    required this.updatedAt,
    this.$__typename = 'SessionSummary',
  });

  factory Query$Sessions$sessions$$SessionsSuccess$sessions.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$runID = json['runID'];
    final l$taskCount = json['taskCount'];
    final l$jobCount = json['jobCount'];
    final l$updatedAt = json['updatedAt'];
    final l$$__typename = json['__typename'];
    return Query$Sessions$sessions$$SessionsSuccess$sessions(
      runID: (l$runID as String),
      taskCount: (l$taskCount as int),
      jobCount: (l$jobCount as int),
      updatedAt: dateTimeFromJson(l$updatedAt),
      $__typename: (l$$__typename as String),
    );
  }

  final String runID;

  final int taskCount;

  final int jobCount;

  final DateTime updatedAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$runID = runID;
    _resultData['runID'] = l$runID;
    final l$taskCount = taskCount;
    _resultData['taskCount'] = l$taskCount;
    final l$jobCount = jobCount;
    _resultData['jobCount'] = l$jobCount;
    final l$updatedAt = updatedAt;
    _resultData['updatedAt'] = dateTimeToJson(l$updatedAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskCount = taskCount;
    final l$jobCount = jobCount;
    final l$updatedAt = updatedAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$runID,
      l$taskCount,
      l$jobCount,
      l$updatedAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$Sessions$sessions$$SessionsSuccess$sessions ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$runID = runID;
    final lOther$runID = other.runID;
    if (l$runID != lOther$runID) {
      return false;
    }
    final l$taskCount = taskCount;
    final lOther$taskCount = other.taskCount;
    if (l$taskCount != lOther$taskCount) {
      return false;
    }
    final l$jobCount = jobCount;
    final lOther$jobCount = other.jobCount;
    if (l$jobCount != lOther$jobCount) {
      return false;
    }
    final l$updatedAt = updatedAt;
    final lOther$updatedAt = other.updatedAt;
    if (l$updatedAt != lOther$updatedAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$Sessions$sessions$$SessionsSuccess$sessions
    on Query$Sessions$sessions$$SessionsSuccess$sessions {
  CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions<
    Query$Sessions$sessions$$SessionsSuccess$sessions
  >
  get copyWith => CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions(
    this,
    (i) => i,
  );
}

abstract class CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions<
  TRes
> {
  factory CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions(
    Query$Sessions$sessions$$SessionsSuccess$sessions instance,
    TRes Function(Query$Sessions$sessions$$SessionsSuccess$sessions) then,
  ) = _CopyWithImpl$Query$Sessions$sessions$$SessionsSuccess$sessions;

  factory CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$Sessions$sessions$$SessionsSuccess$sessions;

  TRes call({
    String? runID,
    int? taskCount,
    int? jobCount,
    DateTime? updatedAt,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$Sessions$sessions$$SessionsSuccess$sessions<TRes>
    implements
        CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions<TRes> {
  _CopyWithImpl$Query$Sessions$sessions$$SessionsSuccess$sessions(
    this._instance,
    this._then,
  );

  final Query$Sessions$sessions$$SessionsSuccess$sessions _instance;

  final TRes Function(Query$Sessions$sessions$$SessionsSuccess$sessions) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskCount = _undefined,
    Object? jobCount = _undefined,
    Object? updatedAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$Sessions$sessions$$SessionsSuccess$sessions(
      runID: runID == _undefined || runID == null
          ? _instance.runID
          : (runID as String),
      taskCount: taskCount == _undefined || taskCount == null
          ? _instance.taskCount
          : (taskCount as int),
      jobCount: jobCount == _undefined || jobCount == null
          ? _instance.jobCount
          : (jobCount as int),
      updatedAt: updatedAt == _undefined || updatedAt == null
          ? _instance.updatedAt
          : (updatedAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$Sessions$sessions$$SessionsSuccess$sessions<TRes>
    implements
        CopyWith$Query$Sessions$sessions$$SessionsSuccess$sessions<TRes> {
  _CopyWithStubImpl$Query$Sessions$sessions$$SessionsSuccess$sessions(
    this._res,
  );

  TRes _res;

  call({
    String? runID,
    int? taskCount,
    int? jobCount,
    DateTime? updatedAt,
    String? $__typename,
  }) => _res;
}

class Query$Sessions$sessions$$GraphError implements Query$Sessions$sessions {
  Query$Sessions$sessions$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Query$Sessions$sessions$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Query$Sessions$sessions$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$Sessions$sessions$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$Sessions$sessions$$GraphError
    on Query$Sessions$sessions$$GraphError {
  CopyWith$Query$Sessions$sessions$$GraphError<
    Query$Sessions$sessions$$GraphError
  >
  get copyWith => CopyWith$Query$Sessions$sessions$$GraphError(this, (i) => i);
}

abstract class CopyWith$Query$Sessions$sessions$$GraphError<TRes> {
  factory CopyWith$Query$Sessions$sessions$$GraphError(
    Query$Sessions$sessions$$GraphError instance,
    TRes Function(Query$Sessions$sessions$$GraphError) then,
  ) = _CopyWithImpl$Query$Sessions$sessions$$GraphError;

  factory CopyWith$Query$Sessions$sessions$$GraphError.stub(TRes res) =
      _CopyWithStubImpl$Query$Sessions$sessions$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$Sessions$sessions$$GraphError<TRes>
    implements CopyWith$Query$Sessions$sessions$$GraphError<TRes> {
  _CopyWithImpl$Query$Sessions$sessions$$GraphError(this._instance, this._then);

  final Query$Sessions$sessions$$GraphError _instance;

  final TRes Function(Query$Sessions$sessions$$GraphError) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$Sessions$sessions$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$Sessions$sessions$$GraphError<TRes>
    implements CopyWith$Query$Sessions$sessions$$GraphError<TRes> {
  _CopyWithStubImpl$Query$Sessions$sessions$$GraphError(this._res);

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}

class Variables$Query$WorkflowJobs {
  factory Variables$Query$WorkflowJobs({
    required String runID,
    String? taskID,
    required int limit,
  }) => Variables$Query$WorkflowJobs._({
    r'runID': runID,
    if (taskID != null) r'taskID': taskID,
    r'limit': limit,
  });

  Variables$Query$WorkflowJobs._(this._$data);

  factory Variables$Query$WorkflowJobs.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    if (data.containsKey('taskID')) {
      final l$taskID = data['taskID'];
      result$data['taskID'] = (l$taskID as String?);
    }
    final l$limit = data['limit'];
    result$data['limit'] = (l$limit as int);
    return Variables$Query$WorkflowJobs._(result$data);
  }

  Map<String, dynamic> _$data;

  String get runID => (_$data['runID'] as String);

  String? get taskID => (_$data['taskID'] as String?);

  int get limit => (_$data['limit'] as int);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$runID = runID;
    result$data['runID'] = l$runID;
    if (_$data.containsKey('taskID')) {
      final l$taskID = taskID;
      result$data['taskID'] = l$taskID;
    }
    final l$limit = limit;
    result$data['limit'] = l$limit;
    return result$data;
  }

  CopyWith$Variables$Query$WorkflowJobs<Variables$Query$WorkflowJobs>
  get copyWith => CopyWith$Variables$Query$WorkflowJobs(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Query$WorkflowJobs ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$runID = runID;
    final lOther$runID = other.runID;
    if (l$runID != lOther$runID) {
      return false;
    }
    final l$taskID = taskID;
    final lOther$taskID = other.taskID;
    if (_$data.containsKey('taskID') != other._$data.containsKey('taskID')) {
      return false;
    }
    if (l$taskID != lOther$taskID) {
      return false;
    }
    final l$limit = limit;
    final lOther$limit = other.limit;
    if (l$limit != lOther$limit) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$limit = limit;
    return Object.hashAll([
      l$runID,
      _$data.containsKey('taskID') ? l$taskID : const {},
      l$limit,
    ]);
  }
}

abstract class CopyWith$Variables$Query$WorkflowJobs<TRes> {
  factory CopyWith$Variables$Query$WorkflowJobs(
    Variables$Query$WorkflowJobs instance,
    TRes Function(Variables$Query$WorkflowJobs) then,
  ) = _CopyWithImpl$Variables$Query$WorkflowJobs;

  factory CopyWith$Variables$Query$WorkflowJobs.stub(TRes res) =
      _CopyWithStubImpl$Variables$Query$WorkflowJobs;

  TRes call({String? runID, String? taskID, int? limit});
}

class _CopyWithImpl$Variables$Query$WorkflowJobs<TRes>
    implements CopyWith$Variables$Query$WorkflowJobs<TRes> {
  _CopyWithImpl$Variables$Query$WorkflowJobs(this._instance, this._then);

  final Variables$Query$WorkflowJobs _instance;

  final TRes Function(Variables$Query$WorkflowJobs) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? limit = _undefined,
  }) => _then(
    Variables$Query$WorkflowJobs._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined) 'taskID': (taskID as String?),
      if (limit != _undefined && limit != null) 'limit': (limit as int),
    }),
  );
}

class _CopyWithStubImpl$Variables$Query$WorkflowJobs<TRes>
    implements CopyWith$Variables$Query$WorkflowJobs<TRes> {
  _CopyWithStubImpl$Variables$Query$WorkflowJobs(this._res);

  TRes _res;

  call({String? runID, String? taskID, int? limit}) => _res;
}

class Query$WorkflowJobs {
  Query$WorkflowJobs({required this.workflowJobs, this.$__typename = 'Query'});

  factory Query$WorkflowJobs.fromJson(Map<String, dynamic> json) {
    final l$workflowJobs = json['workflowJobs'];
    final l$$__typename = json['__typename'];
    return Query$WorkflowJobs(
      workflowJobs: Query$WorkflowJobs$workflowJobs.fromJson(
        (l$workflowJobs as Map<String, dynamic>),
      ),
      $__typename: (l$$__typename as String),
    );
  }

  final Query$WorkflowJobs$workflowJobs workflowJobs;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$workflowJobs = workflowJobs;
    _resultData['workflowJobs'] = l$workflowJobs.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$workflowJobs = workflowJobs;
    final l$$__typename = $__typename;
    return Object.hashAll([l$workflowJobs, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$WorkflowJobs || runtimeType != other.runtimeType) {
      return false;
    }
    final l$workflowJobs = workflowJobs;
    final lOther$workflowJobs = other.workflowJobs;
    if (l$workflowJobs != lOther$workflowJobs) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$WorkflowJobs on Query$WorkflowJobs {
  CopyWith$Query$WorkflowJobs<Query$WorkflowJobs> get copyWith =>
      CopyWith$Query$WorkflowJobs(this, (i) => i);
}

abstract class CopyWith$Query$WorkflowJobs<TRes> {
  factory CopyWith$Query$WorkflowJobs(
    Query$WorkflowJobs instance,
    TRes Function(Query$WorkflowJobs) then,
  ) = _CopyWithImpl$Query$WorkflowJobs;

  factory CopyWith$Query$WorkflowJobs.stub(TRes res) =
      _CopyWithStubImpl$Query$WorkflowJobs;

  TRes call({
    Query$WorkflowJobs$workflowJobs? workflowJobs,
    String? $__typename,
  });
  CopyWith$Query$WorkflowJobs$workflowJobs<TRes> get workflowJobs;
}

class _CopyWithImpl$Query$WorkflowJobs<TRes>
    implements CopyWith$Query$WorkflowJobs<TRes> {
  _CopyWithImpl$Query$WorkflowJobs(this._instance, this._then);

  final Query$WorkflowJobs _instance;

  final TRes Function(Query$WorkflowJobs) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? workflowJobs = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$WorkflowJobs(
      workflowJobs: workflowJobs == _undefined || workflowJobs == null
          ? _instance.workflowJobs
          : (workflowJobs as Query$WorkflowJobs$workflowJobs),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Query$WorkflowJobs$workflowJobs<TRes> get workflowJobs {
    final local$workflowJobs = _instance.workflowJobs;
    return CopyWith$Query$WorkflowJobs$workflowJobs(
      local$workflowJobs,
      (e) => call(workflowJobs: e),
    );
  }
}

class _CopyWithStubImpl$Query$WorkflowJobs<TRes>
    implements CopyWith$Query$WorkflowJobs<TRes> {
  _CopyWithStubImpl$Query$WorkflowJobs(this._res);

  TRes _res;

  call({Query$WorkflowJobs$workflowJobs? workflowJobs, String? $__typename}) =>
      _res;

  CopyWith$Query$WorkflowJobs$workflowJobs<TRes> get workflowJobs =>
      CopyWith$Query$WorkflowJobs$workflowJobs.stub(_res);
}

const documentNodeQueryWorkflowJobs = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.query,
      name: NameNode(value: 'WorkflowJobs'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'runID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'taskID')),
          type: NamedTypeNode(
            name: NameNode(value: 'String'),
            isNonNull: false,
          ),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'limit')),
          type: NamedTypeNode(name: NameNode(value: 'Int'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'workflowJobs'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'runID'),
                value: VariableNode(name: NameNode(value: 'runID')),
              ),
              ArgumentNode(
                name: NameNode(value: 'taskID'),
                value: VariableNode(name: NameNode(value: 'taskID')),
              ),
              ArgumentNode(
                name: NameNode(value: 'limit'),
                value: VariableNode(name: NameNode(value: 'limit')),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'WorkflowJobsSuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'jobs'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'runID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'taskID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'jobID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'jobKind'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'status'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'queue'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'queueTaskID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'duplicate'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'updatedAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Query$WorkflowJobs _parserFn$Query$WorkflowJobs(Map<String, dynamic> data) =>
    Query$WorkflowJobs.fromJson(data);
typedef OnQueryComplete$Query$WorkflowJobs =
    FutureOr<void> Function(Map<String, dynamic>?, Query$WorkflowJobs?);

class Options$Query$WorkflowJobs
    extends graphql.QueryOptions<Query$WorkflowJobs> {
  Options$Query$WorkflowJobs({
    String? operationName,
    required Variables$Query$WorkflowJobs variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$WorkflowJobs? typedOptimisticResult,
    Duration? pollInterval,
    graphql.Context? context,
    OnQueryComplete$Query$WorkflowJobs? onComplete,
    graphql.OnQueryError? onError,
  }) : onCompleteWithParsed = onComplete,
       super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         pollInterval: pollInterval,
         context: context,
         onComplete: onComplete == null
             ? null
             : (data) => onComplete(
                 data,
                 data == null ? null : _parserFn$Query$WorkflowJobs(data),
               ),
         onError: onError,
         document: documentNodeQueryWorkflowJobs,
         parserFn: _parserFn$Query$WorkflowJobs,
       );

  final OnQueryComplete$Query$WorkflowJobs? onCompleteWithParsed;

  @override
  List<Object?> get properties => [
    ...super.onComplete == null
        ? super.properties
        : super.properties.where((property) => property != onComplete),
    onCompleteWithParsed,
  ];
}

class WatchOptions$Query$WorkflowJobs
    extends graphql.WatchQueryOptions<Query$WorkflowJobs> {
  WatchOptions$Query$WorkflowJobs({
    String? operationName,
    required Variables$Query$WorkflowJobs variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$WorkflowJobs? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeQueryWorkflowJobs,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Query$WorkflowJobs,
       );
}

class FetchMoreOptions$Query$WorkflowJobs extends graphql.FetchMoreOptions {
  FetchMoreOptions$Query$WorkflowJobs({
    required graphql.UpdateQuery updateQuery,
    required Variables$Query$WorkflowJobs variables,
  }) : super(
         updateQuery: updateQuery,
         variables: variables.toJson(),
         document: documentNodeQueryWorkflowJobs,
       );
}

extension ClientExtension$Query$WorkflowJobs on graphql.GraphQLClient {
  Future<graphql.QueryResult<Query$WorkflowJobs>> query$WorkflowJobs(
    Options$Query$WorkflowJobs options,
  ) async => await this.query(options);

  graphql.ObservableQuery<Query$WorkflowJobs> watchQuery$WorkflowJobs(
    WatchOptions$Query$WorkflowJobs options,
  ) => this.watchQuery(options);

  void writeQuery$WorkflowJobs({
    required Query$WorkflowJobs data,
    required Variables$Query$WorkflowJobs variables,
    bool broadcast = true,
  }) => this.writeQuery(
    graphql.Request(
      operation: graphql.Operation(document: documentNodeQueryWorkflowJobs),
      variables: variables.toJson(),
    ),
    data: data.toJson(),
    broadcast: broadcast,
  );

  Query$WorkflowJobs? readQuery$WorkflowJobs({
    required Variables$Query$WorkflowJobs variables,
    bool optimistic = true,
  }) {
    final result = this.readQuery(
      graphql.Request(
        operation: graphql.Operation(document: documentNodeQueryWorkflowJobs),
        variables: variables.toJson(),
      ),
      optimistic: optimistic,
    );
    return result == null ? null : Query$WorkflowJobs.fromJson(result);
  }
}

class Query$WorkflowJobs$workflowJobs {
  Query$WorkflowJobs$workflowJobs({required this.$__typename});

  factory Query$WorkflowJobs$workflowJobs.fromJson(Map<String, dynamic> json) {
    switch (json["__typename"] as String) {
      case "WorkflowJobsSuccess":
        return Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess.fromJson(
          json,
        );

      case "GraphError":
        return Query$WorkflowJobs$workflowJobs$$GraphError.fromJson(json);

      default:
        final l$$__typename = json['__typename'];
        return Query$WorkflowJobs$workflowJobs(
          $__typename: (l$$__typename as String),
        );
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$WorkflowJobs$workflowJobs ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$WorkflowJobs$workflowJobs
    on Query$WorkflowJobs$workflowJobs {
  CopyWith$Query$WorkflowJobs$workflowJobs<Query$WorkflowJobs$workflowJobs>
  get copyWith => CopyWith$Query$WorkflowJobs$workflowJobs(this, (i) => i);

  _T when<_T>({
    required _T Function(Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess)
    workflowJobsSuccess,
    required _T Function(Query$WorkflowJobs$workflowJobs$$GraphError)
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "WorkflowJobsSuccess":
        return workflowJobsSuccess(
          this as Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess,
        );

      case "GraphError":
        return graphError(this as Query$WorkflowJobs$workflowJobs$$GraphError);

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess)?
    workflowJobsSuccess,
    _T Function(Query$WorkflowJobs$workflowJobs$$GraphError)? graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "WorkflowJobsSuccess":
        if (workflowJobsSuccess != null) {
          return workflowJobsSuccess(
            this as Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(
            this as Query$WorkflowJobs$workflowJobs$$GraphError,
          );
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Query$WorkflowJobs$workflowJobs<TRes> {
  factory CopyWith$Query$WorkflowJobs$workflowJobs(
    Query$WorkflowJobs$workflowJobs instance,
    TRes Function(Query$WorkflowJobs$workflowJobs) then,
  ) = _CopyWithImpl$Query$WorkflowJobs$workflowJobs;

  factory CopyWith$Query$WorkflowJobs$workflowJobs.stub(TRes res) =
      _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Query$WorkflowJobs$workflowJobs<TRes>
    implements CopyWith$Query$WorkflowJobs$workflowJobs<TRes> {
  _CopyWithImpl$Query$WorkflowJobs$workflowJobs(this._instance, this._then);

  final Query$WorkflowJobs$workflowJobs _instance;

  final TRes Function(Query$WorkflowJobs$workflowJobs) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Query$WorkflowJobs$workflowJobs(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs<TRes>
    implements CopyWith$Query$WorkflowJobs$workflowJobs<TRes> {
  _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs(this._res);

  TRes _res;

  call({String? $__typename}) => _res;
}

class Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess
    implements Query$WorkflowJobs$workflowJobs {
  Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess({
    required this.jobs,
    this.$__typename = 'WorkflowJobsSuccess',
  });

  factory Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$jobs = json['jobs'];
    final l$$__typename = json['__typename'];
    return Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess(
      jobs: (l$jobs as List<dynamic>)
          .map(
            (e) =>
                Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      $__typename: (l$$__typename as String),
    );
  }

  final List<Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs> jobs;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$jobs = jobs;
    _resultData['jobs'] = l$jobs.map((e) => e.toJson()).toList();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$jobs = jobs;
    final l$$__typename = $__typename;
    return Object.hashAll([
      Object.hashAll(l$jobs.map((v) => v)),
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$jobs = jobs;
    final lOther$jobs = other.jobs;
    if (l$jobs.length != lOther$jobs.length) {
      return false;
    }
    for (int i = 0; i < l$jobs.length; i++) {
      final l$jobs$entry = l$jobs[i];
      final lOther$jobs$entry = lOther$jobs[i];
      if (l$jobs$entry != lOther$jobs$entry) {
        return false;
      }
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess
    on Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess {
  CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess<
    Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess
  >
  get copyWith => CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess(
    this,
    (i) => i,
  );
}

abstract class CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess<
  TRes
> {
  factory CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess(
    Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess instance,
    TRes Function(Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess) then,
  ) = _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess;

  factory CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess;

  TRes call({
    List<Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs>? jobs,
    String? $__typename,
  });
  TRes jobs(
    Iterable<Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs>
    Function(
      Iterable<
        CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
          Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs
        >
      >,
    )
    _fn,
  );
}

class _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess<TRes>
    implements
        CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess<TRes> {
  _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess(
    this._instance,
    this._then,
  );

  final Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess _instance;

  final TRes Function(Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess)
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? jobs = _undefined, Object? $__typename = _undefined}) =>
      _then(
        Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess(
          jobs: jobs == _undefined || jobs == null
              ? _instance.jobs
              : (jobs
                    as List<
                      Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs
                    >),
          $__typename: $__typename == _undefined || $__typename == null
              ? _instance.$__typename
              : ($__typename as String),
        ),
      );

  TRes jobs(
    Iterable<Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs>
    Function(
      Iterable<
        CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
          Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs
        >
      >,
    )
    _fn,
  ) => call(
    jobs: _fn(
      _instance.jobs.map(
        (e) =>
            CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess<
  TRes
>
    implements
        CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess<TRes> {
  _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess(
    this._res,
  );

  TRes _res;

  call({
    List<Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs>? jobs,
    String? $__typename,
  }) => _res;

  jobs(_fn) => _res;
}

class Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs {
  Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs({
    required this.runID,
    required this.taskID,
    required this.jobID,
    required this.jobKind,
    required this.status,
    required this.queue,
    required this.queueTaskID,
    required this.duplicate,
    required this.updatedAt,
    this.$__typename = 'WorkflowJob',
  });

  factory Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$runID = json['runID'];
    final l$taskID = json['taskID'];
    final l$jobID = json['jobID'];
    final l$jobKind = json['jobKind'];
    final l$status = json['status'];
    final l$queue = json['queue'];
    final l$queueTaskID = json['queueTaskID'];
    final l$duplicate = json['duplicate'];
    final l$updatedAt = json['updatedAt'];
    final l$$__typename = json['__typename'];
    return Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
      runID: (l$runID as String),
      taskID: (l$taskID as String),
      jobID: (l$jobID as String),
      jobKind: fromJson$Enum$JobKind((l$jobKind as String)),
      status: (l$status as String),
      queue: (l$queue as String),
      queueTaskID: (l$queueTaskID as String),
      duplicate: (l$duplicate as bool),
      updatedAt: dateTimeFromJson(l$updatedAt),
      $__typename: (l$$__typename as String),
    );
  }

  final String runID;

  final String taskID;

  final String jobID;

  final Enum$JobKind jobKind;

  final String status;

  final String queue;

  final String queueTaskID;

  final bool duplicate;

  final DateTime updatedAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$runID = runID;
    _resultData['runID'] = l$runID;
    final l$taskID = taskID;
    _resultData['taskID'] = l$taskID;
    final l$jobID = jobID;
    _resultData['jobID'] = l$jobID;
    final l$jobKind = jobKind;
    _resultData['jobKind'] = toJson$Enum$JobKind(l$jobKind);
    final l$status = status;
    _resultData['status'] = l$status;
    final l$queue = queue;
    _resultData['queue'] = l$queue;
    final l$queueTaskID = queueTaskID;
    _resultData['queueTaskID'] = l$queueTaskID;
    final l$duplicate = duplicate;
    _resultData['duplicate'] = l$duplicate;
    final l$updatedAt = updatedAt;
    _resultData['updatedAt'] = dateTimeToJson(l$updatedAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    final l$jobKind = jobKind;
    final l$status = status;
    final l$queue = queue;
    final l$queueTaskID = queueTaskID;
    final l$duplicate = duplicate;
    final l$updatedAt = updatedAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$runID,
      l$taskID,
      l$jobID,
      l$jobKind,
      l$status,
      l$queue,
      l$queueTaskID,
      l$duplicate,
      l$updatedAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$runID = runID;
    final lOther$runID = other.runID;
    if (l$runID != lOther$runID) {
      return false;
    }
    final l$taskID = taskID;
    final lOther$taskID = other.taskID;
    if (l$taskID != lOther$taskID) {
      return false;
    }
    final l$jobID = jobID;
    final lOther$jobID = other.jobID;
    if (l$jobID != lOther$jobID) {
      return false;
    }
    final l$jobKind = jobKind;
    final lOther$jobKind = other.jobKind;
    if (l$jobKind != lOther$jobKind) {
      return false;
    }
    final l$status = status;
    final lOther$status = other.status;
    if (l$status != lOther$status) {
      return false;
    }
    final l$queue = queue;
    final lOther$queue = other.queue;
    if (l$queue != lOther$queue) {
      return false;
    }
    final l$queueTaskID = queueTaskID;
    final lOther$queueTaskID = other.queueTaskID;
    if (l$queueTaskID != lOther$queueTaskID) {
      return false;
    }
    final l$duplicate = duplicate;
    final lOther$duplicate = other.duplicate;
    if (l$duplicate != lOther$duplicate) {
      return false;
    }
    final l$updatedAt = updatedAt;
    final lOther$updatedAt = other.updatedAt;
    if (l$updatedAt != lOther$updatedAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs
    on Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs {
  CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
    Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs
  >
  get copyWith =>
      CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
  TRes
> {
  factory CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
    Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs instance,
    TRes Function(Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs)
    then,
  ) = _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs;

  factory CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs;

  TRes call({
    String? runID,
    String? taskID,
    String? jobID,
    Enum$JobKind? jobKind,
    String? status,
    String? queue,
    String? queueTaskID,
    bool? duplicate,
    DateTime? updatedAt,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
  TRes
>
    implements
        CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
          TRes
        > {
  _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
    this._instance,
    this._then,
  );

  final Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs _instance;

  final TRes Function(Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs)
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
    Object? jobKind = _undefined,
    Object? status = _undefined,
    Object? queue = _undefined,
    Object? queueTaskID = _undefined,
    Object? duplicate = _undefined,
    Object? updatedAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
      runID: runID == _undefined || runID == null
          ? _instance.runID
          : (runID as String),
      taskID: taskID == _undefined || taskID == null
          ? _instance.taskID
          : (taskID as String),
      jobID: jobID == _undefined || jobID == null
          ? _instance.jobID
          : (jobID as String),
      jobKind: jobKind == _undefined || jobKind == null
          ? _instance.jobKind
          : (jobKind as Enum$JobKind),
      status: status == _undefined || status == null
          ? _instance.status
          : (status as String),
      queue: queue == _undefined || queue == null
          ? _instance.queue
          : (queue as String),
      queueTaskID: queueTaskID == _undefined || queueTaskID == null
          ? _instance.queueTaskID
          : (queueTaskID as String),
      duplicate: duplicate == _undefined || duplicate == null
          ? _instance.duplicate
          : (duplicate as bool),
      updatedAt: updatedAt == _undefined || updatedAt == null
          ? _instance.updatedAt
          : (updatedAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
  TRes
>
    implements
        CopyWith$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs<
          TRes
        > {
  _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$WorkflowJobsSuccess$jobs(
    this._res,
  );

  TRes _res;

  call({
    String? runID,
    String? taskID,
    String? jobID,
    Enum$JobKind? jobKind,
    String? status,
    String? queue,
    String? queueTaskID,
    bool? duplicate,
    DateTime? updatedAt,
    String? $__typename,
  }) => _res;
}

class Query$WorkflowJobs$workflowJobs$$GraphError
    implements Query$WorkflowJobs$workflowJobs {
  Query$WorkflowJobs$workflowJobs$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Query$WorkflowJobs$workflowJobs$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Query$WorkflowJobs$workflowJobs$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$WorkflowJobs$workflowJobs$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$WorkflowJobs$workflowJobs$$GraphError
    on Query$WorkflowJobs$workflowJobs$$GraphError {
  CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError<
    Query$WorkflowJobs$workflowJobs$$GraphError
  >
  get copyWith =>
      CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError(this, (i) => i);
}

abstract class CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError<TRes> {
  factory CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError(
    Query$WorkflowJobs$workflowJobs$$GraphError instance,
    TRes Function(Query$WorkflowJobs$workflowJobs$$GraphError) then,
  ) = _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$GraphError;

  factory CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError.stub(TRes res) =
      _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$GraphError<TRes>
    implements CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError<TRes> {
  _CopyWithImpl$Query$WorkflowJobs$workflowJobs$$GraphError(
    this._instance,
    this._then,
  );

  final Query$WorkflowJobs$workflowJobs$$GraphError _instance;

  final TRes Function(Query$WorkflowJobs$workflowJobs$$GraphError) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$WorkflowJobs$workflowJobs$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$GraphError<TRes>
    implements CopyWith$Query$WorkflowJobs$workflowJobs$$GraphError<TRes> {
  _CopyWithStubImpl$Query$WorkflowJobs$workflowJobs$$GraphError(this._res);

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}

class Variables$Query$SupervisorDecisionHistory {
  factory Variables$Query$SupervisorDecisionHistory({
    required String runID,
    required String taskID,
    required String jobID,
  }) => Variables$Query$SupervisorDecisionHistory._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
  });

  Variables$Query$SupervisorDecisionHistory._(this._$data);

  factory Variables$Query$SupervisorDecisionHistory.fromJson(
    Map<String, dynamic> data,
  ) {
    final result$data = <String, dynamic>{};
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    final l$jobID = data['jobID'];
    result$data['jobID'] = (l$jobID as String);
    return Variables$Query$SupervisorDecisionHistory._(result$data);
  }

  Map<String, dynamic> _$data;

  String get runID => (_$data['runID'] as String);

  String get taskID => (_$data['taskID'] as String);

  String get jobID => (_$data['jobID'] as String);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$runID = runID;
    result$data['runID'] = l$runID;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    final l$jobID = jobID;
    result$data['jobID'] = l$jobID;
    return result$data;
  }

  CopyWith$Variables$Query$SupervisorDecisionHistory<
    Variables$Query$SupervisorDecisionHistory
  >
  get copyWith =>
      CopyWith$Variables$Query$SupervisorDecisionHistory(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Query$SupervisorDecisionHistory ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$runID = runID;
    final lOther$runID = other.runID;
    if (l$runID != lOther$runID) {
      return false;
    }
    final l$taskID = taskID;
    final lOther$taskID = other.taskID;
    if (l$taskID != lOther$taskID) {
      return false;
    }
    final l$jobID = jobID;
    final lOther$jobID = other.jobID;
    if (l$jobID != lOther$jobID) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    return Object.hashAll([l$runID, l$taskID, l$jobID]);
  }
}

abstract class CopyWith$Variables$Query$SupervisorDecisionHistory<TRes> {
  factory CopyWith$Variables$Query$SupervisorDecisionHistory(
    Variables$Query$SupervisorDecisionHistory instance,
    TRes Function(Variables$Query$SupervisorDecisionHistory) then,
  ) = _CopyWithImpl$Variables$Query$SupervisorDecisionHistory;

  factory CopyWith$Variables$Query$SupervisorDecisionHistory.stub(TRes res) =
      _CopyWithStubImpl$Variables$Query$SupervisorDecisionHistory;

  TRes call({String? runID, String? taskID, String? jobID});
}

class _CopyWithImpl$Variables$Query$SupervisorDecisionHistory<TRes>
    implements CopyWith$Variables$Query$SupervisorDecisionHistory<TRes> {
  _CopyWithImpl$Variables$Query$SupervisorDecisionHistory(
    this._instance,
    this._then,
  );

  final Variables$Query$SupervisorDecisionHistory _instance;

  final TRes Function(Variables$Query$SupervisorDecisionHistory) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
  }) => _then(
    Variables$Query$SupervisorDecisionHistory._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
    }),
  );
}

class _CopyWithStubImpl$Variables$Query$SupervisorDecisionHistory<TRes>
    implements CopyWith$Variables$Query$SupervisorDecisionHistory<TRes> {
  _CopyWithStubImpl$Variables$Query$SupervisorDecisionHistory(this._res);

  TRes _res;

  call({String? runID, String? taskID, String? jobID}) => _res;
}

class Query$SupervisorDecisionHistory {
  Query$SupervisorDecisionHistory({
    required this.supervisorDecisionHistory,
    this.$__typename = 'Query',
  });

  factory Query$SupervisorDecisionHistory.fromJson(Map<String, dynamic> json) {
    final l$supervisorDecisionHistory = json['supervisorDecisionHistory'];
    final l$$__typename = json['__typename'];
    return Query$SupervisorDecisionHistory(
      supervisorDecisionHistory:
          Query$SupervisorDecisionHistory$supervisorDecisionHistory.fromJson(
            (l$supervisorDecisionHistory as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Query$SupervisorDecisionHistory$supervisorDecisionHistory
  supervisorDecisionHistory;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$supervisorDecisionHistory = supervisorDecisionHistory;
    _resultData['supervisorDecisionHistory'] = l$supervisorDecisionHistory
        .toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$supervisorDecisionHistory = supervisorDecisionHistory;
    final l$$__typename = $__typename;
    return Object.hashAll([l$supervisorDecisionHistory, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$SupervisorDecisionHistory ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$supervisorDecisionHistory = supervisorDecisionHistory;
    final lOther$supervisorDecisionHistory = other.supervisorDecisionHistory;
    if (l$supervisorDecisionHistory != lOther$supervisorDecisionHistory) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$SupervisorDecisionHistory
    on Query$SupervisorDecisionHistory {
  CopyWith$Query$SupervisorDecisionHistory<Query$SupervisorDecisionHistory>
  get copyWith => CopyWith$Query$SupervisorDecisionHistory(this, (i) => i);
}

abstract class CopyWith$Query$SupervisorDecisionHistory<TRes> {
  factory CopyWith$Query$SupervisorDecisionHistory(
    Query$SupervisorDecisionHistory instance,
    TRes Function(Query$SupervisorDecisionHistory) then,
  ) = _CopyWithImpl$Query$SupervisorDecisionHistory;

  factory CopyWith$Query$SupervisorDecisionHistory.stub(TRes res) =
      _CopyWithStubImpl$Query$SupervisorDecisionHistory;

  TRes call({
    Query$SupervisorDecisionHistory$supervisorDecisionHistory?
    supervisorDecisionHistory,
    String? $__typename,
  });
  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<TRes>
  get supervisorDecisionHistory;
}

class _CopyWithImpl$Query$SupervisorDecisionHistory<TRes>
    implements CopyWith$Query$SupervisorDecisionHistory<TRes> {
  _CopyWithImpl$Query$SupervisorDecisionHistory(this._instance, this._then);

  final Query$SupervisorDecisionHistory _instance;

  final TRes Function(Query$SupervisorDecisionHistory) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? supervisorDecisionHistory = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$SupervisorDecisionHistory(
      supervisorDecisionHistory:
          supervisorDecisionHistory == _undefined ||
              supervisorDecisionHistory == null
          ? _instance.supervisorDecisionHistory
          : (supervisorDecisionHistory
                as Query$SupervisorDecisionHistory$supervisorDecisionHistory),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<TRes>
  get supervisorDecisionHistory {
    final local$supervisorDecisionHistory = _instance.supervisorDecisionHistory;
    return CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory(
      local$supervisorDecisionHistory,
      (e) => call(supervisorDecisionHistory: e),
    );
  }
}

class _CopyWithStubImpl$Query$SupervisorDecisionHistory<TRes>
    implements CopyWith$Query$SupervisorDecisionHistory<TRes> {
  _CopyWithStubImpl$Query$SupervisorDecisionHistory(this._res);

  TRes _res;

  call({
    Query$SupervisorDecisionHistory$supervisorDecisionHistory?
    supervisorDecisionHistory,
    String? $__typename,
  }) => _res;

  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<TRes>
  get supervisorDecisionHistory =>
      CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory.stub(
        _res,
      );
}

const documentNodeQuerySupervisorDecisionHistory = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.query,
      name: NameNode(value: 'SupervisorDecisionHistory'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'runID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'taskID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'jobID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'supervisorDecisionHistory'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'correlation'),
                value: ObjectValueNode(
                  fields: [
                    ObjectFieldNode(
                      name: NameNode(value: 'runID'),
                      value: VariableNode(name: NameNode(value: 'runID')),
                    ),
                    ObjectFieldNode(
                      name: NameNode(value: 'taskID'),
                      value: VariableNode(name: NameNode(value: 'taskID')),
                    ),
                    ObjectFieldNode(
                      name: NameNode(value: 'jobID'),
                      value: VariableNode(name: NameNode(value: 'jobID')),
                    ),
                  ],
                ),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'SupervisorDecisionHistorySuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'decisions'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'signalType'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'action'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'reason'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'occurredAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Query$SupervisorDecisionHistory _parserFn$Query$SupervisorDecisionHistory(
  Map<String, dynamic> data,
) => Query$SupervisorDecisionHistory.fromJson(data);
typedef OnQueryComplete$Query$SupervisorDecisionHistory =
    FutureOr<void> Function(
      Map<String, dynamic>?,
      Query$SupervisorDecisionHistory?,
    );

class Options$Query$SupervisorDecisionHistory
    extends graphql.QueryOptions<Query$SupervisorDecisionHistory> {
  Options$Query$SupervisorDecisionHistory({
    String? operationName,
    required Variables$Query$SupervisorDecisionHistory variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$SupervisorDecisionHistory? typedOptimisticResult,
    Duration? pollInterval,
    graphql.Context? context,
    OnQueryComplete$Query$SupervisorDecisionHistory? onComplete,
    graphql.OnQueryError? onError,
  }) : onCompleteWithParsed = onComplete,
       super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         pollInterval: pollInterval,
         context: context,
         onComplete: onComplete == null
             ? null
             : (data) => onComplete(
                 data,
                 data == null
                     ? null
                     : _parserFn$Query$SupervisorDecisionHistory(data),
               ),
         onError: onError,
         document: documentNodeQuerySupervisorDecisionHistory,
         parserFn: _parserFn$Query$SupervisorDecisionHistory,
       );

  final OnQueryComplete$Query$SupervisorDecisionHistory? onCompleteWithParsed;

  @override
  List<Object?> get properties => [
    ...super.onComplete == null
        ? super.properties
        : super.properties.where((property) => property != onComplete),
    onCompleteWithParsed,
  ];
}

class WatchOptions$Query$SupervisorDecisionHistory
    extends graphql.WatchQueryOptions<Query$SupervisorDecisionHistory> {
  WatchOptions$Query$SupervisorDecisionHistory({
    String? operationName,
    required Variables$Query$SupervisorDecisionHistory variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$SupervisorDecisionHistory? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeQuerySupervisorDecisionHistory,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Query$SupervisorDecisionHistory,
       );
}

class FetchMoreOptions$Query$SupervisorDecisionHistory
    extends graphql.FetchMoreOptions {
  FetchMoreOptions$Query$SupervisorDecisionHistory({
    required graphql.UpdateQuery updateQuery,
    required Variables$Query$SupervisorDecisionHistory variables,
  }) : super(
         updateQuery: updateQuery,
         variables: variables.toJson(),
         document: documentNodeQuerySupervisorDecisionHistory,
       );
}

extension ClientExtension$Query$SupervisorDecisionHistory
    on graphql.GraphQLClient {
  Future<graphql.QueryResult<Query$SupervisorDecisionHistory>>
  query$SupervisorDecisionHistory(
    Options$Query$SupervisorDecisionHistory options,
  ) async => await this.query(options);

  graphql.ObservableQuery<Query$SupervisorDecisionHistory>
  watchQuery$SupervisorDecisionHistory(
    WatchOptions$Query$SupervisorDecisionHistory options,
  ) => this.watchQuery(options);

  void writeQuery$SupervisorDecisionHistory({
    required Query$SupervisorDecisionHistory data,
    required Variables$Query$SupervisorDecisionHistory variables,
    bool broadcast = true,
  }) => this.writeQuery(
    graphql.Request(
      operation: graphql.Operation(
        document: documentNodeQuerySupervisorDecisionHistory,
      ),
      variables: variables.toJson(),
    ),
    data: data.toJson(),
    broadcast: broadcast,
  );

  Query$SupervisorDecisionHistory? readQuery$SupervisorDecisionHistory({
    required Variables$Query$SupervisorDecisionHistory variables,
    bool optimistic = true,
  }) {
    final result = this.readQuery(
      graphql.Request(
        operation: graphql.Operation(
          document: documentNodeQuerySupervisorDecisionHistory,
        ),
        variables: variables.toJson(),
      ),
      optimistic: optimistic,
    );
    return result == null
        ? null
        : Query$SupervisorDecisionHistory.fromJson(result);
  }
}

class Query$SupervisorDecisionHistory$supervisorDecisionHistory {
  Query$SupervisorDecisionHistory$supervisorDecisionHistory({
    required this.$__typename,
  });

  factory Query$SupervisorDecisionHistory$supervisorDecisionHistory.fromJson(
    Map<String, dynamic> json,
  ) {
    switch (json["__typename"] as String) {
      case "SupervisorDecisionHistorySuccess":
        return Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess.fromJson(
          json,
        );

      case "GraphError":
        return Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError.fromJson(
          json,
        );

      default:
        final l$$__typename = json['__typename'];
        return Query$SupervisorDecisionHistory$supervisorDecisionHistory(
          $__typename: (l$$__typename as String),
        );
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$SupervisorDecisionHistory$supervisorDecisionHistory ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$SupervisorDecisionHistory$supervisorDecisionHistory
    on Query$SupervisorDecisionHistory$supervisorDecisionHistory {
  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<
    Query$SupervisorDecisionHistory$supervisorDecisionHistory
  >
  get copyWith =>
      CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory(
        this,
        (i) => i,
      );

  _T when<_T>({
    required _T Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess,
    )
    supervisorDecisionHistorySuccess,
    required _T Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError,
    )
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "SupervisorDecisionHistorySuccess":
        return supervisorDecisionHistorySuccess(
          this
              as Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess,
        );

      case "GraphError":
        return graphError(
          this
              as Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError,
        );

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess,
    )?
    supervisorDecisionHistorySuccess,
    _T Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError,
    )?
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "SupervisorDecisionHistorySuccess":
        if (supervisorDecisionHistorySuccess != null) {
          return supervisorDecisionHistorySuccess(
            this
                as Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(
            this
                as Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError,
          );
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<
  TRes
> {
  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory instance,
    TRes Function(Query$SupervisorDecisionHistory$supervisorDecisionHistory)
    then,
  ) = _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory;

  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<
          TRes
        > {
  _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory(
    this._instance,
    this._then,
  );

  final Query$SupervisorDecisionHistory$supervisorDecisionHistory _instance;

  final TRes Function(Query$SupervisorDecisionHistory$supervisorDecisionHistory)
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory<
          TRes
        > {
  _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory(
    this._res,
  );

  TRes _res;

  call({String? $__typename}) => _res;
}

class Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess
    implements Query$SupervisorDecisionHistory$supervisorDecisionHistory {
  Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess({
    required this.decisions,
    this.$__typename = 'SupervisorDecisionHistorySuccess',
  });

  factory Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$decisions = json['decisions'];
    final l$$__typename = json['__typename'];
    return Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess(
      decisions: (l$decisions as List<dynamic>)
          .map(
            (e) =>
                Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      $__typename: (l$$__typename as String),
    );
  }

  final List<
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
  >
  decisions;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$decisions = decisions;
    _resultData['decisions'] = l$decisions.map((e) => e.toJson()).toList();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$decisions = decisions;
    final l$$__typename = $__typename;
    return Object.hashAll([
      Object.hashAll(l$decisions.map((v) => v)),
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$decisions = decisions;
    final lOther$decisions = other.decisions;
    if (l$decisions.length != lOther$decisions.length) {
      return false;
    }
    for (int i = 0; i < l$decisions.length; i++) {
      final l$decisions$entry = l$decisions[i];
      final lOther$decisions$entry = lOther$decisions[i];
      if (l$decisions$entry != lOther$decisions$entry) {
        return false;
      }
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess
    on
        Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess {
  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess<
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess
  >
  get copyWith =>
      CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess<
  TRes
> {
  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess
    instance,
    TRes Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess,
    )
    then,
  ) = _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess;

  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess;

  TRes call({
    List<
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
    >?
    decisions,
    String? $__typename,
  });
  TRes decisions(
    Iterable<
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
    >
    Function(
      Iterable<
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
          Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
        >
      >,
    )
    _fn,
  );
}

class _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess<
          TRes
        > {
  _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess(
    this._instance,
    this._then,
  );

  final Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess
  _instance;

  final TRes Function(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? decisions = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess(
      decisions: decisions == _undefined || decisions == null
          ? _instance.decisions
          : (decisions
                as List<
                  Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
                >),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  TRes decisions(
    Iterable<
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
    >
    Function(
      Iterable<
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
          Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
        >
      >,
    )
    _fn,
  ) => call(
    decisions: _fn(
      _instance.decisions.map(
        (e) =>
            CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess<
          TRes
        > {
  _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess(
    this._res,
  );

  TRes _res;

  call({
    List<
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
    >?
    decisions,
    String? $__typename,
  }) => _res;

  decisions(_fn) => _res;
}

class Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions {
  Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions({
    required this.signalType,
    required this.action,
    required this.reason,
    required this.occurredAt,
    this.$__typename = 'SupervisorDecision',
  });

  factory Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$signalType = json['signalType'];
    final l$action = json['action'];
    final l$reason = json['reason'];
    final l$occurredAt = json['occurredAt'];
    final l$$__typename = json['__typename'];
    return Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
      signalType: fromJson$Enum$SupervisorSignalType((l$signalType as String)),
      action: fromJson$Enum$SupervisorActionCode((l$action as String)),
      reason: fromJson$Enum$SupervisorReasonCode((l$reason as String)),
      occurredAt: dateTimeFromJson(l$occurredAt),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$SupervisorSignalType signalType;

  final Enum$SupervisorActionCode action;

  final Enum$SupervisorReasonCode reason;

  final DateTime occurredAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$signalType = signalType;
    _resultData['signalType'] = toJson$Enum$SupervisorSignalType(l$signalType);
    final l$action = action;
    _resultData['action'] = toJson$Enum$SupervisorActionCode(l$action);
    final l$reason = reason;
    _resultData['reason'] = toJson$Enum$SupervisorReasonCode(l$reason);
    final l$occurredAt = occurredAt;
    _resultData['occurredAt'] = dateTimeToJson(l$occurredAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$signalType = signalType;
    final l$action = action;
    final l$reason = reason;
    final l$occurredAt = occurredAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$signalType,
      l$action,
      l$reason,
      l$occurredAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$signalType = signalType;
    final lOther$signalType = other.signalType;
    if (l$signalType != lOther$signalType) {
      return false;
    }
    final l$action = action;
    final lOther$action = other.action;
    if (l$action != lOther$action) {
      return false;
    }
    final l$reason = reason;
    final lOther$reason = other.reason;
    if (l$reason != lOther$reason) {
      return false;
    }
    final l$occurredAt = occurredAt;
    final lOther$occurredAt = other.occurredAt;
    if (l$occurredAt != lOther$occurredAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
    on
        Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions {
  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
  >
  get copyWith =>
      CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
  TRes
> {
  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
    instance,
    TRes Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions,
    )
    then,
  ) = _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions;

  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions;

  TRes call({
    Enum$SupervisorSignalType? signalType,
    Enum$SupervisorActionCode? action,
    Enum$SupervisorReasonCode? reason,
    DateTime? occurredAt,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
          TRes
        > {
  _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
    this._instance,
    this._then,
  );

  final Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions
  _instance;

  final TRes Function(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? signalType = _undefined,
    Object? action = _undefined,
    Object? reason = _undefined,
    Object? occurredAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
      signalType: signalType == _undefined || signalType == null
          ? _instance.signalType
          : (signalType as Enum$SupervisorSignalType),
      action: action == _undefined || action == null
          ? _instance.action
          : (action as Enum$SupervisorActionCode),
      reason: reason == _undefined || reason == null
          ? _instance.reason
          : (reason as Enum$SupervisorReasonCode),
      occurredAt: occurredAt == _undefined || occurredAt == null
          ? _instance.occurredAt
          : (occurredAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions<
          TRes
        > {
  _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$SupervisorDecisionHistorySuccess$decisions(
    this._res,
  );

  TRes _res;

  call({
    Enum$SupervisorSignalType? signalType,
    Enum$SupervisorActionCode? action,
    Enum$SupervisorReasonCode? reason,
    DateTime? occurredAt,
    String? $__typename,
  }) => _res;
}

class Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError
    implements Query$SupervisorDecisionHistory$supervisorDecisionHistory {
  Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError
    on Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError {
  CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError<
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError
  >
  get copyWith =>
      CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError<
  TRes
> {
  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError
    instance,
    TRes Function(
      Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError,
    )
    then,
  ) = _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError;

  factory CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError<
          TRes
        > {
  _CopyWithImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError(
    this._instance,
    this._then,
  );

  final Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError
  _instance;

  final TRes Function(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError<
  TRes
>
    implements
        CopyWith$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError<
          TRes
        > {
  _CopyWithStubImpl$Query$SupervisorDecisionHistory$supervisorDecisionHistory$$GraphError(
    this._res,
  );

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}

class Variables$Mutation$ApproveIssueIntake {
  factory Variables$Mutation$ApproveIssueIntake({
    required Input$ApproveIssueIntakeInput input,
  }) => Variables$Mutation$ApproveIssueIntake._({r'input': input});

  Variables$Mutation$ApproveIssueIntake._(this._$data);

  factory Variables$Mutation$ApproveIssueIntake.fromJson(
    Map<String, dynamic> data,
  ) {
    final result$data = <String, dynamic>{};
    final l$input = data['input'];
    result$data['input'] = Input$ApproveIssueIntakeInput.fromJson(
      (l$input as Map<String, dynamic>),
    );
    return Variables$Mutation$ApproveIssueIntake._(result$data);
  }

  Map<String, dynamic> _$data;

  Input$ApproveIssueIntakeInput get input =>
      (_$data['input'] as Input$ApproveIssueIntakeInput);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$input = input;
    result$data['input'] = l$input.toJson();
    return result$data;
  }

  CopyWith$Variables$Mutation$ApproveIssueIntake<
    Variables$Mutation$ApproveIssueIntake
  >
  get copyWith =>
      CopyWith$Variables$Mutation$ApproveIssueIntake(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Mutation$ApproveIssueIntake ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$input = input;
    final lOther$input = other.input;
    if (l$input != lOther$input) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$input = input;
    return Object.hashAll([l$input]);
  }
}

abstract class CopyWith$Variables$Mutation$ApproveIssueIntake<TRes> {
  factory CopyWith$Variables$Mutation$ApproveIssueIntake(
    Variables$Mutation$ApproveIssueIntake instance,
    TRes Function(Variables$Mutation$ApproveIssueIntake) then,
  ) = _CopyWithImpl$Variables$Mutation$ApproveIssueIntake;

  factory CopyWith$Variables$Mutation$ApproveIssueIntake.stub(TRes res) =
      _CopyWithStubImpl$Variables$Mutation$ApproveIssueIntake;

  TRes call({Input$ApproveIssueIntakeInput? input});
}

class _CopyWithImpl$Variables$Mutation$ApproveIssueIntake<TRes>
    implements CopyWith$Variables$Mutation$ApproveIssueIntake<TRes> {
  _CopyWithImpl$Variables$Mutation$ApproveIssueIntake(
    this._instance,
    this._then,
  );

  final Variables$Mutation$ApproveIssueIntake _instance;

  final TRes Function(Variables$Mutation$ApproveIssueIntake) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? input = _undefined}) => _then(
    Variables$Mutation$ApproveIssueIntake._({
      ..._instance._$data,
      if (input != _undefined && input != null)
        'input': (input as Input$ApproveIssueIntakeInput),
    }),
  );
}

class _CopyWithStubImpl$Variables$Mutation$ApproveIssueIntake<TRes>
    implements CopyWith$Variables$Mutation$ApproveIssueIntake<TRes> {
  _CopyWithStubImpl$Variables$Mutation$ApproveIssueIntake(this._res);

  TRes _res;

  call({Input$ApproveIssueIntakeInput? input}) => _res;
}

class Mutation$ApproveIssueIntake {
  Mutation$ApproveIssueIntake({
    required this.approveIssueIntake,
    this.$__typename = 'Mutation',
  });

  factory Mutation$ApproveIssueIntake.fromJson(Map<String, dynamic> json) {
    final l$approveIssueIntake = json['approveIssueIntake'];
    final l$$__typename = json['__typename'];
    return Mutation$ApproveIssueIntake(
      approveIssueIntake:
          Mutation$ApproveIssueIntake$approveIssueIntake.fromJson(
            (l$approveIssueIntake as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Mutation$ApproveIssueIntake$approveIssueIntake approveIssueIntake;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$approveIssueIntake = approveIssueIntake;
    _resultData['approveIssueIntake'] = l$approveIssueIntake.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$approveIssueIntake = approveIssueIntake;
    final l$$__typename = $__typename;
    return Object.hashAll([l$approveIssueIntake, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Mutation$ApproveIssueIntake ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$approveIssueIntake = approveIssueIntake;
    final lOther$approveIssueIntake = other.approveIssueIntake;
    if (l$approveIssueIntake != lOther$approveIssueIntake) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$ApproveIssueIntake
    on Mutation$ApproveIssueIntake {
  CopyWith$Mutation$ApproveIssueIntake<Mutation$ApproveIssueIntake>
  get copyWith => CopyWith$Mutation$ApproveIssueIntake(this, (i) => i);
}

abstract class CopyWith$Mutation$ApproveIssueIntake<TRes> {
  factory CopyWith$Mutation$ApproveIssueIntake(
    Mutation$ApproveIssueIntake instance,
    TRes Function(Mutation$ApproveIssueIntake) then,
  ) = _CopyWithImpl$Mutation$ApproveIssueIntake;

  factory CopyWith$Mutation$ApproveIssueIntake.stub(TRes res) =
      _CopyWithStubImpl$Mutation$ApproveIssueIntake;

  TRes call({
    Mutation$ApproveIssueIntake$approveIssueIntake? approveIssueIntake,
    String? $__typename,
  });
  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<TRes>
  get approveIssueIntake;
}

class _CopyWithImpl$Mutation$ApproveIssueIntake<TRes>
    implements CopyWith$Mutation$ApproveIssueIntake<TRes> {
  _CopyWithImpl$Mutation$ApproveIssueIntake(this._instance, this._then);

  final Mutation$ApproveIssueIntake _instance;

  final TRes Function(Mutation$ApproveIssueIntake) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? approveIssueIntake = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$ApproveIssueIntake(
      approveIssueIntake:
          approveIssueIntake == _undefined || approveIssueIntake == null
          ? _instance.approveIssueIntake
          : (approveIssueIntake
                as Mutation$ApproveIssueIntake$approveIssueIntake),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<TRes>
  get approveIssueIntake {
    final local$approveIssueIntake = _instance.approveIssueIntake;
    return CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake(
      local$approveIssueIntake,
      (e) => call(approveIssueIntake: e),
    );
  }
}

class _CopyWithStubImpl$Mutation$ApproveIssueIntake<TRes>
    implements CopyWith$Mutation$ApproveIssueIntake<TRes> {
  _CopyWithStubImpl$Mutation$ApproveIssueIntake(this._res);

  TRes _res;

  call({
    Mutation$ApproveIssueIntake$approveIssueIntake? approveIssueIntake,
    String? $__typename,
  }) => _res;

  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<TRes>
  get approveIssueIntake =>
      CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake.stub(_res);
}

const documentNodeMutationApproveIssueIntake = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.mutation,
      name: NameNode(value: 'ApproveIssueIntake'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'input')),
          type: NamedTypeNode(
            name: NameNode(value: 'ApproveIssueIntakeInput'),
            isNonNull: true,
          ),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'approveIssueIntake'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'input'),
                value: VariableNode(name: NameNode(value: 'input')),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'ApproveIssueIntakeSuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'decision'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'signalType'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'action'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'reason'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'occurredAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Mutation$ApproveIssueIntake _parserFn$Mutation$ApproveIssueIntake(
  Map<String, dynamic> data,
) => Mutation$ApproveIssueIntake.fromJson(data);
typedef OnMutationCompleted$Mutation$ApproveIssueIntake =
    FutureOr<void> Function(
      Map<String, dynamic>?,
      Mutation$ApproveIssueIntake?,
    );

class Options$Mutation$ApproveIssueIntake
    extends graphql.MutationOptions<Mutation$ApproveIssueIntake> {
  Options$Mutation$ApproveIssueIntake({
    String? operationName,
    required Variables$Mutation$ApproveIssueIntake variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Mutation$ApproveIssueIntake? typedOptimisticResult,
    graphql.Context? context,
    OnMutationCompleted$Mutation$ApproveIssueIntake? onCompleted,
    graphql.OnMutationUpdate<Mutation$ApproveIssueIntake>? update,
    graphql.OnError? onError,
  }) : onCompletedWithParsed = onCompleted,
       super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         onCompleted: onCompleted == null
             ? null
             : (data) => onCompleted(
                 data,
                 data == null
                     ? null
                     : _parserFn$Mutation$ApproveIssueIntake(data),
               ),
         update: update,
         onError: onError,
         document: documentNodeMutationApproveIssueIntake,
         parserFn: _parserFn$Mutation$ApproveIssueIntake,
       );

  final OnMutationCompleted$Mutation$ApproveIssueIntake? onCompletedWithParsed;

  @override
  List<Object?> get properties => [
    ...super.onCompleted == null
        ? super.properties
        : super.properties.where((property) => property != onCompleted),
    onCompletedWithParsed,
  ];
}

class WatchOptions$Mutation$ApproveIssueIntake
    extends graphql.WatchQueryOptions<Mutation$ApproveIssueIntake> {
  WatchOptions$Mutation$ApproveIssueIntake({
    String? operationName,
    required Variables$Mutation$ApproveIssueIntake variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Mutation$ApproveIssueIntake? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeMutationApproveIssueIntake,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Mutation$ApproveIssueIntake,
       );
}

extension ClientExtension$Mutation$ApproveIssueIntake on graphql.GraphQLClient {
  Future<graphql.QueryResult<Mutation$ApproveIssueIntake>>
  mutate$ApproveIssueIntake(
    Options$Mutation$ApproveIssueIntake options,
  ) async => await this.mutate(options);

  graphql.ObservableQuery<Mutation$ApproveIssueIntake>
  watchMutation$ApproveIssueIntake(
    WatchOptions$Mutation$ApproveIssueIntake options,
  ) => this.watchMutation(options);
}

class Mutation$ApproveIssueIntake$approveIssueIntake {
  Mutation$ApproveIssueIntake$approveIssueIntake({required this.$__typename});

  factory Mutation$ApproveIssueIntake$approveIssueIntake.fromJson(
    Map<String, dynamic> json,
  ) {
    switch (json["__typename"] as String) {
      case "ApproveIssueIntakeSuccess":
        return Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess.fromJson(
          json,
        );

      case "GraphError":
        return Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError.fromJson(
          json,
        );

      default:
        final l$$__typename = json['__typename'];
        return Mutation$ApproveIssueIntake$approveIssueIntake(
          $__typename: (l$$__typename as String),
        );
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Mutation$ApproveIssueIntake$approveIssueIntake ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$ApproveIssueIntake$approveIssueIntake
    on Mutation$ApproveIssueIntake$approveIssueIntake {
  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<
    Mutation$ApproveIssueIntake$approveIssueIntake
  >
  get copyWith =>
      CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake(this, (i) => i);

  _T when<_T>({
    required _T Function(
      Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess,
    )
    approveIssueIntakeSuccess,
    required _T Function(
      Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError,
    )
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "ApproveIssueIntakeSuccess":
        return approveIssueIntakeSuccess(
          this
              as Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess,
        );

      case "GraphError":
        return graphError(
          this as Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError,
        );

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(
      Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess,
    )?
    approveIssueIntakeSuccess,
    _T Function(Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError)?
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "ApproveIssueIntakeSuccess":
        if (approveIssueIntakeSuccess != null) {
          return approveIssueIntakeSuccess(
            this
                as Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(
            this as Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError,
          );
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<TRes> {
  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake(
    Mutation$ApproveIssueIntake$approveIssueIntake instance,
    TRes Function(Mutation$ApproveIssueIntake$approveIssueIntake) then,
  ) = _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake;

  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake<TRes>
    implements CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<TRes> {
  _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake(
    this._instance,
    this._then,
  );

  final Mutation$ApproveIssueIntake$approveIssueIntake _instance;

  final TRes Function(Mutation$ApproveIssueIntake$approveIssueIntake) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Mutation$ApproveIssueIntake$approveIssueIntake(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake<TRes>
    implements CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake<TRes> {
  _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake(this._res);

  TRes _res;

  call({String? $__typename}) => _res;
}

class Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess
    implements Mutation$ApproveIssueIntake$approveIssueIntake {
  Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess({
    required this.decision,
    this.$__typename = 'ApproveIssueIntakeSuccess',
  });

  factory Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$decision = json['decision'];
    final l$$__typename = json['__typename'];
    return Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess(
      decision:
          Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision.fromJson(
            (l$decision as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision
  decision;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$decision = decision;
    _resultData['decision'] = l$decision.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$decision = decision;
    final l$$__typename = $__typename;
    return Object.hashAll([l$decision, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$decision = decision;
    final lOther$decision = other.decision;
    if (l$decision != lOther$decision) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess
    on Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess {
  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess<
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess
  >
  get copyWith =>
      CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess<
  TRes
> {
  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess(
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess
    instance,
    TRes Function(
      Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess,
    )
    then,
  ) = _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess;

  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess;

  TRes call({
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision?
    decision,
    String? $__typename,
  });
  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
    TRes
  >
  get decision;
}

class _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess<
  TRes
>
    implements
        CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess<
          TRes
        > {
  _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess(
    this._instance,
    this._then,
  );

  final Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess
  _instance;

  final TRes Function(
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? decision = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess(
      decision: decision == _undefined || decision == null
          ? _instance.decision
          : (decision
                as Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
    TRes
  >
  get decision {
    final local$decision = _instance.decision;
    return CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
      local$decision,
      (e) => call(decision: e),
    );
  }
}

class _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess<
  TRes
>
    implements
        CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess<
          TRes
        > {
  _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess(
    this._res,
  );

  TRes _res;

  call({
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision?
    decision,
    String? $__typename,
  }) => _res;

  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
    TRes
  >
  get decision =>
      CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision.stub(
        _res,
      );
}

class Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision {
  Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision({
    required this.signalType,
    required this.action,
    required this.reason,
    required this.occurredAt,
    this.$__typename = 'SupervisorDecision',
  });

  factory Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$signalType = json['signalType'];
    final l$action = json['action'];
    final l$reason = json['reason'];
    final l$occurredAt = json['occurredAt'];
    final l$$__typename = json['__typename'];
    return Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
      signalType: fromJson$Enum$SupervisorSignalType((l$signalType as String)),
      action: fromJson$Enum$SupervisorActionCode((l$action as String)),
      reason: fromJson$Enum$SupervisorReasonCode((l$reason as String)),
      occurredAt: dateTimeFromJson(l$occurredAt),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$SupervisorSignalType signalType;

  final Enum$SupervisorActionCode action;

  final Enum$SupervisorReasonCode reason;

  final DateTime occurredAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$signalType = signalType;
    _resultData['signalType'] = toJson$Enum$SupervisorSignalType(l$signalType);
    final l$action = action;
    _resultData['action'] = toJson$Enum$SupervisorActionCode(l$action);
    final l$reason = reason;
    _resultData['reason'] = toJson$Enum$SupervisorReasonCode(l$reason);
    final l$occurredAt = occurredAt;
    _resultData['occurredAt'] = dateTimeToJson(l$occurredAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$signalType = signalType;
    final l$action = action;
    final l$reason = reason;
    final l$occurredAt = occurredAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$signalType,
      l$action,
      l$reason,
      l$occurredAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$signalType = signalType;
    final lOther$signalType = other.signalType;
    if (l$signalType != lOther$signalType) {
      return false;
    }
    final l$action = action;
    final lOther$action = other.action;
    if (l$action != lOther$action) {
      return false;
    }
    final l$reason = reason;
    final lOther$reason = other.reason;
    if (l$reason != lOther$reason) {
      return false;
    }
    final l$occurredAt = occurredAt;
    final lOther$occurredAt = other.occurredAt;
    if (l$occurredAt != lOther$occurredAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision
    on
        Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision {
  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision
  >
  get copyWith =>
      CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
  TRes
> {
  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision
    instance,
    TRes Function(
      Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision,
    )
    then,
  ) = _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision;

  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision;

  TRes call({
    Enum$SupervisorSignalType? signalType,
    Enum$SupervisorActionCode? action,
    Enum$SupervisorReasonCode? reason,
    DateTime? occurredAt,
    String? $__typename,
  });
}

class _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
  TRes
>
    implements
        CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
          TRes
        > {
  _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
    this._instance,
    this._then,
  );

  final Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision
  _instance;

  final TRes Function(
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? signalType = _undefined,
    Object? action = _undefined,
    Object? reason = _undefined,
    Object? occurredAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
      signalType: signalType == _undefined || signalType == null
          ? _instance.signalType
          : (signalType as Enum$SupervisorSignalType),
      action: action == _undefined || action == null
          ? _instance.action
          : (action as Enum$SupervisorActionCode),
      reason: reason == _undefined || reason == null
          ? _instance.reason
          : (reason as Enum$SupervisorReasonCode),
      occurredAt: occurredAt == _undefined || occurredAt == null
          ? _instance.occurredAt
          : (occurredAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
  TRes
>
    implements
        CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision<
          TRes
        > {
  _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$ApproveIssueIntakeSuccess$decision(
    this._res,
  );

  TRes _res;

  call({
    Enum$SupervisorSignalType? signalType,
    Enum$SupervisorActionCode? action,
    Enum$SupervisorReasonCode? reason,
    DateTime? occurredAt,
    String? $__typename,
  }) => _res;
}

class Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError
    implements Mutation$ApproveIssueIntake$approveIssueIntake {
  Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError
    on Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError {
  CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError<
    Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError
  >
  get copyWith =>
      CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError<
  TRes
> {
  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError(
    Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError instance,
    TRes Function(Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError)
    then,
  ) = _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError;

  factory CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError<
  TRes
>
    implements
        CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError<
          TRes
        > {
  _CopyWithImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError(
    this._instance,
    this._then,
  );

  final Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError _instance;

  final TRes Function(
    Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError<
  TRes
>
    implements
        CopyWith$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError<
          TRes
        > {
  _CopyWithStubImpl$Mutation$ApproveIssueIntake$approveIssueIntake$$GraphError(
    this._res,
  );

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}

class Variables$Query$ProjectSetups {
  factory Variables$Query$ProjectSetups({required int limit}) =>
      Variables$Query$ProjectSetups._({r'limit': limit});

  Variables$Query$ProjectSetups._(this._$data);

  factory Variables$Query$ProjectSetups.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$limit = data['limit'];
    result$data['limit'] = (l$limit as int);
    return Variables$Query$ProjectSetups._(result$data);
  }

  Map<String, dynamic> _$data;

  int get limit => (_$data['limit'] as int);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$limit = limit;
    result$data['limit'] = l$limit;
    return result$data;
  }

  CopyWith$Variables$Query$ProjectSetups<Variables$Query$ProjectSetups>
  get copyWith => CopyWith$Variables$Query$ProjectSetups(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Query$ProjectSetups ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$limit = limit;
    final lOther$limit = other.limit;
    if (l$limit != lOther$limit) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$limit = limit;
    return Object.hashAll([l$limit]);
  }
}

abstract class CopyWith$Variables$Query$ProjectSetups<TRes> {
  factory CopyWith$Variables$Query$ProjectSetups(
    Variables$Query$ProjectSetups instance,
    TRes Function(Variables$Query$ProjectSetups) then,
  ) = _CopyWithImpl$Variables$Query$ProjectSetups;

  factory CopyWith$Variables$Query$ProjectSetups.stub(TRes res) =
      _CopyWithStubImpl$Variables$Query$ProjectSetups;

  TRes call({int? limit});
}

class _CopyWithImpl$Variables$Query$ProjectSetups<TRes>
    implements CopyWith$Variables$Query$ProjectSetups<TRes> {
  _CopyWithImpl$Variables$Query$ProjectSetups(this._instance, this._then);

  final Variables$Query$ProjectSetups _instance;

  final TRes Function(Variables$Query$ProjectSetups) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? limit = _undefined}) => _then(
    Variables$Query$ProjectSetups._({
      ..._instance._$data,
      if (limit != _undefined && limit != null) 'limit': (limit as int),
    }),
  );
}

class _CopyWithStubImpl$Variables$Query$ProjectSetups<TRes>
    implements CopyWith$Variables$Query$ProjectSetups<TRes> {
  _CopyWithStubImpl$Variables$Query$ProjectSetups(this._res);

  TRes _res;

  call({int? limit}) => _res;
}

class Query$ProjectSetups {
  Query$ProjectSetups({
    required this.projectSetups,
    this.$__typename = 'Query',
  });

  factory Query$ProjectSetups.fromJson(Map<String, dynamic> json) {
    final l$projectSetups = json['projectSetups'];
    final l$$__typename = json['__typename'];
    return Query$ProjectSetups(
      projectSetups: Query$ProjectSetups$projectSetups.fromJson(
        (l$projectSetups as Map<String, dynamic>),
      ),
      $__typename: (l$$__typename as String),
    );
  }

  final Query$ProjectSetups$projectSetups projectSetups;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$projectSetups = projectSetups;
    _resultData['projectSetups'] = l$projectSetups.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$projectSetups = projectSetups;
    final l$$__typename = $__typename;
    return Object.hashAll([l$projectSetups, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$ProjectSetups || runtimeType != other.runtimeType) {
      return false;
    }
    final l$projectSetups = projectSetups;
    final lOther$projectSetups = other.projectSetups;
    if (l$projectSetups != lOther$projectSetups) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups on Query$ProjectSetups {
  CopyWith$Query$ProjectSetups<Query$ProjectSetups> get copyWith =>
      CopyWith$Query$ProjectSetups(this, (i) => i);
}

abstract class CopyWith$Query$ProjectSetups<TRes> {
  factory CopyWith$Query$ProjectSetups(
    Query$ProjectSetups instance,
    TRes Function(Query$ProjectSetups) then,
  ) = _CopyWithImpl$Query$ProjectSetups;

  factory CopyWith$Query$ProjectSetups.stub(TRes res) =
      _CopyWithStubImpl$Query$ProjectSetups;

  TRes call({
    Query$ProjectSetups$projectSetups? projectSetups,
    String? $__typename,
  });
  CopyWith$Query$ProjectSetups$projectSetups<TRes> get projectSetups;
}

class _CopyWithImpl$Query$ProjectSetups<TRes>
    implements CopyWith$Query$ProjectSetups<TRes> {
  _CopyWithImpl$Query$ProjectSetups(this._instance, this._then);

  final Query$ProjectSetups _instance;

  final TRes Function(Query$ProjectSetups) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? projectSetups = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$ProjectSetups(
      projectSetups: projectSetups == _undefined || projectSetups == null
          ? _instance.projectSetups
          : (projectSetups as Query$ProjectSetups$projectSetups),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Query$ProjectSetups$projectSetups<TRes> get projectSetups {
    final local$projectSetups = _instance.projectSetups;
    return CopyWith$Query$ProjectSetups$projectSetups(
      local$projectSetups,
      (e) => call(projectSetups: e),
    );
  }
}

class _CopyWithStubImpl$Query$ProjectSetups<TRes>
    implements CopyWith$Query$ProjectSetups<TRes> {
  _CopyWithStubImpl$Query$ProjectSetups(this._res);

  TRes _res;

  call({
    Query$ProjectSetups$projectSetups? projectSetups,
    String? $__typename,
  }) => _res;

  CopyWith$Query$ProjectSetups$projectSetups<TRes> get projectSetups =>
      CopyWith$Query$ProjectSetups$projectSetups.stub(_res);
}

const documentNodeQueryProjectSetups = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.query,
      name: NameNode(value: 'ProjectSetups'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'limit')),
          type: NamedTypeNode(name: NameNode(value: 'Int'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'projectSetups'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'limit'),
                value: VariableNode(name: NameNode(value: 'limit')),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'ProjectSetupsSuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'projects'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'projectID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'projectName'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'repositories'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: SelectionSetNode(
                                selections: [
                                  FieldNode(
                                    name: NameNode(value: 'repositoryID'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'scmProvider'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'repositoryURL'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'isPrimary'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: '__typename'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                ],
                              ),
                            ),
                            FieldNode(
                              name: NameNode(value: 'boards'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: SelectionSetNode(
                                selections: [
                                  FieldNode(
                                    name: NameNode(value: 'boardID'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'trackerProvider'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'taskboardName'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(
                                      value: 'appliesToAllRepositories',
                                    ),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'repositoryIDs'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: '__typename'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                ],
                              ),
                            ),
                            FieldNode(
                              name: NameNode(value: 'createdAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'updatedAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Query$ProjectSetups _parserFn$Query$ProjectSetups(Map<String, dynamic> data) =>
    Query$ProjectSetups.fromJson(data);
typedef OnQueryComplete$Query$ProjectSetups =
    FutureOr<void> Function(Map<String, dynamic>?, Query$ProjectSetups?);

class Options$Query$ProjectSetups
    extends graphql.QueryOptions<Query$ProjectSetups> {
  Options$Query$ProjectSetups({
    String? operationName,
    required Variables$Query$ProjectSetups variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$ProjectSetups? typedOptimisticResult,
    Duration? pollInterval,
    graphql.Context? context,
    OnQueryComplete$Query$ProjectSetups? onComplete,
    graphql.OnQueryError? onError,
  }) : onCompleteWithParsed = onComplete,
       super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         pollInterval: pollInterval,
         context: context,
         onComplete: onComplete == null
             ? null
             : (data) => onComplete(
                 data,
                 data == null ? null : _parserFn$Query$ProjectSetups(data),
               ),
         onError: onError,
         document: documentNodeQueryProjectSetups,
         parserFn: _parserFn$Query$ProjectSetups,
       );

  final OnQueryComplete$Query$ProjectSetups? onCompleteWithParsed;

  @override
  List<Object?> get properties => [
    ...super.onComplete == null
        ? super.properties
        : super.properties.where((property) => property != onComplete),
    onCompleteWithParsed,
  ];
}

class WatchOptions$Query$ProjectSetups
    extends graphql.WatchQueryOptions<Query$ProjectSetups> {
  WatchOptions$Query$ProjectSetups({
    String? operationName,
    required Variables$Query$ProjectSetups variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Query$ProjectSetups? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeQueryProjectSetups,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Query$ProjectSetups,
       );
}

class FetchMoreOptions$Query$ProjectSetups extends graphql.FetchMoreOptions {
  FetchMoreOptions$Query$ProjectSetups({
    required graphql.UpdateQuery updateQuery,
    required Variables$Query$ProjectSetups variables,
  }) : super(
         updateQuery: updateQuery,
         variables: variables.toJson(),
         document: documentNodeQueryProjectSetups,
       );
}

extension ClientExtension$Query$ProjectSetups on graphql.GraphQLClient {
  Future<graphql.QueryResult<Query$ProjectSetups>> query$ProjectSetups(
    Options$Query$ProjectSetups options,
  ) async => await this.query(options);

  graphql.ObservableQuery<Query$ProjectSetups> watchQuery$ProjectSetups(
    WatchOptions$Query$ProjectSetups options,
  ) => this.watchQuery(options);

  void writeQuery$ProjectSetups({
    required Query$ProjectSetups data,
    required Variables$Query$ProjectSetups variables,
    bool broadcast = true,
  }) => this.writeQuery(
    graphql.Request(
      operation: graphql.Operation(document: documentNodeQueryProjectSetups),
      variables: variables.toJson(),
    ),
    data: data.toJson(),
    broadcast: broadcast,
  );

  Query$ProjectSetups? readQuery$ProjectSetups({
    required Variables$Query$ProjectSetups variables,
    bool optimistic = true,
  }) {
    final result = this.readQuery(
      graphql.Request(
        operation: graphql.Operation(document: documentNodeQueryProjectSetups),
        variables: variables.toJson(),
      ),
      optimistic: optimistic,
    );
    return result == null ? null : Query$ProjectSetups.fromJson(result);
  }
}

class Query$ProjectSetups$projectSetups {
  Query$ProjectSetups$projectSetups({required this.$__typename});

  factory Query$ProjectSetups$projectSetups.fromJson(
    Map<String, dynamic> json,
  ) {
    switch (json["__typename"] as String) {
      case "ProjectSetupsSuccess":
        return Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess.fromJson(
          json,
        );

      case "GraphError":
        return Query$ProjectSetups$projectSetups$$GraphError.fromJson(json);

      default:
        final l$$__typename = json['__typename'];
        return Query$ProjectSetups$projectSetups(
          $__typename: (l$$__typename as String),
        );
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$ProjectSetups$projectSetups ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups$projectSetups
    on Query$ProjectSetups$projectSetups {
  CopyWith$Query$ProjectSetups$projectSetups<Query$ProjectSetups$projectSetups>
  get copyWith => CopyWith$Query$ProjectSetups$projectSetups(this, (i) => i);

  _T when<_T>({
    required _T Function(
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess,
    )
    projectSetupsSuccess,
    required _T Function(Query$ProjectSetups$projectSetups$$GraphError)
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "ProjectSetupsSuccess":
        return projectSetupsSuccess(
          this as Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess,
        );

      case "GraphError":
        return graphError(
          this as Query$ProjectSetups$projectSetups$$GraphError,
        );

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess)?
    projectSetupsSuccess,
    _T Function(Query$ProjectSetups$projectSetups$$GraphError)? graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "ProjectSetupsSuccess":
        if (projectSetupsSuccess != null) {
          return projectSetupsSuccess(
            this as Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(
            this as Query$ProjectSetups$projectSetups$$GraphError,
          );
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Query$ProjectSetups$projectSetups<TRes> {
  factory CopyWith$Query$ProjectSetups$projectSetups(
    Query$ProjectSetups$projectSetups instance,
    TRes Function(Query$ProjectSetups$projectSetups) then,
  ) = _CopyWithImpl$Query$ProjectSetups$projectSetups;

  factory CopyWith$Query$ProjectSetups$projectSetups.stub(TRes res) =
      _CopyWithStubImpl$Query$ProjectSetups$projectSetups;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Query$ProjectSetups$projectSetups<TRes>
    implements CopyWith$Query$ProjectSetups$projectSetups<TRes> {
  _CopyWithImpl$Query$ProjectSetups$projectSetups(this._instance, this._then);

  final Query$ProjectSetups$projectSetups _instance;

  final TRes Function(Query$ProjectSetups$projectSetups) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Query$ProjectSetups$projectSetups(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$ProjectSetups$projectSetups<TRes>
    implements CopyWith$Query$ProjectSetups$projectSetups<TRes> {
  _CopyWithStubImpl$Query$ProjectSetups$projectSetups(this._res);

  TRes _res;

  call({String? $__typename}) => _res;
}

class Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess
    implements Query$ProjectSetups$projectSetups {
  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess({
    required this.projects,
    this.$__typename = 'ProjectSetupsSuccess',
  });

  factory Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$projects = json['projects'];
    final l$$__typename = json['__typename'];
    return Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess(
      projects: (l$projects as List<dynamic>)
          .map(
            (e) =>
                Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      $__typename: (l$$__typename as String),
    );
  }

  final List<Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects>
  projects;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$projects = projects;
    _resultData['projects'] = l$projects.map((e) => e.toJson()).toList();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$projects = projects;
    final l$$__typename = $__typename;
    return Object.hashAll([
      Object.hashAll(l$projects.map((v) => v)),
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$projects = projects;
    final lOther$projects = other.projects;
    if (l$projects.length != lOther$projects.length) {
      return false;
    }
    for (int i = 0; i < l$projects.length; i++) {
      final l$projects$entry = l$projects[i];
      final lOther$projects$entry = lOther$projects[i];
      if (l$projects$entry != lOther$projects$entry) {
        return false;
      }
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess
    on Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess {
  CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess<
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess
  >
  get copyWith =>
      CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess<
  TRes
> {
  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess instance,
    TRes Function(Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess) then,
  ) = _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess;

  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess;

  TRes call({
    List<Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects>?
    projects,
    String? $__typename,
  });
  TRes projects(
    Iterable<Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects>
    Function(
      Iterable<
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
          Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects
        >
      >,
    )
    _fn,
  );
}

class _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess<TRes> {
  _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess(
    this._instance,
    this._then,
  );

  final Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess _instance;

  final TRes Function(Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess)
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? projects = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess(
      projects: projects == _undefined || projects == null
          ? _instance.projects
          : (projects
                as List<
                  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects
                >),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  TRes projects(
    Iterable<Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects>
    Function(
      Iterable<
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
          Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects
        >
      >,
    )
    _fn,
  ) => call(
    projects: _fn(
      _instance.projects.map(
        (e) =>
            CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess<TRes> {
  _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess(
    this._res,
  );

  TRes _res;

  call({
    List<Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects>?
    projects,
    String? $__typename,
  }) => _res;

  projects(_fn) => _res;
}

class Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects {
  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects({
    required this.projectID,
    required this.projectName,
    required this.repositories,
    required this.boards,
    required this.createdAt,
    required this.updatedAt,
    this.$__typename = 'ProjectSetup',
  });

  factory Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$projectID = json['projectID'];
    final l$projectName = json['projectName'];
    final l$repositories = json['repositories'];
    final l$boards = json['boards'];
    final l$createdAt = json['createdAt'];
    final l$updatedAt = json['updatedAt'];
    final l$$__typename = json['__typename'];
    return Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
      projectID: (l$projectID as String),
      projectName: (l$projectName as String),
      repositories: (l$repositories as List<dynamic>)
          .map(
            (e) =>
                Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      boards: (l$boards as List<dynamic>)
          .map(
            (e) =>
                Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      createdAt: dateTimeFromJson(l$createdAt),
      updatedAt: dateTimeFromJson(l$updatedAt),
      $__typename: (l$$__typename as String),
    );
  }

  final String projectID;

  final String projectName;

  final List<
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
  >
  repositories;

  final List<
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
  >
  boards;

  final DateTime createdAt;

  final DateTime updatedAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$projectID = projectID;
    _resultData['projectID'] = l$projectID;
    final l$projectName = projectName;
    _resultData['projectName'] = l$projectName;
    final l$repositories = repositories;
    _resultData['repositories'] = l$repositories
        .map((e) => e.toJson())
        .toList();
    final l$boards = boards;
    _resultData['boards'] = l$boards.map((e) => e.toJson()).toList();
    final l$createdAt = createdAt;
    _resultData['createdAt'] = dateTimeToJson(l$createdAt);
    final l$updatedAt = updatedAt;
    _resultData['updatedAt'] = dateTimeToJson(l$updatedAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$projectID = projectID;
    final l$projectName = projectName;
    final l$repositories = repositories;
    final l$boards = boards;
    final l$createdAt = createdAt;
    final l$updatedAt = updatedAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$projectID,
      l$projectName,
      Object.hashAll(l$repositories.map((v) => v)),
      Object.hashAll(l$boards.map((v) => v)),
      l$createdAt,
      l$updatedAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$projectID = projectID;
    final lOther$projectID = other.projectID;
    if (l$projectID != lOther$projectID) {
      return false;
    }
    final l$projectName = projectName;
    final lOther$projectName = other.projectName;
    if (l$projectName != lOther$projectName) {
      return false;
    }
    final l$repositories = repositories;
    final lOther$repositories = other.repositories;
    if (l$repositories.length != lOther$repositories.length) {
      return false;
    }
    for (int i = 0; i < l$repositories.length; i++) {
      final l$repositories$entry = l$repositories[i];
      final lOther$repositories$entry = lOther$repositories[i];
      if (l$repositories$entry != lOther$repositories$entry) {
        return false;
      }
    }
    final l$boards = boards;
    final lOther$boards = other.boards;
    if (l$boards.length != lOther$boards.length) {
      return false;
    }
    for (int i = 0; i < l$boards.length; i++) {
      final l$boards$entry = l$boards[i];
      final lOther$boards$entry = lOther$boards[i];
      if (l$boards$entry != lOther$boards$entry) {
        return false;
      }
    }
    final l$createdAt = createdAt;
    final lOther$createdAt = other.createdAt;
    if (l$createdAt != lOther$createdAt) {
      return false;
    }
    final l$updatedAt = updatedAt;
    final lOther$updatedAt = other.updatedAt;
    if (l$updatedAt != lOther$updatedAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects
    on Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects {
  CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects
  >
  get copyWith =>
      CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
  TRes
> {
  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects instance,
    TRes Function(
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects,
    )
    then,
  ) = _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects;

  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects;

  TRes call({
    String? projectID,
    String? projectName,
    List<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
    >?
    repositories,
    List<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
    >?
    boards,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? $__typename,
  });
  TRes repositories(
    Iterable<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
    >
    Function(
      Iterable<
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
          Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
        >
      >,
    )
    _fn,
  );
  TRes boards(
    Iterable<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
    >
    Function(
      Iterable<
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
          Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
        >
      >,
    )
    _fn,
  );
}

class _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
          TRes
        > {
  _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
    this._instance,
    this._then,
  );

  final Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects
  _instance;

  final TRes Function(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? projectID = _undefined,
    Object? projectName = _undefined,
    Object? repositories = _undefined,
    Object? boards = _undefined,
    Object? createdAt = _undefined,
    Object? updatedAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
      projectID: projectID == _undefined || projectID == null
          ? _instance.projectID
          : (projectID as String),
      projectName: projectName == _undefined || projectName == null
          ? _instance.projectName
          : (projectName as String),
      repositories: repositories == _undefined || repositories == null
          ? _instance.repositories
          : (repositories
                as List<
                  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
                >),
      boards: boards == _undefined || boards == null
          ? _instance.boards
          : (boards
                as List<
                  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
                >),
      createdAt: createdAt == _undefined || createdAt == null
          ? _instance.createdAt
          : (createdAt as DateTime),
      updatedAt: updatedAt == _undefined || updatedAt == null
          ? _instance.updatedAt
          : (updatedAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  TRes repositories(
    Iterable<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
    >
    Function(
      Iterable<
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
          Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
        >
      >,
    )
    _fn,
  ) => call(
    repositories: _fn(
      _instance.repositories.map(
        (e) =>
            CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );

  TRes boards(
    Iterable<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
    >
    Function(
      Iterable<
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
          Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
        >
      >,
    )
    _fn,
  ) => call(
    boards: _fn(
      _instance.boards.map(
        (e) =>
            CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects<
          TRes
        > {
  _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects(
    this._res,
  );

  TRes _res;

  call({
    String? projectID,
    String? projectName,
    List<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
    >?
    repositories,
    List<
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
    >?
    boards,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? $__typename,
  }) => _res;

  repositories(_fn) => _res;

  boards(_fn) => _res;
}

class Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories {
  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories({
    required this.repositoryID,
    required this.scmProvider,
    required this.repositoryURL,
    required this.isPrimary,
    this.$__typename = 'ProjectRepository',
  });

  factory Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$repositoryID = json['repositoryID'];
    final l$scmProvider = json['scmProvider'];
    final l$repositoryURL = json['repositoryURL'];
    final l$isPrimary = json['isPrimary'];
    final l$$__typename = json['__typename'];
    return Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
      repositoryID: (l$repositoryID as String),
      scmProvider: fromJson$Enum$SCMProvider((l$scmProvider as String)),
      repositoryURL: (l$repositoryURL as String),
      isPrimary: (l$isPrimary as bool),
      $__typename: (l$$__typename as String),
    );
  }

  final String repositoryID;

  final Enum$SCMProvider scmProvider;

  final String repositoryURL;

  final bool isPrimary;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$repositoryID = repositoryID;
    _resultData['repositoryID'] = l$repositoryID;
    final l$scmProvider = scmProvider;
    _resultData['scmProvider'] = toJson$Enum$SCMProvider(l$scmProvider);
    final l$repositoryURL = repositoryURL;
    _resultData['repositoryURL'] = l$repositoryURL;
    final l$isPrimary = isPrimary;
    _resultData['isPrimary'] = l$isPrimary;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$repositoryID = repositoryID;
    final l$scmProvider = scmProvider;
    final l$repositoryURL = repositoryURL;
    final l$isPrimary = isPrimary;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$repositoryID,
      l$scmProvider,
      l$repositoryURL,
      l$isPrimary,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$repositoryID = repositoryID;
    final lOther$repositoryID = other.repositoryID;
    if (l$repositoryID != lOther$repositoryID) {
      return false;
    }
    final l$scmProvider = scmProvider;
    final lOther$scmProvider = other.scmProvider;
    if (l$scmProvider != lOther$scmProvider) {
      return false;
    }
    final l$repositoryURL = repositoryURL;
    final lOther$repositoryURL = other.repositoryURL;
    if (l$repositoryURL != lOther$repositoryURL) {
      return false;
    }
    final l$isPrimary = isPrimary;
    final lOther$isPrimary = other.isPrimary;
    if (l$isPrimary != lOther$isPrimary) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
    on Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories {
  CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
  >
  get copyWith =>
      CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
  TRes
> {
  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
    instance,
    TRes Function(
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories,
    )
    then,
  ) = _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories;

  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories;

  TRes call({
    String? repositoryID,
    Enum$SCMProvider? scmProvider,
    String? repositoryURL,
    bool? isPrimary,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
          TRes
        > {
  _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
    this._instance,
    this._then,
  );

  final Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories
  _instance;

  final TRes Function(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? repositoryID = _undefined,
    Object? scmProvider = _undefined,
    Object? repositoryURL = _undefined,
    Object? isPrimary = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
      repositoryID: repositoryID == _undefined || repositoryID == null
          ? _instance.repositoryID
          : (repositoryID as String),
      scmProvider: scmProvider == _undefined || scmProvider == null
          ? _instance.scmProvider
          : (scmProvider as Enum$SCMProvider),
      repositoryURL: repositoryURL == _undefined || repositoryURL == null
          ? _instance.repositoryURL
          : (repositoryURL as String),
      isPrimary: isPrimary == _undefined || isPrimary == null
          ? _instance.isPrimary
          : (isPrimary as bool),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories<
          TRes
        > {
  _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$repositories(
    this._res,
  );

  TRes _res;

  call({
    String? repositoryID,
    Enum$SCMProvider? scmProvider,
    String? repositoryURL,
    bool? isPrimary,
    String? $__typename,
  }) => _res;
}

class Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards {
  Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards({
    required this.boardID,
    required this.trackerProvider,
    this.taskboardName,
    required this.appliesToAllRepositories,
    required this.repositoryIDs,
    this.$__typename = 'ProjectBoard',
  });

  factory Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$boardID = json['boardID'];
    final l$trackerProvider = json['trackerProvider'];
    final l$taskboardName = json['taskboardName'];
    final l$appliesToAllRepositories = json['appliesToAllRepositories'];
    final l$repositoryIDs = json['repositoryIDs'];
    final l$$__typename = json['__typename'];
    return Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
      boardID: (l$boardID as String),
      trackerProvider: fromJson$Enum$TrackerSourceKind(
        (l$trackerProvider as String),
      ),
      taskboardName: (l$taskboardName as String?),
      appliesToAllRepositories: (l$appliesToAllRepositories as bool),
      repositoryIDs: (l$repositoryIDs as List<dynamic>)
          .map((e) => (e as String))
          .toList(),
      $__typename: (l$$__typename as String),
    );
  }

  final String boardID;

  final Enum$TrackerSourceKind trackerProvider;

  final String? taskboardName;

  final bool appliesToAllRepositories;

  final List<String> repositoryIDs;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$boardID = boardID;
    _resultData['boardID'] = l$boardID;
    final l$trackerProvider = trackerProvider;
    _resultData['trackerProvider'] = toJson$Enum$TrackerSourceKind(
      l$trackerProvider,
    );
    final l$taskboardName = taskboardName;
    _resultData['taskboardName'] = l$taskboardName;
    final l$appliesToAllRepositories = appliesToAllRepositories;
    _resultData['appliesToAllRepositories'] = l$appliesToAllRepositories;
    final l$repositoryIDs = repositoryIDs;
    _resultData['repositoryIDs'] = l$repositoryIDs.map((e) => e).toList();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$boardID = boardID;
    final l$trackerProvider = trackerProvider;
    final l$taskboardName = taskboardName;
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final l$repositoryIDs = repositoryIDs;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$boardID,
      l$trackerProvider,
      l$taskboardName,
      l$appliesToAllRepositories,
      Object.hashAll(l$repositoryIDs.map((v) => v)),
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$boardID = boardID;
    final lOther$boardID = other.boardID;
    if (l$boardID != lOther$boardID) {
      return false;
    }
    final l$trackerProvider = trackerProvider;
    final lOther$trackerProvider = other.trackerProvider;
    if (l$trackerProvider != lOther$trackerProvider) {
      return false;
    }
    final l$taskboardName = taskboardName;
    final lOther$taskboardName = other.taskboardName;
    if (l$taskboardName != lOther$taskboardName) {
      return false;
    }
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final lOther$appliesToAllRepositories = other.appliesToAllRepositories;
    if (l$appliesToAllRepositories != lOther$appliesToAllRepositories) {
      return false;
    }
    final l$repositoryIDs = repositoryIDs;
    final lOther$repositoryIDs = other.repositoryIDs;
    if (l$repositoryIDs.length != lOther$repositoryIDs.length) {
      return false;
    }
    for (int i = 0; i < l$repositoryIDs.length; i++) {
      final l$repositoryIDs$entry = l$repositoryIDs[i];
      final lOther$repositoryIDs$entry = lOther$repositoryIDs[i];
      if (l$repositoryIDs$entry != lOther$repositoryIDs$entry) {
        return false;
      }
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
    on Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards {
  CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
  >
  get copyWith =>
      CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
  TRes
> {
  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
    instance,
    TRes Function(
      Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards,
    )
    then,
  ) = _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards;

  factory CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards;

  TRes call({
    String? boardID,
    Enum$TrackerSourceKind? trackerProvider,
    String? taskboardName,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
          TRes
        > {
  _CopyWithImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
    this._instance,
    this._then,
  );

  final Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards
  _instance;

  final TRes Function(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? boardID = _undefined,
    Object? trackerProvider = _undefined,
    Object? taskboardName = _undefined,
    Object? appliesToAllRepositories = _undefined,
    Object? repositoryIDs = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
      boardID: boardID == _undefined || boardID == null
          ? _instance.boardID
          : (boardID as String),
      trackerProvider: trackerProvider == _undefined || trackerProvider == null
          ? _instance.trackerProvider
          : (trackerProvider as Enum$TrackerSourceKind),
      taskboardName: taskboardName == _undefined
          ? _instance.taskboardName
          : (taskboardName as String?),
      appliesToAllRepositories:
          appliesToAllRepositories == _undefined ||
              appliesToAllRepositories == null
          ? _instance.appliesToAllRepositories
          : (appliesToAllRepositories as bool),
      repositoryIDs: repositoryIDs == _undefined || repositoryIDs == null
          ? _instance.repositoryIDs
          : (repositoryIDs as List<String>),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
  TRes
>
    implements
        CopyWith$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards<
          TRes
        > {
  _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$ProjectSetupsSuccess$projects$boards(
    this._res,
  );

  TRes _res;

  call({
    String? boardID,
    Enum$TrackerSourceKind? trackerProvider,
    String? taskboardName,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
    String? $__typename,
  }) => _res;
}

class Query$ProjectSetups$projectSetups$$GraphError
    implements Query$ProjectSetups$projectSetups {
  Query$ProjectSetups$projectSetups$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Query$ProjectSetups$projectSetups$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Query$ProjectSetups$projectSetups$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Query$ProjectSetups$projectSetups$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Query$ProjectSetups$projectSetups$$GraphError
    on Query$ProjectSetups$projectSetups$$GraphError {
  CopyWith$Query$ProjectSetups$projectSetups$$GraphError<
    Query$ProjectSetups$projectSetups$$GraphError
  >
  get copyWith =>
      CopyWith$Query$ProjectSetups$projectSetups$$GraphError(this, (i) => i);
}

abstract class CopyWith$Query$ProjectSetups$projectSetups$$GraphError<TRes> {
  factory CopyWith$Query$ProjectSetups$projectSetups$$GraphError(
    Query$ProjectSetups$projectSetups$$GraphError instance,
    TRes Function(Query$ProjectSetups$projectSetups$$GraphError) then,
  ) = _CopyWithImpl$Query$ProjectSetups$projectSetups$$GraphError;

  factory CopyWith$Query$ProjectSetups$projectSetups$$GraphError.stub(
    TRes res,
  ) = _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Query$ProjectSetups$projectSetups$$GraphError<TRes>
    implements CopyWith$Query$ProjectSetups$projectSetups$$GraphError<TRes> {
  _CopyWithImpl$Query$ProjectSetups$projectSetups$$GraphError(
    this._instance,
    this._then,
  );

  final Query$ProjectSetups$projectSetups$$GraphError _instance;

  final TRes Function(Query$ProjectSetups$projectSetups$$GraphError) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Query$ProjectSetups$projectSetups$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$GraphError<TRes>
    implements CopyWith$Query$ProjectSetups$projectSetups$$GraphError<TRes> {
  _CopyWithStubImpl$Query$ProjectSetups$projectSetups$$GraphError(this._res);

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}

class Variables$Mutation$UpsertProjectSetup {
  factory Variables$Mutation$UpsertProjectSetup({
    required Input$UpsertProjectSetupInput input,
  }) => Variables$Mutation$UpsertProjectSetup._({r'input': input});

  Variables$Mutation$UpsertProjectSetup._(this._$data);

  factory Variables$Mutation$UpsertProjectSetup.fromJson(
    Map<String, dynamic> data,
  ) {
    final result$data = <String, dynamic>{};
    final l$input = data['input'];
    result$data['input'] = Input$UpsertProjectSetupInput.fromJson(
      (l$input as Map<String, dynamic>),
    );
    return Variables$Mutation$UpsertProjectSetup._(result$data);
  }

  Map<String, dynamic> _$data;

  Input$UpsertProjectSetupInput get input =>
      (_$data['input'] as Input$UpsertProjectSetupInput);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$input = input;
    result$data['input'] = l$input.toJson();
    return result$data;
  }

  CopyWith$Variables$Mutation$UpsertProjectSetup<
    Variables$Mutation$UpsertProjectSetup
  >
  get copyWith =>
      CopyWith$Variables$Mutation$UpsertProjectSetup(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Mutation$UpsertProjectSetup ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$input = input;
    final lOther$input = other.input;
    if (l$input != lOther$input) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$input = input;
    return Object.hashAll([l$input]);
  }
}

abstract class CopyWith$Variables$Mutation$UpsertProjectSetup<TRes> {
  factory CopyWith$Variables$Mutation$UpsertProjectSetup(
    Variables$Mutation$UpsertProjectSetup instance,
    TRes Function(Variables$Mutation$UpsertProjectSetup) then,
  ) = _CopyWithImpl$Variables$Mutation$UpsertProjectSetup;

  factory CopyWith$Variables$Mutation$UpsertProjectSetup.stub(TRes res) =
      _CopyWithStubImpl$Variables$Mutation$UpsertProjectSetup;

  TRes call({Input$UpsertProjectSetupInput? input});
}

class _CopyWithImpl$Variables$Mutation$UpsertProjectSetup<TRes>
    implements CopyWith$Variables$Mutation$UpsertProjectSetup<TRes> {
  _CopyWithImpl$Variables$Mutation$UpsertProjectSetup(
    this._instance,
    this._then,
  );

  final Variables$Mutation$UpsertProjectSetup _instance;

  final TRes Function(Variables$Mutation$UpsertProjectSetup) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? input = _undefined}) => _then(
    Variables$Mutation$UpsertProjectSetup._({
      ..._instance._$data,
      if (input != _undefined && input != null)
        'input': (input as Input$UpsertProjectSetupInput),
    }),
  );
}

class _CopyWithStubImpl$Variables$Mutation$UpsertProjectSetup<TRes>
    implements CopyWith$Variables$Mutation$UpsertProjectSetup<TRes> {
  _CopyWithStubImpl$Variables$Mutation$UpsertProjectSetup(this._res);

  TRes _res;

  call({Input$UpsertProjectSetupInput? input}) => _res;
}

class Mutation$UpsertProjectSetup {
  Mutation$UpsertProjectSetup({
    required this.upsertProjectSetup,
    this.$__typename = 'Mutation',
  });

  factory Mutation$UpsertProjectSetup.fromJson(Map<String, dynamic> json) {
    final l$upsertProjectSetup = json['upsertProjectSetup'];
    final l$$__typename = json['__typename'];
    return Mutation$UpsertProjectSetup(
      upsertProjectSetup:
          Mutation$UpsertProjectSetup$upsertProjectSetup.fromJson(
            (l$upsertProjectSetup as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Mutation$UpsertProjectSetup$upsertProjectSetup upsertProjectSetup;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$upsertProjectSetup = upsertProjectSetup;
    _resultData['upsertProjectSetup'] = l$upsertProjectSetup.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$upsertProjectSetup = upsertProjectSetup;
    final l$$__typename = $__typename;
    return Object.hashAll([l$upsertProjectSetup, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Mutation$UpsertProjectSetup ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$upsertProjectSetup = upsertProjectSetup;
    final lOther$upsertProjectSetup = other.upsertProjectSetup;
    if (l$upsertProjectSetup != lOther$upsertProjectSetup) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup
    on Mutation$UpsertProjectSetup {
  CopyWith$Mutation$UpsertProjectSetup<Mutation$UpsertProjectSetup>
  get copyWith => CopyWith$Mutation$UpsertProjectSetup(this, (i) => i);
}

abstract class CopyWith$Mutation$UpsertProjectSetup<TRes> {
  factory CopyWith$Mutation$UpsertProjectSetup(
    Mutation$UpsertProjectSetup instance,
    TRes Function(Mutation$UpsertProjectSetup) then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup;

  factory CopyWith$Mutation$UpsertProjectSetup.stub(TRes res) =
      _CopyWithStubImpl$Mutation$UpsertProjectSetup;

  TRes call({
    Mutation$UpsertProjectSetup$upsertProjectSetup? upsertProjectSetup,
    String? $__typename,
  });
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes>
  get upsertProjectSetup;
}

class _CopyWithImpl$Mutation$UpsertProjectSetup<TRes>
    implements CopyWith$Mutation$UpsertProjectSetup<TRes> {
  _CopyWithImpl$Mutation$UpsertProjectSetup(this._instance, this._then);

  final Mutation$UpsertProjectSetup _instance;

  final TRes Function(Mutation$UpsertProjectSetup) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? upsertProjectSetup = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$UpsertProjectSetup(
      upsertProjectSetup:
          upsertProjectSetup == _undefined || upsertProjectSetup == null
          ? _instance.upsertProjectSetup
          : (upsertProjectSetup
                as Mutation$UpsertProjectSetup$upsertProjectSetup),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes>
  get upsertProjectSetup {
    final local$upsertProjectSetup = _instance.upsertProjectSetup;
    return CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup(
      local$upsertProjectSetup,
      (e) => call(upsertProjectSetup: e),
    );
  }
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup<TRes>
    implements CopyWith$Mutation$UpsertProjectSetup<TRes> {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup(this._res);

  TRes _res;

  call({
    Mutation$UpsertProjectSetup$upsertProjectSetup? upsertProjectSetup,
    String? $__typename,
  }) => _res;

  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes>
  get upsertProjectSetup =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup.stub(_res);
}

const documentNodeMutationUpsertProjectSetup = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.mutation,
      name: NameNode(value: 'UpsertProjectSetup'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'input')),
          type: NamedTypeNode(
            name: NameNode(value: 'UpsertProjectSetupInput'),
            isNonNull: true,
          ),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'upsertProjectSetup'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'input'),
                value: VariableNode(name: NameNode(value: 'input')),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'UpsertProjectSetupSuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'project'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'projectID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'projectName'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'repositories'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: SelectionSetNode(
                                selections: [
                                  FieldNode(
                                    name: NameNode(value: 'repositoryID'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'scmProvider'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'repositoryURL'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'isPrimary'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: '__typename'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                ],
                              ),
                            ),
                            FieldNode(
                              name: NameNode(value: 'boards'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: SelectionSetNode(
                                selections: [
                                  FieldNode(
                                    name: NameNode(value: 'boardID'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'trackerProvider'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'taskboardName'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(
                                      value: 'appliesToAllRepositories',
                                    ),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: 'repositoryIDs'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                  FieldNode(
                                    name: NameNode(value: '__typename'),
                                    alias: null,
                                    arguments: [],
                                    directives: [],
                                    selectionSet: null,
                                  ),
                                ],
                              ),
                            ),
                            FieldNode(
                              name: NameNode(value: 'createdAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'updatedAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Mutation$UpsertProjectSetup _parserFn$Mutation$UpsertProjectSetup(
  Map<String, dynamic> data,
) => Mutation$UpsertProjectSetup.fromJson(data);
typedef OnMutationCompleted$Mutation$UpsertProjectSetup =
    FutureOr<void> Function(
      Map<String, dynamic>?,
      Mutation$UpsertProjectSetup?,
    );

class Options$Mutation$UpsertProjectSetup
    extends graphql.MutationOptions<Mutation$UpsertProjectSetup> {
  Options$Mutation$UpsertProjectSetup({
    String? operationName,
    required Variables$Mutation$UpsertProjectSetup variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Mutation$UpsertProjectSetup? typedOptimisticResult,
    graphql.Context? context,
    OnMutationCompleted$Mutation$UpsertProjectSetup? onCompleted,
    graphql.OnMutationUpdate<Mutation$UpsertProjectSetup>? update,
    graphql.OnError? onError,
  }) : onCompletedWithParsed = onCompleted,
       super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         onCompleted: onCompleted == null
             ? null
             : (data) => onCompleted(
                 data,
                 data == null
                     ? null
                     : _parserFn$Mutation$UpsertProjectSetup(data),
               ),
         update: update,
         onError: onError,
         document: documentNodeMutationUpsertProjectSetup,
         parserFn: _parserFn$Mutation$UpsertProjectSetup,
       );

  final OnMutationCompleted$Mutation$UpsertProjectSetup? onCompletedWithParsed;

  @override
  List<Object?> get properties => [
    ...super.onCompleted == null
        ? super.properties
        : super.properties.where((property) => property != onCompleted),
    onCompletedWithParsed,
  ];
}

class WatchOptions$Mutation$UpsertProjectSetup
    extends graphql.WatchQueryOptions<Mutation$UpsertProjectSetup> {
  WatchOptions$Mutation$UpsertProjectSetup({
    String? operationName,
    required Variables$Mutation$UpsertProjectSetup variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Mutation$UpsertProjectSetup? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeMutationUpsertProjectSetup,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Mutation$UpsertProjectSetup,
       );
}

extension ClientExtension$Mutation$UpsertProjectSetup on graphql.GraphQLClient {
  Future<graphql.QueryResult<Mutation$UpsertProjectSetup>>
  mutate$UpsertProjectSetup(
    Options$Mutation$UpsertProjectSetup options,
  ) async => await this.mutate(options);

  graphql.ObservableQuery<Mutation$UpsertProjectSetup>
  watchMutation$UpsertProjectSetup(
    WatchOptions$Mutation$UpsertProjectSetup options,
  ) => this.watchMutation(options);
}

class Mutation$UpsertProjectSetup$upsertProjectSetup {
  Mutation$UpsertProjectSetup$upsertProjectSetup({required this.$__typename});

  factory Mutation$UpsertProjectSetup$upsertProjectSetup.fromJson(
    Map<String, dynamic> json,
  ) {
    switch (json["__typename"] as String) {
      case "UpsertProjectSetupSuccess":
        return Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess.fromJson(
          json,
        );

      case "GraphError":
        return Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError.fromJson(
          json,
        );

      default:
        final l$$__typename = json['__typename'];
        return Mutation$UpsertProjectSetup$upsertProjectSetup(
          $__typename: (l$$__typename as String),
        );
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Mutation$UpsertProjectSetup$upsertProjectSetup ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup$upsertProjectSetup
    on Mutation$UpsertProjectSetup$upsertProjectSetup {
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<
    Mutation$UpsertProjectSetup$upsertProjectSetup
  >
  get copyWith =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup(this, (i) => i);

  _T when<_T>({
    required _T Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess,
    )
    upsertProjectSetupSuccess,
    required _T Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError,
    )
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "UpsertProjectSetupSuccess":
        return upsertProjectSetupSuccess(
          this
              as Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess,
        );

      case "GraphError":
        return graphError(
          this as Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError,
        );

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess,
    )?
    upsertProjectSetupSuccess,
    _T Function(Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError)?
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "UpsertProjectSetupSuccess":
        if (upsertProjectSetupSuccess != null) {
          return upsertProjectSetupSuccess(
            this
                as Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(
            this as Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError,
          );
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes> {
  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup(
    Mutation$UpsertProjectSetup$upsertProjectSetup instance,
    TRes Function(Mutation$UpsertProjectSetup$upsertProjectSetup) then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup;

  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes>
    implements CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes> {
  _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup(
    this._instance,
    this._then,
  );

  final Mutation$UpsertProjectSetup$upsertProjectSetup _instance;

  final TRes Function(Mutation$UpsertProjectSetup$upsertProjectSetup) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Mutation$UpsertProjectSetup$upsertProjectSetup(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes>
    implements CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup<TRes> {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup(this._res);

  TRes _res;

  call({String? $__typename}) => _res;
}

class Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess
    implements Mutation$UpsertProjectSetup$upsertProjectSetup {
  Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess({
    required this.project,
    this.$__typename = 'UpsertProjectSetupSuccess',
  });

  factory Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$project = json['project'];
    final l$$__typename = json['__typename'];
    return Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess(
      project:
          Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project.fromJson(
            (l$project as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project
  project;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$project = project;
    _resultData['project'] = l$project.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$project = project;
    final l$$__typename = $__typename;
    return Object.hashAll([l$project, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$project = project;
    final lOther$project = other.project;
    if (l$project != lOther$project) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess
    on Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess {
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess
  >
  get copyWith =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess<
  TRes
> {
  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess
    instance,
    TRes Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess,
    )
    then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess;

  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess;

  TRes call({
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project?
    project,
    String? $__typename,
  });
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
    TRes
  >
  get project;
}

class _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess<
          TRes
        > {
  _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess(
    this._instance,
    this._then,
  );

  final Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess
  _instance;

  final TRes Function(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? project = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess(
      project: project == _undefined || project == null
          ? _instance.project
          : (project
                as Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
    TRes
  >
  get project {
    final local$project = _instance.project;
    return CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
      local$project,
      (e) => call(project: e),
    );
  }
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess<
          TRes
        > {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess(
    this._res,
  );

  TRes _res;

  call({
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project?
    project,
    String? $__typename,
  }) => _res;

  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
    TRes
  >
  get project =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project.stub(
        _res,
      );
}

class Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project {
  Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project({
    required this.projectID,
    required this.projectName,
    required this.repositories,
    required this.boards,
    required this.createdAt,
    required this.updatedAt,
    this.$__typename = 'ProjectSetup',
  });

  factory Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$projectID = json['projectID'];
    final l$projectName = json['projectName'];
    final l$repositories = json['repositories'];
    final l$boards = json['boards'];
    final l$createdAt = json['createdAt'];
    final l$updatedAt = json['updatedAt'];
    final l$$__typename = json['__typename'];
    return Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
      projectID: (l$projectID as String),
      projectName: (l$projectName as String),
      repositories: (l$repositories as List<dynamic>)
          .map(
            (e) =>
                Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      boards: (l$boards as List<dynamic>)
          .map(
            (e) =>
                Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards.fromJson(
                  (e as Map<String, dynamic>),
                ),
          )
          .toList(),
      createdAt: dateTimeFromJson(l$createdAt),
      updatedAt: dateTimeFromJson(l$updatedAt),
      $__typename: (l$$__typename as String),
    );
  }

  final String projectID;

  final String projectName;

  final List<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
  >
  repositories;

  final List<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
  >
  boards;

  final DateTime createdAt;

  final DateTime updatedAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$projectID = projectID;
    _resultData['projectID'] = l$projectID;
    final l$projectName = projectName;
    _resultData['projectName'] = l$projectName;
    final l$repositories = repositories;
    _resultData['repositories'] = l$repositories
        .map((e) => e.toJson())
        .toList();
    final l$boards = boards;
    _resultData['boards'] = l$boards.map((e) => e.toJson()).toList();
    final l$createdAt = createdAt;
    _resultData['createdAt'] = dateTimeToJson(l$createdAt);
    final l$updatedAt = updatedAt;
    _resultData['updatedAt'] = dateTimeToJson(l$updatedAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$projectID = projectID;
    final l$projectName = projectName;
    final l$repositories = repositories;
    final l$boards = boards;
    final l$createdAt = createdAt;
    final l$updatedAt = updatedAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$projectID,
      l$projectName,
      Object.hashAll(l$repositories.map((v) => v)),
      Object.hashAll(l$boards.map((v) => v)),
      l$createdAt,
      l$updatedAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$projectID = projectID;
    final lOther$projectID = other.projectID;
    if (l$projectID != lOther$projectID) {
      return false;
    }
    final l$projectName = projectName;
    final lOther$projectName = other.projectName;
    if (l$projectName != lOther$projectName) {
      return false;
    }
    final l$repositories = repositories;
    final lOther$repositories = other.repositories;
    if (l$repositories.length != lOther$repositories.length) {
      return false;
    }
    for (int i = 0; i < l$repositories.length; i++) {
      final l$repositories$entry = l$repositories[i];
      final lOther$repositories$entry = lOther$repositories[i];
      if (l$repositories$entry != lOther$repositories$entry) {
        return false;
      }
    }
    final l$boards = boards;
    final lOther$boards = other.boards;
    if (l$boards.length != lOther$boards.length) {
      return false;
    }
    for (int i = 0; i < l$boards.length; i++) {
      final l$boards$entry = l$boards[i];
      final lOther$boards$entry = lOther$boards[i];
      if (l$boards$entry != lOther$boards$entry) {
        return false;
      }
    }
    final l$createdAt = createdAt;
    final lOther$createdAt = other.createdAt;
    if (l$createdAt != lOther$createdAt) {
      return false;
    }
    final l$updatedAt = updatedAt;
    final lOther$updatedAt = other.updatedAt;
    if (l$updatedAt != lOther$updatedAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project
    on
        Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project {
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project
  >
  get copyWith =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
  TRes
> {
  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project
    instance,
    TRes Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project,
    )
    then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project;

  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project;

  TRes call({
    String? projectID,
    String? projectName,
    List<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
    >?
    repositories,
    List<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
    >?
    boards,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? $__typename,
  });
  TRes repositories(
    Iterable<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
    >
    Function(
      Iterable<
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
          Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
        >
      >,
    )
    _fn,
  );
  TRes boards(
    Iterable<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
    >
    Function(
      Iterable<
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
          Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
        >
      >,
    )
    _fn,
  );
}

class _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
          TRes
        > {
  _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
    this._instance,
    this._then,
  );

  final Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project
  _instance;

  final TRes Function(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? projectID = _undefined,
    Object? projectName = _undefined,
    Object? repositories = _undefined,
    Object? boards = _undefined,
    Object? createdAt = _undefined,
    Object? updatedAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
      projectID: projectID == _undefined || projectID == null
          ? _instance.projectID
          : (projectID as String),
      projectName: projectName == _undefined || projectName == null
          ? _instance.projectName
          : (projectName as String),
      repositories: repositories == _undefined || repositories == null
          ? _instance.repositories
          : (repositories
                as List<
                  Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
                >),
      boards: boards == _undefined || boards == null
          ? _instance.boards
          : (boards
                as List<
                  Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
                >),
      createdAt: createdAt == _undefined || createdAt == null
          ? _instance.createdAt
          : (createdAt as DateTime),
      updatedAt: updatedAt == _undefined || updatedAt == null
          ? _instance.updatedAt
          : (updatedAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  TRes repositories(
    Iterable<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
    >
    Function(
      Iterable<
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
          Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
        >
      >,
    )
    _fn,
  ) => call(
    repositories: _fn(
      _instance.repositories.map(
        (e) =>
            CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );

  TRes boards(
    Iterable<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
    >
    Function(
      Iterable<
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
          Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
        >
      >,
    )
    _fn,
  ) => call(
    boards: _fn(
      _instance.boards.map(
        (e) =>
            CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
              e,
              (i) => i,
            ),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project<
          TRes
        > {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project(
    this._res,
  );

  TRes _res;

  call({
    String? projectID,
    String? projectName,
    List<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
    >?
    repositories,
    List<
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
    >?
    boards,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? $__typename,
  }) => _res;

  repositories(_fn) => _res;

  boards(_fn) => _res;
}

class Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories {
  Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories({
    required this.repositoryID,
    required this.scmProvider,
    required this.repositoryURL,
    required this.isPrimary,
    this.$__typename = 'ProjectRepository',
  });

  factory Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$repositoryID = json['repositoryID'];
    final l$scmProvider = json['scmProvider'];
    final l$repositoryURL = json['repositoryURL'];
    final l$isPrimary = json['isPrimary'];
    final l$$__typename = json['__typename'];
    return Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
      repositoryID: (l$repositoryID as String),
      scmProvider: fromJson$Enum$SCMProvider((l$scmProvider as String)),
      repositoryURL: (l$repositoryURL as String),
      isPrimary: (l$isPrimary as bool),
      $__typename: (l$$__typename as String),
    );
  }

  final String repositoryID;

  final Enum$SCMProvider scmProvider;

  final String repositoryURL;

  final bool isPrimary;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$repositoryID = repositoryID;
    _resultData['repositoryID'] = l$repositoryID;
    final l$scmProvider = scmProvider;
    _resultData['scmProvider'] = toJson$Enum$SCMProvider(l$scmProvider);
    final l$repositoryURL = repositoryURL;
    _resultData['repositoryURL'] = l$repositoryURL;
    final l$isPrimary = isPrimary;
    _resultData['isPrimary'] = l$isPrimary;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$repositoryID = repositoryID;
    final l$scmProvider = scmProvider;
    final l$repositoryURL = repositoryURL;
    final l$isPrimary = isPrimary;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$repositoryID,
      l$scmProvider,
      l$repositoryURL,
      l$isPrimary,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$repositoryID = repositoryID;
    final lOther$repositoryID = other.repositoryID;
    if (l$repositoryID != lOther$repositoryID) {
      return false;
    }
    final l$scmProvider = scmProvider;
    final lOther$scmProvider = other.scmProvider;
    if (l$scmProvider != lOther$scmProvider) {
      return false;
    }
    final l$repositoryURL = repositoryURL;
    final lOther$repositoryURL = other.repositoryURL;
    if (l$repositoryURL != lOther$repositoryURL) {
      return false;
    }
    final l$isPrimary = isPrimary;
    final lOther$isPrimary = other.isPrimary;
    if (l$isPrimary != lOther$isPrimary) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
    on
        Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories {
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
  >
  get copyWith =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
  TRes
> {
  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
    instance,
    TRes Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories,
    )
    then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories;

  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories;

  TRes call({
    String? repositoryID,
    Enum$SCMProvider? scmProvider,
    String? repositoryURL,
    bool? isPrimary,
    String? $__typename,
  });
}

class _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
          TRes
        > {
  _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
    this._instance,
    this._then,
  );

  final Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories
  _instance;

  final TRes Function(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? repositoryID = _undefined,
    Object? scmProvider = _undefined,
    Object? repositoryURL = _undefined,
    Object? isPrimary = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
      repositoryID: repositoryID == _undefined || repositoryID == null
          ? _instance.repositoryID
          : (repositoryID as String),
      scmProvider: scmProvider == _undefined || scmProvider == null
          ? _instance.scmProvider
          : (scmProvider as Enum$SCMProvider),
      repositoryURL: repositoryURL == _undefined || repositoryURL == null
          ? _instance.repositoryURL
          : (repositoryURL as String),
      isPrimary: isPrimary == _undefined || isPrimary == null
          ? _instance.isPrimary
          : (isPrimary as bool),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories<
          TRes
        > {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$repositories(
    this._res,
  );

  TRes _res;

  call({
    String? repositoryID,
    Enum$SCMProvider? scmProvider,
    String? repositoryURL,
    bool? isPrimary,
    String? $__typename,
  }) => _res;
}

class Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards {
  Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards({
    required this.boardID,
    required this.trackerProvider,
    this.taskboardName,
    required this.appliesToAllRepositories,
    required this.repositoryIDs,
    this.$__typename = 'ProjectBoard',
  });

  factory Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$boardID = json['boardID'];
    final l$trackerProvider = json['trackerProvider'];
    final l$taskboardName = json['taskboardName'];
    final l$appliesToAllRepositories = json['appliesToAllRepositories'];
    final l$repositoryIDs = json['repositoryIDs'];
    final l$$__typename = json['__typename'];
    return Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
      boardID: (l$boardID as String),
      trackerProvider: fromJson$Enum$TrackerSourceKind(
        (l$trackerProvider as String),
      ),
      taskboardName: (l$taskboardName as String?),
      appliesToAllRepositories: (l$appliesToAllRepositories as bool),
      repositoryIDs: (l$repositoryIDs as List<dynamic>)
          .map((e) => (e as String))
          .toList(),
      $__typename: (l$$__typename as String),
    );
  }

  final String boardID;

  final Enum$TrackerSourceKind trackerProvider;

  final String? taskboardName;

  final bool appliesToAllRepositories;

  final List<String> repositoryIDs;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$boardID = boardID;
    _resultData['boardID'] = l$boardID;
    final l$trackerProvider = trackerProvider;
    _resultData['trackerProvider'] = toJson$Enum$TrackerSourceKind(
      l$trackerProvider,
    );
    final l$taskboardName = taskboardName;
    _resultData['taskboardName'] = l$taskboardName;
    final l$appliesToAllRepositories = appliesToAllRepositories;
    _resultData['appliesToAllRepositories'] = l$appliesToAllRepositories;
    final l$repositoryIDs = repositoryIDs;
    _resultData['repositoryIDs'] = l$repositoryIDs.map((e) => e).toList();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$boardID = boardID;
    final l$trackerProvider = trackerProvider;
    final l$taskboardName = taskboardName;
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final l$repositoryIDs = repositoryIDs;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$boardID,
      l$trackerProvider,
      l$taskboardName,
      l$appliesToAllRepositories,
      Object.hashAll(l$repositoryIDs.map((v) => v)),
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$boardID = boardID;
    final lOther$boardID = other.boardID;
    if (l$boardID != lOther$boardID) {
      return false;
    }
    final l$trackerProvider = trackerProvider;
    final lOther$trackerProvider = other.trackerProvider;
    if (l$trackerProvider != lOther$trackerProvider) {
      return false;
    }
    final l$taskboardName = taskboardName;
    final lOther$taskboardName = other.taskboardName;
    if (l$taskboardName != lOther$taskboardName) {
      return false;
    }
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final lOther$appliesToAllRepositories = other.appliesToAllRepositories;
    if (l$appliesToAllRepositories != lOther$appliesToAllRepositories) {
      return false;
    }
    final l$repositoryIDs = repositoryIDs;
    final lOther$repositoryIDs = other.repositoryIDs;
    if (l$repositoryIDs.length != lOther$repositoryIDs.length) {
      return false;
    }
    for (int i = 0; i < l$repositoryIDs.length; i++) {
      final l$repositoryIDs$entry = l$repositoryIDs[i];
      final lOther$repositoryIDs$entry = lOther$repositoryIDs[i];
      if (l$repositoryIDs$entry != lOther$repositoryIDs$entry) {
        return false;
      }
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
    on
        Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards {
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
  >
  get copyWith =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
  TRes
> {
  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
    instance,
    TRes Function(
      Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards,
    )
    then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards;

  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards;

  TRes call({
    String? boardID,
    Enum$TrackerSourceKind? trackerProvider,
    String? taskboardName,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
    String? $__typename,
  });
}

class _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
          TRes
        > {
  _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
    this._instance,
    this._then,
  );

  final Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards
  _instance;

  final TRes Function(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? boardID = _undefined,
    Object? trackerProvider = _undefined,
    Object? taskboardName = _undefined,
    Object? appliesToAllRepositories = _undefined,
    Object? repositoryIDs = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
      boardID: boardID == _undefined || boardID == null
          ? _instance.boardID
          : (boardID as String),
      trackerProvider: trackerProvider == _undefined || trackerProvider == null
          ? _instance.trackerProvider
          : (trackerProvider as Enum$TrackerSourceKind),
      taskboardName: taskboardName == _undefined
          ? _instance.taskboardName
          : (taskboardName as String?),
      appliesToAllRepositories:
          appliesToAllRepositories == _undefined ||
              appliesToAllRepositories == null
          ? _instance.appliesToAllRepositories
          : (appliesToAllRepositories as bool),
      repositoryIDs: repositoryIDs == _undefined || repositoryIDs == null
          ? _instance.repositoryIDs
          : (repositoryIDs as List<String>),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards<
          TRes
        > {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$UpsertProjectSetupSuccess$project$boards(
    this._res,
  );

  TRes _res;

  call({
    String? boardID,
    Enum$TrackerSourceKind? trackerProvider,
    String? taskboardName,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
    String? $__typename,
  }) => _res;
}

class Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError
    implements Mutation$UpsertProjectSetup$upsertProjectSetup {
  Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError
    on Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError {
  CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError<
    Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError
  >
  get copyWith =>
      CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError<
  TRes
> {
  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError instance,
    TRes Function(Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError)
    then,
  ) = _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError;

  factory CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError.stub(
    TRes res,
  ) = _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError<
          TRes
        > {
  _CopyWithImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError(
    this._instance,
    this._then,
  );

  final Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError _instance;

  final TRes Function(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError<
  TRes
>
    implements
        CopyWith$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError<
          TRes
        > {
  _CopyWithStubImpl$Mutation$UpsertProjectSetup$upsertProjectSetup$$GraphError(
    this._res,
  );

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}

class Variables$Subscription$SessionActivity {
  factory Variables$Subscription$SessionActivity({
    required String runID,
    required String taskID,
    required String jobID,
    required int fromOffset,
  }) => Variables$Subscription$SessionActivity._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
    r'fromOffset': fromOffset,
  });

  Variables$Subscription$SessionActivity._(this._$data);

  factory Variables$Subscription$SessionActivity.fromJson(
    Map<String, dynamic> data,
  ) {
    final result$data = <String, dynamic>{};
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    final l$jobID = data['jobID'];
    result$data['jobID'] = (l$jobID as String);
    final l$fromOffset = data['fromOffset'];
    result$data['fromOffset'] = (l$fromOffset as int);
    return Variables$Subscription$SessionActivity._(result$data);
  }

  Map<String, dynamic> _$data;

  String get runID => (_$data['runID'] as String);

  String get taskID => (_$data['taskID'] as String);

  String get jobID => (_$data['jobID'] as String);

  int get fromOffset => (_$data['fromOffset'] as int);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$runID = runID;
    result$data['runID'] = l$runID;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    final l$jobID = jobID;
    result$data['jobID'] = l$jobID;
    final l$fromOffset = fromOffset;
    result$data['fromOffset'] = l$fromOffset;
    return result$data;
  }

  CopyWith$Variables$Subscription$SessionActivity<
    Variables$Subscription$SessionActivity
  >
  get copyWith =>
      CopyWith$Variables$Subscription$SessionActivity(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Variables$Subscription$SessionActivity ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$runID = runID;
    final lOther$runID = other.runID;
    if (l$runID != lOther$runID) {
      return false;
    }
    final l$taskID = taskID;
    final lOther$taskID = other.taskID;
    if (l$taskID != lOther$taskID) {
      return false;
    }
    final l$jobID = jobID;
    final lOther$jobID = other.jobID;
    if (l$jobID != lOther$jobID) {
      return false;
    }
    final l$fromOffset = fromOffset;
    final lOther$fromOffset = other.fromOffset;
    if (l$fromOffset != lOther$fromOffset) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    final l$fromOffset = fromOffset;
    return Object.hashAll([l$runID, l$taskID, l$jobID, l$fromOffset]);
  }
}

abstract class CopyWith$Variables$Subscription$SessionActivity<TRes> {
  factory CopyWith$Variables$Subscription$SessionActivity(
    Variables$Subscription$SessionActivity instance,
    TRes Function(Variables$Subscription$SessionActivity) then,
  ) = _CopyWithImpl$Variables$Subscription$SessionActivity;

  factory CopyWith$Variables$Subscription$SessionActivity.stub(TRes res) =
      _CopyWithStubImpl$Variables$Subscription$SessionActivity;

  TRes call({String? runID, String? taskID, String? jobID, int? fromOffset});
}

class _CopyWithImpl$Variables$Subscription$SessionActivity<TRes>
    implements CopyWith$Variables$Subscription$SessionActivity<TRes> {
  _CopyWithImpl$Variables$Subscription$SessionActivity(
    this._instance,
    this._then,
  );

  final Variables$Subscription$SessionActivity _instance;

  final TRes Function(Variables$Subscription$SessionActivity) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
    Object? fromOffset = _undefined,
  }) => _then(
    Variables$Subscription$SessionActivity._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
      if (fromOffset != _undefined && fromOffset != null)
        'fromOffset': (fromOffset as int),
    }),
  );
}

class _CopyWithStubImpl$Variables$Subscription$SessionActivity<TRes>
    implements CopyWith$Variables$Subscription$SessionActivity<TRes> {
  _CopyWithStubImpl$Variables$Subscription$SessionActivity(this._res);

  TRes _res;

  call({String? runID, String? taskID, String? jobID, int? fromOffset}) => _res;
}

class Subscription$SessionActivity {
  Subscription$SessionActivity({
    required this.sessionActivityStream,
    this.$__typename = 'Subscription',
  });

  factory Subscription$SessionActivity.fromJson(Map<String, dynamic> json) {
    final l$sessionActivityStream = json['sessionActivityStream'];
    final l$$__typename = json['__typename'];
    return Subscription$SessionActivity(
      sessionActivityStream:
          Subscription$SessionActivity$sessionActivityStream.fromJson(
            (l$sessionActivityStream as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Subscription$SessionActivity$sessionActivityStream
  sessionActivityStream;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$sessionActivityStream = sessionActivityStream;
    _resultData['sessionActivityStream'] = l$sessionActivityStream.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$sessionActivityStream = sessionActivityStream;
    final l$$__typename = $__typename;
    return Object.hashAll([l$sessionActivityStream, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Subscription$SessionActivity ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$sessionActivityStream = sessionActivityStream;
    final lOther$sessionActivityStream = other.sessionActivityStream;
    if (l$sessionActivityStream != lOther$sessionActivityStream) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Subscription$SessionActivity
    on Subscription$SessionActivity {
  CopyWith$Subscription$SessionActivity<Subscription$SessionActivity>
  get copyWith => CopyWith$Subscription$SessionActivity(this, (i) => i);
}

abstract class CopyWith$Subscription$SessionActivity<TRes> {
  factory CopyWith$Subscription$SessionActivity(
    Subscription$SessionActivity instance,
    TRes Function(Subscription$SessionActivity) then,
  ) = _CopyWithImpl$Subscription$SessionActivity;

  factory CopyWith$Subscription$SessionActivity.stub(TRes res) =
      _CopyWithStubImpl$Subscription$SessionActivity;

  TRes call({
    Subscription$SessionActivity$sessionActivityStream? sessionActivityStream,
    String? $__typename,
  });
  CopyWith$Subscription$SessionActivity$sessionActivityStream<TRes>
  get sessionActivityStream;
}

class _CopyWithImpl$Subscription$SessionActivity<TRes>
    implements CopyWith$Subscription$SessionActivity<TRes> {
  _CopyWithImpl$Subscription$SessionActivity(this._instance, this._then);

  final Subscription$SessionActivity _instance;

  final TRes Function(Subscription$SessionActivity) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? sessionActivityStream = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Subscription$SessionActivity(
      sessionActivityStream:
          sessionActivityStream == _undefined || sessionActivityStream == null
          ? _instance.sessionActivityStream
          : (sessionActivityStream
                as Subscription$SessionActivity$sessionActivityStream),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Subscription$SessionActivity$sessionActivityStream<TRes>
  get sessionActivityStream {
    final local$sessionActivityStream = _instance.sessionActivityStream;
    return CopyWith$Subscription$SessionActivity$sessionActivityStream(
      local$sessionActivityStream,
      (e) => call(sessionActivityStream: e),
    );
  }
}

class _CopyWithStubImpl$Subscription$SessionActivity<TRes>
    implements CopyWith$Subscription$SessionActivity<TRes> {
  _CopyWithStubImpl$Subscription$SessionActivity(this._res);

  TRes _res;

  call({
    Subscription$SessionActivity$sessionActivityStream? sessionActivityStream,
    String? $__typename,
  }) => _res;

  CopyWith$Subscription$SessionActivity$sessionActivityStream<TRes>
  get sessionActivityStream =>
      CopyWith$Subscription$SessionActivity$sessionActivityStream.stub(_res);
}

const documentNodeSubscriptionSessionActivity = DocumentNode(
  definitions: [
    OperationDefinitionNode(
      type: OperationType.subscription,
      name: NameNode(value: 'SessionActivity'),
      variableDefinitions: [
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'runID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'taskID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'jobID')),
          type: NamedTypeNode(name: NameNode(value: 'String'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
        VariableDefinitionNode(
          variable: VariableNode(name: NameNode(value: 'fromOffset')),
          type: NamedTypeNode(name: NameNode(value: 'Int'), isNonNull: true),
          defaultValue: DefaultValueNode(value: null),
          directives: [],
        ),
      ],
      directives: [],
      selectionSet: SelectionSetNode(
        selections: [
          FieldNode(
            name: NameNode(value: 'sessionActivityStream'),
            alias: null,
            arguments: [
              ArgumentNode(
                name: NameNode(value: 'correlation'),
                value: ObjectValueNode(
                  fields: [
                    ObjectFieldNode(
                      name: NameNode(value: 'runID'),
                      value: VariableNode(name: NameNode(value: 'runID')),
                    ),
                    ObjectFieldNode(
                      name: NameNode(value: 'taskID'),
                      value: VariableNode(name: NameNode(value: 'taskID')),
                    ),
                    ObjectFieldNode(
                      name: NameNode(value: 'jobID'),
                      value: VariableNode(name: NameNode(value: 'jobID')),
                    ),
                  ],
                ),
              ),
              ArgumentNode(
                name: NameNode(value: 'fromOffset'),
                value: VariableNode(name: NameNode(value: 'fromOffset')),
              ),
            ],
            directives: [],
            selectionSet: SelectionSetNode(
              selections: [
                FieldNode(
                  name: NameNode(value: '__typename'),
                  alias: null,
                  arguments: [],
                  directives: [],
                  selectionSet: null,
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'StreamEventSuccess'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'event'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: SelectionSetNode(
                          selections: [
                            FieldNode(
                              name: NameNode(value: 'eventID'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'eventType'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'source'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'payload'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: 'occurredAt'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                            FieldNode(
                              name: NameNode(value: '__typename'),
                              alias: null,
                              arguments: [],
                              directives: [],
                              selectionSet: null,
                            ),
                          ],
                        ),
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
                InlineFragmentNode(
                  typeCondition: TypeConditionNode(
                    on: NamedTypeNode(
                      name: NameNode(value: 'GraphError'),
                      isNonNull: false,
                    ),
                  ),
                  directives: [],
                  selectionSet: SelectionSetNode(
                    selections: [
                      FieldNode(
                        name: NameNode(value: 'code'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'message'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: 'field'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                      FieldNode(
                        name: NameNode(value: '__typename'),
                        alias: null,
                        arguments: [],
                        directives: [],
                        selectionSet: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          FieldNode(
            name: NameNode(value: '__typename'),
            alias: null,
            arguments: [],
            directives: [],
            selectionSet: null,
          ),
        ],
      ),
    ),
  ],
);
Subscription$SessionActivity _parserFn$Subscription$SessionActivity(
  Map<String, dynamic> data,
) => Subscription$SessionActivity.fromJson(data);

class Options$Subscription$SessionActivity
    extends graphql.SubscriptionOptions<Subscription$SessionActivity> {
  Options$Subscription$SessionActivity({
    String? operationName,
    required Variables$Subscription$SessionActivity variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Subscription$SessionActivity? typedOptimisticResult,
    graphql.Context? context,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeSubscriptionSessionActivity,
         parserFn: _parserFn$Subscription$SessionActivity,
       );
}

class WatchOptions$Subscription$SessionActivity
    extends graphql.WatchQueryOptions<Subscription$SessionActivity> {
  WatchOptions$Subscription$SessionActivity({
    String? operationName,
    required Variables$Subscription$SessionActivity variables,
    graphql.FetchPolicy? fetchPolicy,
    graphql.ErrorPolicy? errorPolicy,
    graphql.CacheRereadPolicy? cacheRereadPolicy,
    Object? optimisticResult,
    Subscription$SessionActivity? typedOptimisticResult,
    graphql.Context? context,
    Duration? pollInterval,
    bool? eagerlyFetchResults,
    bool carryForwardDataOnException = true,
    bool fetchResults = false,
  }) : super(
         variables: variables.toJson(),
         operationName: operationName,
         fetchPolicy: fetchPolicy,
         errorPolicy: errorPolicy,
         cacheRereadPolicy: cacheRereadPolicy,
         optimisticResult: optimisticResult ?? typedOptimisticResult?.toJson(),
         context: context,
         document: documentNodeSubscriptionSessionActivity,
         pollInterval: pollInterval,
         eagerlyFetchResults: eagerlyFetchResults,
         carryForwardDataOnException: carryForwardDataOnException,
         fetchResults: fetchResults,
         parserFn: _parserFn$Subscription$SessionActivity,
       );
}

class FetchMoreOptions$Subscription$SessionActivity
    extends graphql.FetchMoreOptions {
  FetchMoreOptions$Subscription$SessionActivity({
    required graphql.UpdateQuery updateQuery,
    required Variables$Subscription$SessionActivity variables,
  }) : super(
         updateQuery: updateQuery,
         variables: variables.toJson(),
         document: documentNodeSubscriptionSessionActivity,
       );
}

extension ClientExtension$Subscription$SessionActivity
    on graphql.GraphQLClient {
  Stream<graphql.QueryResult<Subscription$SessionActivity>>
  subscribe$SessionActivity(Options$Subscription$SessionActivity options) =>
      this.subscribe(options);

  graphql.ObservableQuery<Subscription$SessionActivity>
  watchSubscription$SessionActivity(
    WatchOptions$Subscription$SessionActivity options,
  ) => this.watchQuery(options);
}

class Subscription$SessionActivity$sessionActivityStream {
  Subscription$SessionActivity$sessionActivityStream({
    required this.$__typename,
  });

  factory Subscription$SessionActivity$sessionActivityStream.fromJson(
    Map<String, dynamic> json,
  ) {
    switch (json["__typename"] as String) {
      case "StreamEventSuccess":
        return Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess.fromJson(
          json,
        );

      case "GraphError":
        return Subscription$SessionActivity$sessionActivityStream$$GraphError.fromJson(
          json,
        );

      default:
        final l$$__typename = json['__typename'];
        return Subscription$SessionActivity$sessionActivityStream(
          $__typename: (l$$__typename as String),
        );
    }
  }

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$$__typename = $__typename;
    return Object.hashAll([l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Subscription$SessionActivity$sessionActivityStream ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Subscription$SessionActivity$sessionActivityStream
    on Subscription$SessionActivity$sessionActivityStream {
  CopyWith$Subscription$SessionActivity$sessionActivityStream<
    Subscription$SessionActivity$sessionActivityStream
  >
  get copyWith => CopyWith$Subscription$SessionActivity$sessionActivityStream(
    this,
    (i) => i,
  );

  _T when<_T>({
    required _T Function(
      Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess,
    )
    streamEventSuccess,
    required _T Function(
      Subscription$SessionActivity$sessionActivityStream$$GraphError,
    )
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "StreamEventSuccess":
        return streamEventSuccess(
          this
              as Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess,
        );

      case "GraphError":
        return graphError(
          this
              as Subscription$SessionActivity$sessionActivityStream$$GraphError,
        );

      default:
        return orElse();
    }
  }

  _T maybeWhen<_T>({
    _T Function(
      Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess,
    )?
    streamEventSuccess,
    _T Function(Subscription$SessionActivity$sessionActivityStream$$GraphError)?
    graphError,
    required _T Function() orElse,
  }) {
    switch ($__typename) {
      case "StreamEventSuccess":
        if (streamEventSuccess != null) {
          return streamEventSuccess(
            this
                as Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess,
          );
        } else {
          return orElse();
        }

      case "GraphError":
        if (graphError != null) {
          return graphError(
            this
                as Subscription$SessionActivity$sessionActivityStream$$GraphError,
          );
        } else {
          return orElse();
        }

      default:
        return orElse();
    }
  }
}

abstract class CopyWith$Subscription$SessionActivity$sessionActivityStream<
  TRes
> {
  factory CopyWith$Subscription$SessionActivity$sessionActivityStream(
    Subscription$SessionActivity$sessionActivityStream instance,
    TRes Function(Subscription$SessionActivity$sessionActivityStream) then,
  ) = _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream;

  factory CopyWith$Subscription$SessionActivity$sessionActivityStream.stub(
    TRes res,
  ) = _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream;

  TRes call({String? $__typename});
}

class _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream<TRes>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream<TRes> {
  _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream(
    this._instance,
    this._then,
  );

  final Subscription$SessionActivity$sessionActivityStream _instance;

  final TRes Function(Subscription$SessionActivity$sessionActivityStream) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? $__typename = _undefined}) => _then(
    Subscription$SessionActivity$sessionActivityStream(
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream<TRes>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream<TRes> {
  _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream(
    this._res,
  );

  TRes _res;

  call({String? $__typename}) => _res;
}

class Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess
    implements Subscription$SessionActivity$sessionActivityStream {
  Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess({
    required this.event,
    this.$__typename = 'StreamEventSuccess',
  });

  factory Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$event = json['event'];
    final l$$__typename = json['__typename'];
    return Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess(
      event:
          Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event.fromJson(
            (l$event as Map<String, dynamic>),
          ),
      $__typename: (l$$__typename as String),
    );
  }

  final Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event
  event;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$event = event;
    _resultData['event'] = l$event.toJson();
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$event = event;
    final l$$__typename = $__typename;
    return Object.hashAll([l$event, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$event = event;
    final lOther$event = other.event;
    if (l$event != lOther$event) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess
    on Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess {
  CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess<
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess
  >
  get copyWith =>
      CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess<
  TRes
> {
  factory CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess(
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess
    instance,
    TRes Function(
      Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess,
    )
    then,
  ) = _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess;

  factory CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess.stub(
    TRes res,
  ) = _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess;

  TRes call({
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event?
    event,
    String? $__typename,
  });
  CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
    TRes
  >
  get event;
}

class _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess<
  TRes
>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess<
          TRes
        > {
  _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess(
    this._instance,
    this._then,
  );

  final Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess
  _instance;

  final TRes Function(
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? event = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess(
      event: event == _undefined || event == null
          ? _instance.event
          : (event
                as Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );

  CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
    TRes
  >
  get event {
    final local$event = _instance.event;
    return CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
      local$event,
      (e) => call(event: e),
    );
  }
}

class _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess<
  TRes
>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess<
          TRes
        > {
  _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess(
    this._res,
  );

  TRes _res;

  call({
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event?
    event,
    String? $__typename,
  }) => _res;

  CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
    TRes
  >
  get event =>
      CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event.stub(
        _res,
      );
}

class Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event {
  Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event({
    required this.eventID,
    required this.eventType,
    required this.source,
    required this.payload,
    required this.occurredAt,
    this.$__typename = 'StreamEvent',
  });

  factory Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$eventID = json['eventID'];
    final l$eventType = json['eventType'];
    final l$source = json['source'];
    final l$payload = json['payload'];
    final l$occurredAt = json['occurredAt'];
    final l$$__typename = json['__typename'];
    return Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
      eventID: (l$eventID as String),
      eventType: (l$eventType as String),
      source: fromJson$Enum$StreamEventSource((l$source as String)),
      payload: (l$payload as String),
      occurredAt: dateTimeFromJson(l$occurredAt),
      $__typename: (l$$__typename as String),
    );
  }

  final String eventID;

  final String eventType;

  final Enum$StreamEventSource source;

  final String payload;

  final DateTime occurredAt;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$eventID = eventID;
    _resultData['eventID'] = l$eventID;
    final l$eventType = eventType;
    _resultData['eventType'] = l$eventType;
    final l$source = source;
    _resultData['source'] = toJson$Enum$StreamEventSource(l$source);
    final l$payload = payload;
    _resultData['payload'] = l$payload;
    final l$occurredAt = occurredAt;
    _resultData['occurredAt'] = dateTimeToJson(l$occurredAt);
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$eventID = eventID;
    final l$eventType = eventType;
    final l$source = source;
    final l$payload = payload;
    final l$occurredAt = occurredAt;
    final l$$__typename = $__typename;
    return Object.hashAll([
      l$eventID,
      l$eventType,
      l$source,
      l$payload,
      l$occurredAt,
      l$$__typename,
    ]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$eventID = eventID;
    final lOther$eventID = other.eventID;
    if (l$eventID != lOther$eventID) {
      return false;
    }
    final l$eventType = eventType;
    final lOther$eventType = other.eventType;
    if (l$eventType != lOther$eventType) {
      return false;
    }
    final l$source = source;
    final lOther$source = other.source;
    if (l$source != lOther$source) {
      return false;
    }
    final l$payload = payload;
    final lOther$payload = other.payload;
    if (l$payload != lOther$payload) {
      return false;
    }
    final l$occurredAt = occurredAt;
    final lOther$occurredAt = other.occurredAt;
    if (l$occurredAt != lOther$occurredAt) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event
    on Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event {
  CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event
  >
  get copyWith =>
      CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
  TRes
> {
  factory CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event
    instance,
    TRes Function(
      Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event,
    )
    then,
  ) = _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event;

  factory CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event.stub(
    TRes res,
  ) = _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event;

  TRes call({
    String? eventID,
    String? eventType,
    Enum$StreamEventSource? source,
    String? payload,
    DateTime? occurredAt,
    String? $__typename,
  });
}

class _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
  TRes
>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
          TRes
        > {
  _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
    this._instance,
    this._then,
  );

  final Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event
  _instance;

  final TRes Function(
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? eventID = _undefined,
    Object? eventType = _undefined,
    Object? source = _undefined,
    Object? payload = _undefined,
    Object? occurredAt = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
      eventID: eventID == _undefined || eventID == null
          ? _instance.eventID
          : (eventID as String),
      eventType: eventType == _undefined || eventType == null
          ? _instance.eventType
          : (eventType as String),
      source: source == _undefined || source == null
          ? _instance.source
          : (source as Enum$StreamEventSource),
      payload: payload == _undefined || payload == null
          ? _instance.payload
          : (payload as String),
      occurredAt: occurredAt == _undefined || occurredAt == null
          ? _instance.occurredAt
          : (occurredAt as DateTime),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
  TRes
>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event<
          TRes
        > {
  _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$StreamEventSuccess$event(
    this._res,
  );

  TRes _res;

  call({
    String? eventID,
    String? eventType,
    Enum$StreamEventSource? source,
    String? payload,
    DateTime? occurredAt,
    String? $__typename,
  }) => _res;
}

class Subscription$SessionActivity$sessionActivityStream$$GraphError
    implements Subscription$SessionActivity$sessionActivityStream {
  Subscription$SessionActivity$sessionActivityStream$$GraphError({
    required this.code,
    required this.message,
    this.field,
    this.$__typename = 'GraphError',
  });

  factory Subscription$SessionActivity$sessionActivityStream$$GraphError.fromJson(
    Map<String, dynamic> json,
  ) {
    final l$code = json['code'];
    final l$message = json['message'];
    final l$field = json['field'];
    final l$$__typename = json['__typename'];
    return Subscription$SessionActivity$sessionActivityStream$$GraphError(
      code: fromJson$Enum$GraphErrorCode((l$code as String)),
      message: (l$message as String),
      field: (l$field as String?),
      $__typename: (l$$__typename as String),
    );
  }

  final Enum$GraphErrorCode code;

  final String message;

  final String? field;

  final String $__typename;

  Map<String, dynamic> toJson() {
    final _resultData = <String, dynamic>{};
    final l$code = code;
    _resultData['code'] = toJson$Enum$GraphErrorCode(l$code);
    final l$message = message;
    _resultData['message'] = l$message;
    final l$field = field;
    _resultData['field'] = l$field;
    final l$$__typename = $__typename;
    _resultData['__typename'] = l$$__typename;
    return _resultData;
  }

  @override
  int get hashCode {
    final l$code = code;
    final l$message = message;
    final l$field = field;
    final l$$__typename = $__typename;
    return Object.hashAll([l$code, l$message, l$field, l$$__typename]);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other
            is! Subscription$SessionActivity$sessionActivityStream$$GraphError ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$code = code;
    final lOther$code = other.code;
    if (l$code != lOther$code) {
      return false;
    }
    final l$message = message;
    final lOther$message = other.message;
    if (l$message != lOther$message) {
      return false;
    }
    final l$field = field;
    final lOther$field = other.field;
    if (l$field != lOther$field) {
      return false;
    }
    final l$$__typename = $__typename;
    final lOther$$__typename = other.$__typename;
    if (l$$__typename != lOther$$__typename) {
      return false;
    }
    return true;
  }
}

extension UtilityExtension$Subscription$SessionActivity$sessionActivityStream$$GraphError
    on Subscription$SessionActivity$sessionActivityStream$$GraphError {
  CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError<
    Subscription$SessionActivity$sessionActivityStream$$GraphError
  >
  get copyWith =>
      CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError(
        this,
        (i) => i,
      );
}

abstract class CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError<
  TRes
> {
  factory CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError(
    Subscription$SessionActivity$sessionActivityStream$$GraphError instance,
    TRes Function(
      Subscription$SessionActivity$sessionActivityStream$$GraphError,
    )
    then,
  ) = _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$GraphError;

  factory CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError.stub(
    TRes res,
  ) = _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$GraphError;

  TRes call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  });
}

class _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$GraphError<
  TRes
>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError<
          TRes
        > {
  _CopyWithImpl$Subscription$SessionActivity$sessionActivityStream$$GraphError(
    this._instance,
    this._then,
  );

  final Subscription$SessionActivity$sessionActivityStream$$GraphError
  _instance;

  final TRes Function(
    Subscription$SessionActivity$sessionActivityStream$$GraphError,
  )
  _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? code = _undefined,
    Object? message = _undefined,
    Object? field = _undefined,
    Object? $__typename = _undefined,
  }) => _then(
    Subscription$SessionActivity$sessionActivityStream$$GraphError(
      code: code == _undefined || code == null
          ? _instance.code
          : (code as Enum$GraphErrorCode),
      message: message == _undefined || message == null
          ? _instance.message
          : (message as String),
      field: field == _undefined ? _instance.field : (field as String?),
      $__typename: $__typename == _undefined || $__typename == null
          ? _instance.$__typename
          : ($__typename as String),
    ),
  );
}

class _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$GraphError<
  TRes
>
    implements
        CopyWith$Subscription$SessionActivity$sessionActivityStream$$GraphError<
          TRes
        > {
  _CopyWithStubImpl$Subscription$SessionActivity$sessionActivityStream$$GraphError(
    this._res,
  );

  TRes _res;

  call({
    Enum$GraphErrorCode? code,
    String? message,
    String? field,
    String? $__typename,
  }) => _res;
}
